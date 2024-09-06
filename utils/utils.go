package utils

func SendOrDrop[T any](ch chan T, data T) {
	select {
	case ch <- data:
	default:
	}
}
