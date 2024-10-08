package codecs

import (
	"github.com/pion/rtp"
	"mediaserver-go/thirdparty/ffmpeg/avcodec"
	"mediaserver-go/thirdparty/ffmpeg/avutil"
	"mediaserver-go/utils/types"
)

type Base interface {
	// base info
	MimeType() string
	MediaType() types.MediaType
	AVMediaType() avutil.MediaType
	CodecType() types.CodecType
	AVCodecID() avcodec.CodecID

	// for file
	Extension() string

	// for rtp
	RTPParser(func(codec Codec)) (RTPParser, error)
	RTPPacketizer(pt uint8, ssrc uint32, clockRate uint32) (rtp.Packetizer, error)
	Decoder() Decoder

	CodecFromAVCodecParameters(param *avcodec.AvCodecParameters) (Codec, error)
}
