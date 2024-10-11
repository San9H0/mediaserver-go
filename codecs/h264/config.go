package h264

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/bluenviron/mediacommon/pkg/codecs/h264"
	"mediaserver-go/thirdparty/ffmpeg/avutil"
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

type Config struct {
	sps, pps []byte

	spsInfo     h264.SPS
	profileID   int
	profileComp int
	levelID     int
	width       int
	height      int
	pixelFmt    int
}

func (c *Config) String() string {
	return fmt.Sprintf("H264 width:%d height:%d profile:%d(%x) profileComp:%d(%x) level:%d(%x) pixelFmt:%d",
		c.width, c.height, c.profileID, c.profileID, c.profileComp, c.profileComp, c.levelID, c.levelID, c.pixelFmt)
}

func (c *Config) init(sps, pps []byte) error {
	if len(sps) == 0 || len(pps) == 0 {
		return errInvalidSize
	}
	spsInfo := h264.SPS{}
	if err := spsInfo.Unmarshal(sps); err != nil {
		return fmt.Errorf("failed to unmarshal. %w", err)
	}

	c.spsInfo = spsInfo
	c.sps = bytes.Clone(sps)
	c.pps = bytes.Clone(pps)
	c.profileID = int(sps[1])
	c.profileComp = int(sps[2])
	c.levelID = int(sps[3])
	c.width = spsInfo.Width()
	c.height = spsInfo.Height()
	c.pixelFmt = makePixelFmt(&spsInfo)
	return nil
}

func (c *Config) UnmarshalFromSPSPPS(sps, pps []byte) error {
	return c.init(sps, pps)
}

func (c *Config) UnmarshalFromExtraData(extradata []byte) error {
	if len(extradata) < 11 {
		return errExtraDataShort
	}
	if extradata[0] != 0x01 {
		return errExtraDataInvalid
	}
	spsCount := int(extradata[5] & 0x1F)
	if spsCount == 0 {
		return nil
	}
	offset := 6
	spsLength := int(binary.BigEndian.Uint16(extradata[offset : offset+2]))
	offset += 2

	if len(extradata) < offset+spsLength {
		return fmt.Errorf("extradata length %d offset %d spsLength %d. %w", len(extradata), offset, spsLength, errExtraDataNALU)
	}

	spsNALUnit := extradata[offset : offset+spsLength]
	offset += spsLength

	if offset >= len(extradata) {
		return fmt.Errorf("offset %d >= len(extradata) %d, %w", offset, len(extradata), errExtraDataNALU)
	}

	ppsCount := int(extradata[offset])
	offset++

	if ppsCount == 0 {
		return nil
	}

	ppsLength := int(binary.BigEndian.Uint16(extradata[offset : offset+2]))
	offset += 2

	if len(extradata) < offset+ppsLength {
		return fmt.Errorf("extradata length %d offset %d ppsLength %d. %w", len(extradata), offset, ppsLength, errExtraDataNALU)
	}

	ppsNALUnit := extradata[offset : offset+ppsLength]
	offset += ppsLength

	return c.init(spsNALUnit, ppsNALUnit)
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
func (c *Config) MarshalToExtraData() ([]byte, error) {
	if len(c.sps) == 0 || len(c.pps) == 0 {
		return nil, fmt.Errorf("SPS or PPS is empty. %w", errNeedSPSPPS)
	}
	b := make([]byte, len(c.sps)+len(c.pps)+8+3)
	b[0] = 0x01
	b[1] = c.sps[1]
	b[2] = c.sps[2]
	b[3] = c.sps[3]
	b[4] = 0xfc | 3 // NALU 의 길이는 4비트이다. 4-1=3
	b[5] = 0xe0 | 1 // sps 개수는 1개이다.
	b[6] = byte(len(c.sps) >> 8)
	b[7] = byte(len(c.sps) & 0xff)
	copy(b[8:], c.sps)
	b[8+len(c.sps)] = 1 // pps 개수는 1개이다.
	b[9+len(c.sps)] = byte(len(c.pps) >> 8)
	b[10+len(c.sps)] = byte(len(c.pps) & 0xff)
	copy(b[11+len(c.sps):], c.pps)
	return b, nil
}

func makePixelFmt(spsSet *h264.SPS) int {
	switch spsSet.ChromaFormatIdc {
	case 0:
		return avutil.AV_PIX_FMT_GRAY8
	case 1:
		return avutil.AV_PIX_FMT_YUV420P
	case 2:
		return avutil.AV_PIX_FMT_YUV422P
	case 3:
		return avutil.AV_PIX_FMT_YUV444P
	default:
		return avutil.AV_PIX_FMT_NONE
	}
}
