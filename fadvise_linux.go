//go:build linux
// +build linux

package log

import (
	"os"

	"golang.org/x/sys/unix"
)

func fadvise(f *os.File) error {
	return unix.Fadvise(int(f.Fd()), 0, 0, unix.FADV_DONTNEED)
}
