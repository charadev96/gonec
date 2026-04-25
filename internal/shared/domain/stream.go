package domain

type Packet[T any] struct {
	Msg T
	Err error
}
