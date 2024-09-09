package rtpparser

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/bluenviron/mediacommon/pkg/codecs/h264"
	"github.com/pion/rtp"
	"sync"
	"sync/atomic"
)

var (
	invalidPayloadLength = errors.New("invalid payload length")
)

type H264 struct {
	mu sync.RWMutex

	ready atomic.Bool

	fragments []byte

	sps    []byte
	pps    []byte
	spsSet *h264.SPS
}

func (h *H264) Ready() bool {
	return h.ready.Load()
}

func (h *H264) SPS() []byte {
	return h.sps
}

func (h *H264) PPS() []byte {
	return h.pps
}

func (h *H264) Width() int {
	if h.spsSet != nil {
		return h.spsSet.Width()
	}
	return 0
}

func (h *H264) Height() int {
	if h.spsSet != nil {
		return h.spsSet.Height()
	}
	return 0
}

func (h *H264) FPS() float64 {
	if h.spsSet != nil {
		return h.spsSet.FPS()
	}
	return 0
}

/*
GetAU 함수는 RTP Packet을 받아서 NAL unit을 추출하는 함수이다.
https://datatracker.ietf.org/doc/html/rfc6184#section-5.2
payload의 첫번째 바이트를 가지고 NAL unit을 확인한다.
3가지의 경우가 있음.
1. Single NAL units packet
- 하나의 NAL units 이 하나의 RTP Payload에 있는 패킷.
- 1 ~ 23까지의 값이 들어감.
2. Aggregation Packet
- 여러개의 NAL units 들이 하나의 RTP Payload에 들어가 있는 경우.
- 24 ~ 27 까지의 값이 들어감.
- 24: STAP-A (Single-time aggregation packet)
- 25: STAP-B (Single-time aggregation packet)
- 26: MTAP16 (Multi-time aggregation packet)
- 27: MTAP24 (Multi-time aggregation packet)
3. Fragmentation Unit
- 하나의 NAL units 이 여러개의 RTP Payload에 나눠서 들어가 있는 경우.
- 28 ~ 29 까지의 값이 들어감.
- 28: FU-A (Fragmentation units)
- 29: FU-B (Fragmentation units)
*/
func (h *H264) GetAU(pkt *rtp.Packet) ([][]byte, error) {
	if len(pkt.Payload) < 1 {
		return nil, invalidPayloadLength
	}

	const (
		commonHeaderIdx   = 0
		stapHeaderIdx     = 1
		fragmentHeaderIdx = 1
	)
	naluType := h264.NALUType(pkt.Payload[commonHeaderIdx] & 0x1F)

	switch {
	case 1 <= naluType && naluType <= 23:
		h.extractSPSandPPS(pkt.Payload)
		return [][]byte{pkt.Payload}, nil
	case naluType == h264.NALUTypeSTAPA:
		au := [][]byte{}
		currOffset := stapHeaderIdx
		for currOffset < len(pkt.Payload) {
			naluSize := int(binary.BigEndian.Uint16(pkt.Payload[currOffset:]))
			currOffset += 2

			if len(pkt.Payload) < currOffset+naluSize {
				return nil, errors.New("STAP-A declared size is larger than buffer")
			}

			au = append(au, pkt.Payload[currOffset:currOffset+naluSize])
			currOffset += naluSize
		}
		for _, accessUnit := range au {
			h.extractSPSandPPS(accessUnit)
		}
		return au, nil
	case 25 <= naluType && naluType <= 27:
		panic(fmt.Errorf("not implemented naluType:%d", naluType))
	case naluType == h264.NALUTypeFUA || naluType == h264.NALUTypeFUB:
		if len(pkt.Payload) < 2 {
			return nil, invalidPayloadLength
		}
		s := pkt.Payload[fragmentHeaderIdx] & 0x80
		e := pkt.Payload[fragmentHeaderIdx] & 0x40
		fragmentNALU := pkt.Payload[fragmentHeaderIdx] & 0x1F
		if s != 0 {
			b := fragmentNALU | pkt.Payload[commonHeaderIdx]&0xe0
			h.fragments = append([]byte{}, b)
		}
		h.fragments = append(h.fragments, pkt.Payload[fragmentHeaderIdx+1:]...)
		if e != 0 {
			h.extractSPSandPPS(pkt.Payload)
			return [][]byte{h.fragments}, nil
		}
		return nil, nil
	default:
		fmt.Println("naluType:", naluType, ", ts:", pkt.Timestamp, ", marker:", pkt.Marker)
		return nil, nil
	}
	return nil, nil
}

// extract SPS and PPS without decoding RTP packets
func (h *H264) extractSPSandPPS(payload []byte) {
	if len(payload) < 1 {
		return
	}

	typ := h264.NALUType(payload[0] & 0x1F)

	switch typ {
	case h264.NALUTypeSPS:
		if !bytes.Equal(h.sps, payload) {
			h.sps = payload

			spsSet := &h264.SPS{}
			if err := spsSet.Unmarshal(h.sps); err != nil {
				err := fmt.Errorf("unable to parse H264 SPS: %w", err)
				fmt.Println(err)
				return
			}
			h.spsSet = spsSet
		}
	case h264.NALUTypePPS:
		if !bytes.Equal(h.pps, payload) {
			h.pps = payload
		}
	default:
	}

	if len(h.sps) > 0 && len(h.pps) > 0 {
		h.ready.Store(true)
	}
}
