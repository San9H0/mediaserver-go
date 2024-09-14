package parsers

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"github.com/bluenviron/mediacommon/pkg/codecs/h264"
	"github.com/pion/rtp"
	"mediaserver-go/hubs/codecs"
	"sync"
)

// H264Parser is not thread-safe
type H264Parser struct {
	mu sync.RWMutex

	fragments []byte

	sps []byte
	pps []byte

	codec *codecs.H264
}

func NewH264Parser() *H264Parser {
	return &H264Parser{}
}

// GetCodec 는 sps, pps 를 파싱하지 못한경우 nil 임.
func (h *H264Parser) GetCodec() codecs.Codec {
	return h.codec
}

/*
Parse 함수는 RTP Packet을 받아서 NAL unit을 추출하는 함수이다.
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

func (h *H264Parser) Parse(rtpPacket *rtp.Packet) [][]byte {
	payload := rtpPacket.Payload
	if len(payload) < 1 {
		return nil
	}

	const (
		commonHeaderIdx   = 0
		stapHeaderIdx     = 1
		fragmentHeaderIdx = 1
	)

	naluType := h264.NALUType(payload[commonHeaderIdx] & 0x1F)

	switch {
	case 1 <= naluType && naluType <= 23:
		switch naluType {
		case h264.NALUTypeSEI, h264.NALUTypeFillerData, h264.NALUTypeAccessUnitDelimiter:
			return nil
		case h264.NALUTypeSPS, h264.NALUTypePPS:
			h.setSPSPPS(h.extractSPSPPS(payload))
			fallthrough
		default:
			return [][]byte{payload}
		}
	case naluType == h264.NALUTypeSTAPA: // RTP 를 위한 nalunit
		aus := [][]byte{}
		var sps_, pps_ []byte = nil, nil
		currOffset := stapHeaderIdx
		for currOffset < len(payload) {
			naluSize := int(binary.BigEndian.Uint16(payload[currOffset:]))
			currOffset += 2

			if len(payload) < currOffset+naluSize {
				fmt.Println("STAP-A declared size is larger than buffer 3")
				return nil
			}

			b := make([]byte, len(payload[currOffset:currOffset+naluSize]))
			copy(b, payload[currOffset:currOffset+naluSize])
			aus = append(aus, b)

			currOffset += naluSize
		}
		for _, au := range aus {

			sps, pps := h.extractSPSPPS(au)
			if sps != nil {
				sps_ = sps
			}
			if pps != nil {
				pps_ = pps
			}
		}
		h.setSPSPPS(sps_, pps_)
		return aus
	case 25 <= naluType && naluType <= 27:
		panic(fmt.Errorf("not implemented naluType:%d", naluType))
	case naluType == h264.NALUTypeFUA: // RTP 를 위한 nalunit
		if len(payload) < 2 {
			return nil
		}
		s := payload[fragmentHeaderIdx] & 0x80
		e := payload[fragmentHeaderIdx] & 0x40
		fragmentNALU := h264.NALUType(payload[fragmentHeaderIdx] & 0x1F)
		if fragmentNALU == h264.NALUTypeFillerData {
			return nil
		}
		if s != 0 {
			b := byte(fragmentNALU) | payload[commonHeaderIdx]&0xe0
			h.fragments = append([]byte{}, b)
		}
		h.fragments = append(h.fragments, payload[fragmentHeaderIdx+1:]...)
		if e != 0 {
			if fragmentNALU == h264.NALUTypeSPS || fragmentNALU == h264.NALUTypePPS {
				h.setSPSPPS(h.extractSPSPPS(payload))
			}
			return [][]byte{h.fragments}
		}
		return nil
	default:
		panic(fmt.Errorf("not implemented naluType:%d", naluType))
	}
	return nil
}

// extract SPS and PPS without decoding RTP packets
func (h *H264Parser) extractSPSPPS(payload []byte) ([]byte, []byte) {
	if len(payload) < 1 {
		return nil, nil
	}

	naluType := h264.NALUType(payload[0] & 0x1F)
	var sps, pps []byte
	switch naluType {
	case h264.NALUTypeSPS:
		if !bytes.Equal(h.sps, payload) {
			sps = payload
		}
	case h264.NALUTypePPS:
		if !bytes.Equal(h.pps, payload) {
			pps = payload
		}
	default:
		return nil, nil
	}
	return sps, pps
}

func (h *H264Parser) setSPSPPS(sps, pps []byte) {
	spsChanged, ppsChanged := false, false
	if len(sps) != 0 && !bytes.Equal(h.sps, sps) {
		h.sps = sps
		spsChanged = true
	}
	if len(pps) != 0 && !bytes.Equal(h.pps, pps) {
		h.pps = pps
		ppsChanged = true
	}
	if !spsChanged && !ppsChanged {
		return
	}

	var err error
	h.codec, err = codecs.NewH264(h.sps, h.pps)
	if err != nil {
		fmt.Println("extractSPSandPPS err:", err)
		return
	}
}
