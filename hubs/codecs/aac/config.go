package aac

import "mediaserver-go/hubs/codecs"

type Config struct {
}

func (c *Config) GetCodec() (codecs.Codec, error) {
	return nil, nil
}
