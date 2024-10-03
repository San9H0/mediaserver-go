package av1

import (
	"fmt"
	"github.com/bluenviron/mediacommon/pkg/codecs/av1"
)

type Decoder struct {
}

func (d *Decoder) KeyFrame(payload []byte) bool {
	var header av1.OBUHeader
	if err := header.Unmarshal(payload); err != nil {
		fmt.Println("[TESTDEBUG] obs header unmarshal error: ", err)
		return false
	}
	return false
}
