// +build s390x,linux

package log

import (
	"os"
	"syscall"
)

func fadvise(f *os.File) error {
	_, _, err := syscall.Syscall6(syscall.SYS_FADVISE64, f.Fd(), 0, 0, 6, 0, 0)
	return err
}
