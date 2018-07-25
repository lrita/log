package log

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type dap struct {
	l Level
	d string
}

func (d *dap) Output(level Level, t time.Time, data []byte) {
	d.l = level
	d.d = string(data)
}

func TestGlobalLogger(t *testing.T) {
	d := &dap{}
	assert := assert.New(t)
	check0 := func(l Level) {
		assert.Equal(l, d.l, LevelsToString[l])
		tokens := strings.Split(d.d, " ")
		if assert.Equal(6, len(tokens)) {
			assert.Equal("["+LevelsToString[l]+"]", tokens[2])
			assert.Equal("1", tokens[3])
			assert.Equal("2", tokens[4])
			assert.Equal("3\n", tokens[5])
		}
	}
	check1 := func(l Level) {
		assert.Equal(l, d.l, LevelsToString[l])
		tokens := strings.Split(d.d, " ")
		if assert.Equal(7, len(tokens)) {
			assert.Equal("logger_test.go", filepath.Base(tokens[1]))
			assert.Equal("["+LevelsToString[l]+"]", tokens[3])
			assert.Equal("1", tokens[4])
			assert.Equal("2", tokens[5])
			assert.Equal("3\n", tokens[6])
		}
	}
	check2 := func(l Level) {
		assert.Equal(l, d.l, LevelsToString[l])
		tokens := strings.Split(d.d, " ")
		if assert.Equal(7, len(tokens)) {
			assert.Equal("logger_test.go", tokens[1])
			assert.Equal("["+LevelsToString[l]+"]", tokens[3])
			assert.Equal("1", tokens[4])
			assert.Equal("2", tokens[5])
			assert.Equal("3\n", tokens[6])
		}
	}
	check3 := func(l Level) {
		assert.Equal(l, d.l, LevelsToString[l])
		tokens := strings.Split(d.d, " ")
		if assert.Equal(8, len(tokens)) {
			assert.Equal("logger_test.go", tokens[1])
			assert.Equal("["+LevelsToString[l]+"]", tokens[3])
			assert.Equal("[n]", tokens[4])
			assert.Equal("1", tokens[5])
			assert.Equal("2", tokens[6])
			assert.Equal("3\n", tokens[7])
		}
	}

	SetFormat("%F %T [%l] %m")
	SetAppender(d)
	defer SetAppender(NewConsoleAppender())

	SetLevel(TRACE)
	ExitOnFatal = false

	Trace(1, " 2", " 3")
	check0(TRACE)
	Debug(1, " 2", " 3")
	check0(DEBUG)
	Warn(1, " 2", " 3")
	check0(WARN)
	Info(1, " 2", " 3")
	check0(INFO)
	Error(1, " 2", " 3")
	check0(ERROR)
	Fatal(1, " 2", " 3")
	check0(FATAL)

	SetFormat("%F %C %L [%l] %m")

	Trace(1, " 2", " 3")
	check1(TRACE)
	Debug(1, " 2", " 3")
	check1(DEBUG)
	Warn(1, " 2", " 3")
	check1(WARN)
	Info(1, " 2", " 3")
	check1(INFO)
	Error(1, " 2", " 3")
	check1(ERROR)
	Fatal(1, " 2", " 3")
	check1(FATAL)

	SetFormat("%F %c %L [%l] %m")

	Trace(1, " 2", " 3")
	check2(TRACE)
	Debug(1, " 2", " 3")
	check2(DEBUG)
	Warn(1, " 2", " 3")
	check2(WARN)
	Info(1, " 2", " 3")
	check2(INFO)
	Error(1, " 2", " 3")
	check2(ERROR)
	Fatal(1, " 2", " 3")
	check2(FATAL)

	lg := New("n")
	lg.SetFormat("%F %c %L [%l] [n] %m")

	lg.Trace(1, " 2", " 3")
	check3(TRACE)
	lg.Debug(1, " 2", " 3")
	check3(DEBUG)
	lg.Warn(1, " 2", " 3")
	check3(WARN)
	lg.Info(1, " 2", " 3")
	check3(INFO)
	lg.Error(1, " 2", " 3")
	check3(ERROR)
	lg.Fatal(1, " 2", " 3")
	check3(FATAL)
}

