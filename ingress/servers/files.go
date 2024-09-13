package servers

import (
	"context"
	"mediaserver-go/hubs"
	"mediaserver-go/ingress/sessions"
	"mediaserver-go/utils/dto"
)

type FileServer struct {
	hub *hubs.Hub
}

func NewFileServer(hub *hubs.Hub) (FileServer, error) {
	return FileServer{
		hub: hub,
	}, nil
}

func (f *FileServer) StartSession(streamID string, req dto.IngressFileRequest) (dto.IngressFileResponse, error) {
	stream := hubs.NewStream()
	f.hub.AddStream(streamID, stream)

	fileSession, err := sessions.NewFileSession(req.Path, req.MediaTypes, req.Live, stream)
	if err != nil {
		return dto.IngressFileResponse{}, err
	}

	ctx, cancel := context.WithCancel(context.Background())
	_ = cancel
	go fileSession.Run(ctx)

	return dto.IngressFileResponse{}, nil
}
