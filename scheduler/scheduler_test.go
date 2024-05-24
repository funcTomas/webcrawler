package scheduler

import (
	"net/http"
	"runtime"
	"sync"
	"testing"
	"time"
	"webcrawler/module"
	"webcrawler/toolkit/buffer"
)

var snGen = module.NewSNGenerator(1, 0)

func TestSchedNew(t *testing.T) {
	sched := NewScheduler()
	if sched == nil {
		t.Fatalf("Could not create scheduler")
	}
}

func TestSchedInit(t *testing.T) {
	requestArgs := genRequestArgs([]string{"bing.com"}, 0)
	dataArgs := genDataArgs(10, 2, 1)
	moduleArgs := genSimpleModuleArgs(3, 2, 1, t)
	sched := NewScheduler()
	err := sched.Init(requestArgs, dataArgs, moduleArgs)
	if err != nil {
		t.Fatalf("An error occurs when initializing scheduler: %s", err)
	}
	err = sched.Init(requestArgs, dataArgs, moduleArgs)
	if err != nil {
		t.Fatalf("An error occurs when initializing scheduler: %s", err)
	}
	invalidRequestArgs := genRequestArgs(nil, 0)
	err = sched.Init(invalidRequestArgs, dataArgs, moduleArgs)
	if err == nil {
		t.Fatalf("No error when initializing scheduler with illegal request arguments: %#v", invalidRequestArgs)
	}
	sched = NewScheduler()
	invalidDataArgs := genDataArgs(0, 0, 0)
	err = sched.Init(requestArgs, invalidDataArgs, moduleArgs)
	if err == nil {
		t.Fatalf("No error when initializing scheduler with illegal data arguments: %#v", invalidDataArgs)
	}
	invalidModuleArgsList := []ModuleArgs{
		{},
		genSimpleModuleArgs(-1, 1, 1, t),
		genSimpleModuleArgs(1, -1, 1, t),
		genSimpleModuleArgs(1, 1, -1, t),
		{
			Downloaders: genSimpleDownloaders(2, true, snGen, t),
			Analyzers:   genSimpleAnalyzers(2, false, snGen, t),
			Pipelines:   genSimplePipelines(2, false, snGen, t),
		},
		{
			Downloaders: genSimpleDownloaders(2, false, snGen, t),
			Analyzers:   genSimpleAnalyzers(2, true, snGen, t),
			Pipelines:   genSimplePipelines(2, false, snGen, t),
		},
		{
			Downloaders: genSimpleDownloaders(2, false, snGen, t),
			Analyzers:   genSimpleAnalyzers(2, false, snGen, t),
			Pipelines:   genSimplePipelines(2, true, snGen, t),
		},
	}
	for _, invalidModuleArgs := range invalidModuleArgsList {
		sched = NewScheduler()
		dataArgs := genDataArgs(10, 2, 1)
		err = sched.Init(requestArgs, dataArgs, invalidModuleArgs)
		if err == nil {
			t.Fatalf("No error when creating scheduler with illegal module arguments: %#v", invalidModuleArgs)
		}
	}
	invalidModuleArgsList = []ModuleArgs{
		genSimpleModuleArgs(-2, 1, 1, t),
		genSimpleModuleArgs(1, -2, 1, t),
		genSimpleModuleArgs(1, 1, -2, t),
	}
	for _, invalidModuleArgs := range invalidModuleArgsList {
		sched = NewScheduler()
		err = sched.Init(requestArgs, dataArgs, invalidModuleArgs)
		if err == nil {
			t.Fatalf("An error occurs when initializing scheduler: %s", err)
		}
	}
}

func TestSchedStart(t *testing.T) {
	sched := NewScheduler()
	requestArgs := genRequestArgs([]string{}, 0)
	dataArgs := genDataArgs(10, 2, 1)
	moduleArgs := genSimpleModuleArgs(3, 2, 1, t)
	err := sched.Init(requestArgs, dataArgs, moduleArgs)
	if err != nil {
		t.Fatalf("An error occurs when initializing scheduler: %s", err)
	}
	url := "http://cn.bing.com/search?q=golang"
	firstHTTPReq, err := http.NewRequest("GET", url, nil)
	if err != nil {
		t.Fatalf("An error occurs when creating a HTTP request: %s (url: %s)", err, url)
	}
	err = sched.Start(firstHTTPReq)
	if err != nil {
		t.Fatalf("An error occurs when starting scheduler: %s", err)
	}
	sched.Stop()
	err = sched.Start(nil)
	if err == nil {
		t.Fatalf("No error when starting scheduler with empty HTTP host")
	}
	sched.Stop()
	firstHTTPReq.Host = ""
	err = sched.Start(firstHTTPReq)
	if err == nil {
		t.Fatalf("No error when starting scheduler with empty HTTP host")
	}
	sched.Stop()
}

