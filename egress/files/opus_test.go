package files

import (
	"fmt"
	"mediaserver-go/goav/avformat"
	"testing"
)

func TestWriteOpus_Setup(t *testing.T) {
	avFormatCtx := avformat.NewAvFormatContextNull()
	result := avformat.AvformatAllocOutputContext2(&avFormatCtx, nil, "", "dummy.mp4")
	if result < 0 {
		fmt.Println("Error allocating output context", result)
		t.Error("avformat context allocation failed")
		return
	}

	fmt.Println("success")
}
