package h264

import (
	"encoding/binary"
	"fmt"
	"github.com/bluenviron/mediacommon/pkg/codecs/h264"
	"github.com/pion/rtp"
	"mediaserver-go/hubs/codecs/rtpparsers"
)

type H264Parser struct {
	fragments []byte

	cb rtpparsers.Callback
}

func NewH264Parser(cb func([][]byte) [][]byte) *H264Parser {
	return &H264Parser{
		cb: cb,
	}
}

//func (h *H264Parser) OnCodec(cb func(codec codecs.Codec)) {
//	if cb == nil {
//		h.onCodec.Store(nil)
//		return
//	}
//	h.onCodec.Store(&cb)
//}
//
//func (h *H264Parser) invokeCodec(codec codecs.Codec) {
//	if fn := h.onCodec.Load(); fn != nil {
//		(*fn)(codec)
//	}
//}

func (h *H264Parser) Parse(rtpPacket *rtp.Packet) [][]byte {
	payloads := h.parse(rtpPacket.Payload)
	//h.config.UnmarshalFromNALU(payloads...)
	//if codec, _ := h.config.GetCodec(); codec != nil {
	//	if h.prevCodec != codec {
	//		h.invokeCodec(codec)
	//		h.prevCodec = codec
	//	}
	//}
	return h.cb(payloads)
}

/*
	https://datatracker.ietf.org/doc/html/rfc6184#section-5.2

H264 는 RTP 로 전송될때 24~29까지 추가된 NAL unit을 더 사용한다
- 24: STAP-A (Single-time aggregation packet)
- 25: STAP-B (Single-time aggregation packet)
- 26: MTAP16 (Multi-time aggregation packet)
- 27: MTAP24 (Multi-time aggregation packet)
- 28: FU-A (Fragmentation units)
- 29: FU-B (Fragmentation units)
+---------------+
|0|1|2|3|4|5|6|7|
+-+-+-+-+-+-+-+-+
|F|NRI|  Type   |
+---------------+

첫번째 바이트에 NAL Unit Type 이 들어가 있음.
F: 1 bit
NRI: 2 bit
Type: 5 bit
1. Type 이 1 ~ 23 까지의 값이 들어가 있으면 Single NAL unit packet. 하나의 패킷에 하나의 NAL unit.
2. Type 이 24 ~ 27 까지의 값이 들어가 있으면 Aggregation Packet. 여러개의 NAL unit 들이 하나의 RTP Payload에 들어가 있는 경우.
2 바이트의 NALU size 가 들어가 있음. [size(2byte)][nalunit(1byte)][unit...(n-1byte)] 형태로 들어가있음.
3. Type 이 28 ~ 29 까지의 값이 들어가 있으면 Fragmentation Unit. 하나의 NAL unit 이 여러개의 RTP Payload에 나눠서 들어가 있는 경우.
+---------------+
|0|1|2|3|4|5|6|7|
+-+-+-+-+-+-+-+-+
|S|E|R|  Type   |
+---------------+
FU 인경우 NAL Unit 다음 패킷별로 위 포멧으로 NAL Unit 이 나눠서 들어가게 된다.
S: 1 bit. Start bit. 첫번째 패킷인지 여부.
E: 1 bit. End bit. 마지막 패킷인지 여부.
R: 1 bit. Reserved bit.
Type: 5 bit. NAL Unit Type.
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
			return [][]byte{h.fragments}
		}
		return nil
	default:
		panic(fmt.Errorf("not implemented naluType:%d", naluType))
	}
	return nil
}
