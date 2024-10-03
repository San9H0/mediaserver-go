package hubs

import (
	"errors"
	"go.uber.org/zap"
	"golang.org/x/exp/maps"
	"mediaserver-go/hubs/tracks"
	"mediaserver-go/hubs/transcoders"
	"mediaserver-go/utils/log"
	"sync"
	"sync/atomic"
	"time"

	"mediaserver-go/codecs"
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

	tracks map[string]Track

	typ codecs.Base

	set            bool
	codecset       chan codecs.Codec
	codec          codecs.Codec
	transcodeCodec codecs.Codec
	transcoder     *transcoders.VideoTranscoder
}

func NewHubSource(typ codecs.Base) *HubSource {
	return &HubSource{
		typ:      typ,
		codecset: make(chan codecs.Codec),
		tracks:   make(map[string]Track),
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
		us = []units.Unit{unit}
		//us = t.transcoder.Transcode(unit)
	} else {
		us = []units.Unit{unit}
	}

	t.mu.RLock()
	defer t.mu.RUnlock()

	for _, track := range t.tracks {
		for _, u := range us {
			utils.SendOrDrop(track.InputCh(), u)
		}
	}
}

func (t *HubSource) GetTrack(codec codecs.Codec) Track {
	t.mu.Lock()
	defer t.mu.Unlock()

	iTrack, ok := t.tracks[codec.String()]
	if !ok {
		log.Logger.Info("GetTrack",
			zap.Strings("keys", maps.Keys(t.tracks)), // for deugging
			zap.String("codec", codec.String()),
		)
		if !t.codec.Equals(codec) {
			transcoder := transcoders.NewAudioTranscoder(t.codec, codec)
			if err := transcoder.Setup(); err != nil {
				log.Logger.Error("transcoder setup failed", zap.Error(err))
				return nil
			}

			transcoderTrack := tracks.NewTranscoderTrack(transcoder)
			go transcoderTrack.Run()

			log.Logger.Info("NewAudioTranscoder",
				zap.String("sourceCodec", t.codec.String()),
				zap.String("targetCodec", codec.String()),
			)
			t.tracks[codec.String()] = transcoderTrack
			iTrack = transcoderTrack
		} else {
			track := tracks.NewTrack(codec)
			go track.Run()
			log.Logger.Info("No Transcoder",
				zap.String("sourceCodec", t.codec.String()),
				zap.String("targetCodec", codec.String()))
			t.tracks[codec.String()] = track
			iTrack = track
		}
	}
	return iTrack
}
