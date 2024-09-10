package sessions

import (
	"bytes"
	"context"
	"fmt"
	"github.com/bluenviron/mediacommon/pkg/codecs/h264"
	rtpcodecs "github.com/pion/rtp/codecs"
	pion "github.com/pion/webrtc/v3"
	"golang.org/x/image/vp8"
	_ "golang.org/x/image/vp8"
	"io"
	"mediaserver-go/hubs"
	"mediaserver-go/hubs/codecs"
	"mediaserver-go/parser/codecparser"
	"mediaserver-go/utils"
	"mediaserver-go/utils/generators"
	"mediaserver-go/utils/types"
	"mediaserver-go/utils/units"
)

type Codec interface {
	Setup(ctx context.Context) error
	WritePacket(unit units.Unit) error
	Finish()
}

type WebRTCSession struct {
	id    string
	token string

	api               *pion.API
	pc                *pion.PeerConnection
	onTrack           chan OnTrack
	onConnectionState chan pion.PeerConnectionState

	stream *hubs.Stream
}

func NewWebRTCSession(offer, token string, api *pion.API, stream *hubs.Stream) (WebRTCSession, error) {
	fmt.Println("sdp offer:", offer)
	onTrack := make(chan OnTrack, 10)
	onConnectionState := make(chan pion.PeerConnectionState, 10)
	id, err := generators.GenerateID()
	if err != nil {
		return WebRTCSession{}, err
	}

	pc, err := api.NewPeerConnection(pion.Configuration{
		SDPSemantics: pion.SDPSemanticsUnifiedPlan,
	})
	if err != nil {
		return WebRTCSession{}, err
	}

	if err := pc.SetRemoteDescription(pion.SessionDescription{
		Type: pion.SDPTypeOffer,
		SDP:  offer,
	}); err != nil {
		return WebRTCSession{}, err
	}

	candCh := make(chan *pion.ICECandidate, 10)
	pc.OnICECandidate(func(candidate *pion.ICECandidate) {
		if candidate == nil {
			close(candCh)
			return
		}
		candCh <- candidate
	})
	pc.OnConnectionStateChange(func(connectionState pion.PeerConnectionState) {
		utils.SendOrDrop(onConnectionState, connectionState)
	})
	pc.OnTrack(func(remote *pion.TrackRemote, receiver *pion.RTPReceiver) {
		utils.SendOrDrop(onTrack, OnTrack{
			remote:   remote,
			receiver: receiver,
		})
	})

	sd, err := pc.CreateAnswer(&pion.AnswerOptions{})
	if err != nil {
		return WebRTCSession{}, err
	}

	if err := pc.SetLocalDescription(sd); err != nil {
		return WebRTCSession{}, err
	}

	for range candCh {
	}

	return WebRTCSession{
		id:                id,
		token:             token,
		api:               api,
		pc:                pc,
		onTrack:           onTrack,
		onConnectionState: onConnectionState,
		stream:            stream,
	}, nil
}

func (w *WebRTCSession) Answer() string {
	fmt.Println("sdp answer:", w.pc.LocalDescription().SDP)
	return w.pc.LocalDescription().SDP
}

type OnTrack struct {
	remote   *pion.TrackRemote
	receiver *pion.RTPReceiver
}

func (w *WebRTCSession) Run(ctx context.Context) error {
	fmt.Println("[TESTDEBUG] Session Started")
	defer func() {
		fmt.Println("[TESTDEBUG] Session Closing")
		w.pc.Close()
		fmt.Println("[TESTDEBUG] Session Closed")
		w.stream.Close()
	}()
	for {
		select {
		case <-ctx.Done():
			return nil
		case onTrack := <-w.onTrack:
			mediaType := types.MediaTypeFromPion(onTrack.remote.Kind())
			codecType := types.CodecTypeFromMimeType(onTrack.remote.Codec().MimeType)
			target := hubs.NewTrack(mediaType, codecType)
			if onTrack.remote.Kind() == pion.RTPCodecTypeVideo {
				w.stream.AddTrack(target)
			}

			go w.readRTP(onTrack.remote, onTrack.receiver, target)
			go w.readRTCP(onTrack.remote, onTrack.receiver)
		case connectionState := <-w.onConnectionState:
			fmt.Println("conn:", connectionState.String())
			switch connectionState {
			case pion.PeerConnectionStateDisconnected, pion.PeerConnectionStateFailed:
				return io.EOF
			default:
			}
		}
	}
}

