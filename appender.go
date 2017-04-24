package log

import (
	"io"
	"os"
	"path"
	"sync"
	"time"
)

var (
	HourlySuffix = ".20060102-15"
	DailySuffix  = ".20060102"
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
	rtfn     func(time.Time) (time.Time, string)
	file     *os.File
}

func hourly() time.Time {
	n := time.Now()
	y, m, d := n.Date()
	h, _, _ := n.Clock()
	return time.Date(y, m, d, h+1, 0, 0, 0, time.Local)
}

func daily() time.Time {
	y, m, d := time.Now().Date()
	return time.Date(y, m, d+1, 0, 0, 0, 0, time.Local)
}

func NewHourlyRotateAppender(filename string) (*RotateAppender, error) {
	a := &RotateAppender{
		filename: path.Clean(filename),
		rt:       hourly(),
	}

	a.rtfn = func(t time.Time) (time.Time, string) {
		return hourly(), t.Format(HourlySuffix)
	}

	return a.open()
}

func NewDailyRotateAppender(filename string) (*RotateAppender, error) {
	a := &RotateAppender{
		filename: path.Clean(filename),
		rt:       daily(),
	}

	a.rtfn = func(t time.Time) (time.Time, string) {
		return daily(), t.Format(DailySuffix)
	}

	return a.open()
}

func (a *RotateAppender) open() (*RotateAppender, error) {
	err := os.MkdirAll(path.Dir(a.filename), 0755)
	if err != nil && !os.IsExist(err) {
		return nil, err
	}
	a.file, err = os.OpenFile(a.filename,
		os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	return a, err
}

func (a *RotateAppender) Close() error {
	a.mu.Lock()
	e := a.file.Close()
	a.file = nil
	a.mu.Unlock()
	return e
}

func (a *RotateAppender) Output(_ Level, t time.Time, data []byte) {
	a.mu.Lock()
	if t.After(a.rt) {
		var suffix string
		a.rt, suffix = a.rtfn(a.rt)
		filename := a.filename + suffix
		err := a.file.Close()
		if err != nil {
			println("appender close ", a.filename, "error: ", err)
		}
		if err = os.Rename(a.filename, filename); err != nil {
			println("appender rename ", filename, "error: ", err)
		}

		a.file, err = os.OpenFile(a.filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			println("appender open ", a.filename, "error: ", err)
		}
	}
	if a.file == nil {
		a.mu.Unlock()
		return
	}
	a.file.Write(data)
	a.mu.Unlock()
}
