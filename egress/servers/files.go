package servers

import (
	"context"
	"errors"
	"go.uber.org/zap"
	"mediaserver-go/egress/sessions"
	"mediaserver-go/hubs"
	"mediaserver-go/utils/buffers"
	"mediaserver-go/utils/dto"
	"mediaserver-go/utils/log"
	"time"
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

	filteredSourceTracks, err := filterMediaTypesInStream(stream, req.MediaTypes)
	if err != nil {
		return dto.EgressFileResponse{}, err
	}

	fileSession, err := sessions.NewFileSession(req.Path, filteredSourceTracks, buffers.NewMemoryBuffer())
	if err != nil {
		return dto.EgressFileResponse{}, err
	}

	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Duration(req.Interval)*time.Millisecond)
		defer cancel()
		if err := fileSession.Run(ctx); err != nil {
			log.Logger.Warn("file session error", zap.Error(err))
		}
	}()
	return dto.EgressFileResponse{}, nil
}