func TestSchedStop(t *testing.T) {
	sched := NewScheduler()
	requestArgs := genRequestArgs([]string{}, 0)
	dataArgs := genDataArgs(10, 2, 1)
	moduleArgs := genSimpleModuleArgs(3, 2, 1, t)
	err := sched.Init(requestArgs, dataArgs, moduleArgs)
	if err != nil {
		t.Fatalf("An error occurs when initializing scheduler: %s", err)
	}
	url := "http://cn.bing.com/search?q=golang"
	firstHTTPReq, err := http.NewRequest("GET", url, nil)
	if err != nil {
		t.Fatalf("An error occurs when creating a HTTP request: %s (url: %s)", err, url)
	}
	err = sched.Start(firstHTTPReq)
	if err != nil {
		t.Fatalf("An error occurs when starting scheduler: %s", err)
	}
	if err = sched.Stop(); err != nil {
		t.Fatalf("An error occurs when stopping scheduler: %s", err)
	}
}

func TestSchedStatus(t *testing.T) {
	requestArgs := genRequestArgs([]string{"bing.com"}, 0)
	dataArgs := genDataArgs(10, 2, 1)
	moduleArgs := genSimpleModuleArgs(3, 2, 1, t)
	sched := NewScheduler()
	url := "http://cn.bing.com/search?q=golang"
	firstHTTPReq, err := http.NewRequest("GET", url, nil)
	if err != nil {
		t.Fatalf("An error occur when creating a HTTP request: %s (url: %s)", err, url)
	}
	if err = sched.Start(firstHTTPReq); err == nil {
		t.Fatal("No error when starting the scheduler before initialize")
	}
	if err = sched.Stop(); err == nil {
		t.Fatal("No error when stopping schduler before initialize")
	}
	if err = sched.Init(requestArgs, dataArgs, moduleArgs); err != nil {
		t.Fatalf("An error occurs when initializing the scheduler: %s", err)
	}
	if err = sched.Init(requestArgs, dataArgs, moduleArgs); err != nil {
		t.Fatalf("An error occurs when repeatedly initializing the scheduler: %s", err)
	}
	if err = sched.Stop(); err == nil {
		t.Fatal("No error when stop scheduler after initialize")
	}
	if err = sched.Start(firstHTTPReq); err != nil {
		t.Fatalf("An error occurs when starting scheduler after initialize: %s", err)
	}
	if err = sched.Start(firstHTTPReq); err == nil {
		t.Fatal("No error when repeatedly start scheduler")
	}

	if err = sched.Init(requestArgs, dataArgs, moduleArgs); err == nil {
		t.Fatal("No error when initializing scheduler after start")
	}
	if err = sched.Stop(); err != nil {
		t.Fatalf("An error occurs when stopping scheduler after start: %s", err)
	}
	if err = sched.Stop(); err == nil {
		t.Fatalf("No error when repeatedly stop scheduler")
	}
	if err = sched.Init(requestArgs, dataArgs, moduleArgs); err != nil {
		t.Fatalf("An error occurs when initializing scheduler after stop: %s", err)
	}
}

