package codecs

import (
	"github.com/pion/rtp"
	"mediaserver-go/thirdparty/ffmpeg/avcodec"
	"mediaserver-go/thirdparty/ffmpeg/avutil"
	"mediaserver-go/utils/types"
)

type CodecType interface {
	MimeType() string
	MediaType() types.MediaType
	AVMediaType() avutil.MediaType
	CodecType() types.CodecType
	AVCodecID() avcodec.CodecID
	RTPParser(cb func([][]byte) [][]byte) (RTPParser, error)
	RTPPacketizer(pt uint8, ssrc uint32, clockRate uint32) (rtp.Packetizer, error)
	CodecFromAVCodecParameters(param *avcodec.AvCodecParameters) (Codec, error)
}
