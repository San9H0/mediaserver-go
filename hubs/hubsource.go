package hubs

import (
	"errors"
	"go.uber.org/zap"
	"golang.org/x/exp/maps"
	"mediaserver-go/hubs/transcoders"
	"mediaserver-go/utils/log"
	"sync"
	"sync/atomic"
	"time"

	"mediaserver-go/hubs/codecs"
	"mediaserver-go/utils"
	"mediaserver-go/utils/types"
	"mediaserver-go/utils/units"
)

var (
	errFailedToSetTranscodeCodec = errors.New("failed to set transcode codec")
)

type HubSource struct {
	mu     sync.RWMutex
	closed atomic.Bool

	tracks map[string]*Track

	typ codecs.CodecType

	set            bool
	codecset       chan codecs.Codec
	codec          codecs.Codec
	transcodeCodec codecs.Codec
	transcoder     *transcoders.VideoTranscoder
}

func NewHubSource(typ codecs.CodecType) *HubSource {
	return &HubSource{
		typ:      typ,
		codecset: make(chan codecs.Codec),
		tracks:   make(map[string]*Track),
	}
}

func (t *HubSource) Close() {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.closed.Swap(true) {
		return
	}

	for _, track := range t.tracks {
		track.Close()
	}

	if t.transcoder != nil {
		t.transcoder.Close()
	}
}

func (t *HubSource) MediaType() types.MediaType {
	return t.typ.MediaType()
}

func (t *HubSource) CodecType() types.CodecType {
	return t.typ.CodecType()
}

func (t *HubSource) SetCodec(c codecs.Codec) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.closed.Load() {
		return
	}

	log.Logger.Info("SetCodec called", zap.Any("codec", c.CodecType()))
	t.codec = c
	if t.set {
		return
	}
	t.set = true
	close(t.codecset)
}

func (t *HubSource) SetTranscodeCodec(c codecs.Codec) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.closed.Load() {
		return errors.New("hub source closed")
	}

	log.Logger.Info("SetTranscodeCodec called", zap.Any("codec", c.CodecType()))
	t.transcodeCodec = c

	t.transcoder = transcoders.NewVideoTranscoder()
	if err := t.transcoder.Setup(t.codec, t.transcodeCodec); err != nil {
		log.Logger.Error("transcoder setup failed", zap.Error(err))
	}
	return nil
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
	var us []units.Unit
	if t.transcoder != nil {
		us = t.transcoder.Transcode(unit)
	} else {
		us = []units.Unit{unit}
	}

	t.mu.RLock()
	defer t.mu.RUnlock()

	for _, track := range t.tracks {
		for _, u := range us {
			utils.SendOrDrop(track.ch, u)
		}
	}
}

func (t *HubSource) GetTrack(codec codecs.Codec) *Track {
	t.mu.Lock()
	defer t.mu.Unlock()

	track, ok := t.tracks[codec.String()]
	if !ok {
		keys := maps.Keys(t.tracks)
		log.Logger.Info("GetTrack",
			zap.Strings("keys", keys),
			zap.String("codec", codec.String()),
		)
		track = NewTrack(codec.Type())
		track.SetCodec(codec)
		if !t.codec.Equals(codec) {
			log.Logger.Info("NewAudioTranscoder",
				zap.String("sourceCodec", t.codec.String()),
				zap.String("targetCodec", codec.String()),
			)
			tanscoder := transcoders.NewAudioTranscoder()
			if err := tanscoder.Setup(t.codec, codec); err != nil {
				log.Logger.Error("transcoder setup failed", zap.Error(err))
				return nil
			}
			track.transcoder = tanscoder
		} else {
			log.Logger.Info("No Transcoder",
				zap.String("sourceCodec", t.codec.String()),
				zap.String("targetCodec", codec.String()),
			)
		}
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
