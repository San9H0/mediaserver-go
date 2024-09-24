package codecs

import (
	"fmt"
	"github.com/pion/sdp/v3"
	pion "github.com/pion/webrtc/v3"
	"mediaserver-go/ffmpeg/goav/avcodec"
	"mediaserver-go/ffmpeg/goav/avutil"
	"mediaserver-go/hubs/codecs/bitstreamfilter"
	"mediaserver-go/hubs/engines"
	"mediaserver-go/utils/types"
	"strings"
)

var _ AudioCodec = (*Opus)(nil)

type Opus struct {
	sampleRate int
	channels   int
	sampleFmt  int
}

type OpusParameters struct {
	Channels   int
	SampleRate int
	SampleFmt  int
}

func NewOpus(o OpusParameters) *Opus {
	return &Opus{
		channels:   o.Channels,
		sampleFmt:  o.SampleFmt,
		sampleRate: o.SampleRate,
	}
}

func (o *Opus) Clone() Codec {
	return &Opus{
		channels:   o.channels,
		sampleFmt:  o.sampleFmt,
		sampleRate: o.sampleRate,
	}
}

func (o *Opus) Equals(codec Codec) bool {
	if codec == nil {
		return false
	}
	opusCodec, ok := codec.(*Opus)
	if !ok {
		return false
	}
	if o.CodecType() != opusCodec.CodecType() || o.MediaType() != opusCodec.MediaType() {
		return false
	}

	if o.Channels() != opusCodec.Channels() || o.SampleFormat() != opusCodec.SampleFormat() || o.SampleRate() != opusCodec.SampleRate() {
		return false
	}
	return true
}

func (o *Opus) String() string {
	return fmt.Sprintf("Opus. SampleRate: %d, Channels: %d, SampleFmt: %d", o.sampleRate, o.channels, o.sampleFmt)
}

func (o *Opus) CodecType() types.CodecType {
	return types.CodecTypeOpus
}

func (o *Opus) MediaType() types.MediaType {
	return types.MediaTypeAudio
}

func (o *Opus) Channels() int {
	return o.channels
}

func (o *Opus) SampleFormat() int {
	return o.sampleFmt
}

func (o *Opus) SampleRate() int {
	return o.sampleRate
}

func (o *Opus) ExtraData() []byte {
	return nil
}

func (o *Opus) SetCodecContext(codecCtx *avcodec.CodecContext) {
	codecCtx.SetCodecID(types.CodecIDFromType(types.CodecTypeOpus))
	codecCtx.SetCodecType(types.MediaTypeToFFMPEG(types.MediaTypeAudio))
	codecCtx.SetSampleRate(o.sampleRate)
	avutil.AvChannelLayoutDefault(codecCtx.ChLayout(), o.channels)
	codecCtx.SetSampleFmt(avutil.AvSampleFormat(o.sampleFmt))
}

func (o *Opus) AvCodecFifoAlloc() *avutil.AvAudioFifo {
	return avutil.AvAudioFifoAlloc(avutil.AvSampleFormat(o.sampleFmt), o.channels, o.sampleRate)
}

func (o *Opus) WebRTCCodecCapability() (pion.RTPCodecCapability, error) {
	return pion.RTPCodecCapability{
		MimeType:     types.MimeTypeFromCodecType(o.CodecType()),
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
