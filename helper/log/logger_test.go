package log

import (
	"os"
	"testing"
	"webcrawler/helper/log/base"
)

func TestLogger(t *testing.T) {
	logger := DLogger()
	if logger == nil {
		t.Fatal("the default logger is invalid")
	}
	if logger.Name() != "logrus" {
		t.Fatalf("Inconsistent logger type: expected: %s, actual: %s", "logrus", logger.Name())
	}
	t.Logf("The default logger: %#v\n", logger)

	loggerType := base.TYPE_LOGRUS
	loggerLevel := base.LEVEL_INFO
	logFormat := base.FORMAT_JSON
	options := []base.Option{
		base.OptWithLocation{Value: true},
	}
	logger = Logger(
		loggerType,
		loggerLevel,
		logFormat,
		os.Stdout,
		options,
	)
	if logger == nil {
		t.Fatal("the logrus logger is invalid")
	}
	if logger.Name() != "logrus" {
		t.Fatalf("Inconsistent logger type: expected: %s, actual: %s", "logrus", logger.Name())
	}
	if logger.Level() != base.LEVEL_INFO {
		t.Fatalf("Inconsistent logger level: expected: %d, actual: %d", loggerLevel, logger.Level())
	}
	if logger.Format() != logFormat {
		t.Fatalf("Inconsistent logger format: expected: %s, actual: %s", logFormat, logger.Format())
	}
	t.Logf("the logrus logger: %#v\n", logger)
}
