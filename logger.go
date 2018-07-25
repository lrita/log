package log

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"sync/atomic"
	"time"
	"unsafe"

	"github.com/lrita/ratelimit"
)

// ExitOnFatal decides whether or not to exit when fatal log printing.
var ExitOnFatal = true

type Logger interface {
	// New return a new log handler which inherit its appender and formater
	New(name string) Logger
	// Level return the logger current log-level
	Level() Level
	// SetLevel set the logger current log-level
	SetLevel(level Level)
	// SetAppender the given log-level to use the special appender.
	// If non-given log-level, all log-level use it
	SetAppender(appender Appender, levels ...Level)
	// SetRatelimit the give limit(QPS) rate to the logger.
	SetRatelimit(limit int64, levels ...Level)
	// SetFormat the given log-level to use the special format.
	// If non-given log-level, all log-level use it
	// fmt is a pattern-string, default is "%F %T [%l] %m"
	// %m => the log message and its arguments formatted with `fmt.Sprintf` or `fmt.Sprint`
	// %l => the log-level string
	// %C => the caller with full file path
	// %c => the caller with short file path
	// %L => the line number of caller
	// %% => '%'
	// %n => '\n'
	// %F => the date formatted like "2006-01-02"
	// %D => the date formatted like "01/02/06"
	// %T => the time formatted like 24h style "15:04:05"
	// %a => the short name of weekday like "Mon"
	// %A => the full name of weekday like "Monday"
	// %b => the short name of month like "Jan"
	// %B => the full name of month like "January"
	// %d => the datetime formatted like RFC3339 "2006-01-02T15:04:05Z07:00"
	SetFormat(fmt string, levels ...Level)
	// SetCallDepth set callee stack depth
	SetCallDepth(d int)
	// IsDebugEnabled indicates whether debug level is enabled
	IsDebugEnabled() bool

	Fatal(v ...interface{})
	Error(v ...interface{})
	Info(v ...interface{})
	Warn(v ...interface{})
	Debug(v ...interface{})
	Trace(v ...interface{})

	Fatalf(fmt string, v ...interface{})
	Errorf(fmt string, v ...interface{})
	Infof(fmt string, v ...interface{})
	Warnf(fmt string, v ...interface{})
	Debugf(fmt string, v ...interface{})
	Tracef(fmt string, v ...interface{})
}

type logger struct {
	l        sync.Mutex
	name     string
	meta     unsafe.Pointer
	children []*logger
}

const (
	detachlvl = 1 << iota
	detachapp
	detachfmt
	detachlmt
)

type meta struct {
	detach    uint8
	level     Level
	calldepth int
	appenders map[Level]Appender
	formats   map[Level]string
	limits    map[Level]*ratelimit.Bucket
}

func (m *meta) clone() *meta {
	mm := &meta{
		detach:    m.detach,
		level:     m.level,
		calldepth: m.calldepth,
		appenders: make(map[Level]Appender),
		formats:   make(map[Level]string),
		limits:    make(map[Level]*ratelimit.Bucket),
	}
	for level, app := range m.appenders {
		mm.appenders[level] = app
	}
	for level, fmt := range m.formats {
		mm.formats[level] = fmt
	}
	for level, l := range m.limits {
		mm.limits[level] = l
	}
	return mm
}

var (
	log = &logger{
		name: "",
		meta: unsafe.Pointer(&meta{
			level:     DEBUG,
			calldepth: 1,
			appenders: make(map[Level]Appender),
			formats:   make(map[Level]string),
		}),
	}
	pool = sync.Pool{
		New: func() interface{} {
			b := make([]byte, 256)
			return &b
		},
	}
)

func init() {
	log.SetLevel(DEBUG)
	log.SetFormat("%F %T [%l] %m")
	log.SetAppender(NewConsoleAppender())
}

func (l *logger) New(name string) Logger {
	l.l.Lock()
	m := (*meta)(atomic.LoadPointer(&l.meta)).clone()
	m.detach = 0
	m.calldepth = 0
	child := &logger{
		name: name,
		meta: unsafe.Pointer(m),
	}
	l.children = append(l.children, child)
	l.l.Unlock()
	return child
}

func (l *logger) Level() Level {
	return (*meta)(atomic.LoadPointer(&l.meta)).level
}

