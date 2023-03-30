package rotor

type (
	eventType byte

	event struct {
		data      []byte
		eventType eventType
	}
)

const (
	eventClose eventType = 1 << iota
	eventSync
	eventWrite
)
