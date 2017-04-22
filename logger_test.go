package log

import (
	"io/ioutil"
	"testing"
	"time"
)

func TestGlobalLogger(t *testing.T) {
	SetLevel(TRACE)
	ExitOnFatal = false

	Trace(1, " 2", "3")
	Debug(1, " 2", "3")
	Warn(1, " 2", "3")
	Info(1, " 2", "3")
	Error(1, " 2", "3")
	Fatal(1, " 2", "3")

	SetFormat("%F %C %L {%l} %m")

	Trace(1, " 2", "3")
	Debug(1, " 2", "3")
	Warn(1, " 2", "3")
	Info(1, " 2", "3")
	Error(1, " 2", "3")
	Fatal(1, " 2", "3")

	SetFormat("%F %c %L {%l} %m")

	Trace(1, " 2", "3")
	Debug(1, " 2", "3")
	Warn(1, " 2", "3")
	Info(1, " 2", "3")
	Error(1, " 2", "3")
	Fatal(1, " 2", "3")

	lg := New("n")
	lg.SetFormat("%F %c %L {%l} [new logger n] %m")

	lg.Trace(1, " 2", "3")
	lg.Debug(1, " 2", "3")
	lg.Warn(1, " 2", "3")
	lg.Info(1, " 2", "3")
	lg.Error(1, " 2", "3")
	lg.Fatal(1, " 2", "3")
}

func TestLoggerSetLevel(t *testing.T) {
	ExitOnFatal = false
	println("=== Set TRACE ===")
	SetLevel(TRACE)
	Trace(1, " 2", "3")
	Debug(1, " 2", "3")
	Warn(1, " 2", "3")
	Info(1, " 2", "3")
	Error(1, " 2", "3")
	Fatal(1, " 2", "3")

	println("=== Set DEBUG ===")
	SetLevel(DEBUG)
	Trace(1, " 2", "3")
	Debug(1, " 2", "3")
	Warn(1, " 2", "3")
	Info(1, " 2", "3")
	Error(1, " 2", "3")
	Fatal(1, " 2", "3")

	println("=== Set WARN ===")
	SetLevel(WARN)
	Trace(1, " 2", "3")
	Debug(1, " 2", "3")
	Warn(1, " 2", "3")
	Info(1, " 2", "3")
	Error(1, " 2", "3")
	Fatal(1, " 2", "3")

	println("=== Set INFO ===")
	SetLevel(INFO)
	Trace(1, " 2", "3")
	Debug(1, " 2", "3")
	Warn(1, " 2", "3")
	Info(1, " 2", "3")
	Error(1, " 2", "3")
	Fatal(1, " 2", "3")

	println("=== Set ERROR ===")
	SetLevel(ERROR)
	Trace(1, " 2", "3")
	Debug(1, " 2", "3")
	Warn(1, " 2", "3")
	Info(1, " 2", "3")
	Error(1, " 2", "3")
	Fatal(1, " 2", "3")

	println("=== Set FATAL ===")
	SetLevel(FATAL)
	Trace(1, " 2", "3")
	Debug(1, " 2", "3")
	Warn(1, " 2", "3")
	Info(1, " 2", "3")
	Error(1, " 2", "3")
	Fatal(1, " 2", "3")
}

type null struct{}

func (n *null) Output(level Level, t time.Time, data []byte) {
	ioutil.Discard.Write(data)
}

func BenchmarkLogger(b *testing.B) {
	SetAppender(&null{})
	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			Infof("BenchmarkLogger running %s %d", "go go go", 12345678)
		}
	})
}
