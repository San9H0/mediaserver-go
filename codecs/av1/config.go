package av1

import (
	"github.com/bluenviron/mediacommon/pkg/codecs/av1"
)

type Config struct {
	header av1.SequenceHeader
}

func NewAV1Config() *Config {
	return &Config{}
}

func (c *Config) UnmarshalSequenceHeader(data []byte) error {
	var header av1.SequenceHeader
	if err := header.Unmarshal(data); err != nil {
		return err
	}
	c.header = header
	return nil
}

func (c *Config) Width() int {
	return c.header.Width()
}

func (c *Config) Height() int {
	return c.header.Height()
}
