package pipeline

import (
	"fmt"
	werr "webcrawler/errors"
	"webcrawler/helper/log"
	"webcrawler/module"
	"webcrawler/module/stub"
)

var logger = log.DLogger()

func New(mid module.MID, itemProcessors []module.ProcessItem, scoreCalculator module.CalculateScore) (module.Pipeline, error) {
	moduleBase, err := stub.NewModuleInternal(mid, scoreCalculator)
	if err != nil {
		return nil, err
	}
	if itemProcessors == nil {
		return nil, genParameterError("nil item processor list")
	}
	if len(itemProcessors) == 0 {
		return nil, genParameterError("empty item processor list")
	}
	var innerProcessors []module.ProcessItem
	for i, pipeline := range itemProcessors {
		if pipeline == nil {
			err := genParameterError(fmt.Sprintf("nil item processor[%d]", i))
			return nil, err
		}
		innerProcessors = append(innerProcessors, pipeline)
	}
	return &myPipeline{
		ModuleInternal: moduleBase,
		itemProcessors: innerProcessors,
	}, nil
}

type myPipeline struct {
	stub.ModuleInternal
	itemProcessors []module.ProcessItem
	failFast       bool
}

func (pipeline *myPipeline) ItemProcessors() []module.ProcessItem {
	processors := make([]module.ProcessItem, len(pipeline.itemProcessors))
	copy(processors, pipeline.itemProcessors)
	return processors
}

func (pipeline *myPipeline) Send(item module.Item) []error {
	pipeline.IncrHandlingNumber()
	defer pipeline.DecrHandlingNumber()
	pipeline.IncrCalledCount()
	var errs []error
	if item == nil {
		err := genParameterError("nil item")
		errs = append(errs, err)
		return errs
	}
	pipeline.IncrAcceptedCount()
	logger.Infof("Process item %+v...\n", item)
	var currentItem = item
	for _, processor := range pipeline.itemProcessors {
		processedItem, err := processor(currentItem)
		if err != nil {
			errs = append(errs, err)
			if pipeline.failFast {
				break
			}
		}
		if processedItem != nil {
			currentItem = processedItem
		}
	}
	if len(errs) == 0 {
		pipeline.IncrCompletedCount()
	}
	return errs
}

func (pipeline *myPipeline) FailFast() bool {
	return pipeline.failFast
}

func (pipeline *myPipeline) SetFailFast(failFast bool) {
	pipeline.failFast = failFast
}

type extraSummaryStruct struct {
	FailFast        bool `json:"fail_fast"`
	ProcessorNumber int  `json:"processor_number"`
}

func (pipeline *myPipeline) Summary() module.SummaryStruct {
	summary := pipeline.ModuleInternal.Summary()
	summary.Extra = extraSummaryStruct{
		FailFast:        pipeline.failFast,
		ProcessorNumber: len(pipeline.itemProcessors),
	}
	return summary
}

func genParameterError(errMsg string) error {
	return werr.NewCrawlerErrorBy(werr.ERROR_TYPE_PIPELINE, werr.NewIllegalParameterError(errMsg))
}

func genError(errMsg string) error {
	return werr.NewCrawlerError(werr.ERROR_TYPE_PIPELINE, errMsg)
}
