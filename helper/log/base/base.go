package base

import "webcrawler/helper/log/field"

type Option interface {
	Name() string
}

type OptWithLocation struct {
	Value bool
}

func (opt OptWithLocation) Name() string {
	return "with location"
}

type MyLogger interface {
	Name() string
	Level() LogLevel
	Format() LogFormat
	Options() []Option

	Debug(v ...interface{})
	Debugf(format string, v ...interface{})
	Debugln(v ...interface{})
	Error(v ...interface{})
	Errorf(format string, v ...interface{})
	Errorln(v ...interface{})
	Fatal(v ...interface{})
	Fatalf(format string, v ...interface{})
	Fatalln(v ...interface{})
	Info(v ...interface{})
	Infof(format string, v ...interface{})
	Infoln(v ...interface{})
	Panic(v ...interface{})
	Panicf(format string, v ...interface{})
	Panicln(v ...interface{})
	Warn(v ...interface{})
	Warnf(format string, v ...interface{})
	Warnln(v ...interface{})

	WithFields(fields ...field.Field) MyLogger
}
