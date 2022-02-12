package plog

import (
	"errors"
	"fmt"
	"io"
	"log"
)

var ErrorBadLevel = errors.New("unrecognized log level")

type LogLevel int

func (l *LogLevel) String() string {
	return LogLevelStrings[int(*l)]
}

func (l *LogLevel) Set(val string) error {
	for i, v := range LogLevelStrings {
		if v == val {
			*l = LogLevel(i)
			return nil
		}
	}

	return fmt.Errorf("%w. Options: %v", ErrorBadLevel, LogLevelStrings)
}

const (
	LogLevelDebug LogLevel = iota
	LogLevelInfo
	LogLevelWarn
	LogLevelError
)

var LogLevelStrings = []string{"debug", "info", "warn", "error"}

func RequireLevel(current LogLevel, minimum LogLevel, f func()) {
	// For example, if current level is Info(1), and the minimum level is Debug (0), then it should not be printed.
	// but if current level is Info (1), and the minimum is Warn (2), then it should be printed.
	if current > minimum {
		return
	}

	f()
}

type Logger struct {
	*log.Logger
	level LogLevel
}

func New(level LogLevel, w io.Writer) *Logger {
	return &Logger{
		level:  level,
		Logger: log.New(w, "", log.Default().Flags()),
	}
}

func (l *Logger) Debug(v ...interface{}) {
	v = append([]interface{}{"[debug]"}, v...)
	RequireLevel(l.level, LogLevelDebug, func() {
		l.Logger.Print(v...)
	})
}

func (l *Logger) Debugf(format string, v ...interface{}) {
	format = "[debug]" + format
	RequireLevel(l.level, LogLevelDebug, func() {
		l.Logger.Printf(format, v...)
	})
}

func (l *Logger) Debugln(v ...interface{}) {
	v = append([]interface{}{"[debug]"}, v...)
	RequireLevel(l.level, LogLevelDebug, func() {
		l.Logger.Println(v...)
	})
}

func (l *Logger) Info(v ...interface{}) {
	v = append([]interface{}{"[info]"}, v...)
	RequireLevel(l.level, LogLevelInfo, func() {
		l.Logger.Print(v...)
	})
}

func (l *Logger) Infof(format string, v ...interface{}) {
	format = "[info]" + format
	RequireLevel(l.level, LogLevelInfo, func() {
		l.Logger.Printf(format, v...)
	})
}

func (l *Logger) Infoln(v ...interface{}) {
	v = append([]interface{}{"[info]"}, v...)
	RequireLevel(l.level, LogLevelInfo, func() {
		l.Logger.Println(v...)
	})
}

func (l *Logger) Warn(v ...interface{}) {
	v = append([]interface{}{"[warn]"}, v...)
	RequireLevel(l.level, LogLevelWarn, func() {
		l.Logger.Print(v...)
	})
}

func (l *Logger) Warnf(format string, v ...interface{}) {
	format = "[warn]" + format
	RequireLevel(l.level, LogLevelWarn, func() {
		l.Logger.Printf(format, v...)
	})
}

func (l *Logger) Warnln(v ...interface{}) {
	v = append([]interface{}{"[warn]"}, v...)
	RequireLevel(l.level, LogLevelWarn, func() {
		l.Logger.Println(v...)
	})
}

func (l *Logger) Error(v ...interface{}) {
	v = append([]interface{}{"[error]"}, v...)
	RequireLevel(l.level, LogLevelError, func() {
		l.Logger.Print(v...)
	})
}

func (l *Logger) Errorf(format string, v ...interface{}) {
	format = "[error]" + format
	RequireLevel(l.level, LogLevelError, func() {
		l.Logger.Printf(format, v...)
	})
}

func (l *Logger) Errorln(v ...interface{}) {
	v = append([]interface{}{"[error]"}, v...)
	RequireLevel(l.level, LogLevelError, func() {
		l.Logger.Println(v...)
	})
}

func (l *Logger) Fatal(v ...interface{}) {
	v = append([]interface{}{"[fatal]"}, v...)
	RequireLevel(l.level, LogLevelError, func() {
		l.Logger.Fatal(v...)
	})
}

func (l *Logger) Fatalf(format string, v ...interface{}) {
	format = "[fatal]" + format
	RequireLevel(l.level, LogLevelError, func() {
		l.Logger.Fatalf(format, v...)
	})
}

func (l *Logger) Fatalln(v ...interface{}) {
	v = append([]interface{}{"[fatal]"}, v...)
	RequireLevel(l.level, LogLevelError, func() {
		l.Logger.Fatalln(v...)
	})
}

func (l *Logger) Panic(v ...interface{}) {
	v = append([]interface{}{"[panic]"}, v...)
	RequireLevel(l.level, LogLevelError, func() {
		l.Logger.Panic(v...)
	})
}

func (l *Logger) Panicf(format string, v ...interface{}) {
	format = "[panic]" + format
	RequireLevel(l.level, LogLevelError, func() {
		l.Logger.Panicf(format, v...)
	})
}

func (l *Logger) Panicln(v ...interface{}) {
	v = append([]interface{}{"[panicln]"}, v...)
	RequireLevel(l.level, LogLevelError, func() {
		l.Logger.Panicln(v...)
	})
}
