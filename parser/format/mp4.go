package format

import (
	"encoding/binary"
)

/*
bits
8   version ( always 0x01 )
8   avc profile ( sps[0][1] )
8   avc compatibility ( sps[0][2] )
8   avc level ( sps[0][3] )
6   reserved ( all bits on )
2   NALULengthSizeMinusOne
3   reserved ( all bits on )
5   number of SPS NALUs (usually 1)
repeated once per SPS:

	16         SPS size
	variable   SPS NALU data

8   number of PPS NALUs (usually 1)
repeated once per PPS:

	16       PPS size
	variable PPS NALU data
*/
func ExtraDataForAVCC(sps, pps []byte) []byte {
	if len(sps) < 1 || len(pps) < 1 {
		return nil
	}
	b := make([]byte, len(sps)+len(pps)+8+3)
	b[0] = 0x01
	b[1] = sps[1]
	b[2] = sps[2]
	b[3] = sps[3]
	b[4] = 0xfc | 3 // NALU 의 길이는 4비트이다. 4-1=3
	b[5] = 0xe0 | 1 // sps 개수는 1개이다.
	b[6] = byte(len(sps) >> 8)
	b[7] = byte(len(sps) & 0xff)
	copy(b[8:], sps)
	b[8+len(sps)] = 1 // pps 개수는 1개이다.
	b[9+len(sps)] = byte(len(pps) >> 8)
	b[10+len(sps)] = byte(len(pps) & 0xff)
	copy(b[11+len(sps):], pps)
	return b
}

func SPSPPSFromAVCCExtraData(data []byte) (sps []byte, pps []byte) {
	if len(data) < 11 {
		return nil, nil
	}
	if data[0] != 0x01 {
		return nil, nil
	}
	spsLen := int(data[6])<<8 | int(data[7])
	sps = data[8 : 8+spsLen]
	ppsLen := int(data[9+spsLen])<<8 | int(data[10+spsLen])
	pps = data[11+spsLen : 11+spsLen+ppsLen]
	return sps, pps
}

// GetAUFromAVCC extracts Access Units from AVCC format.
func GetAUFromAVCC(payload []byte) [][]byte {
	if len(payload) < 1 {
		return nil
	}

	const (
		avcSizeLength = 4
	)

	var aus [][]byte
	offset := uint32(0)
	for {
		if offset >= uint32(len(payload))-avcSizeLength {
			break
		}
		auLength := binary.BigEndian.Uint32(payload[offset:])
		offset += avcSizeLength
		au := payload[offset : auLength+offset]
		offset += auLength
		aus = append(aus, au)
	}

	return aus
}
