package opus

import (
	"fmt"
	"github.com/pion/sdp/v3"
	pion "github.com/pion/webrtc/v3"
	"mediaserver-go/codecs"
	"mediaserver-go/codecs/bitstreamfilter"
	"mediaserver-go/hubs/engines"
	"mediaserver-go/thirdparty/ffmpeg/avcodec"
	"mediaserver-go/thirdparty/ffmpeg/avutil"
	"mediaserver-go/utils/types"
	"strings"
)

type Opus struct {
	Base

	config *Config
}

func NewOpus(config *Config) *Opus {
	return &Opus{
		Base:   Base{},
		config: config,
	}
}

func (o *Opus) Equals(codec codecs.Codec) bool {
	if codec == nil {
		return false
	}
	opusCodec, ok := codec.(*Opus)
	if !ok {
		return false
	}
	if o.Channels() != opusCodec.Channels() || o.SampleFormat() != opusCodec.SampleFormat() || o.SampleRate() != opusCodec.SampleRate() {
		return false
	}
	return true
}

func (o *Opus) String() string {
	return o.MimeType()
}

func (o *Opus) GetBase() codecs.Base {
	return o.Base
}

func (o *Opus) CodecType() types.CodecType {
	return types.CodecTypeOpus
}

func (o *Opus) MediaType() types.MediaType {
	return types.MediaTypeAudio
}

func (o *Opus) Channels() int {
	return o.config.Channels()
}

func (o *Opus) SampleFormat() int {
	return o.config.SampleFormat()
}

func (o *Opus) SampleRate() int {
	return o.config.SampleRate()
}

func (o *Opus) ExtraData() []byte {
	return nil
}

func (o *Opus) SetCodecContext(codecCtx *avcodec.CodecContext) {
	codecCtx.SetCodecID(o.AVCodecID())
	codecCtx.SetCodecType(o.AVMediaType())
	codecCtx.SetSampleRate(o.SampleRate())
	avutil.AvChannelLayoutDefault(codecCtx.ChLayout(), o.Channels())
	codecCtx.SetSampleFmt(avutil.AvSampleFormat(o.SampleFormat()))
}

func (o *Opus) AvCodecFifoAlloc() *avutil.AvAudioFifo {
	return avutil.AvAudioFifoAlloc(avutil.AvSampleFormat(o.SampleFormat()), o.Channels(), o.SampleRate())
}

func (o *Opus) WebRTCCodecCapability() (pion.RTPCodecCapability, error) {
	return pion.RTPCodecCapability{
		MimeType:     o.MimeType(),
		ClockRate:    uint32(o.SampleRate()),
		Channels:     uint16(o.Channels()),
		SDPFmtpLine:  "minptime=20;maxaveragebitrate=96000;stereo=1;sprop-stereo=1;useinbandfec=1",
		RTCPFeedback: nil,
	}, nil
}

func (o *Opus) RTPCodecCapability(targetPort int) (engines.RTPCodecParameters, error) {
	payloadType := 111
	return engines.RTPCodecParameters{
		PayloadType: uint8(payloadType),
		ClockRate:   48000,
		CodecType:   o.CodecType(),
		MediaDescription: sdp.MediaDescription{
			MediaName: sdp.MediaName{
				Media: o.MediaType().String(),
				Port: sdp.RangedPort{
					Value: targetPort,
				},
				Protos:  []string{"RTP", "AVP"},
				Formats: []string{fmt.Sprintf("%d", payloadType)},
			},
			Attributes: []sdp.Attribute{
				{
					Key:   "rtpmap",
					Value: fmt.Sprintf("%d %s/%d/%d", payloadType, strings.ToLower(string(o.CodecType())), o.SampleRate(), o.Channels()),
				},
				{
					Key:   "fmtp",
					Value: fmt.Sprintf("%d minptime=10;maxaveragebitrate=96000;stereo=1;sprop-stereo=1;useinbandfec=1", payloadType),
				},
			},
		},
	}, nil
}

func (o *Opus) GetBitStreamFilter() bitstreamfilter.BitStreamFilter {
	return bitstreamfilter.NewBitStream(o.CodecType())
}
