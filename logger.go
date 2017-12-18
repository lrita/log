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
)

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
	children []Logger
}

type meta struct {
	detach    bool
	level     Level
	calldepth int
	appenders map[Level]Appender
	formats   map[Level]string
	parent    *logger
}

func (m *meta) appender(level Level) Appender {
	app := m.appenders[level]
	if app == nil && m.parent != nil {
		app = m.parent.appender(level)
	}
	return app
}

func (m *meta) format(level Level) string {
	f := m.formats[level]
	if f == "" && m.parent != nil {
		f = m.parent.format(level)
	}
	return f
}

func (m *meta) Level() Level {
	if m.detach {
		return m.level
	}
	return m.parent.Level()
}

var log = &logger{
	name: "",
	meta: unsafe.Pointer(&meta{
		detach:    true,
		level:     DEBUG,
		calldepth: 1,
		appenders: make(map[Level]Appender),
		formats:   make(map[Level]string),
	}),
}

func init() {
	log.SetLevel(DEBUG)
	log.SetFormat("%F %T [%l] %m")
	log.SetAppender(NewConsoleAppender())
}

func (l *logger) New(name string) Logger {
	l.l.Lock()
	lg := &logger{
		name: name,
		meta: unsafe.Pointer(&meta{
			calldepth: 0,
			appenders: make(map[Level]Appender),
			formats:   make(map[Level]string),
			parent:    l,
		}),
	}
	l.children = append(l.children, lg)
	l.l.Unlock()
	return lg
}

func (l *logger) Level() Level {
	m := (*meta)(atomic.LoadPointer(&l.meta))
	if m.detach {
		return m.level
	}
	return m.parent.Level()
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

func (l *logger) SetLevel(level Level) {
	l.l.Lock()
	m := *(*meta)(atomic.LoadPointer(&l.meta))
	m.detach = true
	m.level = level
	atomic.StorePointer(&l.meta, unsafe.Pointer(&m))
	l.l.Unlock()
}

func (l *logger) SetAppender(appender Appender, levels ...Level) {
	l.l.Lock()
	m := *(*meta)(atomic.LoadPointer(&l.meta))
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
	l.l.Unlock()
}

func (l *logger) SetFormat(fmt string, levels ...Level) {
	l.l.Lock()
	m := *(*meta)(atomic.LoadPointer(&l.meta))
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
	l.l.Unlock()
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
	if level > m.Level() {
		return
	}

	app := m.appender(level)
	if app == nil {
		return
	}

	var (
		ok     bool
		line   int
		caller string
		buf    = make([]byte, 0, 256)
		format = m.format(level)
		tm     = time.Now()
		n      = len(format)
	)

	for i := 0; i < n; i++ {
		lasti := i
		for i < n && format[i] != '%' {
			i++
		}
		if i > lasti {
			buf = append(buf, format[lasti:i]...)
		}
		if i >= n { // done processing format string
			break
		}

		i++ // skip '%'

		switch format[i] {
		case 'm':
			if f != "" {
				buf = append(buf, fmt.Sprintf(f, v...)...)
			} else {
				buf = append(buf, fmt.Sprint(v...)...)
			}
		case 'l':
			buf = append(buf, LevelsToString[level]...)
		case 'C':
			if caller == "" {
				_, caller, line, ok = runtime.Caller(m.calldepth + 2)
				if !ok {
					caller = "???"
				}
			}
			buf = append(buf, caller...)
		case 'c':
			if caller == "" {
				_, caller, line, ok = runtime.Caller(m.calldepth + 2)
				if !ok {
					caller = "???"
				}
			}
			buf = append(buf, filepath.Base(caller)...)
		case 'L':
			if caller == "" {
				_, caller, line, ok = runtime.Caller(m.calldepth + 2)
				if !ok {
					caller = "???"
				}
			}
			itoa(&buf, line, -1)
		case '%':
			buf = append(buf, '%')
		case 'n':
			buf = append(buf, '\n')
		case 'F':
			buf = tm.AppendFormat(buf, "2006-01-02")
		case 'D':
			buf = tm.AppendFormat(buf, "01/02/06")
		case 'd':
			buf = tm.AppendFormat(buf, time.RFC3339)
		case 'T':
			buf = tm.AppendFormat(buf, "15:04:05")
		case 'a':
			buf = tm.AppendFormat(buf, "Mon")
		case 'A':
			buf = tm.AppendFormat(buf, "Monday")
		case 'b':
			buf = tm.AppendFormat(buf, "Jan")
		case 'B':
			buf = tm.AppendFormat(buf, "January")
		}
	}

	if len(buf) == 0 || buf[len(buf)-1] != '\n' {
		buf = append(buf, '\n')
	}

	app.Output(level, tm, buf)

	if level == FATAL && ExitOnFatal {
		os.Exit(-1)
	}
}

func (l *logger) appender(level Level) Appender {
	return (*meta)(atomic.LoadPointer(&l.meta)).appender(level)
}

func (l *logger) format(level Level) string {
	return (*meta)(atomic.LoadPointer(&l.meta)).format(level)
}
