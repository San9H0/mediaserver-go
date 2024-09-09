package codecparser

import "github.com/bluenviron/mediacommon/pkg/codecs/h264"

// extract SPS and PPS without decoding RTP packets
func H264ExtractParams(payload []byte) ([]byte, []byte) {
	if len(payload) < 1 {
		return nil, nil
	}

	typ := h264.NALUType(payload[0] & 0x1F)

	switch typ {
	case h264.NALUTypeSPS:
		return payload, nil

	case h264.NALUTypePPS:
		return nil, payload

	case h264.NALUTypeSTAPA:
		payload = payload[1:]
		var sps []byte
		var pps []byte

		for len(payload) > 0 {
			if len(payload) < 2 {
				break
			}

			size := uint16(payload[0])<<8 | uint16(payload[1])
			payload = payload[2:]

			if size == 0 {
				break
			}

			if int(size) > len(payload) {
				return nil, nil
			}

			nalu := payload[:size]
			payload = payload[size:]

			typ = h264.NALUType(nalu[0] & 0x1F)

			switch typ {
			case h264.NALUTypeSPS:
				sps = nalu

			case h264.NALUTypePPS:
				pps = nalu
			}
		}

		return sps, pps

	default:
		return nil, nil
	}
}
