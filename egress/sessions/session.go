package sessions

import (
	"context"
	"go.uber.org/zap"

	"golang.org/x/sync/errgroup"

	"mediaserver-go/hubs"
	"mediaserver-go/utils/log"
	"mediaserver-go/utils/types"
	"mediaserver-go/utils/units"
)

type Handler[T any] interface {
	NegotiatedTracks() []*hubs.Track
	OnClosed(ctx context.Context) error

	OnTrack(ctx context.Context, track *hubs.Track) (T, error)
	OnVideo(ctx context.Context, handle T, unit units.Unit) error
	OnAudio(ctx context.Context, handle T, unit units.Unit) error
}

type Session[T any] struct {
	handler   Handler[T]
	hubTracks []*hubs.Track
}

func NewSession[T any](handler Handler[T]) *Session[T] {
	return &Session[T]{
		handler:   handler,
		hubTracks: handler.NegotiatedTracks(),
	}
}

func (s *Session[T]) Run(ctx context.Context) error {
	log.Logger.Info("session started")
	defer func() {
		log.Logger.Info("session stopped")
		if err := s.handler.OnClosed(ctx); err != nil {
			log.Logger.Error("failed to close session", zap.Error(err))
		}
	}()
	g, ctx := errgroup.WithContext(ctx)
	for _, track := range s.hubTracks {
		track := track
		consumerCh := track.AddConsumer()
		g.Go(func() error {
			defer func() {
				track.RemoveConsumer(consumerCh)
			}()
			handle, err := s.handler.OnTrack(ctx, track)
			if err != nil {
				return err
			}
			for {
				select {
				case <-ctx.Done():
					return nil
				case unit, ok := <-consumerCh:
					if !ok {
						return nil
					}
					if track.MediaType() == types.MediaTypeVideo {
						if err := s.handler.OnVideo(ctx, handle, unit); err != nil {
							return err
						}
					} else if track.MediaType() == types.MediaTypeAudio {
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
