package scheduler

import "webcrawler/module"

type Args interface {
	Check() error
}

type RequestArgs struct {
	AcceptedDomains []string `json:"accepted_primary_domains"`
	MaxDepth        uint32   `json:"max_depth"`
}

func (args *RequestArgs) Check() error {
	if args.AcceptedDomains == nil {
		return genError("nil accepted primary domain list")
	}
	return nil
}

func (args *RequestArgs) Same(another *RequestArgs) bool {
	if another == nil {
		return false
	}
	if another.MaxDepth != args.MaxDepth {
		return false
	}
	anotherDomains := another.AcceptedDomains
	if len(anotherDomains) != len(args.AcceptedDomains) {
		return false
	}
	for i, domain := range anotherDomains {
		if domain != args.AcceptedDomains[i] {
			return false
		}
	}
	return true
}

type DataArgs struct {
	ReqBufferCap         uint32 `json:"req_buffer_cap"`
	ReqMaxBufferNumber   uint32 `json:"req_max_buffer_number"`
	RespBufferCap        uint32 `json:"resp_buffer_cap"`
	RespMaxBufferNumber  uint32 `json:"resp_max_buffer_number"`
	ItemBufferCap        uint32 `json:"item_buffer_cap"`
	ItemMaxBufferNumber  uint32 `json:"item_max_buffer_number"`
	ErrorBufferCap       uint32 `json:"error_buffer_cap"`
	ErrorMaxBufferNumber uint32 `json:"error_max_buffer_number"`
}

func (args *DataArgs) Check() error {
	if args.ReqBufferCap == 0 {
		return genError("zero request buffer cap")
	}
	if args.ReqMaxBufferNumber == 0 {
		return genError("zero request max buffer number")
	}
	if args.RespBufferCap == 0 {
		return genError("zero responce buffer cap")
	}
	if args.RespMaxBufferNumber == 0 {
		return genError("zero response max buffer number")
	}
	if args.ItemBufferCap == 0 {
		return genError("zero item buffer cap")
	}
	if args.ItemMaxBufferNumber == 0 {
		return genError("zero item max buffer number")
	}
	if args.ErrorBufferCap == 0 {
		return genError("zero error buffer cap")
	}
	if args.ErrorMaxBufferNumber == 0 {
		return genError("zero error max buffer number")
	}
	return nil
}

type ModuleArgsSummary struct {
	DownloaderListSize int `json:"downloader_list_size"`
	AnalyzerListSize   int `json:"analyzer_list_size"`
	PipelineListSize   int `json:"pipeline_list_size"`
}

type ModuleArgs struct {
	Downloaders []module.Downloader
	Analyzers   []module.Analyzer
	Pipelines   []module.Pipleline
}

func (args *ModuleArgs) Check() error {
	if len(args.Downloaders) == 0 {
		return genError("empty downloaders list")
	}
	if len(args.Analyzers) == 0 {
		return genError("empty analyzers list")
	}
	if len(args.Pipelines) == 0 {
		return genError("empty pipelines list")
	}
	return nil
}

func (args *ModuleArgs) Summary() ModuleArgsSummary {
	return ModuleArgsSummary{
		DownloaderListSize: len(args.Downloaders),
		AnalyzerListSize:   len(args.Analyzers),
		PipelineListSize:   len(args.Pipelines),
	}
}
