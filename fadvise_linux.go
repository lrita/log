// +build linux
// +build 386 amd64

package log

import (
	"os"
	"syscall"
)

func fadvise(f *os.File) error {
	_, _, err := syscall.Syscall6(syscall.SYS_FADVISE64, f.Fd(), 0, 0, 4, 0, 0)
	return err
}
