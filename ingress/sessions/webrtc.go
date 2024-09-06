package sessions

import (
	"context"
	"fmt"
	pion "github.com/pion/webrtc/v3"
	"mediaserver-go/utils"
	"mediaserver-go/utils/generators"
)

type Session struct {
	id    string
	token string

	api               *pion.API
	pc                *pion.PeerConnection
	onTrack           chan OnTrack
	onConnectionState chan pion.PeerConnectionState
}

func NewSession(offer, token string, api *pion.API) (Session, error) {
	onTrack := make(chan OnTrack, 10)
	onConnectionState := make(chan pion.PeerConnectionState, 10)
	id, err := generators.GenerateID()
	if err != nil {
		return Session{}, err
	}

	pc, err := api.NewPeerConnection(pion.Configuration{
		SDPSemantics: pion.SDPSemanticsUnifiedPlan,
	})
	if err != nil {
		return Session{}, err
	}

	if err := pc.SetRemoteDescription(pion.SessionDescription{
		Type: pion.SDPTypeOffer,
		SDP:  offer,
	}); err != nil {
		return Session{}, err
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
		return Session{}, err
	}

	if err := pc.SetLocalDescription(sd); err != nil {
		return Session{}, err
	}

	for range candCh {
	}

	return Session{
		id:                id,
		token:             token,
		api:               api,
		pc:                pc,
		onTrack:           onTrack,
		onConnectionState: onConnectionState,
	}, nil
}

func (s *Session) Answer() string {
	return s.pc.LocalDescription().SDP
}

type OnTrack struct {
	remote   *pion.TrackRemote
	receiver *pion.RTPReceiver
}

func (s *Session) Run(ctx context.Context) error {
	fmt.Println("[TESTDEBUG] Session Started")
	defer func() {
		fmt.Println("[TESTDEBUG] Session Closed")
		s.pc.Close()
	}()
	for {
		select {
		case <-ctx.Done():
			return nil
		case onTrack := <-s.onTrack:
			go s.readRTP(onTrack.remote, onTrack.receiver)
			go s.readRTCP(onTrack.remote, onTrack.receiver)
		case connectionState := <-s.onConnectionState:
			fmt.Println("conn:", connectionState.String())
			// TODO control connection state
		}
	}
}

func (s *Session) readRTP(remote *pion.TrackRemote, receiver *pion.RTPReceiver) error {
	for {
		rtpPacket, _, err := remote.ReadRTP()
		if err != nil {
			return err
		}
		_ = rtpPacket
	}
}

func (s *Session) readRTCP(remote *pion.TrackRemote, receiver *pion.RTPReceiver) error {
	for {
		rtcpPackets, _, err := receiver.ReadRTCP()
		if err != nil {
			return err
		}
		_ = rtcpPackets
	}

}
