package sessions

import (
	"context"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
	"mediaserver-go/hubs"
	"mediaserver-go/utils/units"

	"mediaserver-go/utils/log"
	"mediaserver-go/utils/types"
)

type Handler2[T any] interface {
	OnClosed(ctx context.Context) error
	OnSource(ctx context.Context, source *hubs.HubSource) (T, hubs.Track, error)

	OnVideo(ctx context.Context, handle T, unit units.Unit) error
	OnAudio(ctx context.Context, handle T, unit units.Unit) error
}

type Session2[T any] struct {
	handler   Handler2[T]
	stream    *hubs.Stream
	hubTracks []hubs.Track
}

func NewSession2[T any](handler Handler2[T], stream *hubs.Stream) *Session2[T] {
	return &Session2[T]{
		handler: handler,
		stream:  stream,
	}
}

func (s *Session2[T]) Run(ctx context.Context) error {
	log.Logger.Info("session started")
	defer func() {
		log.Logger.Info("session stopped")
		if err := s.handler.OnClosed(ctx); err != nil {
			log.Logger.Error("failed to close session", zap.Error(err))
		}
	}()
	g, ctx := errgroup.WithContext(ctx)

	for source := range s.stream.Subscribe() {
		handle, track, err := s.handler.OnSource(ctx, source)
		if err != nil {
			continue
		}
		consumerCh := track.AddConsumer()
		g.Go(func() error {
			defer track.RemoveConsumer(consumerCh)
			for {
				select {
				case <-ctx.Done():
					return nil
				case unit, ok := <-consumerCh:
					if !ok {
						return nil
					}
					if track.GetCodec().MediaType() == types.MediaTypeVideo {
						if err := s.handler.OnVideo(ctx, handle, unit); err != nil {
							return err
						}
					} else if track.GetCodec().MediaType() == types.MediaTypeAudio {
						if err := s.handler.OnAudio(ctx, handle, unit); err != nil {
							return err
						}
					}
				}
			}
		})
	}

	return g.Wait()
}