type la struct {
	m map[Level]int
}

func (a *la) Output(level Level, t time.Time, data []byte) {
	a.m[level]++
}

func (a *la) check(t *testing.T, level Level, vv int) {
	for l, v := range a.m {
		if l <= level {
			if v != vv {
				t.Errorf("%v: expect %d, got %d", LevelsToString[l], vv, v)
			}
		} else {
			if v >= vv {
				t.Errorf("%v: expect lt %d, got %d", LevelsToString[l], vv, v)
			}
		}
	}
}

func TestLoggerSetLevel(t *testing.T) {
	a := &la{m: make(map[Level]int, len(StringToLevels))}
	for l := range LevelsToString {
		a.m[l] = 0
	}
	tt := []struct {
		l Level
		n int
	}{
		{l: TRACE, n: 1},
		{l: DEBUG, n: 2},
		{l: INFO, n: 3},
		{l: WARN, n: 4},
		{l: ERROR, n: 5},
		{l: FATAL, n: 6},
	}

	log.SetAppender(a)
	ExitOnFatal = false

	for _, v := range tt {
		SetLevel(v.l)
		Trace(1, " 2", "3")
		Debug(1, " 2", "3")
		Warn(1, " 2", "3")
		Info(1, " 2", "3")
		Error(1, " 2", "3")
		Fatal(1, " 2", "3")
		a.check(t, v.l, v.n)
	}
}

type ha struct {
	count int
	data  map[Level][]byte
}

func (c *ha) Output(level Level, t time.Time, data []byte) {
	c.count++
	if d, ok := c.data[level]; ok {
		if !bytes.Equal(d, data) {
			panic("format is not equal")
		}
	} else {
		dd := make([]byte, len(data))
		copy(dd, data)
		c.data[level] = dd
	}
}

func TestLoggerInherit(t *testing.T) {
	var (
		ha0 = &ha{data: make(map[Level][]byte)}
		ha1 = &ha{data: make(map[Level][]byte)}
	)

	defer SetAppender(NewConsoleAppender())
	ExitOnFatal = false
	SetAppender(ha0)
	SetFormat("%F %a %l %m")
	SetLevel(TRACE)
	log0 := New("log0")
	log1 := New("log1")
	log2 := log0.New("log2")
	log0.SetAppender(ha1, DEBUG, ERROR)
	log0.SetFormat("%a %l %m", DEBUG, ERROR)

	for _, l := range []Logger{log, log0, log1, log2} {
		l.Trace("trace message")
		l.Debug("debug message")
		l.Info("info message")
		l.Warn("warn message")
		l.Error("error message")
		l.Fatal("fatal message")
	}
}

type null struct{}

func (n *null) Output(level Level, t time.Time, data []byte) {
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

var (
	bench0, bench1, bench2, bench3, bench4 Logger
)

func init() {
	bench0 = New("bench1")
	bench1 = bench0.New("bench1")
	bench2 = bench1.New("bench2")
	bench3 = bench1.New("bench3")
	bench4 = bench1.New("bench4")
	bench0.SetAppender(&null{})
	bench0.SetLevel(TRACE)
}

func benmarkLoggerWithMultiInherit(b *testing.B, p int) {
	b.SetParallelism(p)
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			bench0.Info("benmarkLoggerWithMultiInherit")
			bench1.Info("benmarkLoggerWithMultiInherit")
			bench2.Info("benmarkLoggerWithMultiInherit")
			bench3.Info("benmarkLoggerWithMultiInherit")
			bench4.Info("benmarkLoggerWithMultiInherit")
		}
	})
}

func BenchmarkLoggerWithMultiInherit1(b *testing.B) {
	benmarkLoggerWithMultiInherit(b, 1)
}

func BenchmarkLoggerWithMultiInherit10(b *testing.B) {
	benmarkLoggerWithMultiInherit(b, 10)
}

func BenchmarkLoggerWithMultiInherit20(b *testing.B) {
	benmarkLoggerWithMultiInherit(b, 20)
}
