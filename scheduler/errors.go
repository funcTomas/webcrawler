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

func genParameterError(errMsg string) error {
	return errors.NewCrawlerErrorBy(errors.ERROR_TYPE_SCHEDULER, errors.NewIllegalParameterError(errMsg))
}

func sendError(err error, mid module.MID, errBufferPool buffer.Pool) bool {
	if err == nil || errBufferPool == nil || errBufferPool.Closed() {
		return false
	}
	var crawlerError errors.CrawlerError
	var ok bool
	crawlerError, ok = err.(errors.CrawlerError)
	if !ok {
		var moduleType module.Type
		var errorType errors.ErrorType
		ok, moduleType := module.GetType(mid)
		if !ok {
			errorType = errors.ERROR_TYPE_SCHEDULER
		} else {
			switch moduleType {
			case module.TYPE_DOWNLOADER:
				errorType = errors.ERROR_TYPE_DOWNLOADER
			case module.TYPE_ANALYZER:
				errorType = errors.ERROR_TYPE_ANALYZER
			case module.TYPE_PIPELINE:
				errorType = errors.ERROR_TYPE_PIPELINE
			}
		}
		crawlerError = errors.NewCrawlerError(errorType, err.Error())
	}
	if errBufferPool.Closed() {
		return false
	}
	go func(crawlerError errors.CrawlerError) {
		if err := errBufferPool.Put(crawlerError); err != nil {
			logger.Warnf("The error buffer pool was closed. Ignore error sending")
		}
	}(crawlerError)
	return true
}
