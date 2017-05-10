package log

type Level int8

const (
	FATAL Level = iota
	ERROR
	WARN
	INFO
	DEBUG
	TRACE
)

var StringToLevels = map[string]Level{
	"TRACE": TRACE,
	"DEBUG": DEBUG,
	"INFO":  INFO,
	"WARN":  WARN,
	"ERROR": ERROR,
	"FATAL": FATAL,
}

var LevelsToString = map[Level]string{
	TRACE: "TRACE",
	DEBUG: "DEBUG",
	INFO:  "INFO",
	WARN:  "WARN",
	ERROR: "ERROR",
	FATAL: "FATAL",
}
