package utils

import (
	"crypto/rand"
	"encoding/binary"
	"log"
)

func SendOrDrop[T any](ch chan T, data T) {
	select {
	case ch <- data:
	default:
	}
}

func RandomUint16() uint16 {
	var b [2]byte
	_, err := rand.Read(b[:])
	if err != nil {
		log.Fatalf("failed to generate random number: %v", err)
		return 0
	}

	return binary.BigEndian.Uint16(b[:])
}

func RandomUint32() uint32 {
	var b [4]byte
	_, err := rand.Read(b[:])
	if err != nil {
		log.Fatalf("failed to generate random number: %v", err)
		return 0
	}

	return binary.BigEndian.Uint32(b[:])
}

func GetPointer[T any](v T) *T {
	return &v
}
