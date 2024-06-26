package scheduler

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
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

func NewScheduler() Scheduler {
	return &myScheduler{}
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
	cancelFunc        context.CancelFunc
	status            Status
	statusLock        sync.RWMutex
	summary           SchedSummary
}

func (sched *myScheduler) Init(requestArgs RequestArgs, dataArgs DataArgs, moduleArgs ModuleArgs) (err error) {
	logger.Info("Check status for initialization")
	var oldStatus Status
	oldStatus, err = sched.checkAndSetStatus(SCHED_STATUS_INITIALIZING)
	if err != nil {
		return
	}
	defer func() {
		sched.statusLock.Lock()
		if err != nil {
			sched.status = oldStatus
		} else {
			sched.status = SCHED_STATUS_INITIALIZED
		}
		sched.statusLock.Unlock()
	}()
	logger.Info("Check request arguments...")
	if err = requestArgs.Check(); err != nil {
		return err
	}
	logger.Info("Check data arguments...")
	if err = dataArgs.Check(); err != nil {
		return err
	}
	logger.Info("Data arguments are valid")
	logger.Info("Check module arguments...")
	if err = moduleArgs.Check(); err != nil {
		return err
	}
	logger.Info("Module arguments are valid")
	logger.Info("Initialize fields of scheduler")
	if sched.registrar == nil {
		sched.registrar = module.NewRegistrar()
	} else {
		sched.registrar.Clear()
	}
	sched.maxDepth = requestArgs.MaxDepth
	logger.Infof("-- Max depth: %d", sched.maxDepth)
	sched.acceptedDomainMap = sync.Map{}
	for _, domain := range requestArgs.AcceptedDomains {
		sched.acceptedDomainMap.Store(domain, struct{}{})
	}
	logger.Infof("-- Accepted primary domains: %v", requestArgs.AcceptedDomains)
	sched.urlMap = sync.Map{}
	sched.initBufferPool(dataArgs)
	sched.resetContext()
	sched.summary = newSchedSummary(requestArgs, dataArgs, moduleArgs, sched)
	logger.Info("Register modules")
	if err = sched.registerModules(moduleArgs); err != nil {
		return err
	}
	logger.Info("Scheduler has been initialized")
	return nil
}

func (sched *myScheduler) Start(firstHTTPReq *http.Request) (err error) {
	defer func() {
		if p := recover(); p != nil {
			errMsg := fmt.Sprintf("Fatal scheduler error %s", p)
			logger.Fatal(errMsg)
			err = genError(errMsg)
		}
	}()
	logger.Info("Start scheduler...")
	logger.Info("Check status for start...")
	var oldStatus Status
	oldStatus, err = sched.checkAndSetStatus(SCHED_STATUS_STARTING)
	defer func() {
		sched.statusLock.Lock()
		if err != nil {
			sched.status = oldStatus
		} else {
			sched.status = SCHED_STATUS_STARTED
		}
		sched.statusLock.Unlock()
	}()
	if err != nil {
		return
	}
	logger.Info("Check first HTTP request...")
	if firstHTTPReq == nil {
		err = genParameterError("nil first HTTP request")
		return
	}
	logger.Info("The first HTTP request is valid")
	logger.Info("Get the primary domain...")
	logger.Infof("-- Host: %s", firstHTTPReq.Host)
	var primaryDomain string
	primaryDomain, err = getPrimaryDomain(firstHTTPReq.Host)
	if err != nil {
		return
	}
	logger.Infof("-- Primary domain: %s", primaryDomain)
	sched.acceptedDomainMap.Store(primaryDomain, struct{}{})
	if err = sched.checkBufferPoolForStart(); err != nil {
		return
	}
	sched.download()
	sched.analyze()
	sched.pick()
	logger.Info("Scheduler has been started")
	firstReq := module.NewRequest(firstHTTPReq, 0)
	sched.sendReq(firstReq)
	return nil
}

func (sched *myScheduler) Stop() (err error) {
	logger.Info("Stop scheduler...")
	logger.Info("Check status for stop")
	var oldStatus Status
	oldStatus, err = sched.checkAndSetStatus(SCHED_STATUS_STOPPING)
	defer func() {
		sched.statusLock.Lock()
		if err != nil {
			sched.status = oldStatus
		} else {
			sched.status = SCHED_STATUS_STOPPED
		}
		sched.statusLock.Unlock()
	}()
	if err != nil {
		return
	}
	sched.cancelFunc()
	sched.reqBufferPool.Close()
	sched.respBufferPool.Close()
	sched.itemBufferPool.Close()
	sched.errorBufferPool.Close()
	logger.Info("Scheduler has been stopped")
	return nil
}

func (sched *myScheduler) Status() Status {
	var status Status
	sched.statusLock.RLock()
	status = sched.status
	sched.statusLock.RUnlock()
	return status
}

