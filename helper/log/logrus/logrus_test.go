package logrus

import (
	"os"
	"testing"
	"webcrawler/helper/log/base"
	"webcrawler/helper/log/field"
)

func TestLogrusLogger(t *testing.T) {
	defer func() {
		if p := recover(); p != nil {
			switch i := p.(type) {
			case error, string:
				t.Fatalf("Fatal error: %s \n", i)
			default:
				t.Fatalf("Fatal error: %#v\n", i)
			}
		}
	}()
	loggers := []base.MyLogger{}
	loggers = append(loggers, NewLogger())
	loggers = append(loggers,
		NewLoggerBy(
			base.LEVEL_DEBUG,
			base.FORMAT_JSON,
			os.Stderr,
			[]base.Option{
				base.OptWithLocation{Value: true},
			},
		))

	for i, logger := range loggers {
		t.Logf("The tested logger [%d]: %#v", i, logger)
		log(logger)
	}
}

func log(logger base.MyLogger) {
	logger = logger.WithFields(
		field.Bool("bool", false),
		field.Int64("int64", 123456),
		field.Float64("float64", 123.456),
		field.String("string", "logrus"),
		field.Object("object", interface{}("1234abcd3dd")),
	)
	logger.Info("Info log logrus")
	logger.Infoln("Info log ln logrus")
	logger.Error("Error log logrus")
	logger.Errorln("Error ln log logrus")
	logger.Warn("Warn log logrus")
	logger.Warnln("Warn ln log logrus")
	logger.Errorf("%s error log f", "ERrorf")
}
