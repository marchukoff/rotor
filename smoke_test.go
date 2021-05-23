package rotor

import (
	"io"
	"testing"
)

func TestRotor_ImplementsWriteCloser(t *testing.T) {
	var _ io.WriteCloser = new(Rotor)
}
