package rtppackets

import (
	"encoding/binary"
	"fmt"
	"github.com/bluenviron/mediacommon/pkg/codecs/h264"
	"github.com/pion/rtp"
	"mediaserver-go/hubs/codecs"
	"mediaserver-go/parsers/format"
	"sync/atomic"
)

type H264Parser struct {
	prevCodec codecs.Codec
	config    *codecs.H264Config

	fragments []byte

	onCodec atomic.Pointer[func(codec codecs.Codec)]
}

func NewH264Parser(config *codecs.H264Config) *H264Parser {
	return &H264Parser{
		config: config,
	}
}

func (h *H264Parser) OnCodec(cb func(codec codecs.Codec)) {
	if cb == nil {
		h.onCodec.Store(nil)
		return
	}
	h.onCodec.Store(&cb)
}

func (h *H264Parser) invokeCodec(codec codecs.Codec) {
	if fn := h.onCodec.Load(); fn != nil {
		(*fn)(codec)
	}
}

func (h *H264Parser) Parse(rtpPacket *rtp.Packet) [][]byte {
	payloads := h.parse(rtpPacket.Payload)
	h.config.UnmarshalFromNALU(payloads...)
	if codec, _ := h.config.GetCodec(); codec != nil {
		if h.prevCodec != codec {
			h.invokeCodec(codec)
			h.prevCodec = codec
		}
	}
	return payloads
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

func (h *H264Parser) parse(rtpPayload []byte) [][]byte {
	if len(rtpPayload) < 1 {
		return nil
	}

	const (
		commonHeaderIdx   = 0
		stapHeaderIdx     = 1
		fragmentHeaderIdx = 1
	)

	naluType := h264.NALUType(rtpPayload[commonHeaderIdx] & 0x1F)

	switch {
	case 1 <= naluType && naluType <= 23:
		if format.DropNalUnit(naluType) {
			return nil
		}
		return [][]byte{rtpPayload}
	case naluType == h264.NALUTypeSTAPA: // RTP 를 위한 nalunit
		var aus [][]byte
		currOffset := stapHeaderIdx
		for currOffset < len(rtpPayload) {
			naluSize := int(binary.BigEndian.Uint16(rtpPayload[currOffset:]))
			currOffset += 2

			if len(rtpPayload) < currOffset+naluSize {
				fmt.Println("STAP-A declared size is larger than buffer 3")
				return nil
			}

			au := make([]byte, len(rtpPayload[currOffset:currOffset+naluSize]))
			copy(au, rtpPayload[currOffset:currOffset+naluSize])
			currOffset += naluSize
			if format.DropNalUnit(h264.NALUType(au[0] & 0x1F)) {
				continue
			}
			aus = append(aus, au)
		}
		return aus
	case 25 <= naluType && naluType <= 27:
		panic(fmt.Errorf("not implemented naluType:%d", naluType))
	case naluType == h264.NALUTypeFUA: // RTP 를 위한 nalunit
		if len(rtpPayload) < 2 {
			return nil
		}
		s := rtpPayload[fragmentHeaderIdx] & 0x80
		e := rtpPayload[fragmentHeaderIdx] & 0x40
		fragmentNALU := h264.NALUType(rtpPayload[fragmentHeaderIdx] & 0x1F)
		if fragmentNALU == h264.NALUTypeFillerData {
			return nil
		}
		if s != 0 {
			b := byte(fragmentNALU) | rtpPayload[commonHeaderIdx]&0xe0
			h.fragments = append([]byte{}, b)
		}
		h.fragments = append(h.fragments, rtpPayload[fragmentHeaderIdx+1:]...)
		if e != 0 {
			if format.DropNalUnit(h264.NALUType(h.fragments[0] & 0x1F)) {
				return nil
			}
			return [][]byte{h.fragments}
		}
		return nil
	default:
		panic(fmt.Errorf("not implemented naluType:%d", naluType))
	}
	return nil
}
