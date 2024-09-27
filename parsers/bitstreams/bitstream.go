package bitstreams

type Bitstream interface {
	SetBitStream(unit []byte) []byte
}
