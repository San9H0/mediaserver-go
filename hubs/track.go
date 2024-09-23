package hubs

import (
	"errors"
	"go.uber.org/zap"
	"mediaserver-go/hubs/codecs"
	"mediaserver-go/utils"
	"mediaserver-go/utils/log"
	"mediaserver-go/utils/types"
	"mediaserver-go/utils/units"
	"sync"
	"time"
)

type Track struct {
	mu sync.RWMutex

	ch        chan units.Unit
	closed    bool
	consumers []chan units.Unit

	mediaType types.MediaType
	codecType types.CodecType

	codecset chan codecs.Codec
	codec    codecs.Codec
}

func NewTrack(mediaType types.MediaType, codecType types.CodecType) *Track {
	return &Track{
		ch:        make(chan units.Unit, 100),
		mediaType: mediaType,
		codecType: codecType,
		codecset:  make(chan codecs.Codec),
	}
}

func (t *Track) Run() {
	for {
		select {
		case unit := <-t.ch:
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

func (t *Track) MediaType() types.MediaType {
	return t.mediaType
}

func (t *Track) CodecType() types.CodecType {
	return t.codecType
}

func (t *Track) SetCodec(c codecs.Codec) {
	t.mu.Lock()
	defer t.mu.Unlock()

	log.Logger.Info("SetCodec called", zap.Any("codec", c.CodecType()))
	t.codec = c
	if t.closed {
		return
	}
	t.closed = true
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
