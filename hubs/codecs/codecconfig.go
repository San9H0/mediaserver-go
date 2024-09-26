package codecs

type Config interface {
	GetCodec() (Codec, error)
}
