package sessions

import (
	"context"
	"fmt"
	"github.com/bluenviron/mediacommon/pkg/codecs/h264"
	"github.com/pion/rtcp"
	"github.com/pion/rtp"
	pion "github.com/pion/webrtc/v3"
	_ "golang.org/x/image/vp8"
	"io"
	"mediaserver-go/goav/avutil"
	"mediaserver-go/hubs"
	"mediaserver-go/hubs/codecs"
	"mediaserver-go/hubs/parsers"
	"mediaserver-go/utils"
	"mediaserver-go/utils/generators"
	"mediaserver-go/utils/types"
	"mediaserver-go/utils/units"
	"sync"
	"time"
)

type WHIPSession struct {
	id    string
	token string

	api               *pion.API
	pc                *pion.PeerConnection
	onTrack           chan OnTrack
	onConnectionState chan pion.PeerConnectionState

	stream *hubs.Stream
}

func NewWHIPSession(offer, token string, api *pion.API, stream *hubs.Stream) (WHIPSession, error) {
	fmt.Println("sdp offer:", offer)
	onTrack := make(chan OnTrack, 10)
	onConnectionState := make(chan pion.PeerConnectionState, 10)
	id, err := generators.GenerateID()
	if err != nil {
		return WHIPSession{}, err
	}

	pc, err := api.NewPeerConnection(pion.Configuration{
		SDPSemantics: pion.SDPSemanticsUnifiedPlan,
	})
	if err != nil {
		return WHIPSession{}, err
	}

	if err := pc.SetRemoteDescription(pion.SessionDescription{
		Type: pion.SDPTypeOffer,
		SDP:  offer,
	}); err != nil {
		return WHIPSession{}, err
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
		return WHIPSession{}, err
	}

	if err := pc.SetLocalDescription(sd); err != nil {
		return WHIPSession{}, err
	}

	for range candCh {
	}

	return WHIPSession{
		id:                id,
		token:             token,
		api:               api,
		pc:                pc,
		onTrack:           onTrack,
		onConnectionState: onConnectionState,
		stream:            stream,
	}, nil
}

func (w *WHIPSession) Answer() string {
	fmt.Println("sdp answer:", w.pc.LocalDescription().SDP)
	return w.pc.LocalDescription().SDP
}

type OnTrack struct {
	remote   *pion.TrackRemote
	receiver *pion.RTPReceiver
}

func (w *WHIPSession) Run(ctx context.Context) error {
	fmt.Println("[TESTDEBUG] Session Started")
	defer func() {
		fmt.Println("[TESTDEBUG] Session Closing")
		w.pc.Close()
		fmt.Println("[TESTDEBUG] Session Closed")
		w.stream.Close()
	}()
	pliTicker := time.NewTicker(1000 * time.Millisecond)
	defer pliTicker.Stop()
	videoSSRC := uint32(0)
	for {
		select {
		case <-ctx.Done():
			return nil
		case onTrack := <-w.onTrack:
			mediaType := types.MediaTypeFromPion(onTrack.remote.Kind())
			codecType := types.CodecTypeFromMimeType(onTrack.remote.Codec().MimeType)
			target := hubs.NewTrack(mediaType, codecType)
			if onTrack.remote.Kind() == pion.RTPCodecTypeVideo {
				videoSSRC = uint32(onTrack.remote.SSRC())
				w.stream.AddTrack(target)
			} else {
				w.stream.AddTrack(target)
				opusCodec := codecs.NewOpus()
				opusCodec.SetMetaData(codecs.OpusMetadata{
					CodecType:  types.CodecTypeOpus,
					MediaType:  types.MediaTypeAudio,
					SampleRate: int(onTrack.remote.Codec().ClockRate),
					Channels:   int(onTrack.remote.Codec().Channels),
					SampleFmt:  int(avutil.AV_SAMPLE_FMT_S16),
				})
				target.SetAudioCodec(opusCodec)
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
		case <-pliTicker.C:
			if videoSSRC == 0 {
				continue
			}
			if err := w.pc.WriteRTCP([]rtcp.Packet{
				&rtcp.PictureLossIndication{
					SenderSSRC: 0, MediaSSRC: videoSSRC,
				},
			}); err != nil {
				fmt.Println("write rtcp err:", err)
			}
		}
	}
}

func (w *WHIPSession) readRTP(remote *pion.TrackRemote, receiver *pion.RTPReceiver, target *hubs.Track) error {
	startTS := uint32(0)
	prevTS := uint32(0)
	duration := 0
	h264Parser2 := parsers.NewH264Parser()
	var once sync.Once
	for {
		buf := make([]byte, types.ReadBufferSize)
		n, _, err := remote.Read(buf)
		if err != nil {
			fmt.Println("read rtp err:", err)
			return err
		}
		rtpPacket := rtp.Packet{}
		if err := rtpPacket.Unmarshal(buf[:n]); err != nil {
			fmt.Println("unmarshal rtp err:", err)
			continue
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
			aus := h264Parser2.Parse(rtpPacket.Payload)
			codec := h264Parser2.GetCodec()
			if codec == nil {
				continue
			}

			once.Do(func() {
				target.SetVideoCodec(codec)
			})

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
		}
		if types.CodecTypeFromMimeType(remote.Codec().MimeType) == types.CodecTypeVP8 {
			panic("not implemented")
		}
		if types.CodecTypeFromMimeType(remote.Codec().MimeType) == types.CodecTypeOpus {
			target.Write(units.Unit{
				RTPPacket: &rtpPacket,
			})
			continue
			//
			//target.Write(units.Unit{
			//	Payload:  rtpPacket.Payload,
			//	PTS:      int64(pts),
			//	DTS:      int64(pts),
			//	Duration: int64(duration),
			//	TimeBase: int(remote.Codec().ClockRate),
			//	Flags:    0,
			//})
		}
		prevTS = rtpPacket.Timestamp
	}
}

func (w *WHIPSession) readRTCP(remote *pion.TrackRemote, receiver *pion.RTPReceiver) error {
	for {
		rtcpPackets, _, err := receiver.ReadRTCP()
		if err != nil {
			return err
		}
		_ = rtcpPackets
	}

}
