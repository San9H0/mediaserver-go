package opus

import "mediaserver-go/hubs/codecs"

type OpusConfig struct {
	parameters Parameters

	codec codecs.Codec
}

func NewOpusConfig(parameters Parameters) *OpusConfig {
	return &OpusConfig{
		parameters: parameters,
		codec:      NewOpus(parameters),
	}
}

func (o *OpusConfig) GetCodec() (codecs.Codec, error) {
	return o.codec, nil
}
