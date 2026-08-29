//go:build !linux

package terminal

import (
	"errors"
	"os"
)

func MakeRaw(file *os.File) (func(), error) {
	return nil, errors.New("interactive watch is currently supported on Linux")
}
