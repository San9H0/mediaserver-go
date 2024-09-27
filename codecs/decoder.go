package codecs

type Decoder interface {
	KeyFrame(payload []byte) bool
}
