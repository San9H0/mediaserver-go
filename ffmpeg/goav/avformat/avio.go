package avformat

import "C"
import (
	"errors"
	"fmt"
	"io"
	"sync"
	"unsafe"
)

//#cgo pkg-config: libavformat
//#include <libavformat/avformat.h>
//#include <libavformat/avio.h>
//extern int read_packet(void *, uint8_t *, int);
//extern int write_packet(void *, uint8_t *, int);
//extern int write_packet2(void *, uint8_t *, int);
//extern int64_t seek(void *, int64_t, int);
import "C"

const (
	AVIO_FLAG_READ       = C.AVIO_FLAG_READ
	AVIO_FLAG_WRITE      = C.AVIO_FLAG_WRITE
	AVIO_FLAG_READ_WRITE = C.AVIO_FLAG_READ_WRITE
)

var ContextBufferMap sync.Map

func AvIOOpen(avioCtx **AvIOContext, url string, flags int) int {
	curl := (*C.char)(nil)
	if len(url) > 0 {
		curl = C.CString(url)
		defer C.free(unsafe.Pointer(curl))
	}
	return int(C.avio_open((**C.struct_AVIOContext)(unsafe.Pointer(avioCtx)), curl, C.int(flags)))
}

func AVIoAllocContext(fmtCtx *FormatContext, buffer io.ReadWriteSeeker, buf *uint8, bufSize int, writeFlag int, seekable bool) *AvIOContext {
	readCb := (*[0]byte)(C.read_packet)
	writeCb := (*[0]byte)(C.write_packet)
	seekCb := (*[0]byte)(C.seek)
	if writeFlag == 0 {
		writeCb = nil
	} else {
		readCb = nil
	}
	if !seekable {
		seekCb = nil
	}
	avioCtx := (*AvIOContext)(
		C.avio_alloc_context(
			(*C.uchar)(unsafe.Pointer(buf)),
			C.int(bufSize),
			C.int(writeFlag),
			unsafe.Pointer(fmtCtx),
			readCb,
			writeCb,
			seekCb,
		),
	)
	ContextBufferMap.Store(fmtCtx, buffer)

	return avioCtx
}

func AVIOOpenDynBuf() *AvIOContext {
	avIOCtx := (*AvIOContext)(unsafe.Pointer((*C.struct_AVDictionary)(nil)))
	C.avio_open_dyn_buf((**C.struct_AVIOContext)(unsafe.Pointer(&avIOCtx)))
	return avIOCtx
}

func AVIOCloseDynBuf(avIOCtx *AvIOContext) []byte {
	var buf *uint8
	bufferLen := int(C.avio_close_dyn_buf((*C.struct_AVIOContext)(avIOCtx), (**C.uint8_t)(unsafe.Pointer(&buf))))
	if bufferLen > 0 {
		b := C.GoBytes(unsafe.Pointer(buf), C.int(bufferLen))
		C.av_freep(unsafe.Pointer(&buf))
		return b
	}

	return nil
}

func AVIOCloseP(pb **AvIOContext) int {
	return int(C.avio_closep((**C.struct_AVIOContext)(unsafe.Pointer(pb))))
}

func AvIoContextFree(avIOCtx *AvIOContext) {
	fmtCtx := (*FormatContext)((*avIOCtx).opaque)
	ContextBufferMap.Delete(fmtCtx)
	C.avio_context_free((**C.struct_AVIOContext)(unsafe.Pointer(&avIOCtx)))
}

//export read_packet
func read_packet(opaque unsafe.Pointer, buf *C.uint8_t, buf_size C.int) C.int {
	fmt.Println("[TESTDEBUG] read_packet called with buf_size:", buf_size)
	ctx_ptr := (*FormatContext)(opaque)
	value, ok := ContextBufferMap.Load(ctx_ptr)
	if !ok {
		return -1
	}
	reader, ok := value.(io.Reader)
	if !ok {
		return -1
	}
	n, err := reader.Read(C.GoBytes(unsafe.Pointer(buf), C.int(buf_size)))
	if err != nil {
		if errors.Is(err, io.EOF) {
			return 0
		}
		return -1
	}
	return C.int(n)
}

//export write_packet2
func write_packet2(opaque unsafe.Pointer, buf *C.uint8_t, buf_size C.int) C.int {
	fmt.Println("[TESTDEBUG] write_packet called with buf_size:", buf_size)
	ctx_ptr := (*interface{})(opaque)
	value, ok := ContextBufferMap.Load(ctx_ptr)
	if !ok {
		return -1
	}
	writer, ok := value.(io.Writer)
	if !ok {
		return -1
	}
	n, err := writer.Write(C.GoBytes(unsafe.Pointer(buf), C.int(buf_size)))
	if err != nil {
		return -1
	}
	return C.int(n)
}

//export write_packet
func write_packet(opaque unsafe.Pointer, buf *C.uint8_t, buf_size C.int) C.int {
	fmt.Println("[TESTDEBUG] write_packet called with buf_size:", buf_size)
	ctx_ptr := (*FormatContext)(opaque)
	value, ok := ContextBufferMap.Load(ctx_ptr)
	if !ok {
		return -1
	}
	writer, ok := value.(io.Writer)
	if !ok {
		return -1
	}
	n, err := writer.Write(C.GoBytes(unsafe.Pointer(buf), C.int(buf_size)))
	if err != nil {
		return -1
	}
	return C.int(n)
}

//export seek
func seek(opaque unsafe.Pointer, pos C.int64_t, whence C.int) C.int64_t {
	fmt.Println("[TESTDEBUG] seek called with pos:", pos, "whence:", whence)
	ctx_ptr := (*FormatContext)(opaque)
	value, ok := ContextBufferMap.Load(ctx_ptr)
	if !ok {
		return -1
	}
	seeker, ok := value.(io.Seeker)
	if !ok {
		return -1
	}
	n, err := seeker.Seek(int64(pos), int(whence))
	if err != nil {
		return -1
	}
	return C.int64_t(n)
}
