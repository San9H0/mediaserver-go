package avcodec

//#cgo pkg-config: libavcodec libavutil
//#include <libavformat/avformat.h>
//#include <libavcodec/avcodec.h>
//#include <libavutil/avutil.h>
import "C"
import (
	"mediaserver-go/thirdparty/ffmpeg/avutil"
	"unsafe"
)

func (cc *CodecContext) AvCodecOpen2(codec *Codec, options **Dictionary) int {
	return int(C.avcodec_open2((*C.struct_AVCodecContext)(unsafe.Pointer(cc)), (*C.struct_AVCodec)(unsafe.Pointer(codec)), (**C.struct_AVDictionary)(unsafe.Pointer(options))))
}

func (cc *CodecContext) AvCodecSendPacket(packet *Packet) int {
	return int(C.avcodec_send_packet((*C.struct_AVCodecContext)(unsafe.Pointer(cc)), (*C.struct_AVPacket)(unsafe.Pointer(packet))))
}

func (cc *CodecContext) AvCodecReceivePacket(packet *Packet) int {
	return int(C.avcodec_receive_packet((*C.struct_AVCodecContext)(unsafe.Pointer(cc)), (*C.struct_AVPacket)(unsafe.Pointer(packet))))
}

func (cc *CodecContext) AvCodecSendFrame(frame *avutil.Frame) int {
	return int(C.avcodec_send_frame((*C.struct_AVCodecContext)(unsafe.Pointer(cc)), (*C.struct_AVFrame)(unsafe.Pointer(frame))))
}

func (cc *CodecContext) FrameSize() int {
	return int(cc.frame_size)
}

func (cc *CodecContext) CodecID() CodecID {
	return CodecID(cc.codec_id)
}

func (cc *CodecContext) SetCodecID(codecID CodecID) {
	cc.codec_id = C.enum_AVCodecID(codecID)
}

func (cc *CodecContext) Profile() int {
	return int(cc.profile)
}

func (cc *CodecContext) SetProfile(profile int) {
	cc.profile = C.int(profile)
}

func (cc *CodecContext) Level() int {
	return int(cc.level)
}

func (cc *CodecContext) SetLevel(level int) {
	cc.level = C.int(level)
}

func (cc *CodecContext) CodecType() avutil.MediaType {
	return avutil.MediaType(cc.codec_type)
}

func (cc *CodecContext) SetCodecType(mediaType avutil.MediaType) {
	cc.codec_type = C.enum_AVMediaType(mediaType)
}

func (cc *CodecContext) CodecTag() uint32 {
	return uint32(cc.codec_tag)
}

func (cc *CodecContext) SampleRate() int {
	return int(cc.sample_rate)
}

func (cc *CodecContext) SetSampleRate(sampleRate int) {
	cc.sample_rate = C.int(sampleRate)
}

func (cc *CodecContext) Width() int {
	return int(cc.width)
}

func (cc *CodecContext) SetWidth(w int) {
	cc.width = C.int(w)
}

func (cc *CodecContext) Height() int {
	return int(cc.height)
}

func (cc *CodecContext) SetHeight(h int) {
	cc.height = C.int(h)
}

func (cc *CodecContext) BitRate() int64 {
	return int64(cc.bit_rate)
}

func (cc *CodecContext) SetBitRate(bitRate int64) {
	cc.bit_rate = C.int64_t(bitRate)
}

func (cc *CodecContext) TimeBase() avutil.Rational {
	return avutil.NewRational(int(cc.time_base.num), int(cc.time_base.den))
}

func (cc *CodecContext) SetTimeBase(rational avutil.Rational) {
	cc.time_base.num = C.int(rational.Num())
	cc.time_base.den = C.int(rational.Den())
}

func (cc *CodecContext) FrameRate() avutil.Rational {
	return avutil.NewRational(int(cc.framerate.num), int(cc.framerate.den))
}

func (cc *CodecContext) SetFrameRate(r avutil.Rational) {
	cc.framerate.num = C.int(r.Num())
	cc.framerate.den = C.int(r.Den())
}

func (cc *CodecContext) MaxBFrames() int {
	return int(cc.max_b_frames)
}

func (cc *CodecContext) SetMaxBFrames(b int) {
	cc.max_b_frames = C.int(b)
}

func (cc *CodecContext) GOP() int {
	return int(cc.gop_size)
}

func (cc *CodecContext) SetGOP(gopSize int) {
	cc.gop_size = C.int(gopSize)
}

func (cc *CodecContext) PixelFormat() avutil.PixelFormat {
	return avutil.PixelFormat(cc.pix_fmt)
}

func (cc *CodecContext) SetPixelFormat(pixFmt avutil.PixelFormat) {
	cc.pix_fmt = C.enum_AVPixelFormat(pixFmt)
}

func (cc *CodecContext) SetExtraData(data []byte) {
	cc.extradata = (*C.uint8_t)(C.CBytes(data))
	cc.extradata_size = C.int(len(data))
}

func (cc *CodecContext) ExtraData() []byte {
	return C.GoBytes(unsafe.Pointer(cc.extradata), C.int(cc.extradata_size))
}

func (cc *CodecContext) ChangeExtraData(data []byte) {
	cc.extradata = (*C.uint8_t)(C.CBytes(data))
	cc.extradata_size = C.int(len(data))
}

func (cc *CodecContext) SetSampleFmt(sampleFmt avutil.AvSampleFormat) {
	cc.sample_fmt = (C.enum_AVSampleFormat)(sampleFmt)
}

func (cc *CodecContext) SampleFmt() avutil.AvSampleFormat {
	return avutil.AvSampleFormat(cc.sample_fmt)
}

func (cc *CodecContext) ChLayout() *avutil.AvChannelLayout {
	return (*avutil.AvChannelLayout)(unsafe.Pointer(&cc.ch_layout))
}

func (cc *CodecContext) AvCodecReceiveFrame(frame *avutil.Frame) int {
	return int(C.avcodec_receive_frame((*C.struct_AVCodecContext)(unsafe.Pointer(cc)), (*C.struct_AVFrame)(unsafe.Pointer(frame))))
}

func (cc *CodecContext) AvCodecParametersToContext(param *AvCodecParameters) int {
	return int(C.avcodec_parameters_to_context((*C.struct_AVCodecContext)(unsafe.Pointer(cc)), (*C.struct_AVCodecParameters)(unsafe.Pointer(param))))
}

func (cc *CodecContext) PrivData() unsafe.Pointer {
	return unsafe.Pointer(cc.priv_data)
}

func (cc *CodecContext) SetThreadCount(count int) {
	cc.thread_count = C.int(count)
}
