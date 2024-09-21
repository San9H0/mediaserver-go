package servers

import (
	"context"
	"errors"
	"mediaserver-go/egress/sessions"
	"mediaserver-go/egress/sessions/images"
	"mediaserver-go/hubs"
	"mediaserver-go/utils/dto"
	"mediaserver-go/utils/types"
)

type ImageServer struct {
	hub *hubs.Hub
}

func NewImageServer(hub *hubs.Hub) (ImageServer, error) {
	return ImageServer{}, nil
}

func (i *ImageServer) StartSession(streamID string, req dto.ImagesRequest) (dto.ImagesResponse, error) {
	stream, ok := i.hub.GetStream(streamID)
	if !ok {
		return dto.ImagesResponse{}, errors.New("stream not found")
	}

	filteredSourceTracks, err := filterMediaTypesInStream(stream, []types.MediaType{types.MediaTypeVideo})
	if err != nil {
		return dto.ImagesResponse{}, err
	}

	handler := images.NewHandler(req.Encoding)
	if err := handler.Init(context.Background(), filteredSourceTracks); err != nil {
		return dto.ImagesResponse{}, err
	}

	sess := sessions.NewSession[*images.TrackContext](handler)
	go func() {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		sess.Run(ctx)
	}()
	return dto.ImagesResponse{}, nil
}