func (l *logger) SetCallDepth(d int) {
	l.l.Lock()
	m := *(*meta)(atomic.LoadPointer(&l.meta))
	m.calldepth = d
	atomic.StorePointer(&l.meta, unsafe.Pointer(&m))
	l.l.Unlock()
}

func (l *logger) IsDebugEnabled() bool {
	return l.Level() >= DEBUG
}

func (l *logger) setLevelInternal(detach bool, level Level) {
	l.l.Lock()
	defer l.l.Unlock()
	m := *(*meta)(atomic.LoadPointer(&l.meta))
	if detach {
		m.detach |= detachlvl
	} else if m.detach&detachlvl != 0 {
		return
	}
	m.level = level
	atomic.StorePointer(&l.meta, unsafe.Pointer(&m))
	for _, child := range l.children {
		child.setLevelInternal(false, level)
	}
}

func (l *logger) SetLevel(level Level) {
	l.setLevelInternal(true, level)
}

func (l *logger) setAppenderInternal(detach bool, appender Appender, levels ...Level) {
	l.l.Lock()
	defer l.l.Unlock()
	m := *(*meta)(atomic.LoadPointer(&l.meta))
	if detach {
		m.detach |= detachapp
	} else if m.detach&detachapp != 0 {
		return
	}
	m.appenders = make(map[Level]Appender, len(LevelsToString))
	if len(levels) == 0 {
		for level := range LevelsToString {
			m.appenders[level] = appender
		}
	} else {
		m0 := (*meta)(atomic.LoadPointer(&l.meta))
		for l, a := range m0.appenders {
			m.appenders[l] = a
		}
		for _, level := range levels {
			m.appenders[level] = appender
		}
	}
	atomic.StorePointer(&l.meta, unsafe.Pointer(&m))
	for _, child := range l.children {
		child.setAppenderInternal(false, appender, levels...)
	}
}

func (l *logger) SetAppender(appender Appender, levels ...Level) {
	l.setAppenderInternal(true, appender, levels...)
}

func (l *logger) setFormatInternal(detach bool, fmt string, levels ...Level) {
	l.l.Lock()
	defer l.l.Unlock()
	m := *(*meta)(atomic.LoadPointer(&l.meta))
	if detach {
		m.detach |= detachfmt
	} else if m.detach&detachfmt != 0 {
		return
	}
	m.formats = make(map[Level]string, len(LevelsToString))
	if len(levels) == 0 {
		for level := range LevelsToString {
			m.formats[level] = fmt
		}
	} else {
		m0 := (*meta)(atomic.LoadPointer(&l.meta))
		for l, f := range m0.formats {
			m.formats[l] = f
		}
		for _, level := range levels {
			m.formats[level] = fmt
		}
	}
	atomic.StorePointer(&l.meta, unsafe.Pointer(&m))
	for _, child := range l.children {
		child.setFormatInternal(false, fmt, levels...)
	}
}

func (l *logger) SetFormat(fmt string, levels ...Level) {
	l.setFormatInternal(true, fmt, levels...)
}

func (l *logger) setRatelimitInternal(detach bool, bucket *ratelimit.Bucket, levels ...Level) {
	l.l.Lock()
	defer l.l.Unlock()
	m := *(*meta)(atomic.LoadPointer(&l.meta))
	if detach {
		m.detach |= detachlmt
	} else if m.detach&detachlmt != 0 {
		return
	}
	m.limits = make(map[Level]*ratelimit.Bucket, len(LevelsToString))
	if len(levels) == 0 {
		for level := range LevelsToString {
			m.limits[level] = bucket
		}
	} else {
		m0 := (*meta)(atomic.LoadPointer(&l.meta))
		for l, b := range m0.limits {
			m.limits[l] = b
		}
		for _, level := range levels {
			m.limits[level] = bucket
		}
	}
	atomic.StorePointer(&l.meta, unsafe.Pointer(&m))
	for _, child := range l.children {
		child.setRatelimitInternal(false, bucket, levels...)
	}
}

func (l *logger) SetRatelimit(limit int64, levels ...Level) {
	bucket := ratelimit.NewBucketWithRate(float64(limit), 1)
	l.setRatelimitInternal(true, bucket, levels...)
}

// Cheap integer to fixed-width decimal ASCII.  Give a negative width to avoid zero-padding.
func itoa(buf *[]byte, i int, wid int) {
	// Assemble decimal in reverse order.
	var b [20]byte
	bp := len(b) - 1
	for i >= 10 || wid > 1 {
		wid--
		q := i / 10
		b[bp] = byte('0' + i - q*10)
		bp--
		i = q
	}
	// i < 10
	b[bp] = byte('0' + i)
	*buf = append(*buf, b[bp:]...)
}

