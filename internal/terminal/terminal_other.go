//go:build !linux && !darwin

package terminal

import (
	"errors"
	"os"
)

func MakeRaw(file *os.File) (func(), error) {
	return nil, errors.New("interactive watch is currently supported on Linux and macOS")
}

func Size(file *os.File) (cols, rows int, err error) {
	return 0, 0, errors.New("terminal size detection is currently supported on Linux and macOS")
}
