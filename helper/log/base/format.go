package base

type LogFormat string

const (
	FORMAT_TEXT LogFormat = "text"
	FORMAT_JSON LogFormat = "json"
)

const TIMESTAMP_FORMAT = "2006-01-02T15:04:05.999"
