package av1

import (
	"fmt"
	"github.com/bluenviron/mediacommon/pkg/codecs/av1"
	"mediaserver-go/utils"
)

type Config struct {
	seqData []byte
	header  av1.SequenceHeader
}

func NewAV1Config() *Config {
	return &Config{}
}

func (c *Config) UnmarshalSequenceHeader(data []byte) error {
	var header av1.SequenceHeader
	if err := header.Unmarshal(data); err != nil {
		return err
	}

	//if b, err := json.MarshalIndent(header, "", "  "); err == nil {
	//	fmt.Println("[TESTDEBUG] av1 SequenceHeader...:", string(b))
	//}

	c.header = header
	c.seqData = data
	return nil
}

func (c *Config) Width() int {
	return c.header.Width()
}

func (c *Config) Height() int {
	return c.header.Height()
}

func (c *Config) MarshalToExtraData() ([]byte, error) {
	if c.seqData == nil {
		return nil, fmt.Errorf("SequenceHeader is nil")
	}
	data := make([]byte, 0, len(c.seqData)+4)
	version := 1
	// marker and version
	data = append(data, byte(0x80|(version&0x7F)))
	// seq profile and level
	data = append(data, (c.header.SeqProfile&0x07)<<5|c.header.SeqLevelIdx[0]&0x1F)

	ext := byte(0)
	tier := utils.GetIntFromBool(c.header.SeqTier[0])
	depth0 := utils.GetIntFromBool(c.header.ColorConfig.BitDepth > 8)
	depth1 := utils.GetIntFromBool(c.header.ColorConfig.BitDepth == 12)
	monochrome := utils.GetIntFromBool(c.header.ColorConfig.MonoChrome)
	subSampleX := utils.GetIntFromBool(c.header.ColorConfig.SubsamplingX)
	subSampleY := utils.GetIntFromBool(c.header.ColorConfig.SubsamplingY)
	samplePosition := c.header.ColorConfig.ChromaSamplePosition
	ext |= byte(tier&0x01) << 7
	ext |= byte(depth0&0x01) << 6
	ext |= byte(depth1&0x01) << 5
	ext |= byte(monochrome&0x01) << 4
	ext |= byte(subSampleX&0x01) << 3
	ext |= byte(subSampleY&0x01) << 2
	ext |= byte(samplePosition & 0x03)
	data = append(data, ext)
	data = append(data, 0x00) // padding
	data = append(data, c.seqData...)
	return data, nil
}
