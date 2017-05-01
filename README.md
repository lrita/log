## log
It is a logger for golang, which has met many of my needs. It cat set different formatter with
the same appender. Also it can set multi-appender for a logger instance.

All the new logger inherited the global logger. If the new logger is not set the formatter
or appender, it will inherited the global logger's formatter and appender.

## Level
This logger has 6 log-level.

```
FATAL
ERROR
INFO
WARN
DEBUG
TRACE
```

We can use the `SetLevel` or `Logger.SetLevel` to chang the log-level for each logger instance.

## Fomatter
We can set a format layout to the each level for the logger instance. The global logger default
format layout is `%F %T [%l] %m`. The pattern is:

```
    %m => the log message and its arguments formatted with `fmt.Sprintf` or `fmt.Sprint`
    %l => the log-level string
    %C => the caller with full file path
    %c => the caller with short file path
    %L => the line number of caller
    %% => '%'
    %n => '\n'
    %F => the date formatted like "2006-01-02"
    %D => the date formatted like "01/02/06"
    %T => the time formatted like 24h style "15:04:05"
    %a => the short name of weekday like "Mon"
    %A => the full name of weekday like "Monday"
    %b => the short name of month like "Jan"
    %B => the full name of month like "January"
    %d => the datetime formatted like RFC3339 "2006-01-02T15:04:05Z07:00"
```


## Install

```
go get -u github.com/lrita/log
```
