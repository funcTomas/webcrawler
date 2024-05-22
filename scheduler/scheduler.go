package scheduler

import (
	"context"
	"net/http"
	"sync"
	"webcrawler/helper/log"
	"webcrawler/module"
	"webcrawler/toolkit/buffer"
)

var logger = log.DLogger()

type Scheduler interface {
	Init(requestArgs RequestArgs, dataArgs DataArgs, moduleArgs ModuleArgs) (err error)
	Start(firstHTTPReq *http.Request) (err error)
	Stop() (err error)
	Status() Status
	ErrorChan() <-chan error
	Idle() bool
	Summary() SchedSummary
}

type myScheduler struct {
	maxDepth          uint32
	acceptedDomainMap sync.Map
	registrar         module.Registrar
	reqBufferPool     buffer.Pool
	respBufferPool    buffer.Pool
	itemBufferPool    buffer.Pool
	errorBufferPool   buffer.Pool
	urlMap            sync.Map
	ctx               context.Context
	cancleFunc        context.CancelFunc
	status            Status
	statusLock        sync.Mutex
	summary           SchedSummary
}
