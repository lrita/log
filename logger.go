package log

import (
	"fmt"
	"os"
	"path"
	"runtime"
	"sync"
	"time"
)

var ExitOnFatal = true

type Logger interface {
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
	l         sync.RWMutex
	calldepth int
	level     Level
	name      string
	enabled   map[Level]bool
	appenders map[Level]Appender
	formats   map[Level]string
	children  []Logger
	parent    *logger
}

var log = &logger{
	calldepth: 1,
	level:     DEBUG,
	name:      "",
	enabled:   make(map[Level]bool),
	appenders: make(map[Level]Appender),
	formats:   make(map[Level]string),
}

func init() {
	log.SetLevel(DEBUG)
	log.SetFormat("%F %T [%l] %m")
	log.SetAppender(NewConsoleAppender())
}

func (l *logger) cloneEnabled() map[Level]bool {
	enabled := make(map[Level]bool)
	for level, v := range l.enabled {
		enabled[level] = v
	}
	return enabled
}

func (l *logger) cloneAppenders() map[Level]Appender {
	appenders := make(map[Level]Appender)
	for level, appender := range l.appenders {
		appenders[level] = appender
	}
	return appenders
}

func (l *logger) cloneFormats() map[Level]string {
	formats := make(map[Level]string)
	for level, format := range l.formats {
		formats[level] = format
	}
	return formats
}

func (l *logger) New(name string) Logger {
	l.l.Lock()
	lg := &logger{
		calldepth: 0,
		level:     l.level,
		name:      name,
		enabled:   l.cloneEnabled(),
		appenders: l.cloneAppenders(),
		formats:   l.cloneFormats(),
		parent:    l,
	}
	l.children = append(l.children, lg)
	l.l.Unlock()
	return lg
}

func (l *logger) Level() Level {
	l.l.RLock()
	lvl := l.level
	l.l.RUnlock()
	return lvl
}

func (l *logger) SetCallDepth(d int) {
	l.l.Lock()
	l.calldepth = d
	l.l.Unlock()
}

func (l *logger) SetLevel(level Level) {
	l.l.Lock()
	l.level = level
	for k := range LevelsToString {
		if k <= level {
			l.enabled[k] = true
		} else {
			l.enabled[k] = false
		}
	}
	l.l.Unlock()
}

func (l *logger) SetAppender(appender Appender, levels ...Level) {
	l.l.Lock()
	if len(levels) == 0 {
		for level := range LevelsToString {
			l.appenders[level] = appender
		}
	} else {
		for _, level := range levels {
			l.appenders[level] = appender
		}
	}
	l.l.Unlock()
}

func (l *logger) SetFormat(fmt string, levels ...Level) {
	l.l.Lock()
	if len(levels) == 0 {
		for level := range LevelsToString {
			l.formats[level] = fmt
		}
	} else {
		for _, level := range levels {
			l.formats[level] = fmt
		}
	}
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
	l.l.RLock()
	ok := l.enabled[level]
	calldepth := l.calldepth
	l.l.RUnlock()
	if !ok {
		return
	}

	app := l.appender(level)
	if app == nil {
		return
	}

	var (
		line   int
		caller string
		buf    = make([]byte, 0, 256)
		format = l.format(level)
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
				_, caller, line, ok = runtime.Caller(calldepth + 2)
				if !ok {
					caller = "???"
				}
			}
			buf = append(buf, caller...)
		case 'c':
			if caller == "" {
				_, caller, line, ok = runtime.Caller(calldepth + 2)
				if !ok {
					caller = "???"
				}
			}
			buf = append(buf, path.Base(caller)...)
		case 'L':
			if caller == "" {
				_, caller, line, ok = runtime.Caller(calldepth + 2)
				if !ok {
					caller = "???"
				}
			}
			itoa(&buf, line, -1)
		case '%':
			buf = append(buf, '%')
		case 'n':
			buf = append(buf, 'n')
		case 'F':
			buf = append(buf, tm.Format("2006-01-02")...)
		case 'D':
			buf = append(buf, tm.Format("01/02/06")...)
		case 'd':
			buf = append(buf, tm.Format(time.RFC3339)...)
		case 'T':
			buf = append(buf, tm.Format("15:04:05")...)
		case 'a':
			buf = append(buf, tm.Format("Mon")...)
		case 'A':
			buf = append(buf, tm.Format("Monday")...)
		case 'b':
			buf = append(buf, tm.Format("Jan")...)
		case 'B':
			buf = append(buf, tm.Format("January")...)
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
	l.l.RLock()
	app := l.appenders[level]
	l.l.RUnlock()
	if app == nil && l.parent != nil {
		app = l.parent.appender(level)
	}
	return app
}

func (l *logger) format(level Level) string {
	l.l.RLock()
	f := l.formats[level]
	l.l.RUnlock()
	if f == "" && l.parent != nil {
		f = l.parent.format(level)
	}
	return f
}
