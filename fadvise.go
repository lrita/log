// +build !linux

package log

import "os"

func fadvise(_ *os.File) error {
	return nil
}
