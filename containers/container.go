package containers

import "strings"

type Writer struct {
}

func NewWriter() *Writer {
	return &Writer{}
}

func (w *Writer) Init(mimeType string) {
	parts := strings.Split(strings.ToLower(mimeType), "/")
	if len(parts) != 2 {
		return
	}

	switch parts[0] {
	case "video":
	case "audio":
	default:
		return
	}

	switch parts[1] {
	case "":
	case "webm":
	}
}
