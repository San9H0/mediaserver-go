// Use of this source code is governed by a MIT license that can be found in the LICENSE file.
// Giorgis (habtom@giorgis.io)

package avformat

//#cgo pkg-config: libavformat
//#include <libavformat/avformat.h>
import "C"
import (
	"mediaserver-go/thirdparty/ffmpeg/avcodec"
	"reflect"
	"unsafe"

	"mediaserver-go/thirdparty/ffmpeg/avutil"
)

func (f *FormatContext) Chapters() **AvChapter {
	return (**AvChapter)(unsafe.Pointer(f.chapters))
}

func (f *FormatContext) AudioCodec() *avcodec.Codec {
	return (*avcodec.Codec)(unsafe.Pointer(f.audio_codec))
}

func (f *FormatContext) SubtitleCodec() *avcodec.Codec {
	return (*avcodec.Codec)(unsafe.Pointer(f.subtitle_codec))
}

func (f *FormatContext) VideoCodec() *avcodec.Codec {
	return (*avcodec.Codec)(unsafe.Pointer(f.video_codec))
}

func (f *FormatContext) Metadata() *avutil.Dictionary {
	return (*avutil.Dictionary)(unsafe.Pointer(f.metadata))
}

//func (ctxt *Context) Internal() *AvFormatInternal {
//	return (*AvFormatInternal)(unsafe.Pointer(ctxt.internal))
//}

func (f *FormatContext) Pb() *AvIOContext {
	return (*AvIOContext)(unsafe.Pointer(f.pb))
}

func (f *FormatContext) InterruptCallback() AvIOInterruptCB {
	return AvIOInterruptCB(f.interrupt_callback)
}

func (f *FormatContext) Programs() []*AvProgram {
	header := reflect.SliceHeader{
		Data: uintptr(unsafe.Pointer(f.programs)),
		Len:  int(f.NbPrograms()),
		Cap:  int(f.NbPrograms()),
	}

	return *((*[]*AvProgram)(unsafe.Pointer(&header)))
}

func (f *FormatContext) Streams() []*Stream {
	header := reflect.SliceHeader{
		Data: uintptr(unsafe.Pointer(f.streams)),
		Len:  int(f.NbStreams()),
		Cap:  int(f.NbStreams()),
	}

	return *((*[]*Stream)(unsafe.Pointer(&header)))
}

//func (ctxt *Context) Filename() string {
//	return C.GoString((*C.char)(unsafe.Pointer(&ctxt.filename[0])))
//}
//
// func (ctxt *Context) CodecWhitelist() string {
// 	return C.GoString(ctxt.codec_whitelist)
// }
//
// func (ctxt *Context) FormatWhitelist() string {
// 	return C.GoString(ctxt.format_whitelist)
// }

func (f *FormatContext) AudioCodecId() CodecId {
	return CodecId(f.audio_codec_id)
}

func (f *FormatContext) SubtitleCodecId() CodecId {
	return CodecId(f.subtitle_codec_id)
}

func (f *FormatContext) VideoCodecId() CodecId {
	return CodecId(f.video_codec_id)
}

func (f *FormatContext) DurationEstimationMethod() AvDurationEstimationMethod {
	return AvDurationEstimationMethod(f.duration_estimation_method)
}

func (f *FormatContext) AudioPreload() int {
	return int(f.audio_preload)
}

func (f *FormatContext) AvioFlags() int {
	return int(f.avio_flags)
}

func (f *FormatContext) AvoidNegativeTs() int {
	return int(f.avoid_negative_ts)
}

func (f *FormatContext) BitRate() int {
	return int(f.bit_rate)
}

func (f *FormatContext) CtxFlags() int {
	return int(f.ctx_flags)
}

func (f *FormatContext) Debug() int {
	return int(f.debug)
}

func (f *FormatContext) ErrorRecognition() int {
	return int(f.error_recognition)
}

