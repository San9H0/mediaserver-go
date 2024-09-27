package aac

import (
	"errors"
	pion "github.com/pion/webrtc/v3"
	"mediaserver-go/codecs"
	"mediaserver-go/codecs/bitstreamfilter"
	"mediaserver-go/hubs/engines"
	"mediaserver-go/thirdparty/ffmpeg/avcodec"
	"mediaserver-go/thirdparty/ffmpeg/avutil"
)

type AAC struct {
	Base

	config *Config
}

func NewAAC(config *Config) *AAC {
	return &AAC{
		Base:   Base{},
		config: config,
	}
}

func (a *AAC) Equals(codec codecs.Codec) bool {
	if codec == nil {
		return false
	}
	aacCodec, ok := codec.(*AAC)
	if !ok {
		return false
	}
	if a.Channels() != aacCodec.Channels() || a.SampleRate() != aacCodec.SampleRate() || a.SampleFormat() != aacCodec.SampleFormat() {
		return false
	}
	return true
}
func (a *AAC) String() string {
	return a.MimeType()
}

func (a *AAC) Channels() int {
	return a.config.Channels
}

func (a *AAC) SampleFormat() int {
	return a.config.SampleFormat
}

func (a *AAC) SampleRate() int {
	return a.config.SampleRate
}

func (a *AAC) ExtraData() []byte {
	return nil
}

func (a *AAC) WebRTCCodecCapability() (pion.RTPCodecCapability, error) {
	return pion.RTPCodecCapability{}, errors.New("unsupported webrtc codec")
}

func (a *AAC) SetCodecContext(codecCtx *avcodec.CodecContext) {
	codecCtx.SetCodecID(a.AVCodecID())
	codecCtx.SetCodecType(a.AVMediaType())
	codecCtx.SetSampleRate(a.SampleRate())
	avutil.AvChannelLayoutDefault(codecCtx.ChLayout(), a.Channels())
	codecCtx.SetSampleFmt(avutil.AvSampleFormat(a.SampleFormat()))
}

func (a *AAC) AvCodecFifoAlloc() *avutil.AvAudioFifo {
	return avutil.AvAudioFifoAlloc(avutil.AvSampleFormat(a.SampleFormat()), a.Channels(), a.SampleRate())
}

func (a *AAC) RTPCodecCapability(targetPort int) (engines.RTPCodecParameters, error) {
	return engines.RTPCodecParameters{}, errors.New("not implemented")
}

func (a *AAC) BitStreamFilter(b []byte) [][]byte {
	return [][]byte{b}
}

func (a *AAC) GetBitStreamFilter() bitstreamfilter.BitStreamFilter {
	return bitstreamfilter.NewBitStream(a.CodecType())
}
