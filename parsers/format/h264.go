package format

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/bluenviron/mediacommon/pkg/codecs/h264"
)

var (
	errExtraDataInvalid = errors.New("invalid extra data")
	errExtraDataShort   = errors.New("extra data too short")
	errExtraDataNALU    = errors.New("invalid NALU length")

	errInvalidSize = errors.New("invalid size")
	errNaluSize    = errors.New("nalu size is too short")
	errNeedSPSPPS  = errors.New("need SPS and PPS")

	errFailedMarshal = errors.New("failed to marshal")
)

type H264Config struct {
	ProfileID   int
	ProfileComp int
	LevelID     int
	SPS         []byte
	PPS         []byte
}

func (e *H264Config) UnmarshalFromConfig(body []byte) error {
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

func (e *H264Config) UnmarshalFromNALU(nalus ...[]byte) error {
	if len(nalus) < 2 {
		return errInvalidSize
	}
	for _, nalu := range nalus {
		switch h264.NALUType(nalu[0] & 0x1F) {
		case h264.NALUTypeSPS:
			if len(nalu) != 0 && !bytes.Equal(e.SPS, nalu) {
				e.SPS = nalu
				e.ProfileID = int(e.SPS[1])
				e.ProfileComp = int(e.SPS[2])
				e.LevelID = int(e.SPS[3])
			}
		case h264.NALUTypePPS:
			if len(nalu) != 0 && !bytes.Equal(e.PPS, nalu) {
				e.PPS = nalu
			}
		}
	}
	return nil
}

func (e *H264Config) UnmarshalFromSPSPPS(sps, pps []byte) error {
	if len(sps) == 0 || len(pps) == 0 {
		return errNaluSize
	}

	if len(sps) != 0 && !bytes.Equal(e.SPS, sps) {
		e.SPS = sps
		e.ProfileID = int(e.SPS[1])
		e.ProfileComp = int(e.SPS[2])
		e.LevelID = int(e.SPS[3])
	}

	if len(pps) != 0 && !bytes.Equal(e.PPS, pps) {
		e.PPS = pps
	}
	return nil
}

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
func (e *H264Config) Marshal() ([]byte, error) {
	if len(e.SPS) == 0 || len(e.PPS) == 0 {
		return nil, fmt.Errorf("SPS or PPS is empty. %w", errNeedSPSPPS)
	}
	b := make([]byte, len(e.SPS)+len(e.PPS)+8+3)
	b[0] = 0x01
	b[1] = e.SPS[1]
	b[2] = e.SPS[2]
	b[3] = e.SPS[3]
	b[4] = 0xfc | 3 // NALU 의 길이는 4비트이다. 4-1=3
	b[5] = 0xe0 | 1 // sps 개수는 1개이다.
	b[6] = byte(len(e.SPS) >> 8)
	b[7] = byte(len(e.SPS) & 0xff)
	copy(b[8:], e.SPS)
	b[8+len(e.SPS)] = 1 // pps 개수는 1개이다.
	b[9+len(e.SPS)] = byte(len(e.SPS) >> 8)
	b[10+len(e.SPS)] = byte(len(e.SPS) & 0xff)
	copy(b[11+len(e.SPS):], e.SPS)
	return b, nil
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
