package codecs

import (
	"mediaserver-go/utils/types"
	"sync"
)

var _ AudioCodec = (*Opus)(nil)

type Opus struct {
	mu sync.RWMutex

	meta OpusMetadata
}

func NewOpus() *Opus {
	return &Opus{}
}

func (a *Opus) SampleRate() int {
	a.mu.RLock()
	defer a.mu.RUnlock()

	return a.meta.SampleRate
}

type OpusMetadata struct {
	CodecType  types.CodecType
	MediaType  types.MediaType
	SampleRate int
	Channels   int
	SampleFmt  int
}

func (a *Opus) SetMetaData(meta OpusMetadata) {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.meta = meta
}

func (a *Opus) CodecType() types.CodecType {
	a.mu.RLock()
	defer a.mu.RUnlock()

	return a.meta.CodecType
}

func (a *Opus) MediaType() types.MediaType {
	a.mu.RLock()
	defer a.mu.RUnlock()

	return a.meta.MediaType
}

func (a *Opus) Channels() int {
	a.mu.RLock()
	defer a.mu.RUnlock()

	return a.meta.Channels
}

func (a *Opus) SampleFormat() int {
	a.mu.RLock()
	defer a.mu.RUnlock()

	return a.meta.SampleFmt
}
