package hubs

import (
	"errors"
	"go.uber.org/zap"
	"mediaserver-go/utils/log"
	"sync"
	"time"

	"mediaserver-go/hubs/codecs"
	"mediaserver-go/utils"
	"mediaserver-go/utils/types"
	"mediaserver-go/utils/units"
)

type HubSource struct {
	mu sync.RWMutex

	tracks map[string]*Track

	mediaType types.MediaType
	codecType types.CodecType

	set      bool
	codecset chan codecs.Codec
	codec    codecs.Codec
}

func NewHubSource(mediaType types.MediaType, codecType types.CodecType) *HubSource {
	return &HubSource{
		mediaType: mediaType,
		codecType: codecType,
		codecset:  make(chan codecs.Codec),
		tracks:    make(map[string]*Track),
	}
}

func (t *HubSource) Close() {
	t.mu.Lock()
	defer t.mu.Unlock()

	for _, track := range t.tracks {
		track.Close()
	}
}

func (t *HubSource) MediaType() types.MediaType {
	return t.mediaType
}

func (t *HubSource) CodecType() types.CodecType {
	return t.codecType
}

func (t *HubSource) SetCodec(c codecs.Codec) {
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

func (t *HubSource) Codec() (codecs.Codec, error) {
	select {
	case <-t.codecset:
	case <-time.After(500 * time.Millisecond):
		return nil, errors.New("video codec not set")
	}

	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.codec, nil
}

func (t *HubSource) VideoCodec() (codecs.VideoCodec, error) {
	select {
	case <-t.codecset:
	case <-time.After(500 * time.Millisecond):
		return nil, errors.New("video codec not set")
	}

	t.mu.RLock()
	defer t.mu.RUnlock()
	c, ok := t.codec.(codecs.VideoCodec)
	if !ok {
		return nil, errors.New("invalid codec")
	}
	return c, nil
}

func (t *HubSource) AudioCodec() (codecs.AudioCodec, error) {
	select {
	case <-t.codecset:
	case <-time.After(500 * time.Millisecond):
		return nil, errors.New("video codec not set")
	}

	t.mu.RLock()
	defer t.mu.RUnlock()
	c, ok := t.codec.(codecs.AudioCodec)
	if !ok {
		return nil, errors.New("invalid codec")
	}
	return c, nil
}

func (t *HubSource) Write(unit units.Unit) {
	t.mu.RLock()
	defer t.mu.RUnlock()

	for _, track := range t.tracks {
		utils.SendOrDrop(track.ch, unit)
	}
}

func (t *HubSource) GetTrack(codec codecs.Codec) *Track {
	t.mu.Lock()
	defer t.mu.Unlock()

	track, ok := t.tracks[codec.String()]
	if !ok {
		track = NewTrack(t.mediaType, codec.CodecType())
		track.SetCodec(codec)
		go track.Run()
		t.tracks[codec.String()] = track
	}
	return track
}

//func (t *HubSource) AddTranscodeTrack(targetCodec codecs.Codec) (*HubSource, error) {
//	sourceCodec, err := t.Codec()
//	if err != nil {
//		return nil, err
//	}
//	if sourceCodec == targetCodec {
//		return t, nil
//	}
//
//	transcoder := NewAudioTranscoder()
//	if err := transcoder.SetupAudio(sourceCodec, targetCodec); err != nil {
//		return nil, err
//	}
//
//	return track, nil
//}
