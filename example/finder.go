package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"
	lib "webcrawler/example/internal"
	"webcrawler/example/monitor"
	"webcrawler/helper/log"
	sched "webcrawler/scheduler"
)

var (
	firstURL string
	domains  string
	depth    uint
	dirPath  string
)

var logger = log.DLogger()

func init() {
	flag.StringVar(&firstURL, "first", "http://baidu.com",
		"The first URL which you want to access.")
	flag.StringVar(&domains, "domains", "baidu.com",
		"The primary domains which you accepted. "+
			"please using comma-separated multiple domains.")
	flag.UintVar(&depth, "depth", 3, "The depth for crawling.")
	flag.StringVar(&dirPath, "dir", "./pictures",
		"The path which you want to save the image files")
}

func Usage() {
	fmt.Fprintf(os.Stderr, "Usage of %s: \n", os.Args[0])
	fmt.Fprintf(os.Stderr, "\tfinder [flags] \n")
	fmt.Fprintf(os.Stderr, "Flags:\n")
	flag.PrintDefaults()
}

func main() {
	flag.Usage = Usage
	flag.Parse()
	scheduler := sched.NewScheduler()
	domainParts := strings.Split(domains, ",")
	acceptDomains := []string{}
	for _, domain := range domainParts {
		domain = strings.TrimSpace(domain)
		if domain != "" {
			acceptDomains = append(acceptDomains, domain)
		}
	}
	reqArgs := sched.RequestArgs{
		AcceptedDomains: acceptDomains,
		MaxDepth:        uint32(depth),
	}
	dataArgs := sched.DataArgs{
		ReqBufferCap:         50,
		ReqMaxBufferNumber:   1000,
		RespBufferCap:        50,
		RespMaxBufferNumber:  10,
		ItemBufferCap:        50,
		ItemMaxBufferNumber:  100,
		ErrorBufferCap:       50,
		ErrorMaxBufferNumber: 1,
	}
	downloaders, err := lib.GetDownloaders(1)
	if err != nil {
		logger.Fatalf("An error occurs when creating downloaders: %s", err)
	}
	analyzers, err := lib.GetAnalyzers(1)
	if err != nil {
		logger.Fatalf("An error occurs when creating analyzers: %s", err)
	}
	pipelines, err := lib.GetPipelines(1, dirPath)
	if err != nil {
		logger.Fatalf("An error occurs when creating pipelines: %s", err)
	}
	moduleArgs := sched.ModuleArgs{
		Downloaders: downloaders,
		Analyzers:   analyzers,
		Pipelines:   pipelines,
	}

	err = scheduler.Init(
		reqArgs,
		dataArgs,
		moduleArgs,
	)
	if err != nil {
		logger.Fatalf("An error occurs when initializing scheduler: %s", err)
	}
	checkInterval := time.Second
	summarizeInterval := 100 * time.Millisecond
	maxIdleCnt := uint(5)
	checkCountChan := monitor.Monitor(
		scheduler,
		checkInterval,
		summarizeInterval,
		maxIdleCnt,
		true,
		lib.Record,
	)
	firstHTTPReq, err := http.NewRequest("GET", firstURL, nil)
	if err != nil {
		logger.Fatalln(err)
		return
	}
	err = scheduler.Start(firstHTTPReq)
	if err != nil {
		logger.Fatalf("An error occurs when starting scheduler: %s", err)
	}
	<-checkCountChan
}
