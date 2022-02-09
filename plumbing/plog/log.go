package plog

import "io"

func Debug(v ...interface{}) {
	DefaultLogger.Debug(v...)
}

func Debugf(format string, v ...interface{}) {
	DefaultLogger.Debugf(format, v...)
}

func Debugln(v ...interface{}) {
	DefaultLogger.Debugln(v...)
}

func Info(v ...interface{}) {
	DefaultLogger.Info(v...)
}

func Infof(format string, v ...interface{}) {
	DefaultLogger.Infof(format, v...)
}

func Infoln(v ...interface{}) {
	DefaultLogger.Infoln(v...)
}

func Warn(v ...interface{}) {
	DefaultLogger.Warn(v...)
}

func Warnf(format string, v ...interface{}) {
	DefaultLogger.Warnf(format, v...)
}

func Warnln(v ...interface{}) {
	DefaultLogger.Warnln(v...)
}

func Error(v ...interface{}) {
	DefaultLogger.Error(v...)
}

func Errorf(format string, v ...interface{}) {
	DefaultLogger.Errorf(format, v...)
}

func Errorln(v ...interface{}) {
	DefaultLogger.Errorln(v...)
}

func Fatal(v ...interface{}) {
	DefaultLogger.Fatal(v...)
}

func Fatalf(format string, v ...interface{}) {
	DefaultLogger.Fatalf(format, v...)
}

func Fatalln(v ...interface{}) {
	DefaultLogger.Fatalln(v...)
}

func Panic(v ...interface{}) {
	DefaultLogger.Panic(v...)
}

func Panicf(format string, v ...interface{}) {
	DefaultLogger.Panicf(format, v...)
}

func Panicln(v ...interface{}) {
	DefaultLogger.Panicln(v...)
}

func Writer() io.Writer {
	return DefaultLogger.Writer()
}
