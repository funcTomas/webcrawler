package scheduler

import (
	"webcrawler/errors"
	"webcrawler/module"
	"webcrawler/toolkit/buffer"
)

func genError(errMsg string) error {
	return errors.NewCrawlerError(errors.ERROR_TYPE_SCHEDULER, errMsg)
}

func genErrorByError(err error) error {
	return errors.NewCrawlerError(errors.ERROR_TYPE_SCHEDULER, err.Error())
}

func sendError(err error, mid module.MID, errBufferPool buffer.Pool) bool
