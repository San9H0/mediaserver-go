package avutil

// #cgo pkg-config: libavutil
// #include <libavutil/avutil.h>
// #include <libavutil/frame.h>
// #include <libavutil/opt.h>
import "C"
import "unsafe"

func AvOptSetInt(obj unsafe.Pointer, name string, value int, searchFlags int) int {
	cname := (*C.char)(nil)
	if name != "" {
		cname = C.CString(name)
		defer C.free(unsafe.Pointer(cname))
	}
	return int(C.av_opt_set_int(obj, cname, C.int64_t(value), C.int(searchFlags)))
}

func AvOptSetSampleFmt(obj unsafe.Pointer, name string, fmt AvSampleFormat, searchFlags int) int {
	cname := (*C.char)(nil)
	if name != "" {
		cname = C.CString(name)
		defer C.free(unsafe.Pointer(cname))
	}
	return int(C.av_opt_set_sample_fmt(obj, cname, (C.enum_AVSampleFormat)(fmt), C.int(searchFlags)))
}

func AvOptSetChlayout(obj unsafe.Pointer, name string, chLayout *AvChannelLayout, searchFlags int) int {
	cname := (*C.char)(nil)
	if name != "" {
		cname = C.CString(name)
		defer C.free(unsafe.Pointer(cname))
	}
	return int(C.av_opt_set_chlayout(obj, cname, (*C.struct_AVChannelLayout)(unsafe.Pointer(chLayout)), C.int(searchFlags)))
}
