package aac

import (
	"errors"
	"fmt"
	pion "github.com/pion/webrtc/v3"
	"mediaserver-go/hubs/codecs"
	"mediaserver-go/hubs/codecs/bitstreamfilter"
	"mediaserver-go/hubs/engines"
	"mediaserver-go/thirdparty/ffmpeg/avcodec"
	"mediaserver-go/thirdparty/ffmpeg/avutil"
	"mediaserver-go/utils/types"
)

type AAC struct {
	sampleRate int
	channels   int
	sampleFmt  int
}

type Parameters struct {
	SampleRate int
	Channels   int
	SampleFmt  int
}

func NewAAC(a Parameters) *AAC {
	return &AAC{
		channels:   a.Channels,
		sampleFmt:  a.SampleFmt,
		sampleRate: a.SampleRate,
	}
}

func (a *AAC) MimeType() string {
	return "audio/aac"
}

func (a *AAC) Equals(codec codecs.Codec) bool {
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

func (a *AAC) Clone() codecs.Codec {
	return &AAC{
		channels:   a.channels,
		sampleFmt:  a.sampleFmt,
		sampleRate: a.sampleRate,
	}
}

func (a *AAC) Type() codecs.CodecType {
	return Type{}
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

func (a *AAC) ExtraData() []byte {
	return nil
}

func (a *AAC) WebRTCCodecCapability() (pion.RTPCodecCapability, error) {
	return pion.RTPCodecCapability{}, errors.New("unsupported webrtc codec")
}

func (a *AAC) SetCodecContext(codecCtx *avcodec.CodecContext) {
	codecCtx.SetCodecID(a.Type().AVCodecID())
	codecCtx.SetCodecType(a.Type().AVMediaType())
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

func (a *AAC) BitStreamFilter(b []byte) [][]byte {
	return [][]byte{b}
}

func (a *AAC) GetBitStreamFilter() bitstreamfilter.BitStreamFilter {
	return bitstreamfilter.NewBitStream(a.CodecType())
}
