package logrus

import (
	"io"
	"os"
	"webcrawler/helper/log/base"
	"webcrawler/helper/log/field"

	"github.com/sirupsen/logrus"
)

type loggerLogrus struct {
	level           base.LogLevel
	format          base.LogFormat
	optWithLocation base.OptWithLocation
	inner           *logrus.Entry
}

func (l *loggerLogrus) Name() string {
	return "logrus"
}

func (l *loggerLogrus) Level() base.LogLevel {
	return l.level
}

func (l *loggerLogrus) Format() base.LogFormat {
	return l.format
}

func (l *loggerLogrus) Options() []base.Option {
	return []base.Option{l.optWithLocation}
}

// Debug implements base.MyLogger.
func (l *loggerLogrus) Debug(v ...interface{}) {
	l.getInner().Debug(v...)
}

// Debugf implements base.MyLogger.
func (l *loggerLogrus) Debugf(format string, v ...interface{}) {
	l.getInner().Debugf(format, v...)
}

// Debugln implements base.MyLogger.
func (l *loggerLogrus) Debugln(v ...interface{}) {
	l.getInner().Debugln(v...)
}

// Error implements base.MyLogger.
func (l *loggerLogrus) Error(v ...interface{}) {
	l.getInner().Error(v...)
}

// Errorf implements base.MyLogger.
func (l *loggerLogrus) Errorf(format string, v ...interface{}) {
	l.getInner().Errorf(format, v...)
}

// Errorln implements base.MyLogger.
func (l *loggerLogrus) Errorln(v ...interface{}) {
	l.getInner().Errorln(v...)
}

// Fatal implements base.MyLogger.
func (l *loggerLogrus) Fatal(v ...interface{}) {
	l.getInner().Fatal(v...)
}

// Fatalf implements base.MyLogger.
func (l *loggerLogrus) Fatalf(format string, v ...interface{}) {
	l.getInner().Fatalf(format, v...)
}

// Fatalln implements base.MyLogger.
func (l *loggerLogrus) Fatalln(v ...interface{}) {
	l.getInner().Fatalln(v...)
}

// Info implements base.MyLogger.
func (l *loggerLogrus) Info(v ...interface{}) {
	l.getInner().Info(v...)
}

// Infof implements base.MyLogger.
func (l *loggerLogrus) Infof(format string, v ...interface{}) {
	l.getInner().Infof(format, v...)
}

// Infoln implements base.MyLogger.
func (l *loggerLogrus) Infoln(v ...interface{}) {
	l.getInner().Infoln(v...)
}

// Panic implements base.MyLogger.
func (l *loggerLogrus) Panic(v ...interface{}) {
	l.getInner().Panic(v...)
}

// Panicf implements base.MyLogger.
func (l *loggerLogrus) Panicf(format string, v ...interface{}) {
	l.getInner().Panicf(format, v...)
}

// Panicln implements base.MyLogger.
func (l *loggerLogrus) Panicln(v ...interface{}) {
	l.getInner().Panicln(v...)
}

// Warn implements base.MyLogger.
func (l *loggerLogrus) Warn(v ...interface{}) {
	l.getInner().Warn(v...)
}

// Warnf implements base.MyLogger.
func (l *loggerLogrus) Warnf(format string, v ...interface{}) {
	l.getInner().Warnf(format, v...)
}

// Warnln implements base.MyLogger.
func (l *loggerLogrus) Warnln(v ...interface{}) {
	l.getInner().Warnln(v...)
}

// WithFields implements base.MyLogger.
func (l *loggerLogrus) WithFields(fields ...field.Field) base.MyLogger {
	fieldsLen := len(fields)
	if fieldsLen == 0 {
		return l
	}
	logrusFields := make(map[string]interface{}, fieldsLen)
	for _, curField := range fields {
		logrusFields[curField.Name()] = curField.Value()
	}
	return &loggerLogrus{
		level:           l.level,
		format:          l.format,
		optWithLocation: l.optWithLocation,
		inner:           l.inner.WithFields(logrusFields),
	}
}

func NewLogger() base.MyLogger {
	return NewLoggerBy(base.LEVEL_INFO, base.FORMAT_TEXT, os.Stdout, nil)
}

func NewLoggerBy(level base.LogLevel, format base.LogFormat, writer io.Writer, options []base.Option) base.MyLogger {
	var logrusLevel logrus.Level
	switch level {
	default:
		level = base.LEVEL_INFO
		logrusLevel = logrus.InfoLevel
	case base.LEVEL_DEBUG:
		logrusLevel = logrus.DebugLevel
	case base.LEVEL_WARN:
		logrusLevel = logrus.WarnLevel
	case base.LEVEL_ERROR:
		logrusLevel = logrus.ErrorLevel
	case base.LEVEL_FATAL:
		logrusLevel = logrus.FatalLevel
	case base.LEVEL_PANIC:
		logrusLevel = logrus.PanicLevel
	}
	var optWithLocation base.OptWithLocation
	for _, opt := range options {
		if opt.Name() == "with location" {
			optWithLocation, _ = opt.(base.OptWithLocation)
		}
	}
	return &loggerLogrus{
		level:           level,
		format:          format,
		optWithLocation: optWithLocation,
		inner:           initInnerLogger(logrusLevel, format, writer),
	}
}

func initInnerLogger(level logrus.Level, format base.LogFormat, writer io.Writer) *logrus.Entry {
	innerLogger := logrus.New()

	switch format {
	case base.FORMAT_JSON:
		innerLogger.Formatter = &logrus.JSONFormatter{
			TimestampFormat: base.TIMESTAMP_FORMAT,
		}
	default:
		innerLogger.Formatter = &logrus.TextFormatter{
			FullTimestamp:   true,
			TimestampFormat: base.TIMESTAMP_FORMAT,
			DisableSorting:  false,
		}
	}
	innerLogger.Level = level
	innerLogger.Out = writer
	return logrus.NewEntry(innerLogger)
}

func (l *loggerLogrus) getInner() *logrus.Entry {
	inner := l.inner
	if l.optWithLocation.Value {
		inner = WithLocation(inner)
	}
	return inner
}

func WithLocation(entry *logrus.Entry) *logrus.Entry {
	funcPath, fileName, line := base.GetInvokerLocation(4)
	return entry.WithField(
		"location", map[string]interface{}{
			"func_path": funcPath,
			"file_name": fileName,
			"line":      line,
		},
	)

}
