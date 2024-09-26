package hubs

import (
	"errors"
	"go.uber.org/zap"
	"mediaserver-go/hubs/codecs"
	"mediaserver-go/hubs/transcoders"
	"mediaserver-go/utils"
	"mediaserver-go/utils/log"
	"mediaserver-go/utils/types"
	"mediaserver-go/utils/units"
	"sync"
	"time"
)

type Track struct {
	mu sync.RWMutex

	transcoder *transcoders.AudioTranscoder

	typ       codecs.CodecType
	ch        chan units.Unit
	consumers []chan units.Unit

	set      bool
	codecset chan codecs.Codec
	codec    codecs.Codec
}

func NewTrack(typ codecs.CodecType) *Track {
	return &Track{
		typ:      typ,
		ch:       make(chan units.Unit, 100),
		codecset: make(chan codecs.Codec),
	}
}

func (t *Track) Run() {
	defer func() {
		log.Logger.Info("Track closed", zap.String("mimeType", t.typ.MimeType()))
		if t.transcoder != nil {
			t.transcoder.Close()
		}
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

func (t *Track) Close() {
	t.mu.Lock()
	defer t.mu.Unlock()

	consumer := t.consumers
	t.consumers = nil
	for _, c := range consumer {
		close(c)
	}
	close(t.ch)
}

func (t *Track) Type() codecs.CodecType {
	return t.typ
}

func (t *Track) MediaType() types.MediaType {
	return t.typ.MediaType()
}

func (t *Track) CodecType() types.CodecType {
	return t.typ.CodecType()
}

func (t *Track) SetCodec(c codecs.Codec) {
	t.mu.Lock()
	defer t.mu.Unlock()

	log.Logger.Info("SetCodec called", zap.Any("codec", c.CodecType()))
	t.codec = c
	if t.set {
		return
	}
	t.set = true
	close(t.codecset)
}

func (t *Track) Codec() (codecs.Codec, error) {
	select {
	case <-t.codecset:
	case <-time.After(500 * time.Millisecond):
		return nil, errors.New("video codec not set")
	}

	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.codec, nil
}

func (t *Track) Write(unit units.Unit) {
	t.mu.RLock()
	defer t.mu.RUnlock()

	for _, c := range t.consumers {
		utils.SendOrDrop(c, unit)
	}
}

func (t *Track) AddConsumer() chan units.Unit {
	t.mu.Lock()
	defer t.mu.Unlock()

	consumerCh := make(chan units.Unit, 100)
	t.consumers = append(t.consumers, consumerCh)
	return consumerCh
}

func (t *Track) RemoveConsumer(consumerCh chan units.Unit) {
	t.mu.Lock()
	defer t.mu.Unlock()

	for i, c := range t.consumers {
		if c == consumerCh {
			t.consumers = append(t.consumers[:i], t.consumers[i+1:]...)
			close(c)
			return
		}
	}
}
