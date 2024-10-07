package tracks

import (
	"context"
	"go.uber.org/zap"
	"mediaserver-go/codecs"
	"mediaserver-go/utils"
	"mediaserver-go/utils/log"
	"mediaserver-go/utils/units"
	"sync"
)

type Track struct {
	mu sync.RWMutex

	cancel    context.CancelFunc
	rid       string
	codec     codecs.Codec
	ch        chan units.Unit
	consumers []chan units.Unit

	set bool

	stats *Stats
}

func NewTrack(codec codecs.Codec, rid string) *Track {
	return &Track{
		codec: codec,
		rid:   rid,
		ch:    make(chan units.Unit, 100),
		stats: NewStats(),
	}
}

func (t *Track) RID() string {
	return t.rid
}

func (t *Track) GetStats() *Stats {
	return t.stats
}

func (t *Track) InputCh() chan units.Unit {
	return t.ch
}

func (t *Track) Run() {
	ctx, cancel := context.WithCancel(context.Background())
	defer func() {
		cancel()
		log.Logger.Info("Track closed", zap.String("mimeType", t.codec.MimeType()))
	}()

	go t.stats.Run(ctx)

	for {
		select {
		case unit, ok := <-t.ch:
			if !ok {
				return
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

func (t *Track) GetCodec() codecs.Codec {
	return t.codec
}

func (t *Track) Write(unit units.Unit) {
	t.stats.update(unit)

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
