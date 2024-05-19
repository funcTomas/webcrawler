package scheduler

import "webcrawler/helper/log"

var logger = log.DLogger()

type Scheduler interface {
	Init(requestArgs RequestArgs)
}
