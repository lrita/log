package log

// New return a sub logger of global logger
func New(name string) Logger {
	return log.New(name)
}

// SetLevel log level for global logger
func SetLevel(level Level) {
	log.SetLevel(level)
}

// SetAppender set append for global logger
func SetAppender(appender Appender, levels ...Level) {
	log.SetAppender(appender, levels...)
}

// SetFormat set format-string for global logger
func SetFormat(fmt string, levels ...Level) {
	log.SetFormat(fmt, levels...)
}

// SetCallDepth set callee stack depth
func SetCallDepth(d int) {
	log.SetCallDepth(d + 1)
}

// IsDebugEnabled indicates whether debug level is enabled
func IsDebugEnabled() bool {
	return log.IsDebugEnabled()
}

func Fatal(v ...interface{}) {
	log.Fatal(v...)
}

func Error(v ...interface{}) {
	log.Error(v...)
}

func Info(v ...interface{}) {
	log.Info(v...)
}

func Warn(v ...interface{}) {
	log.Warn(v...)
}

func Debug(v ...interface{}) {
	log.Debug(v...)
}

func Trace(v ...interface{}) {
	log.Trace(v...)
}

func Fatalf(fmt string, v ...interface{}) {
	log.Fatalf(fmt, v...)
}

func Errorf(fmt string, v ...interface{}) {
	log.Errorf(fmt, v...)
}

func Infof(fmt string, v ...interface{}) {
	log.Infof(fmt, v...)
}

func Warnf(fmt string, v ...interface{}) {
	log.Warnf(fmt, v...)
}

func Debugf(fmt string, v ...interface{}) {
	log.Debugf(fmt, v...)
}

func Tracef(fmt string, v ...interface{}) {
	log.Tracef(fmt, v...)
}
