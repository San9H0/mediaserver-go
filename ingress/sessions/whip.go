package sessions

import (
	"context"
	"fmt"
	"github.com/pion/interceptor"
	"mediaserver-go/codecs/factory"
	"mediaserver-go/ingress/sessions/rtpinbounder"
	"sync"
	"time"

	"github.com/pion/rtcp"
	pion "github.com/pion/webrtc/v3"
	"go.uber.org/zap"
	_ "golang.org/x/image/vp8"

	"mediaserver-go/codecs"
	"mediaserver-go/hubs"
	"mediaserver-go/utils"
	"mediaserver-go/utils/log"
)

type WHIPSession struct {
	token string

	api               *pion.API
	pc                *pion.PeerConnection
	onTrack           chan OnTrack
	onConnectionState chan pion.PeerConnectionState

	stream *hubs.Stream
}

func NewWHIPSession(offer, token string, api *pion.API, stream *hubs.Stream) (WHIPSession, error) {
	onTrack := make(chan OnTrack, 10)
	onConnectionState := make(chan pion.PeerConnectionState, 10)

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
		token:             token,
		api:               api,
		pc:                pc,
		onTrack:           onTrack,
		onConnectionState: onConnectionState,
		stream:            stream,
	}, nil
}

func (w *WHIPSession) Answer() string {
	return w.pc.LocalDescription().SDP
}

type OnTrack struct {
	remote   *pion.TrackRemote
	receiver *pion.RTPReceiver
}

func (w *WHIPSession) Run(ctx context.Context) error {
	defer func() {
		w.pc.Close()
		w.stream.Close()
	}()

	var once sync.Once
	for {
		select {
		case <-ctx.Done():
			return nil
		case onTrack := <-w.onTrack:
			log.Logger.Info("whip ontrack",
				zap.Uint16("pt", uint16(onTrack.remote.Codec().PayloadType)),
				zap.String("mimetype", onTrack.remote.Codec().MimeType),
				zap.String("kind", onTrack.remote.Kind().String()),
				zap.Uint32("ssrc", uint32(onTrack.remote.SSRC())),
				zap.String("streamID", onTrack.remote.StreamID()),
				zap.String("trackID", onTrack.remote.ID()),
				zap.String("rid", onTrack.remote.RID()),
			)

			base, err := factory.NewBase(onTrack.remote.Codec().MimeType)
			if err != nil {
				return err
			}

			stats := rtpinbounder.NewStats(onTrack.remote.Codec().ClockRate, uint32(onTrack.remote.SSRC()))

			hubSource := hubs.NewHubSource(base, onTrack.remote.RID())
			w.stream.AddSource(hubSource)

			parser, err := base.RTPParser(func(codec codecs.Codec) {
				hubSource.SetCodec(codec)
			})
			if err != nil {
				return err
			}
			inbounder := rtpinbounder.NewInbounder(parser, int(onTrack.remote.Codec().ClockRate), func(buf []byte) (int, error) {
				n, _, err := onTrack.remote.Read(buf)
				return n, err
			})
			go inbounder.Run(ctx, hubSource, stats)
			if onTrack.remote.Kind() == pion.RTPCodecTypeVideo {
				once.Do(func() {
					go w.sendPLI(ctx, stats)
				})
			}
			go w.sendReceiverReport(ctx, stats)
			go w.readRTCP(onTrack.remote, onTrack.receiver, stats)
		case connectionState := <-w.onConnectionState:
			fmt.Println("conn:", connectionState.String())
			switch connectionState {
			case pion.PeerConnectionStateDisconnected, pion.PeerConnectionStateFailed:
				return nil
			default:
			}
		}
	}
}

func (w *WHIPSession) readRTCP(remote *pion.TrackRemote, receiver *pion.RTPReceiver, stats *rtpinbounder.Stats) error {
	readRTCPFunc := receiver.ReadRTCP
	if remote.RID() != "" {
		readRTCPFunc = func() ([]rtcp.Packet, interceptor.Attributes, error) {
			return receiver.ReadSimulcastRTCP(remote.RID())
		}
	}

	for {
		rtcpPackets, _, err := readRTCPFunc()
		if err != nil {
			return err
		}
		for _, irtcpPacket := range rtcpPackets {
			switch rtcpPacket := irtcpPacket.(type) {
			case *rtcp.SenderReport:
				stats.UpdateSR(rtcpPacket)
			default:
			}
		}
	}
}

func (w *WHIPSession) sendReceiverReport(ctx context.Context, stats *rtpinbounder.Stats) {
	ticker := time.NewTicker(1000 * time.Millisecond)
	defer ticker.Stop()

	prevBytes := uint32(0)
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			totalBytes := stats.GetTotalBytes()
			bps := 8 * (totalBytes - prevBytes)
			prevBytes = totalBytes
			_ = bps

			if err := w.pc.WriteRTCP([]rtcp.Packet{stats.GetReceiverReport(), stats.GetRemB()}); err != nil {
				log.Logger.Warn("write rtcp err", zap.Error(err))
				return
			}
		}
	}
}

func (w *WHIPSession) sendPLI(ctx context.Context, stats *rtpinbounder.Stats) {
	ticker := time.NewTicker(1000 * time.Millisecond)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := w.pc.WriteRTCP([]rtcp.Packet{
				&rtcp.PictureLossIndication{
					SenderSSRC: 0, MediaSSRC: stats.SSRC,
				},
			}); err != nil {
				log.Logger.Warn("write rtcp err", zap.Error(err))
				return
			}
		}
	}
}
