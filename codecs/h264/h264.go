package h264

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/pion/sdp/v3"
	pion "github.com/pion/webrtc/v3"
	"mediaserver-go/codecs"
	"mediaserver-go/hubs/engines"
	"mediaserver-go/parsers/format"
	"mediaserver-go/thirdparty/ffmpeg/avcodec"
	"mediaserver-go/thirdparty/ffmpeg/avutil"
	"strings"
)

var (
	errInvalidSPSLength     = errors.New("invalid SPS length")
	errInvalidPPSLength     = errors.New("invalid PPS length")
	errFailedToUnmarshalSPS = errors.New("failed to unmarshal SPS")
)

type H264 struct {
	Base
	config *Config
}

func NewH264(config *Config) codecs.Codec {
	return &H264{
		Base:   Base{},
		config: config,
	}
}

func (h *H264) String() string {
	return fmt.Sprintf("%s. width:%d,height:%d", h.MimeType(), h.Width(), h.Height())
}

func (h *H264) HLSMIME() string {
	return fmt.Sprintf("avc1.%02X%02X%02x", h.profileIDC(), h.constraintFlag(), h.level())
}

func (h *H264) GetBase() codecs.Base {
	return h.Base
}

func (h *H264) Equals(codec codecs.Codec) bool {
	if codec == nil {
		return false
	}
	h264Codec, ok := codec.(*H264)
	if !ok {
		return false
	}
	if h.Width() != h264Codec.Width() || h.Height() != h264Codec.Height() || h.PixelFormat() != h264Codec.PixelFormat() {
		return false
	}
	if !bytes.Equal(h.SPS(), h264Codec.SPS()) {
		return false
	}
	if !bytes.Equal(h.PPS(), h264Codec.PPS()) {
		return false
	}
	return true
}

func (h *H264) Width() int {
	return h.config.width
}

func (h *H264) Height() int {
	return h.config.height
}

func (h *H264) ClockRate() uint32 {
	return 90000
}

func (h *H264) FPS() float64 {
	return 30
}

func (h *H264) PixelFormat() int {
	return h.config.pixelFmt
}

// ExtraData use readonly
func (h *H264) ExtraData() []byte {
	b, _ := h.config.MarshalToExtraData()
	return b
}

// SPS use readonly
func (h *H264) SPS() []byte {
	return h.config.sps
}

// PPS use readonly
func (h *H264) PPS() []byte {
	return h.config.pps
}

func (h *H264) profileIDC() uint8 {
	return uint8(h.config.profileID)
}

func (h *H264) constraintFlag() uint8 {
	return uint8(h.config.profileComp)
}

func (h *H264) level() uint8 {
	return uint8(h.config.levelID)
}

func (h *H264) profile() string {
	return fmt.Sprintf("%02x%02x%02x", h.config.profileID, h.config.profileComp, h.config.levelID)
}

func (h *H264) SetCodecContext(codecCtx *avcodec.CodecContext, transcodeInfo *codecs.VideoTranscodeInfo) {
	fmt.Println("[TESTDEBUG] h264 setTimebase:", h.FPS())
	codecCtx.SetCodecID(h.AVCodecID())
	codecCtx.SetCodecType(h.AVMediaType())
	codecCtx.SetWidth(h.Width())
	codecCtx.SetHeight(h.Height())
	codecCtx.SetTimeBase(avutil.NewRational(1, 30))
	codecCtx.SetPixelFormat(avutil.PixelFormat(h.PixelFormat()))
	codecCtx.SetProfile(int(h.profileIDC()))
	codecCtx.SetLevel(int(h.level()))
	codecCtx.SetExtraData(h.ExtraData())

	if transcodeInfo != nil {
		fmt.Println("[TESTDEBUG] h.transcodingInfo.GOPSize:", transcodeInfo.GOPSize, ", fps:", transcodeInfo.FPS, ", maxbframe:", transcodeInfo.MaxBFrameSize)
		codecCtx.SetGOP(transcodeInfo.GOPSize)
		codecCtx.SetFrameRate(avutil.NewRational(transcodeInfo.FPS, 1))
		codecCtx.SetMaxBFrames(transcodeInfo.MaxBFrameSize)
		avutil.AvOptSet(codecCtx.PrivData(), "ref", "1", 0)
		avutil.AvOptSet(codecCtx.PrivData(), "rc-lookahead", "0", 0)
		avutil.AvOptSet(codecCtx.PrivData(), "mbtree", "0", 0)
	}
}

func (h *H264) WebRTCCodecCapability() (pion.RTPCodecCapability, error) {
	return pion.RTPCodecCapability{
		MimeType:     h.MimeType(),
		ClockRate:    h.ClockRate(),
		Channels:     0,
		SDPFmtpLine:  fmt.Sprintf("level-asymmetry-allowed=1;packetization-mode=1;profile-level-id=%s", h.profile()),
		RTCPFeedback: nil,
	}, nil
}

func (h *H264) RTPCodecCapability(targetPort int) (engines.RTPCodecParameters, error) {
	payloadType := 98
	return engines.RTPCodecParameters{
		PayloadType: uint8(payloadType),
		ClockRate:   90000,
		CodecType:   h.CodecType(),
		MediaDescription: sdp.MediaDescription{
			MediaName: sdp.MediaName{
				Media: h.MediaType().String(),
				Port: sdp.RangedPort{
					Value: targetPort,
				},
				Protos:  []string{"RTP", "AVP"},
				Formats: []string{fmt.Sprintf("%d", payloadType)},
			},
			Attributes: []sdp.Attribute{
				{
					Key:   "rtpmap",
					Value: fmt.Sprintf("%d %s/%d", payloadType, strings.ToUpper(string(h.CodecType())), h.ClockRate()),
				},
				{
					Key:   "fmtp",
					Value: fmt.Sprintf("%d %s", payloadType, fmt.Sprintf("level-asymmetry-allowed=1;packetization-mode=1;profile-level-id=%s", h.profile())),
				},
			},
		},
	}, nil
}

//func (h *H264) HLSCodecCapability() (string, error) {
//str := avutil.AvFourcc2str(stream.CodecParameters().CodecTag())
//codec, err := h.negotidated[i].Codec()
//if err != nil {
//	continue
//}
//videoCodec, ok := codec.(codecs.VideoCodec)
//if !ok {
//	continue
//}
//profile := stream.CodecParameters().Profile()
//constraintFlags := videoCodec.ExtraData()[2]
//level := stream.CodecParameters().Level()
//return fmt.Sprintf("%s.%02X%02X%02x",
//	str,
//	profile, constraintFlags, level)
//}

func (h *H264) BitStreamFilter2(data []byte) [][]byte {
	return format.GetAUFromAVC(data)
}

func (h *H264) BitStreamFilter(data []byte) [][]byte {
	return format.GetAUFromAVC(data)
}
