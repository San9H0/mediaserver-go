package sessions

import (
	"context"
	"fmt"
	pion "github.com/pion/webrtc/v3"
	"io"
	"mediaserver-go/egress/files"
	"mediaserver-go/hubs"
	"mediaserver-go/utils"
	"mediaserver-go/utils/buffers"
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

	hubManager *hubs.Manager
}

func NewWebRTCSession(offer, token string, api *pion.API) (WebRTCSession, error) {
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
	}, nil
}

func (s *WebRTCSession) Answer() string {
	fmt.Println("sdp answer:", s.pc.LocalDescription().SDP)
	return s.pc.LocalDescription().SDP
}

type OnTrack struct {
	remote   *pion.TrackRemote
	receiver *pion.RTPReceiver
}

func (s *WebRTCSession) Run(ctx context.Context, hubManager *hubs.Manager) error {
	fmt.Println("[TESTDEBUG] Session Started")
	defer func() {
		fmt.Println("[TESTDEBUG] Session Closing")
		s.pc.Close()
		fmt.Println("[TESTDEBUG] Session Closed")
	}()
	for {
		select {
		case <-ctx.Done():
			return nil
		case onTrack := <-s.onTrack:
			hub, err := hubManager.NewHub(s.id, types.MediaTypeFromPion(onTrack.remote.Kind()))
			if err != nil {
				fmt.Println("err:", err)
				continue
			}
			var codec Codec
			codec = files.NewOpus(48000, buffers.NewMemoryBuffer())
			if onTrack.remote.Kind() == pion.RTPCodecTypeVideo {
				codec = files.NewH264(90000, buffers.NewMemoryBuffer())
			}
			go s.readRTP(onTrack.remote, onTrack.receiver, codec, hub)
			go s.readRTCP(onTrack.remote, onTrack.receiver)
		case connectionState := <-s.onConnectionState:
			fmt.Println("conn:", connectionState.String())
			switch connectionState {
			case pion.PeerConnectionStateDisconnected, pion.PeerConnectionStateFailed:
				return io.EOF
			default:
			}
		}
	}
}

func (s *WebRTCSession) readRTP(remote *pion.TrackRemote, receiver *pion.RTPReceiver, codec Codec, hub hubs.Hub) error {
	startTS := uint32(0)
	for {
		rtpPacket, _, err := remote.ReadRTP()
		if err != nil {
			codec.Finish()
			return err
		}
		codec.Setup(context.Background())
		if startTS == 0 {
			startTS = rtpPacket.Timestamp
		}
		pts := rtpPacket.Timestamp - startTS

		unit := units.Unit{
			Payload:   rtpPacket.Payload,
			Timestamp: int64(pts),
		}
		codec.WritePacket(unit)

		hub.PushUnit(unit)
	}
}

func (s *WebRTCSession) readRTCP(remote *pion.TrackRemote, receiver *pion.RTPReceiver) error {
	for {
		rtcpPackets, _, err := receiver.ReadRTCP()
		if err != nil {
			return err
		}
		_ = rtcpPackets
	}

}
