package hubs

import (
	"errors"
	"mediaserver-go/hubs/codecs"
	"mediaserver-go/utils"
	"mediaserver-go/utils/types"
	"mediaserver-go/utils/units"
	"sync"
)

type Track struct {
	mu sync.RWMutex

	ready     bool
	consumers []chan units.Unit

	mediaType types.MediaType
	codecType types.CodecType

	codec codecs.Codec
}

func NewTrack(mediaType types.MediaType, codecType types.CodecType) *Track {
	return &Track{
		mediaType: mediaType,
		codecType: codecType,
	}
}

func (t *Track) Close() {
	t.mu.Lock()
	defer t.mu.Unlock()

	for _, c := range t.consumers {
		close(c)
	}
	t.consumers = nil
}

func (t *Track) MediaType() types.MediaType {
	return t.mediaType
}

func (t *Track) CodecType() types.CodecType {
	return t.codecType
}

func (t *Track) IsReady() bool {
	t.mu.RLock()
	defer t.mu.RUnlock()

	return t.ready
}

func (t *Track) SetVideoCodec(c codecs.VideoCodec) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.ready {
		return
	}
	t.ready = true
	t.codec = c
}

func (t *Track) SetAudioCodec(c codecs.AudioCodec) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.ready {
		return
	}
	t.ready = true
	t.codec = c
}

func (t *Track) VideoCodec() (codecs.VideoCodec, error) {
	t.mu.RLock()
	defer t.mu.RUnlock()
	if t.codec == nil {
		return nil, nil
	}
	c, ok := t.codec.(codecs.VideoCodec)
	if !ok {
		return nil, errors.New("invalid codec")
	}
	return c, nil
}

func (t *Track) AudioCodec() (codecs.AudioCodec, error) {
	t.mu.RLock()
	defer t.mu.RUnlock()
	if t.codec == nil {
		return nil, nil
	}
	c, ok := t.codec.(codecs.AudioCodec)
	if !ok {
		return nil, errors.New("invalid codec")
	}
	return c, nil
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
