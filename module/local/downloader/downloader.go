package downloader

import (
	"net/http"
	werr "webcrawler/errors"
	"webcrawler/helper/log"
	"webcrawler/module"
	"webcrawler/module/stub"
)

var logger = log.DLogger()

func New(mid module.MID, client *http.Client, scoreCalculator module.CalculateScore) (module.Downloader, error) {
	moduleBase, err := stub.NewModuleInternal(mid, scoreCalculator)
	if err != nil {
		return nil, err
	}
	if client == nil {
		return nil, genParameterError("nil http client")
	}
	return &myDownloader{
		ModuleInternal: moduleBase,
		httpClient:     *client,
	}, nil
}

type myDownloader struct {
	stub.ModuleInternal
	httpClient http.Client
}

func (downloader *myDownloader) Download(req *module.Request) (*module.Response, error) {
	downloader.IncrHandlingNumber()
	defer downloader.DecrHandlingNumber()
	downloader.IncrCalledCount()
	if req == nil {
		return nil, genParameterError("nil request")
	}
	httpReq := req.HTTPReq()
	if httpReq == nil {
		return nil, genParameterError("nil HTTP request")
	}
	downloader.IncrAcceptedCount()
	logger.Infof("Do the request (URL: %s, depth: %d)... \n", httpReq.URL, req.Depth())
	httpResp, err := downloader.httpClient.Do(httpReq)
	if err != nil {
		return nil, err
	}
	downloader.IncrCompletedCount()
	return module.NewResponse(httpResp, req.Depth()), nil
}

func genParameterError(errMsg string) error {
	return werr.NewCrawlerErrorBy(werr.ERROR_TYPE_DOWNLOADER, werr.NewIllegalParameterError(errMsg))
}

func genError(errMsg string) error {
	return werr.NewCrawlerError(werr.ERROR_TYPE_DOWNLOADER, errMsg)
}
