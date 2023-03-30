package rotor

import (
	"io"
	"testing"
)

func TestRotor_ImplementsWriteCloser(_ *testing.T) {
	var _ io.WriteCloser = new(Rotor)
}