func (sched *myScheduler) ErrorChan() <-chan error {
	errBuffer := sched.errorBufferPool
	errCh := make(chan error, errBuffer.BufferCap())
	go func(errBuffer buffer.Pool, errCh chan error) {
		for {
			if sched.canceled() {
				close(errCh)
				break
			}
			datum, err := errBuffer.Get()
			if err != nil {
				logger.Warnln("The error buffer pool was closed. Break error reception")
				close(errCh)
				break
			}
			err, ok := datum.(error)
			if !ok {
				errMsg := fmt.Sprintf("incorrect error type: %T", datum)
				sendError(errors.New(errMsg), "", sched.errorBufferPool)
				continue
			}
			if sched.canceled() {
				close(errCh)
				break
			}
			errCh <- err
		}
	}(errBuffer, errCh)
	return errCh
}

func (sched *myScheduler) Idle() bool {
	moduleMap := sched.registrar.GetAll()
	for _, module := range moduleMap {
		if module.HandlingNumber() > 0 {
			return false
		}
	}
	if sched.reqBufferPool.Total() > 0 ||
		sched.respBufferPool.Total() > 0 ||
		sched.itemBufferPool.Total() > 0 {
		return false
	}
	return true
}

func (sched *myScheduler) Summary() SchedSummary {
	return sched.summary
}

func (sched *myScheduler) checkAndSetStatus(wantedStatus Status) (oldStatus Status, err error) {
	sched.statusLock.Lock()
	defer sched.statusLock.Unlock()
	oldStatus = sched.status
	err = checkStatus(oldStatus, wantedStatus, nil)
	if err == nil {
		sched.status = wantedStatus
	}
	return
}

func (sched *myScheduler) initBufferPool(dataArgs DataArgs) {
	if sched.reqBufferPool != nil && !sched.reqBufferPool.Closed() {
		sched.reqBufferPool.Close()
	}
	sched.reqBufferPool, _ = buffer.NewPool(dataArgs.ReqBufferCap, dataArgs.ReqMaxBufferNumber)
	logger.Infof("-- Request buffer pool: bufferCap: %d, maxBufferNumber: %d",
		sched.reqBufferPool.BufferCap(), sched.reqBufferPool.MaxBufferNumber())
	if sched.respBufferPool != nil && !sched.respBufferPool.Closed() {
		sched.respBufferPool.Close()
	}
	sched.respBufferPool, _ = buffer.NewPool(dataArgs.RespBufferCap, dataArgs.RespMaxBufferNumber)
	logger.Infof("-- Response buffer pool: bufferCap: %d, maxBufferNumber: %d",
		sched.respBufferPool.BufferCap(), sched.respBufferPool.MaxBufferNumber())
	if sched.itemBufferPool != nil && !sched.itemBufferPool.Closed() {
		sched.itemBufferPool.Close()
	}
	sched.itemBufferPool, _ = buffer.NewPool(dataArgs.ItemBufferCap, dataArgs.ItemMaxBufferNumber)
	logger.Infof("-- Item buffer pool: bufferCap: %d, maxBufferNumber: %d",
		sched.itemBufferPool.BufferCap(), sched.itemBufferPool.MaxBufferNumber())
	if sched.errorBufferPool != nil && !sched.errorBufferPool.Closed() {
		sched.errorBufferPool.Close()
	}
	sched.errorBufferPool, _ = buffer.NewPool(dataArgs.ErrorBufferCap, dataArgs.ErrorMaxBufferNumber)
	logger.Infof("-- Error buffer pool: bufferCap: %d, maxBufferNumber: %d",
		sched.errorBufferPool.BufferCap(), sched.errorBufferPool.MaxBufferNumber())
}

func (sched *myScheduler) resetContext() {
	sched.ctx, sched.cancelFunc = context.WithCancel(context.Background())
}

func (sched *myScheduler) registerModules(moduleArgs ModuleArgs) error {
	for _, d := range moduleArgs.Downloaders {
		if d == nil {
			continue
		}
		ok, err := sched.registrar.Register(d)
		if err != nil {
			return genErrorByError(err)
		}
		if !ok {
			errMsg := fmt.Sprintf("Could not register downloader instance with MID %q", d.ID())
			return genError(errMsg)
		}
	}
	logger.Infof("All downloads have benn registered (number: %d)", len(moduleArgs.Downloaders))
	for _, a := range moduleArgs.Analyzers {
		if a == nil {
			continue
		}
		ok, err := sched.registrar.Register(a)
		if err != nil {
			return genErrorByError(err)
		}
		if !ok {
			errMsg := fmt.Sprintf("Could not register analyzer instance with MID %q", a.ID())
			return genError(errMsg)
		}
	}
	logger.Infof("All analyzers have benn registered (number: %d)", len(moduleArgs.Analyzers))
	for _, p := range moduleArgs.Pipelines {
		if p == nil {
			continue
		}
		ok, err := sched.registrar.Register(p)
		if err != nil {
			return genErrorByError(err)
		}
		if !ok {
			errMsg := fmt.Sprintf("Could not register pipeline instance with MID %q", p.ID())
			return genError(errMsg)
		}
	}
	logger.Infof("All pipelines have benn registered (number: %d)", len(moduleArgs.Pipelines))
	return nil
}

