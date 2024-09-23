package codecs

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

	codec Codec
}

func NewH264Config() *H264Config {
	return &H264Config{}
}

func (h *H264Config) GetCodec() (Codec, error) {
	return h.codec, nil
}

func (h *H264Config) UnmarshalFromConfig(body []byte) error {
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

	if (len(spsNALUnit) != 0 && !bytes.Equal(h.SPS, spsNALUnit)) || (len(ppsNALUnit) != 0 && !bytes.Equal(h.PPS, ppsNALUnit)) {
		h.SPS = spsNALUnit
		h.PPS = ppsNALUnit
		h.ProfileID = int(body[1])
		h.ProfileComp = int(body[2])
		h.LevelID = int(body[3])
		h.codec, _ = NewH264(h.SPS, h.PPS)
	}
	return nil
}

func (h *H264Config) UnmarshalFromNALU(nalus ...[]byte) error {
	if len(nalus) < 2 {
		return errInvalidSize
	}
	var sps, pps []byte
	for _, nalu := range nalus {
		switch h264.NALUType(nalu[0] & 0x1F) {
		case h264.NALUTypeSPS:
			if len(nalu) != 0 && !bytes.Equal(h.SPS, nalu) {
				sps = nalu
			}
		case h264.NALUTypePPS:
			if len(nalu) != 0 && !bytes.Equal(h.PPS, nalu) {
				pps = nalu
			}
		}
	}

	if (len(sps) != 0 && !bytes.Equal(h.SPS, sps)) || (len(pps) != 0 && !bytes.Equal(h.PPS, pps)) {
		h.SPS = sps
		h.PPS = pps
		h.ProfileID = int(sps[1])
		h.ProfileComp = int(sps[2])
		h.LevelID = int(sps[3])
		h.codec, _ = NewH264(h.SPS, h.PPS)
	}
	return nil
}

func (h *H264Config) UnmarshalFromSPSPPS(sps, pps []byte) error {
	if len(sps) == 0 || len(pps) == 0 {
		return errNaluSize
	}

	if (len(sps) != 0 && !bytes.Equal(h.SPS, sps)) || (len(pps) != 0 && !bytes.Equal(h.PPS, pps)) {
		h.SPS = sps
		h.PPS = pps
		h.ProfileID = int(sps[1])
		h.ProfileComp = int(sps[2])
		h.LevelID = int(sps[3])
		codec, _ := NewH264FromConfig(h)
		h.codec = codec
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
func (h *H264Config) Marshal() ([]byte, error) {
	if len(h.SPS) == 0 || len(h.PPS) == 0 {
		return nil, fmt.Errorf("SPS or PPS is empty. %w", errNeedSPSPPS)
	}
	b := make([]byte, len(h.SPS)+len(h.PPS)+8+3)
	b[0] = 0x01
	b[1] = h.SPS[1]
	b[2] = h.SPS[2]
	b[3] = h.SPS[3]
	b[4] = 0xfc | 3 // NALU 의 길이는 4비트이다. 4-1=3
	b[5] = 0xe0 | 1 // sps 개수는 1개이다.
	b[6] = byte(len(h.SPS) >> 8)
	b[7] = byte(len(h.SPS) & 0xff)
	copy(b[8:], h.SPS)
	b[8+len(h.SPS)] = 1 // pps 개수는 1개이다.
	b[9+len(h.SPS)] = byte(len(h.SPS) >> 8)
	b[10+len(h.SPS)] = byte(len(h.SPS) & 0xff)
	copy(b[11+len(h.SPS):], h.SPS)
	return b, nil
}
