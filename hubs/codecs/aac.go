package codecs

import (
	"errors"
	"fmt"
	pion "github.com/pion/webrtc/v3"
	"mediaserver-go/ffmpeg/goav/avcodec"
	"mediaserver-go/ffmpeg/goav/avutil"
	"mediaserver-go/hubs/engines"
	"mediaserver-go/utils/types"
)

var _ AudioCodec = (*AAC)(nil)

type AAC struct {
	sampleRate int
	channels   int
	sampleFmt  int
}

type AACParameters struct {
	SampleRate int
	Channels   int
	SampleFmt  int
}

func NewAAC(a AACParameters) *AAC {
	return &AAC{
		channels:   a.Channels,
		sampleFmt:  a.SampleFmt,
		sampleRate: a.SampleRate,
	}
}

func (a *AAC) Equals(codec Codec) bool {
	if codec == nil {
		return false
	}
	aac, ok := codec.(*AAC)
	if !ok {
		return false
	}

	if a.CodecType() != aac.CodecType() || a.MediaType() != aac.MediaType() {
		return false
	}
	if a.Channels() != aac.Channels() || a.SampleFormat() != aac.SampleFormat() || a.SampleRate() != aac.SampleRate() {
		return false
	}
	return true
}

func (a *AAC) String() string {
	return fmt.Sprintf("AAC. SampleRate: %d, Channels: %d, SampleFmt: %d", a.sampleRate, a.channels, a.sampleFmt)
}

func (a *AAC) CodecType() types.CodecType {
	return types.CodecTypeAAC
}

func (a *AAC) MediaType() types.MediaType {
	return types.MediaTypeAudio
}

func (a *AAC) Channels() int {
	return a.channels
}

func (a *AAC) SampleFormat() int {
	return a.sampleFmt
}

func (a *AAC) SampleRate() int {
	return a.sampleRate
}

func (a *AAC) WebRTCCodecCapability() (pion.RTPCodecCapability, error) {
	return pion.RTPCodecCapability{}, errUnsupportedWebRTC
}

func (a *AAC) SetCodecContext(codecCtx *avcodec.CodecContext) {
	codecCtx.SetCodecID(types.CodecIDFromType(a.CodecType()))
	codecCtx.SetCodecType(types.MediaTypeToFFMPEG(a.MediaType()))
	codecCtx.SetSampleRate(a.SampleRate())
	avutil.AvChannelLayoutDefault(codecCtx.ChLayout(), a.Channels())
	codecCtx.SetSampleFmt(avutil.AvSampleFormat(a.SampleFormat()))
}

func (a *AAC) AvCodecFifoAlloc() *avutil.AvAudioFifo {
	return avutil.AvAudioFifoAlloc(avutil.AvSampleFormat(a.sampleFmt), a.channels, a.sampleRate)
}

func (a *AAC) RTPCodecCapability(targetPort int) (engines.RTPCodecParameters, error) {
	return engines.RTPCodecParameters{}, errors.New("not implemented")
}