func (sched *myScheduler) checkBufferPoolForStart() error {
	if sched.reqBufferPool == nil {
		return genError("nil request buffer pool")
	}
	if sched.reqBufferPool != nil && sched.reqBufferPool.Closed() {
		sched.reqBufferPool, _ = buffer.NewPool(sched.reqBufferPool.BufferCap(), sched.reqBufferPool.MaxBufferNumber())
	}
	if sched.respBufferPool == nil {
		return genError("nil response buffer pool")
	}
	if sched.respBufferPool != nil && sched.respBufferPool.Closed() {
		sched.respBufferPool, _ = buffer.NewPool(sched.respBufferPool.BufferCap(), sched.respBufferPool.MaxBufferNumber())
	}
	if sched.itemBufferPool == nil {
		return genError("nil item buffer pool")
	}
	if sched.itemBufferPool != nil && sched.itemBufferPool.Closed() {
		sched.itemBufferPool, _ = buffer.NewPool(sched.itemBufferPool.BufferCap(), sched.itemBufferPool.MaxBufferNumber())
	}
	if sched.errorBufferPool == nil {
		return genError("nil error buffer pool")
	}
	if sched.errorBufferPool != nil && sched.errorBufferPool.Closed() {
		sched.errorBufferPool, _ = buffer.NewPool(sched.errorBufferPool.BufferCap(), sched.errorBufferPool.MaxBufferNumber())
	}
	return nil
}

func (sched *myScheduler) download() {
	go func() {
		for {
			if sched.canceled() {
				break
			}
			datum, err := sched.reqBufferPool.Get()
			if err != nil {
				logger.Warnln("The request buffer pool was closed. Break request reception")
				break
			}
			req, ok := datum.(*module.Request)
			if !ok {
				errMsg := fmt.Sprintf("incorrect request type: %T", datum)
				sendError(errors.New(errMsg), "", sched.errorBufferPool)
			}
			sched.downloadOne(req)
		}
	}()
}

func (sched *myScheduler) downloadOne(req *module.Request) {
	if req == nil {
		return
	}
	if sched.canceled() {
		return
	}
	m, err := sched.registrar.Get(module.TYPE_DOWNLOADER)
	if err != nil || m == nil {
		errMsg := fmt.Sprintf("could not get a downloader: %s", err)
		sendError(errors.New(errMsg), "", sched.errorBufferPool)
		sched.sendReq(req)
		return
	}
	downloader, ok := m.(module.Downloader)
	if !ok {
		errMsg := fmt.Sprintf("incorrect downloader type: %T (MID: %s)", m, m.ID())
		sendError(errors.New(errMsg), m.ID(), sched.errorBufferPool)
		sched.sendReq(req)
		return
	}
	resp, err := downloader.Download(req)
	if resp != nil {
		sendResp(resp, sched.respBufferPool)
	}
	if err != nil {
		sendError(err, m.ID(), sched.errorBufferPool)
	}
}

func (sched *myScheduler) sendReq(req *module.Request) bool {
	if req == nil {
		return false
	}
	if sched.canceled() {
		return false
	}
	httpReq := req.HTTPReq()
	if httpReq == nil {
		logger.Warn("Ignore the request, Its HTTP request is invalid")
		return false
	}
	reqURL := httpReq.URL
	if reqURL == nil {
		logger.Warnf("Ignore the request, Its URL is invalid")
		return false
	}
	scheme := strings.ToLower(reqURL.Scheme)
	if scheme != "http" && scheme != "https" {
		logger.Warnf("Ignore the request, Its URL scheme is %q, but should be http or https (URL: %s)", scheme, reqURL)
		return false
	}
	if _, ok := sched.urlMap.Load(reqURL.String()); ok {
		logger.Warnf("Ignore the request, Its URL is repeated. (URL: %s)", reqURL)
		return false
	}
	pd, _ := getPrimaryDomain(httpReq.Host)
	if _, ok := sched.acceptedDomainMap.Load(pd); !ok {
		if pd == "bing.net" {
			panic(httpReq.URL)
		}
		logger.Warnf("Ignore the request. Its host %q is not in the primary domain map (url: %s)", httpReq.Host, reqURL)
		return false
	}
	if req.Depth() > sched.maxDepth {
		logger.Warnf("Ignore the request. Its depth reaches the max %d (URL: %s)", req.Depth(), reqURL)
		return false
	}
	go func(req *module.Request) {
		if err := sched.reqBufferPool.Put(req); err != nil {
			logger.Warnln("The request buffer pool was closed. Ignore the request sending")
		}
	}(req)
	sched.urlMap.Store(reqURL.String(), struct{}{})
	return true
}

