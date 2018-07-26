package log

import (
	"os"
	"testing"
	"time"
)

func TestHourlyRotateAppender(t *testing.T) {
	const filename = "a.log"
	app, err := NewHourlyRotateAppender(filename)
	if err != nil {
		t.Fatalf("new hourly rotate appender error %v", err)
	}

	log := New("t")

	defer func() {
		app.Close()
		os.Remove(filename)
	}()

	log.SetAppender(app)
	log.Errorf("test string : %v", "only for test")
}

func TestHourlyRotateBufAppender(t *testing.T) {
	const filename = "a.log"
	app, err := NewHourlyRotateBufAppender(filename, 4096)
	if err != nil {
		t.Fatalf("new hourly rotate appender error %v", err)
	}

	defer func() {
		app.Close()
		os.Remove(filename)
	}()

	app.Output(DEBUG, time.Now(), []byte("1111\n"))
	if err := app.Flush(); err != nil {
		t.Errorf("appender flush error: %v", err)
	}
	if err := app.close(); err != nil {
		t.Fatalf("appender close error: %v", err)
	}

	app.file, err = os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		t.Fatalf("appender open file %q error: %v", filename, err)
	}
	app.reset(app.file)
	app.Output(DEBUG, time.Now(), []byte("2222\n"))
}

func BenchmarkRotateAppenderBuf0(b *testing.B) {
	const filename = "a.log"
	app, err := NewHourlyRotateAppender(filename)
	if err != nil {
		b.Fatalf("new hourly rotate appender error %v", err)
	}

	defer func() {
		app.Close()
		os.Remove(filename)
	}()

	tt := time.Now()
	data := []byte("appender benchmark test data content information...")

	b.ReportAllocs()
	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			app.Output(DEBUG, tt, data)
		}
	})
}

func BenchmarkRotateAppenderBuf4k(b *testing.B) {
	const filename = "a.log"
	app, err := NewHourlyRotateBufAppender(filename, 4096)
	if err != nil {
		b.Fatalf("new hourly rotate appender error %v", err)
	}

	defer func() {
		app.Close()
		os.Remove(filename)
	}()

	tt := time.Now()
	data := []byte("appender benchmark test data content information...")

	b.ReportAllocs()
	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			app.Output(DEBUG, tt, data)
		}
	})
}

func BenchmarkRotateAppenderBuf8k(b *testing.B) {
	const filename = "a.log"
	app, err := NewHourlyRotateBufAppender(filename, 8192)
	if err != nil {
		b.Fatalf("new hourly rotate appender error %v", err)
	}

	defer func() {
		app.Close()
		os.Remove(filename)
	}()

	tt := time.Now()
	data := []byte("appender benchmark test data content information...")

	b.ReportAllocs()
	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			app.Output(DEBUG, tt, data)
		}
	})
}

func BenchmarkRotateAppenderBuf16k(b *testing.B) {
	const filename = "a.log"
	app, err := NewHourlyRotateBufAppender(filename, 1024*16)
	if err != nil {
		b.Fatalf("new hourly rotate appender error %v", err)
	}

	defer func() {
		app.Close()
		os.Remove(filename)
	}()

	tt := time.Now()
	data := []byte("appender benchmark test data content information...")

	b.ReportAllocs()
	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			app.Output(DEBUG, tt, data)
		}
	})
}
