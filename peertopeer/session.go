package peertopeer

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"time"
)

type Handler interface {
	HandleMessage(sess *Session, data []byte) error
}

type Session struct {
	ID      string
	conn    *websocket.Conn
	writeCh chan []byte

	handler Handler
}

func NewSession(conn *websocket.Conn, handler Handler) *Session {
	return &Session{
		ID:      uuid.NewString(),
		conn:    conn,
		writeCh: make(chan []byte),
		handler: handler,
	}
}

func (s *Session) Run(ctx context.Context) error {
	defer s.conn.Close()

	go s.write(ctx)
	return s.read(ctx)
}

func (s *Session) read(ctx context.Context) error {
	for {
		messageType, data, err := s.conn.ReadMessage()
		if err != nil {
			break // 오류가 발생하면 루프 종료
		}
		fmt.Println("messageType:", messageType, ", msg:", string(data))

		s.handler.HandleMessage(s, data)
	}
	return nil
}

func (s *Session) write(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return nil
		case msg := <-s.writeCh:
			if err := s.conn.WriteMessage(websocket.TextMessage, msg); err != nil {
				return err
			}

		}
	}
}

func (s *Session) WriteData(ctx context.Context, b []byte) error {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	select {
	case <-ctx.Done():
		return errors.New("timeout")
	case s.writeCh <- b:
	}
	return nil
}

func (s *Session) Write(ctx context.Context, msg interface{}) error {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	b, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	select {
	case <-ctx.Done():
		return errors.New("timeout")
	case s.writeCh <- b:
	}
	return nil
}