func (f *FormatContext) EventFlags() int {
	return int(f.event_flags)
}

func (f *FormatContext) Flags() int {
	return int(f.flags)
}

func (f *FormatContext) FlushPackets() int {
	return int(f.flush_packets)
}

func (f *FormatContext) FormatProbesize() int {
	return int(f.format_probesize)
}

func (f *FormatContext) FpsProbeSize() int {
	return int(f.fps_probe_size)
}

func (f *FormatContext) IoRepositioned() int {
	return int(f.io_repositioned)
}

func (f *FormatContext) Keylen() int {
	return int(f.keylen)
}

func (f *FormatContext) MaxChunkDuration() int {
	return int(f.max_chunk_duration)
}

func (f *FormatContext) MaxChunkSize() int {
	return int(f.max_chunk_size)
}

func (f *FormatContext) MaxDelay() int {
	return int(f.max_delay)
}

func (f *FormatContext) MaxTsProbe() int {
	return int(f.max_ts_probe)
}

func (f *FormatContext) MetadataHeaderPadding() int {
	return int(f.metadata_header_padding)
}

func (f *FormatContext) ProbeScore() int {
	return int(f.probe_score)
}

func (f *FormatContext) Seek2any() int {
	return int(f.seek2any)
}

func (f *FormatContext) StrictStdCompliance() int {
	return int(f.strict_std_compliance)
}

//func (ctxt *Context) TsId() int {
//	return int(ctxt.ts_id)
//}

func (f *FormatContext) UseWallclockAsTimestamps() int {
	return int(f.use_wallclock_as_timestamps)
}

func (f *FormatContext) Duration() int64 {
	return int64(f.duration)
}

func (f *FormatContext) MaxAnalyzeDuration2() int64 {
	return int64(f.max_analyze_duration)
}

func (f *FormatContext) MaxInterleaveDelta() int64 {
	return int64(f.max_interleave_delta)
}

func (f *FormatContext) OutputTsOffset() int64 {
	return int64(f.output_ts_offset)
}

func (f *FormatContext) Probesize2() int64 {
	return int64(f.probesize)
}

func (f *FormatContext) SkipInitialBytes() int64 {
	return int64(f.skip_initial_bytes)
}

func (f *FormatContext) StartTime() int64 {
	return int64(f.start_time)
}

func (f *FormatContext) StartTimeRealtime() int64 {
	return int64(f.start_time_realtime)
}

func (f *FormatContext) Iformat() *InputFormat {
	return (*InputFormat)(unsafe.Pointer(f.iformat))
}

func (f *FormatContext) Oformat() *OutputFormat {
	return (*OutputFormat)(unsafe.Pointer(f.oformat))
}

// func (ctxt *Context) DumpSeparator() uint8 {
// 	return uint8(ctxt.dump_separator)
// }

func (f *FormatContext) CorrectTsOverflow() int {
	return int(f.correct_ts_overflow)
}

func (f *FormatContext) MaxIndexSize() uint {
	return uint(f.max_index_size)
}

func (f *FormatContext) MaxPictureBuffer() uint {
	return uint(f.max_picture_buffer)
}

func (f *FormatContext) NbChapters() uint {
	return uint(f.nb_chapters)
}

func (f *FormatContext) NbPrograms() uint {
	return uint(f.nb_programs)
}

func (f *FormatContext) NbStreams() uint {
	return uint(f.nb_streams)
}

func (f *FormatContext) PacketSize() uint {
	return uint(f.packet_size)
}

func (f *FormatContext) Probesize() uint {
	return uint(f.probesize)
}

func (f *FormatContext) SetPb(pb *AvIOContext) {
	f.pb = (*C.struct_AVIOContext)(unsafe.Pointer(pb))
}

func (f *FormatContext) Pb2() **AvIOContext {
	return (**AvIOContext)(unsafe.Pointer(&f.pb))
}
