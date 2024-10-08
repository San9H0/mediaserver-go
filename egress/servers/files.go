package servers

import (
	"context"
	"errors"
	"time"

	"mediaserver-go/egress/sessions"
	"mediaserver-go/egress/sessions/files"
	"mediaserver-go/hubs"
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

func (f *FileServer) StartSession(streamID string, req dto.EgressFileRequest) (dto.EgressFileResponse, error) {
	stream, ok := f.hub.GetStream(streamID)
	if !ok {
		return dto.EgressFileResponse{}, errors.New("stream not found")
	}

	filteredSources, err := filterMediaTypesInStream(stream, req.MediaTypes)
	if err != nil {
		return dto.EgressFileResponse{}, err
	}

	handler := files.NewHandler(req.Path)
	if err := handler.Init(context.Background(), filteredSources); err != nil {
		return dto.EgressFileResponse{}, err
	}

	sess := sessions.NewSession[*files.TrackContext](handler)
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Duration(req.Interval)*time.Millisecond)
		defer cancel()
		sess.Run(ctx)
	}()

	return dto.EgressFileResponse{}, nil
}
