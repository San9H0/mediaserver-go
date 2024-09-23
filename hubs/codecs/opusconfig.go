package codecs

type OpusConfig struct {
	parameters OpusParameters

	codec Codec
}

func NewOpusConfig(parameters OpusParameters) *OpusConfig {
	return &OpusConfig{
		parameters: parameters,
		codec:      NewOpus(parameters),
	}
}

func (o *OpusConfig) GetCodec() (Codec, error) {
	return o.codec, nil
}
