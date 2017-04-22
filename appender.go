package log

import (
	"io"
	"os"
	"path"
	"sync"
	"time"
)

type Appender interface {
	Output(level Level, t time.Time, data []byte)
}

type console struct {
	io.Writer
	mu sync.Mutex
}

func NewConsoleAppender() Appender {
	return &console{Writer: os.Stdout}
}

func (c *console) Output(level Level, t time.Time, data []byte) {
	c.mu.Lock()
	c.Write(data)
	c.mu.Unlock()
}

type RotateAppender struct {
	mu       sync.Mutex
	rt       time.Time
	filename string
	file     *os.File
}

func (a *RotateAppender) Open() error {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.open()
}

func (a *RotateAppender) open() error {
	err := os.MkdirAll(path.Dir(a.filename), 0755)
	if err != nil && err != os.ErrExist {
		return err
	}
	a.file, err = os.OpenFile(a.filename, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	return err
}

func (a *RotateAppender) Close() {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.file.Close()
}

//func (a *RotateAppender) Output(level Level, t time.Time, data []byte) {
//	a.mu.Lock()

//	a.Write(data)
//	a.mu.Unlock()
//}
