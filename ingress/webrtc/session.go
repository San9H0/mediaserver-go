package webrtc

import (
	"context"
	"fmt"
	pion "github.com/pion/webrtc/v3"
	"mediaserver-go/utils/generators"
)

type Session struct {
	id    string
	token string

	pc *pion.PeerConnection
}

func NewSession(token string) (Session, error) {
	id, err := generators.GenerateID()
	if err != nil {
		return Session{}, err
	}
	return Session{
		id:    id,
		token: token,
	}, nil
}

func (s *Session) Run(ctx context.Context) {
	fmt.Println("[TESTDEBUG] Session Started")
	<-ctx.Done()
	fmt.Println("[TESTDEBUG] Session Closed")
}
