package logrus

import (
	"io"
	"os"
	"webcrawler/helper/log/base"

	"github.com/sirupsen/logrus"
)

type Logrus struct {
	level          base.LogLevel
	format         base.LogFormat
	opWithLocation base.OptWithLocation
	inner          *logrus.Entry
}

func NewLogger() base.MyLogger {
	return NewLoggerBy(base.LEVEL_INFO, base.FORMAT_TEXT, os.Stdout, nil)
}

func NewLoggerBy(
	level base.LogLevel,
	format base.LogFormat,
	writer io.Writer,
	options []base.Option) base.MyLogger {
	return &loggerLogrus{}
}