func isVP8KeyFrame(vp8Packet *rtpcodecs.VP8Packet) bool {
	if len(vp8Packet.Payload) == 0 {
		return false
	}
	return vp8Packet.Payload[0]&0x01 == 0 && vp8Packet.S == 1
}

func (w *WebRTCSession) readRTP(remote *pion.TrackRemote, receiver *pion.RTPReceiver, target *hubs.Track) error {
	startTS := uint32(0)
	prevTS := uint32(0)
	duration := 0
	var h264parser codecparser.H264
	for {
		rtpPacket, _, err := remote.ReadRTP()
		if err != nil {
			return err
		}
		if startTS == 0 {
			startTS = rtpPacket.Timestamp
		}
		pts := rtpPacket.Timestamp - startTS

		if rtpPacket.Timestamp != prevTS {
			if prevTS == 0 {
				duration = 0
			} else {
				duration = int(rtpPacket.Timestamp - prevTS)
			}
		}

		if types.CodecTypeFromMimeType(remote.Codec().MimeType) == types.CodecTypeH264 {
			aus := h264parser.GetAU(rtpPacket.Payload)
			if h264parser.SPS() == nil || h264parser.PPS() == nil {
				continue
			}

			h264Codec := codecs.NewH264()
			h264Codec.SetMetaData(codecs.H264Metadata{
				CodecType: types.CodecTypeH264,
				MediaType: types.MediaTypeVideo,
				Width:     h264parser.Width(),
				Height:    h264parser.Height(),
				FPS:       h264parser.FPS(),
				PixelFmt:  h264parser.PixelFmt(),
				SPS:       h264parser.SPS(),
				PPS:       h264parser.PPS(),
			})
			target.SetVideoCodec(h264Codec)
			for _, au := range aus {
				naluType := h264.NALUType(au[0] & 0x1F)
				flags := 0
				if naluType == h264.NALUTypeIDR {
					flags = 1
				}
				target.Write(units.Unit{
					Payload:  au,
					PTS:      int64(pts),
					DTS:      int64(pts),
					Duration: int64(duration),
					TimeBase: int(remote.Codec().ClockRate),
					Flags:    flags,
				})
			}
		} else if types.CodecTypeFromMimeType(remote.Codec().MimeType) == types.CodecTypeVP8 {
			p := rtpcodecs.VP8Packet{}
			if _, err := p.Unmarshal(rtpPacket.Payload); err == nil {
				if isVP8KeyFrame(&p) {
					fmt.Println("key frame")
					vp8Decoder := vp8.NewDecoder()
					vp8Decoder.Init(bytes.NewReader(p.Payload), len(p.Payload))
					if vp8FrameHeader, err := vp8Decoder.DecodeFrameHeader(); err == nil {
						fmt.Println("vp8 frame header width:", vp8FrameHeader.Width, "height:", vp8FrameHeader.Height, "key frame:", vp8FrameHeader.KeyFrame)
					}
				}
			}
		}
		prevTS = rtpPacket.Timestamp
	}
}

func (w *WebRTCSession) readRTCP(remote *pion.TrackRemote, receiver *pion.RTPReceiver) error {
	for {
		rtcpPackets, _, err := receiver.ReadRTCP()
		if err != nil {
			return err
		}
		_ = rtcpPackets
	}

}
