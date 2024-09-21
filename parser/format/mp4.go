package format

import (
	"encoding/binary"
	"fmt"
	"github.com/bluenviron/mediacommon/pkg/codecs/h264"
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
func ExtraDataForAVC(sps, pps []byte) []byte {
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

var (
	errExtraDataInvalid = fmt.Errorf("invalid extra data")
	errExtraDataShort   = fmt.Errorf("extra data too short")
	errExtraDataNALU    = fmt.Errorf("invalid NALU length")
)

type ExtraData struct {
	ProfileID   int
	ProfileComp int
	LevelID     int
	SPS         []byte
	PPS         []byte
}

func (e *ExtraData) Unmarshal(body []byte) error {
	if len(body) < 11 {
		return errExtraDataShort
	}
	if body[0] != 0x01 {
		return errExtraDataInvalid
	}
	//naluLengthSize := int((body[4] & 0x03) + 1)
	spsCount := int(body[5] & 0x1F)
	if spsCount == 0 {
		return nil
	}
	offset := 6
	spsLength := int(binary.BigEndian.Uint16(body[offset : offset+2]))
	offset += 2

	if len(body) < offset+spsLength {
		return fmt.Errorf("body length %d offset %d spsLength %d. %w", len(body), offset, spsLength, errExtraDataNALU)
	}

	spsNALUnit := body[offset : offset+spsLength]
	offset += spsLength

	if offset >= len(body) {
		return fmt.Errorf("offset %d >= len(body) %d, %w", offset, len(body), errExtraDataNALU)
	}

	ppsCount := int(body[offset])
	offset++

	if ppsCount == 0 {
		return nil
	}

	ppsLength := int(binary.BigEndian.Uint16(body[offset : offset+2]))
	offset += 2

	if len(body) < offset+ppsLength {
		return fmt.Errorf("body length %d offset %d ppsLength %d. %w", len(body), offset, ppsLength, errExtraDataNALU)
	}

	ppsNALUnit := body[offset : offset+ppsLength]
	offset += ppsLength

	e.ProfileID = int(body[1])
	e.ProfileComp = int(body[2])
	e.LevelID = int(body[3])
	e.SPS = spsNALUnit
	e.PPS = ppsNALUnit
	return nil
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

func DropNalUnit(naluType h264.NALUType) bool {
	switch naluType {
	case h264.NALUTypeSEI, h264.NALUTypeFillerData, h264.NALUTypeAccessUnitDelimiter:
		return true
	}
	return false
}

// GetAUFromAVC extracts Access Units from AVCC format.
func GetAUFromAVC(payload []byte) [][]byte {
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
		nalUnit := h264.NALUType(au[0] & 0x1F)
		if DropNalUnit(nalUnit) {
			continue
		}
		aus = append(aus, au)
	}

	return aus
}
