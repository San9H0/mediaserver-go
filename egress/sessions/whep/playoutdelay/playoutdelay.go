package playoutdelay

import (
	"time"
)

const capacityTime = 5

type Handler struct {
	id             int
	use            bool
	targetMin      int
	payload        []byte
	updateInterval time.Duration

	curMin     int
	updateTime time.Time
}

func NewHandler(id int, use bool) *Handler {
	return &Handler{
		id:             id,
		use:            use,
		targetMin:      60,
		payload:        make([]byte, 3),
		updateInterval: 300 * time.Millisecond,
		updateTime:     time.Now(),
	}
}

func (p *Handler) SetTargetMin(min int) {
	p.targetMin = min
}

func (p *Handler) GetPayload() (int, []byte, bool) {
	prevMin := p.curMin
	if p.use {
		if p.curMin < p.targetMin && time.Since(p.updateTime) > p.updateInterval {
			p.curMin++
		}
	} else {
		p.curMin = 0
	}

	curMax := p.curMin + capacityTime

	if prevMin != p.curMin {
		p.updateTime = time.Now()
		p.payload[0] = byte((p.curMin >> 4) & 0xFF)
		p.payload[1] = byte(((p.curMin << 4) & 0xF0) | ((curMax >> 8) & 0x0F))
		p.payload[2] = byte(curMax & 0xFF)
	}
	return p.id, p.payload, true
}
