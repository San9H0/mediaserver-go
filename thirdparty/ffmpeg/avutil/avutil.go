// Use of this source code is governed by a MIT license that can be found in the LICENSE file.
// Giorgis (habtom@giorgis.io)

// Package avutil is a utility library to aid portable multimedia programming.
// It contains safe portable string functions, random number generators, data structures,
// additional mathematics functions, cryptography and multimedia related functionality.
// Some generic features and utilities provided by the libavutil library
package avutil

import "C"
import "unsafe"

//#cgo pkg-config: libavutil
//#include <libavutil/avutil.h>
//#include <libavutil/samplefmt.h>
//#include <stdlib.h>
//#include <stdbool.h>
//char * fn_av_fourcc2str(int fourcc) {
//  return av_fourcc2str(fourcc);
//}
//char *fn_av_err2str(int errnum) {
//  return av_err2str(errnum);
//}
//bool is_again(int errnum) {
//  return errnum == AVERROR(EAGAIN) || errnum == AVERROR_EOF;
//}
import "C"

type (
	Options       C.struct_AVOptions
	AvTree        C.struct_AVTree
	Rational      C.struct_AVRational
	MediaType     C.enum_AVMediaType
	AvPictureType C.enum_AVPictureType
	PixelFormat   C.enum_AVPixelFormat
	File          C.FILE
	Frame         C.struct_AVFrame
	AvAudioFifo   C.struct_AVAudioFifo
)

const (
	AV_PICTURE_TYPE_NONE AvPictureType = C.AV_PICTURE_TYPE_NONE ///< Undefined
	AV_PICTURE_TYPE_I    AvPictureType = C.AV_PICTURE_TYPE_I    ///< Intra
	AV_PICTURE_TYPE_P    AvPictureType = C.AV_PICTURE_TYPE_P    ///< Predicted
	AV_PICTURE_TYPE_B    AvPictureType = C.AV_PICTURE_TYPE_B    ///< Bi-dir predicted
	AV_PICTURE_TYPE_S    AvPictureType = C.AV_PICTURE_TYPE_S    ///< S(GMC)-VOP MPEG-4
	AV_PICTURE_TYPE_SI   AvPictureType = C.AV_PICTURE_TYPE_SI   ///< Switching Intra
	AV_PICTURE_TYPE_SP   AvPictureType = C.AV_PICTURE_TYPE_SP   ///< Switching Predicted
	AV_PICTURE_TYPE_BI   AvPictureType = C.AV_PICTURE_TYPE_BI   ///< BI type
)

const (
	AVMEDIA_TYPE_UNKNOWN    MediaType = C.AVMEDIA_TYPE_UNKNOWN
	AVMEDIA_TYPE_VIDEO      MediaType = C.AVMEDIA_TYPE_VIDEO
	AVMEDIA_TYPE_AUDIO      MediaType = C.AVMEDIA_TYPE_AUDIO
	AVMEDIA_TYPE_DATA       MediaType = C.AVMEDIA_TYPE_DATA
	AVMEDIA_TYPE_SUBTITLE   MediaType = C.AVMEDIA_TYPE_SUBTITLE
	AVMEDIA_TYPE_ATTACHMENT MediaType = C.AVMEDIA_TYPE_ATTACHMENT
	AVMEDIA_TYPE_NB         MediaType = C.AVMEDIA_TYPE_NB
)

// Return the LIBAvUTIL_VERSION_INT constant.
func AvutilVersion() uint {
	return uint(C.avutil_version())
}

// Return the libavutil build-time configuration.
func AvutilConfiguration() string {
	return C.GoString(C.avutil_configuration())
}

// Return the libavutil license.
func AvutilLicense() string {
	return C.GoString(C.avutil_license())
}

// Return a string describing the media_type enum, NULL if media_type is unknown.
func AvGetMediaTypeString(mt MediaType) string {
	return C.GoString(C.av_get_media_type_string((C.enum_AVMediaType)(mt)))
}

// Return a single letter to describe the given picture type pict_type.
func AvGetPictureTypeChar(pt AvPictureType) string {
	return string(C.av_get_picture_type_char((C.enum_AVPictureType)(pt)))
}

// Return x default pointer in case p is NULL.
func AvXIfNull(p, x int) {
	C.av_x_if_null(unsafe.Pointer(&p), unsafe.Pointer(&x))
}

// Compute the length of an integer list.
func AvIntListLengthForSize(e uint, l int, t uint64) uint {
	return uint(C.av_int_list_length_for_size(C.uint(e), unsafe.Pointer(&l), (C.uint64_t)(t)))
}

//// Open a file using a UTF-8 filename.
//func AvFopenUtf8(p, m string) *File {
//	f := C.av_fopen_utf8(C.CString(p), C.CString(m))
//	return (*File)(f)
//}

// Return the fractional representation of the internal time base.
func AvGetTimeBaseQ() Rational {
	return (Rational)(C.av_get_time_base_q())
}

func AvFourcc2str(fourcc int) string {
	return C.GoString(C.fn_av_fourcc2str(C.int(fourcc)))
}

func AvErr2str(errnum int) string {
	return C.GoString(C.fn_av_err2str(C.int(errnum)))
}

func AvAgain(errnum int) bool {
	return bool(C.is_again(C.int(errnum)))
}
