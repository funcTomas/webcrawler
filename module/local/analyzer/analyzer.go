package analyzer

import (
	"fmt"
	werr "webcrawler/errors"
	"webcrawler/helper/log"
	"webcrawler/module"
	"webcrawler/module/stub"
	"webcrawler/module/toolkit/reader"
)

var logger = log.DLogger()

func New(mid module.MID, respParsers []module.ParseResponse, scoreCalculator module.CalculateScore) (module.Analyzer, error) {
	moduleBase, err := stub.NewModuleInternal(mid, scoreCalculator)
	if err != nil {
		return nil, err
	}
	if respParsers == nil {
		return nil, genParameterError("nil response parsers")
	}
	if len(respParsers) == 0 {
		return nil, genParameterError("empty response parsers")
	}
	var innerParsers []module.ParseResponse
	for i, parser := range respParsers {
		if parser == nil {
			return nil, genParameterError(fmt.Sprintf("nil response parser [%d]", i))
		}
		innerParsers = append(innerParsers, parser)
	}
	return &myAnalyzer{
		ModuleInternal: moduleBase,
		respParsers:    innerParsers,
	}, nil
}

type myAnalyzer struct {
	stub.ModuleInternal
	respParsers []module.ParseResponse
}

func (analyzer *myAnalyzer) RespParsers() []module.ParseResponse {
	parsers := make([]module.ParseResponse, len(analyzer.respParsers))
	copy(parsers, analyzer.respParsers)
	return parsers
}

func (analyzer *myAnalyzer) Analyze(resp *module.Response) (dataList []module.Data, errorList []error) {
	analyzer.IncrHandlingNumber()
	defer analyzer.DecrHandlingNumber()
	analyzer.IncrCalledCount()
	if resp == nil {
		errorList = append(errorList, genParameterError("nil response"))
		return
	}
	httpResp := resp.HTTPResp()
	if httpResp == nil {
		errorList = append(errorList, genParameterError("nil HTTP response"))
		return
	}
	httpReq := httpResp.Request
	if httpReq == nil {
		errorList = append(errorList, genParameterError("nil HTTP request"))
		return
	}
	var reqURL = httpReq.URL
	if reqURL == nil {
		errorList = append(errorList, genParameterError("nil HTTP request URL"))
		return
	}
	analyzer.IncrAcceptedCount()
	respDepth := resp.Depth()
	logger.Infof("Parse the reponse (URL: %s, depth: %d)...\n", reqURL, respDepth)
	originalRespBody := httpResp.Body
	if originalRespBody != nil {
		defer originalRespBody.Close()
	}
	multipleReader, err := reader.NewMultipleReader(originalRespBody)
	if err != nil {
		errorList = append(errorList, genError(err.Error()))
		return
	}
	dataList = []module.Data{}
	for _, respParser := range analyzer.respParsers {
		httpResp.Body = multipleReader.Reader()
		pDataList, pErrorList := respParser(httpResp, respDepth)
		for _, pData := range pDataList {
			if pData == nil {
				continue
			}
			dataList = appendDataList(dataList, pData, respDepth)
		}
		for _, pError := range pErrorList {
			if pError == nil {
				continue
			}
			errorList = append(errorList, pError)
		}
	}
	if len(errorList) == 0 {
		analyzer.IncrCompletedCount()
	}
	return dataList, errorList
}

func appendDataList(dataList []module.Data, data module.Data, respDepth uint32) []module.Data {
	if data == nil {
		return dataList
	}
	req, ok := data.(*module.Request)
	if !ok {
		return append(dataList, data)
	}
	newDepth := respDepth + 1
	if req.Depth() != newDepth {
		req = module.NewRequest(req.HTTPReq(), newDepth)
	}
	return append(dataList, req)
}

func genParameterError(errMsg string) error {
	return werr.NewCrawlerErrorBy(werr.ERROR_TYPE_ANALYZER, werr.NewIllegalParameterError(errMsg))
}

func genError(errMsg string) error {
	return werr.NewCrawlerError(werr.ERROR_TYPE_ANALYZER, errMsg)
}
