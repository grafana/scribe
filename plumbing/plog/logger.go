package plog

import (
	"errors"
	"io"
	"log"
)

type LogLevel int

func (l *LogLevel) String() string {
	return LogLevelStrings[*l]
}

func (l *LogLevel) Set(val string) error {
	for i, v := range LogLevelStrings {
		if v == val {
			*l = LogLevel(i)
		}
	}

	return errors.New("log level not found")
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
	RequireLevel(l.level, LogLevelDebug, func() {
		l.Logger.Print(v...)
	})
}

func (l *Logger) Debugf(format string, v ...interface{}) {
	RequireLevel(l.level, LogLevelDebug, func() {
		l.Logger.Printf(format, v...)
	})
}

func (l *Logger) Debugln(v ...interface{}) {
	RequireLevel(l.level, LogLevelDebug, func() {
		l.Logger.Println(v...)
	})
}

func (l *Logger) Info(v ...interface{}) {
	RequireLevel(l.level, LogLevelInfo, func() {
		l.Logger.Print(v...)
	})
}

func (l *Logger) Infof(format string, v ...interface{}) {
	RequireLevel(l.level, LogLevelInfo, func() {
		l.Logger.Printf(format, v...)
	})
}

func (l *Logger) Infoln(v ...interface{}) {
	RequireLevel(l.level, LogLevelInfo, func() {
		l.Logger.Println(v...)
	})
}

func (l *Logger) Warn(v ...interface{}) {
	RequireLevel(l.level, LogLevelWarn, func() {
		l.Logger.Print(v...)
	})
}

func (l *Logger) Warnf(format string, v ...interface{}) {
	RequireLevel(l.level, LogLevelWarn, func() {
		l.Logger.Printf(format, v...)
	})
}

func (l *Logger) Warnln(v ...interface{}) {
	RequireLevel(l.level, LogLevelWarn, func() {
		l.Logger.Println(v...)
	})
}

func (l *Logger) Error(v ...interface{}) {
	RequireLevel(l.level, LogLevelError, func() {
		l.Logger.Print(v...)
	})
}

func (l *Logger) Errorf(format string, v ...interface{}) {
	RequireLevel(l.level, LogLevelError, func() {
		l.Logger.Printf(format, v...)
	})
}

func (l *Logger) Errorln(v ...interface{}) {
	RequireLevel(l.level, LogLevelError, func() {
		l.Logger.Println(v...)
	})
}

func (l *Logger) Fatal(v ...interface{}) {
	RequireLevel(l.level, LogLevelError, func() {
		l.Logger.Fatal(v...)
	})
}

func (l *Logger) Fatalf(format string, v ...interface{}) {
	RequireLevel(l.level, LogLevelError, func() {
		l.Logger.Fatalf(format, v...)
	})
}

func (l *Logger) Fatalln(v ...interface{}) {
	RequireLevel(l.level, LogLevelError, func() {
		l.Logger.Fatalln(v...)
	})
}

func (l *Logger) Panic(v ...interface{}) {
	RequireLevel(l.level, LogLevelError, func() {
		l.Logger.Panic(v...)
	})
}

func (l *Logger) Panicf(format string, v ...interface{}) {
	RequireLevel(l.level, LogLevelError, func() {
		l.Logger.Panicf(format, v...)
	})
}

func (l *Logger) Panicln(v ...interface{}) {
	RequireLevel(l.level, LogLevelError, func() {
		l.Logger.Panicln(v...)
	})
}
