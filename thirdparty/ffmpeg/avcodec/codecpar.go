package avcodec

import "C"

//#cgo pkg-config: libavformat libavcodec libavutil libswresample
//#include <stdio.h>
//#include <stdlib.h>
//#include <inttypes.h>
//#include <stdint.h>
//#include <string.h>
//#include <libavformat/avformat.h>
//#include <libavcodec/avcodec.h>
//#include <libavutil/channel_layout.h>
//#include <libavutil/avutil.h>
//#include <libavutil/rational.h>
import "C"
import (
	"mediaserver-go/thirdparty/ffmpeg/avutil"
	"unsafe"
)

func (cp *AvCodecParameters) CodecTag() int {
	return int(cp.codec_tag)
}

func (cp *AvCodecParameters) CodecID() CodecID {
	return CodecID(cp.codec_id)
}

func (cp *AvCodecParameters) SetCodecID(codecID CodecID) {
	cp.codec_id = C.enum_AVCodecID(codecID)
}

func (cp *AvCodecParameters) CodecType() avutil.MediaType {
	return avutil.MediaType(cp.codec_type)
}

func (cp *AvCodecParameters) SetCodecType(mediaType avutil.MediaType) {
	cp.codec_type = C.enum_AVMediaType(mediaType)
}

func (cp *AvCodecParameters) SampleRate() int {
	return int(cp.sample_rate)
}

func (cp *AvCodecParameters) SetSampleRate(sampleRate int) {
	cp.sample_rate = C.int(sampleRate)
}

func (cp *AvCodecParameters) Channels() int {
	return int(cp.ch_layout.nb_channels)
}

func (cp *AvCodecParameters) ChLayout() *avutil.AvChannelLayout {
	return (*avutil.AvChannelLayout)(unsafe.Pointer(&cp.ch_layout))
}

func (cp *AvCodecParameters) AvChannelLayoutDefault(channel int) { // SetChLayout
	C.av_channel_layout_default(&cp.ch_layout, C.int(channel))
}

func (cp *AvCodecParameters) SetWidth(w int) {
	cp.width = C.int(w)
}

func (cp *AvCodecParameters) Width() int {
	return int(cp.width)
}

func (cp *AvCodecParameters) SetHeight(h int) {
	cp.height = C.int(h)
}

func (cp *AvCodecParameters) Height() int {
	return int(cp.height)
}

func (cp *AvCodecParameters) SetFormat(f int) {
	cp.format = C.int(f)
}

func (cp *AvCodecParameters) Format() int {
	return int(cp.format)
}

func (cp *AvCodecParameters) SetBitRate(br int) {
	cp.bit_rate = C.int64_t(br)
}

func (cp *AvCodecParameters) BitRate() int64 {
	return int64(cp.bit_rate)
}

func AvCodecParametersCopy(dst, src *AvCodecParameters) int {
	return int(C.avcodec_parameters_copy((*C.struct_AVCodecParameters)(dst), (*C.struct_AVCodecParameters)(src)))
}

func (cp *AvCodecParameters) Profile() int {
	return int(cp.profile)
}

func (cp *AvCodecParameters) SetProfile(profile int) {
	cp.profile = C.int(profile)
}

func (cp *AvCodecParameters) Level() int {
	return int(cp.level)
}

func (cp *AvCodecParameters) SetLevel(level int) {
	cp.level = C.int(level)
}

func (cp *AvCodecParameters) SetTimeBase(r Rational) {
	cp.framerate.num = r.num
	cp.framerate.den = r.den
}

func (cp *AvCodecParameters) ExtraData() []byte {
	return C.GoBytes(unsafe.Pointer(cp.extradata), C.int(cp.extradata_size))
}

func (cp *AvCodecParameters) SetExtraData(data []byte) {
	cp.extradata = (*C.uint8_t)(C.CBytes(data))
	cp.extradata_size = C.int(len(data))
}

func AvCodecParametersFromContext(codecParameters *AvCodecParameters, codecCtx *CodecContext) int {
	return int(C.avcodec_parameters_from_context((*C.struct_AVCodecParameters)(unsafe.Pointer(codecParameters)), (*C.struct_AVCodecContext)(unsafe.Pointer(codecCtx))))
}
