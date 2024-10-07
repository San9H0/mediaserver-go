package tracks

import (
	"context"
	"go.uber.org/zap"
	"mediaserver-go/hubs/transcoders"
	"mediaserver-go/utils/log"
)

type TranscoderTrack struct {
	*Track

	transcoder *transcoders.AudioTranscoder
}

func NewTranscoderTrack(transcoder *transcoders.AudioTranscoder, rid string) *TranscoderTrack {
	return &TranscoderTrack{
		Track:      NewTrack(transcoder.Target(), rid),
		transcoder: transcoder,
	}
}

func (t *TranscoderTrack) Run() {
	ctx, cancel := context.WithCancel(context.Background())
	defer func() {
		cancel()
		t.transcoder.Close()
		log.Logger.Info("Transcoder closed",
			zap.String("source", t.transcoder.Source().MimeType()),
			zap.String("target", t.transcoder.Target().MimeType()),
		)
	}()
	go t.stats.Run(ctx)
	for {
		select {
		case unit, ok := <-t.ch:
			if !ok {
				return
			}
			for _, u := range t.transcoder.Transcode(unit) {
				t.Write(u)
			}
		}
	}
}
