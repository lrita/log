package log

type Level int8

const (
	FATAL Level = iota
	ERROR
	INFO
	WARN
	DEBUG
	TRACE
)

var StringToLevels = map[string]Level{
	"TRACE": TRACE,
	"DEBUG": DEBUG,
	"WARN":  WARN,
	"INFO":  INFO,
	"ERROR": ERROR,
	"FATAL": FATAL,
}

var LevelsToString = map[Level]string{
	TRACE: "TRACE",
	DEBUG: "DEBUG",
	WARN:  "WARN",
	INFO:  "INFO",
	ERROR: "ERROR",
	FATAL: "FATAL",
}
