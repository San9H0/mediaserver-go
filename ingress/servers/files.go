package servers

import (
	"context"
	"mediaserver-go/dto"
	"mediaserver-go/hubs"
	"mediaserver-go/ingress/sessions"
	"time"
)

type FileServer struct {
	hub *hubs.Hub
}

func NewFileServer(hubManager *hubs.Hub) (FileServer, error) {
	return FileServer{
		hub: hubManager,
	}, nil
}

func (f *FileServer) StartSession(req dto.IngressFileRequest) (dto.IngressFileResponse, error) {
	streamID := "FileServerID"
	stream := hubs.NewStream()
	f.hub.AddStream(streamID, stream)

	fileSession, err := sessions.NewFileSession(req.Path, req.MediaTypes, req.Live, stream)
	if err != nil {
		return dto.IngressFileResponse{}, err
	}

	ctx, cancel := context.WithCancel(context.Background())
	_ = cancel
	go fileSession.Run(ctx)

	go func() {
		time.Sleep(10 * time.Second)
		cancel()
	}()

	return dto.IngressFileResponse{}, nil
}
