package tracks

import (
	"go.uber.org/zap"
	"mediaserver-go/hubs/transcoders"
	"mediaserver-go/utils/log"
	"mediaserver-go/utils/units"
)

type TranscoderTrack struct {
	*Track

	transcoder *transcoders.AudioTranscoder

	ch        chan units.Unit
	consumers []chan units.Unit

	set bool
}

func NewTranscoderTrack(transcoder *transcoders.AudioTranscoder) *TranscoderTrack {
	return &TranscoderTrack{
		Track:      NewTrack(transcoder.Target()),
		transcoder: transcoder,
		ch:         make(chan units.Unit, 100),
	}
}

func (t *TranscoderTrack) Run() {
	defer func() {
		if t.transcoder != nil {
			t.transcoder.Close()
		}
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
			if t.transcoder != nil {
				for _, u := range t.transcoder.Transcode(unit) {
					t.Write(u)
				}
				continue
			}
			t.Write(unit)
		}
	}
}
