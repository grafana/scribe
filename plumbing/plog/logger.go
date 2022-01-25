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
			return nil
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

func WithPrefix(logger *log.Logger, prefix string, f func(*log.Logger)) {
	logger.SetPrefix(prefix + " ")
	f(logger)
	logger.SetPrefix("")
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
	WithPrefix(l.Logger, "[debug]", func(logger *log.Logger) {
		RequireLevel(l.level, LogLevelDebug, func() {
			logger.Print(v...)
		})
	})
}

func (l *Logger) Debugf(format string, v ...interface{}) {
	WithPrefix(l.Logger, "[debug]", func(logger *log.Logger) {
		RequireLevel(l.level, LogLevelDebug, func() {
			logger.Printf(format, v...)
		})
	})
}

func (l *Logger) Debugln(v ...interface{}) {
	WithPrefix(l.Logger, "[debug]", func(logger *log.Logger) {
		RequireLevel(l.level, LogLevelDebug, func() {
			logger.Println(v...)
		})
	})
}

func (l *Logger) Info(v ...interface{}) {
	WithPrefix(l.Logger, "[info]", func(logger *log.Logger) {
		RequireLevel(l.level, LogLevelInfo, func() {
			logger.Print(v...)
		})
	})
}

func (l *Logger) Infof(format string, v ...interface{}) {
	WithPrefix(l.Logger, "[info]", func(logger *log.Logger) {
		RequireLevel(l.level, LogLevelInfo, func() {
			logger.Printf(format, v...)
		})
	})
}

func (l *Logger) Infoln(v ...interface{}) {
	WithPrefix(l.Logger, "[info]", func(logger *log.Logger) {
		RequireLevel(l.level, LogLevelInfo, func() {
			logger.Println(v...)
		})
	})
}

func (l *Logger) Warn(v ...interface{}) {
	WithPrefix(l.Logger, "[warn]", func(logger *log.Logger) {
		RequireLevel(l.level, LogLevelWarn, func() {
			logger.Print(v...)
		})
	})
}

func (l *Logger) Warnf(format string, v ...interface{}) {
	WithPrefix(l.Logger, "[warn]", func(logger *log.Logger) {
		RequireLevel(l.level, LogLevelWarn, func() {
			logger.Printf(format, v...)
		})
	})
}

func (l *Logger) Warnln(v ...interface{}) {
	WithPrefix(l.Logger, "[warn]", func(logger *log.Logger) {
		RequireLevel(l.level, LogLevelWarn, func() {
			logger.Println(v...)
		})
	})
}

func (l *Logger) Error(v ...interface{}) {
	WithPrefix(l.Logger, "[error]", func(logger *log.Logger) {
		RequireLevel(l.level, LogLevelError, func() {
			logger.Print(v...)
		})
	})
}

func (l *Logger) Errorf(format string, v ...interface{}) {
	WithPrefix(l.Logger, "[error]", func(logger *log.Logger) {
		RequireLevel(l.level, LogLevelError, func() {
			logger.Printf(format, v...)
		})
	})
}

func (l *Logger) Errorln(v ...interface{}) {
	WithPrefix(l.Logger, "[error]", func(logger *log.Logger) {
		RequireLevel(l.level, LogLevelError, func() {
			logger.Println(v...)
		})
	})
}

func (l *Logger) Fatal(v ...interface{}) {
	WithPrefix(l.Logger, "[fatal]", func(logger *log.Logger) {
		RequireLevel(l.level, LogLevelError, func() {
			logger.Fatal(v...)
		})
	})
}

func (l *Logger) Fatalf(format string, v ...interface{}) {
	WithPrefix(l.Logger, "[fatal]", func(logger *log.Logger) {
		RequireLevel(l.level, LogLevelError, func() {
			logger.Fatalf(format, v...)
		})
	})
}

func (l *Logger) Fatalln(v ...interface{}) {
	WithPrefix(l.Logger, "[fatal]", func(logger *log.Logger) {
		RequireLevel(l.level, LogLevelError, func() {
			logger.Fatalln(v...)
		})
	})
}

func (l *Logger) Panic(v ...interface{}) {
	WithPrefix(l.Logger, "[panic]", func(logger *log.Logger) {
		RequireLevel(l.level, LogLevelError, func() {
			logger.Panic(v...)
		})
	})
}

func (l *Logger) Panicf(format string, v ...interface{}) {
	WithPrefix(l.Logger, "[panic]", func(logger *log.Logger) {
		RequireLevel(l.level, LogLevelError, func() {
			logger.Panicf(format, v...)
		})
	})
}

func (l *Logger) Panicln(v ...interface{}) {
	WithPrefix(l.Logger, "[panic]", func(logger *log.Logger) {
		RequireLevel(l.level, LogLevelError, func() {
			logger.Panicln(v...)
		})
	})
}
