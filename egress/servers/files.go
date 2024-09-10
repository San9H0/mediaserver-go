package servers

import (
	"context"
	"errors"
	"mediaserver-go/dto"
	"mediaserver-go/egress/sessions"
	"mediaserver-go/hubs"
	"mediaserver-go/utils/buffers"
)

type FileServer struct {
	hub *hubs.Hub
}

func NewFileServer(hub *hubs.Hub) (FileServer, error) {
	return FileServer{
		hub: hub,
	}, nil
}

func (f *FileServer) StartSession(streamID string, req dto.EgressFileRequest) (dto.EgressFileResponse, error) {
	stream, ok := f.hub.GetStream(streamID)
	if !ok {
		return dto.EgressFileResponse{}, errors.New("stream not found")
	}

	fileSession, err := sessions.NewFileSession(req.Path, stream.Tracks(), buffers.NewMemoryBuffer())
	if err != nil {
		return dto.EgressFileResponse{}, err
	}
	ctx, cancel := context.WithCancel(context.Background())
	_ = cancel

	go fileSession.Run(ctx)
	return dto.EgressFileResponse{}, nil
}