func (l *logger) Fatal(v ...interface{}) {
	l.Log(FATAL, "", v...)
}

func (l *logger) Error(v ...interface{}) {
	l.Log(ERROR, "", v...)
}

func (l *logger) Info(v ...interface{}) {
	l.Log(INFO, "", v...)
}

func (l *logger) Warn(v ...interface{}) {
	l.Log(WARN, "", v...)
}

func (l *logger) Debug(v ...interface{}) {
	l.Log(DEBUG, "", v...)
}

func (l *logger) Trace(v ...interface{}) {
	l.Log(TRACE, "", v...)
}

func (l *logger) Fatalf(fmt string, v ...interface{}) {
	l.Log(FATAL, fmt, v...)
}

func (l *logger) Errorf(fmt string, v ...interface{}) {
	l.Log(ERROR, fmt, v...)
}

func (l *logger) Infof(fmt string, v ...interface{}) {
	l.Log(INFO, fmt, v...)
}

func (l *logger) Warnf(fmt string, v ...interface{}) {
	l.Log(WARN, fmt, v...)
}

func (l *logger) Debugf(fmt string, v ...interface{}) {
	l.Log(DEBUG, fmt, v...)
}

func (l *logger) Tracef(fmt string, v ...interface{}) {
	l.Log(TRACE, fmt, v...)
}

func (l *logger) Log(level Level, f string, v ...interface{}) {
	m := (*meta)(atomic.LoadPointer(&l.meta))
	if level > m.level {
		return
	}

	app := m.appenders[level]
	if app == nil {
		return
	}

	if limit := m.limits[level]; limit != nil && limit.TakeAvailable(1) == 0 {
		return
	}

	var (
		ok     bool
		line   int
		caller string
		b      = pool.Get().(*[]byte)
		format = m.formats[level]
		tm     = time.Now()
		n      = len(format)
	)

	*b = (*b)[:0]

	for i := 0; i < n; i++ {
		lasti := i
		for i < n && format[i] != '%' {
			i++
		}
		if i > lasti {
			*b = append(*b, format[lasti:i]...)
		}
		if i >= n { // done processing format string
			break
		}

		i++ // skip '%'

		switch format[i] {
		case 'm':
			if f != "" {
				*b = append(*b, fmt.Sprintf(f, v...)...)
			} else {
				*b = append(*b, fmt.Sprint(v...)...)
			}
		case 'l':
			*b = append(*b, LevelsToString[level]...)
		case 'C':
			if caller == "" {
				_, caller, line, ok = runtime.Caller(m.calldepth + 2)
				if !ok {
					caller = "???"
				}
			}
			*b = append(*b, caller...)
		case 'c':
			if caller == "" {
				_, caller, line, ok = runtime.Caller(m.calldepth + 2)
				if !ok {
					caller = "???"
				}
			}
			*b = append(*b, filepath.Base(caller)...)
		case 'L':
			if caller == "" {
				_, caller, line, ok = runtime.Caller(m.calldepth + 2)
				if !ok {
					caller = "???"
				}
			}
			itoa(b, line, -1)
		case '%':
			*b = append(*b, '%')
		case 'n':
			*b = append(*b, '\n')
		case 'F':
			*b = tm.AppendFormat(*b, "2006-01-02")
		case 'D':
			*b = tm.AppendFormat(*b, "01/02/06")
		case 'd':
			*b = tm.AppendFormat(*b, time.RFC3339)
		case 'T':
			*b = tm.AppendFormat(*b, "15:04:05")
		case 'a':
			*b = tm.AppendFormat(*b, "Mon")
		case 'A':
			*b = tm.AppendFormat(*b, "Monday")
		case 'b':
			*b = tm.AppendFormat(*b, "Jan")
		case 'B':
			*b = tm.AppendFormat(*b, "January")
		}
	}

	if len(*b) == 0 || (*b)[len(*b)-1] != '\n' {
		*b = append(*b, '\n')
	}

	app.Output(level, tm, *b)
	pool.Put(b)

	if level == FATAL && ExitOnFatal {
		if flusher, ok := app.(Flusher); ok {
			flusher.Flush()
		}
		os.Exit(-1)
	}
}
