package buffers

import (
	"errors"
	"io"
)

var _ io.ReadWriteSeeker = &MemoryBuffer{}

type MemoryBuffer struct {
	buf    []byte
	cursor int64
}

func NewMemoryBuffer() *MemoryBuffer {
	return &MemoryBuffer{
		buf:    make([]byte, 0, 1024*1024*1024),
		cursor: 0,
	}
}

func (m *MemoryBuffer) Read(dst []byte) (int, error) {
	start := m.cursor
	if start >= int64(len(m.buf)) {
		return 0, io.EOF
	}
	n := copy(dst, m.buf[start:])
	m.cursor = m.cursor + int64(n)
	return n, nil
}

func (m *MemoryBuffer) Write(src []byte) (int, error) {
	start := m.cursor
	expected := start + int64(len(src))
	if expected > int64(len(m.buf)) {
		m.extendToSize(expected)
	}
	n := copy(m.buf[start:], src)
	m.cursor = m.cursor + int64(n)
	return n, nil
}

func (m *MemoryBuffer) Seek(offset int64, whence int) (int64, error) {
	baseOffset := int64(0)
	switch whence {
	case io.SeekStart:
		baseOffset = 0
	case io.SeekCurrent:
		baseOffset = m.cursor
	case io.SeekEnd:
		baseOffset = int64(len(m.buf))
	default:
		return m.cursor, errors.New("invalid whence")
	}
	newOffset := baseOffset + offset
	if newOffset < 0 {
		return m.cursor, errors.New("can't seek to negative offset")
	}
	if newOffset <= int64(len(m.buf)) {
		m.cursor = newOffset
		return m.cursor, nil
	}
	m.extendToSize(newOffset)
	m.cursor = newOffset
	return newOffset, nil
}

func (m *MemoryBuffer) extendToSize(size int64) {
	if size <= int64(cap(m.buf)) {
		m.buf = m.buf[:size]
		return
	}

	extraSize := 2 * (size - int64(len(m.buf)))
	extraSizeLimit := int64(1 << 30)
	if extraSize >= extraSizeLimit {
		extraSize = extraSizeLimit
	}
	newBuf := make([]byte, size, size+extraSize)
	copy(newBuf, m.buf)
	m.buf = newBuf
}
