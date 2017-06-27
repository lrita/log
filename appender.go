package log

import (
	"bufio"
	"io"
	"os"
	"path/filepath"
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
	w        io.Writer
	file     *os.File
}

func hourly() time.Time {
	return time.Now().Add(time.Hour).Truncate(time.Hour)
}

func daily() time.Time {
	y, m, d := time.Now().Date()
	return time.Date(y, m, d+1, 0, 0, 0, 0, time.Local)
}

func NewHourlyRotateAppender(filename string) (*RotateAppender, error) {
	return NewHourlyRotateBufAppender(filename, 0)
}

func NewHourlyRotateBufAppender(filename string, bufsize int) (*RotateAppender, error) {
	a := &RotateAppender{
		filename: filepath.Clean(filename),
		rt:       hourly(),
	}

	a.rtfn = func(t time.Time) (time.Time, string) {
		return hourly(), t.Add(-time.Hour).Format(HourlySuffix)
	}

	return a.open(bufsize)
}

func NewDailyRotateAppender(filename string) (*RotateAppender, error) {
	return NewDailyRotateBufAppender(filename, 0)
}

func NewDailyRotateBufAppender(filename string, bufsize int) (*RotateAppender, error) {
	a := &RotateAppender{
		filename: filepath.Clean(filename),
		rt:       daily(),
	}

	a.rtfn = func(t time.Time) (time.Time, string) {
		return daily(), t.Add(-24 * time.Hour).Format(DailySuffix)
	}

	return a.open(bufsize)
}

func (a *RotateAppender) open(bufsize int) (*RotateAppender, error) {
	err := os.MkdirAll(filepath.Dir(a.filename), 0755)
	if err != nil && !os.IsExist(err) {
		return nil, err
	}
	a.file, err = os.OpenFile(a.filename,
		os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if bufsize > 0 {
		a.w = bufio.NewWriterSize(a.file, bufsize)
	} else {
		a.w = a.file
	}
	return a, err
}

func (a *RotateAppender) Close() error {
	a.mu.Lock()
	e := a.close()
	a.mu.Unlock()
	return e
}

func (a *RotateAppender) close() error {
	var e1, e2 error
	if bw, ok := a.w.(*bufio.Writer); ok {
		if e1 = bw.Flush(); e1 != nil {
			println("appender close bufio flush error: ", e1.Error())
		}
	}

	// ignore error
	a.file.Sync()
	fadvise(a.file)

	if e2 = a.file.Close(); e2 != nil {
		println("appender close filename: ", a.filename, "error: ", e2.Error())
	} else {
		a.file = nil
	}

	if e1 != nil {
		return e1
	} else if e2 != nil {
		return e2
	}
	return nil
}

func (a *RotateAppender) reset(file *os.File) {
	if bw, ok := a.w.(*bufio.Writer); ok {
		bw.Reset(file)
	} else {
		a.w = file
	}
}

func (a *RotateAppender) Output(_ Level, t time.Time, data []byte) {
	a.mu.Lock()
	if t.After(a.rt) {
		var suffix string
		a.rt, suffix = a.rtfn(a.rt)
		filename := a.filename + suffix
		err := a.close()
		if err != nil {
			println("appender close ", a.filename, "error: ", err.Error())
		}
		if err = os.Rename(a.filename, filename); err != nil {
			println("appender rename ", filename, "error: ", err.Error())
		}

		a.file, err = os.OpenFile(a.filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			println("appender open ", a.filename, "error: ", err.Error())
		}
		a.reset(a.file)
	}
	if a.file == nil {
		a.mu.Unlock()
		return
	}
	a.w.Write(data)
	a.mu.Unlock()
}