func TestSchedSimple(t *testing.T) {
	requestArgs := genRequestArgs([]string{}, 0)
	dataArgs := genDataArgs(10, 2, 1)
	moduleArgs := genSimpleModuleArgs(3, 2, 1, t)
	url := "http://cn.bing.com/search?q=golang"
	firstHTTPReq, err := http.NewRequest("GET", url, nil)
	if err != nil {
		t.Fatalf("An error occurs when creating a HTTP request: %s (url: %s)", err, url)
	}
	sched := NewScheduler()
	err = sched.Init(requestArgs, dataArgs, moduleArgs)
	if err != nil {
		t.Fatalf("An error occurs when initializing scheduler: %s", err)
	}
	err = sched.Start(firstHTTPReq)
	if err != nil {
		t.Fatalf("An error occurs when starting scheduler: %s", err)
	}
	go func() {
		errChan := sched.ErrorChan()
		for {
			err, ok := <-errChan
			if !ok {
				break
			}
			t.Errorf("An error occurs when running scheduler: %s", err)
		}
	}()
	var count int
	max := 5
	tickCh := time.Tick(time.Second)
	for range tickCh {
		if sched.Idle() {
			count++
			logger.Infof("Increase idle count, and value is %d", count)
		}
		if count >= max {
			logger.Infof("The idle count is equal or greater than %d", max)
			break
		}
	}
	if err = sched.Stop(); err != nil {
		t.Fatalf("An error occurs when stopping scheduler: %s", err)
	}
	_, ok := <-sched.ErrorChan()
	if ok {
		t.Fatalf("The error channel has not been closed in stopped scheduler")
	}
	if _, ok := <-sched.ErrorChan(); ok {
		t.Logf("Closed error channel")
	}
	logger.Infof("-- Final summary:\n %s", sched.Summary())
}

func TestSchedSendReq(t *testing.T) {
	requestArgs := genRequestArgs([]string{}, 0)
	dataArgs := genDataArgs(10, 2, 1)
	moduleArgs := genSimpleModuleArgs(3, 2, 1, t)
	sched := NewScheduler()
	err := sched.Init(requestArgs, dataArgs, moduleArgs)
	if err != nil {
		t.Fatalf("An error occurs when initializing scheduler: %s", err)
	}
	url := "http://cn.bing.com/search?q=golang"
	firstHTTPReq, err := http.NewRequest("GET", url, nil)
	if err != nil {
		t.Fatalf("An error occurs when creating a HTTP request: %s (url: %s)", err, url)
	}
	err = sched.Start(firstHTTPReq)
	if err != nil {
		t.Fatalf("An error occurs when starting scheduler: %s", err)
	}
	mySched := sched.(*myScheduler)
	if mySched.sendReq(nil) {
		t.Fatalf("It can still send nil request")
	}
	url = "http://cn.bing.com/images/serach?q=golang"
	httpReq, err := http.NewRequest("GET", url, nil)
	if err != nil {
		t.Fatalf("An error occurs when creating a HTTP request: %s (url: %s)", err, url)
	}
	req := module.NewRequest(httpReq, 0)
	if !mySched.sendReq(req) {
		t.Fatalf("Could not send request")
	}
	if mySched.sendReq(req) {
		t.Fatalf("It can still send req repeatedly")
	}
	mySched.urlMap = sync.Map{}
	httpReq.URL.Scheme = "tcp"
	if mySched.sendReq(req) {
		t.Fatalf("It can still send request with unsupported URL scheme")
	}
	httpReq.URL = nil
	if mySched.sendReq(req) {
		t.Fatalf("It can still send request with nil URL")
	}
	sched.Stop()
	time.Sleep(time.Millisecond * 500)
	if mySched.sendReq(nil) {
		t.Fatal("It can still send request in stopped scheduler")
	}
}

func TestSendResp(t *testing.T) {
	buffer, _ := buffer.NewPool(10, 2)
	if sendResp(nil, buffer) {
		t.Fatalf("It can still send nil response")
	}
	httpReq, _ := http.NewRequest("GET", "https://github.com/gopcp", nil)
	httpResp := &http.Response{
		Request: httpReq,
		Body:    nil,
	}
	resp := module.NewResponse(httpResp, 0)
	buffer.Close()
	done := sendResp(resp, buffer)
	runtime.Gosched()
	if done {
		t.Fatalf("It can stil send response to closed buffer")
	}
}

func TestSendItem(t *testing.T) {
	buffer, _ := buffer.NewPool(10, 2)
	if sendItem(nil, buffer) {
		t.Fatalf("It can still send nil item")
	}
	item := module.Item(map[string]interface{}{})
	buffer.Close()
	done := sendItem(item, buffer)
	runtime.Gosched()
	if done {
		t.Fatalf("It can still send item to closed buffer")
	}
}