func (sched *myScheduler) canceled() bool {
	select {
	case <-sched.ctx.Done():
		return true
	default:
		return false
	}
}

func (sched *myScheduler) analyze() {
	go func() {
		for {
			if sched.canceled() {
				break
			}
			datum, err := sched.respBufferPool.Get()
			if err != nil {
				logger.Warnln("The response buffer pool was closed. Break response reception")
				break
			}
			resp, ok := datum.(*module.Response)
			if !ok {
				errMsg := fmt.Sprintf("incorrect response type: %T", datum)
				sendError(errors.New(errMsg), "", sched.errorBufferPool)
			}
			sched.analyzeOne(resp)
		}
	}()
}

func (sched *myScheduler) analyzeOne(resp *module.Response) {
	if resp == nil {
		return
	}
	if sched.canceled() {
		return
	}
	m, err := sched.registrar.Get(module.TYPE_ANALYZER)
	if err != nil || m == nil {
		errMsg := fmt.Sprintf("could not get an analyzer: %s", err)
		sendError(errors.New(errMsg), "", sched.errorBufferPool)
		sendResp(resp, sched.respBufferPool)
		return
	}
	analyzer, ok := m.(module.Analyzer)
	if !ok {
		errMsg := fmt.Sprintf("incorrect analyzer type: %T (MID: %s)", m, m.ID())
		sendError(errors.New(errMsg), m.ID(), sched.errorBufferPool)
		sendResp(resp, sched.respBufferPool)
		return
	}
	dataList, errs := analyzer.Analyze(resp)
	for _, data := range dataList {
		if data == nil {
			continue
		}
		switch d := data.(type) {
		case *module.Request:
			sched.sendReq(d)
		case module.Item:
			sendItem(d, sched.itemBufferPool)
		default:
			errMsg := fmt.Sprintf("Unsupported data type: %T (data: %#v)", d, d)
			sendError(errors.New(errMsg), m.ID(), sched.errorBufferPool)
		}
	}
	for _, err := range errs {
		sendError(err, m.ID(), sched.errorBufferPool)
	}
}

func sendResp(resp *module.Response, respBufferPool buffer.Pool) bool {
	if resp == nil || respBufferPool == nil || respBufferPool.Closed() {
		return false
	}
	go func(resp *module.Response) {
		if err := respBufferPool.Put(resp); err != nil {
			logger.Warnf("The reponse buffer pool was closed. Ignore request sending")
		}
	}(resp)
	return true
}

func sendItem(item module.Item, itemBufferPool buffer.Pool) bool {
	if item == nil || itemBufferPool == nil || itemBufferPool.Closed() {
		return false
	}
	go func(item module.Item) {
		if err := itemBufferPool.Put(item); err != nil {
			logger.Warnln("The item buffer pool was closed. Ignore item sending")
		}
	}(item)
	return true
}

func (sched *myScheduler) pick() {
	go func() {
		for {
			if sched.canceled() {
				break
			}
			datum, err := sched.itemBufferPool.Get()
			if err != nil {
				logger.Warnln("The item buffer pool was closed. Break item reception")
				break
			}
			item, ok := datum.(module.Item)
			if !ok {
				errMsg := fmt.Sprintf("incorrect item type: %T", datum)
				sendError(errors.New(errMsg), "", sched.errorBufferPool)
			}
			sched.pickOne(item)
		}
	}()
}

func (sched *myScheduler) pickOne(item module.Item) {
	if sched.canceled() {
		return
	}
	m, err := sched.registrar.Get(module.TYPE_PIPELINE)
	if err != nil || m == nil {
		errMsg := fmt.Sprintf("could not get a pipeline: %s", err)
		sendError(errors.New(errMsg), "", sched.errorBufferPool)
		sendItem(item, sched.itemBufferPool)
		return
	}
	pipeline, ok := m.(module.Pipleline)
	if !ok {
		errMsg := fmt.Sprintf("incorrect pipeline type: %T (MID: %s)", m, m.ID())
		sendError(errors.New(errMsg), m.ID(), sched.errorBufferPool)
		sendItem(item, sched.itemBufferPool)
		return
	}
	errs := pipeline.Send(item)
	for _, err := range errs {
		sendError(err, m.ID(), sched.errorBufferPool)
	}
}
