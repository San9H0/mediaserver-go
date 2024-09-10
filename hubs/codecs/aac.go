package codecs

import (
	"mediaserver-go/utils/types"
	"sync"
)

var _ AudioCodec = (*AAC)(nil)

type AAC struct {
	mu sync.RWMutex

	meta AACMetadata
}

func NewAAC() *AAC {
	return &AAC{}
}

func (a *AAC) SampleRate() int {
	a.mu.RLock()
	defer a.mu.RUnlock()

	return a.meta.SampleRate
}

type AACMetadata struct {
	CodecType  types.CodecType
	MediaType  types.MediaType
	SampleRate int
	Channels   int
	SampleFmt  int
}

func (a *AAC) SetMetaData(meta AACMetadata) {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.meta = meta
}

func (a *AAC) CodecType() types.CodecType {
	a.mu.RLock()
	defer a.mu.RUnlock()

	return a.meta.CodecType
}

func (a *AAC) MediaType() types.MediaType {
	a.mu.RLock()
	defer a.mu.RUnlock()

	return a.meta.MediaType
}

func (a *AAC) Channels() int {
	a.mu.RLock()
	defer a.mu.RUnlock()

	return a.meta.Channels
}

func (a *AAC) SampleFormat() int {
	a.mu.RLock()
	defer a.mu.RUnlock()

	return a.meta.SampleFmt
}
