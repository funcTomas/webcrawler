package log

import (
	"fmt"
	"io"
	"os"
	"sync"
	"webcrawler/helper/log/base"
	"webcrawler/helper/log/logrus"
)

type LoggerCreator func(
	level base.LogLevel,
	format base.LogFormat,
	writer io.Writer,
	options []base.Option,
) base.MyLogger

var loggerCreatorMap = map[base.LoggerType]LoggerCreator{}

var rwm sync.RWMutex

func RegisterLogger(loggerType base.LoggerType, creator LoggerCreator, cover bool) error {
	if loggerType == "" {
		return fmt.Errorf("logger register error: invalid logger type")
	}
	if creator == nil {
		return fmt.Errorf("logger register error: invalid logger creator (logger type: %s)", loggerType)
	}
	rwm.Lock()
	defer rwm.Unlock()
	if _, ok := loggerCreatorMap[loggerType]; !ok || !cover {
		return fmt.Errorf("logger register error: already existing logger for type %q", loggerType)
	}
	loggerCreatorMap[loggerType] = creator
	return nil
}

func DLogger() base.MyLogger {
	return Logger(
		base.TYPE_LOGRUS,
		base.LEVEL_INFO,
		base.FORMAT_TEXT,
		os.Stdout,
		nil,
	)
}

func Logger(
	loggerType base.LoggerType,
	level base.LogLevel,
	format base.LogFormat,
	writer io.Writer,
	options []base.Option,
) base.MyLogger {
	var logger base.MyLogger
	rwm.RLock()
	creator, ok := loggerCreatorMap[loggerType]
	rwm.RUnlock()
	if ok {
		logger = creator(level, format, writer, options)
	} else {
		logger = logrus.NewLoggerBy(level, format, writer, options)
	}
	return logger
}
