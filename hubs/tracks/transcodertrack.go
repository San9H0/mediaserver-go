package tracks

import (
	"go.uber.org/zap"
	"mediaserver-go/hubs/transcoders"
	"mediaserver-go/utils/log"
)

type TranscoderTrack struct {
	*Track

	transcoder *transcoders.AudioTranscoder
}

func NewTranscoderTrack(transcoder *transcoders.AudioTranscoder) *TranscoderTrack {
	return &TranscoderTrack{
		Track:      NewTrack(transcoder.Target()),
		transcoder: transcoder,
	}
}

func (t *TranscoderTrack) Run() {
	defer func() {
		t.transcoder.Close()
		log.Logger.Info("Transcoder closed",
			zap.String("source", t.transcoder.Source().MimeType()),
			zap.String("target", t.transcoder.Target().MimeType()),
		)
	}()
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
