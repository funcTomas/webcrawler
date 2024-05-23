package scheduler

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"
	"webcrawler/module"
	"webcrawler/module/local/analyzer"
	"webcrawler/module/local/downloader"
	"webcrawler/module/local/pipeline"

	"github.com/PuerkitoBio/goquery"
)

func TestArgsRequest(t *testing.T) {
	reqArgs := RequestArgs{
		AcceptedDomains: []string{},
		MaxDepth:        0,
	}
	if err := reqArgs.Check(); err != nil {
		t.Fatalf("Inconsistent check result, expected: %v, actual: %v", nil, err)
	}
	reqArgs = RequestArgs{
		AcceptedDomains: nil,
		MaxDepth:        0,
	}
	if err := reqArgs.Check(); err == nil {
		t.Fatalf("Inconsistent check result, expected: %v, actual: %v", err, nil)
	}
	one := RequestArgs{
		AcceptedDomains: []string{"bing.com"},
		MaxDepth:        0,
	}
	another := RequestArgs{
		AcceptedDomains: []string{"bing.com"},
		MaxDepth:        0,
	}
	same := one.Same(&another)
	if !same {
		t.Fatalf("Inconsistent request arguments sameness, expected: %v, actual: %v", true, same)
	}
	same = one.Same(nil)
	if same {
		t.Fatalf("Inconsistent request arguments sameness, expected: %v, actual: %v", false, same)
	}
	another = RequestArgs{
		AcceptedDomains: []string{"bing.net", "bing.com"},
		MaxDepth:        0,
	}
	same = one.Same(&another)
	if same {
		t.Fatalf("Inconsistent request arguments sameness, expected: %v, actual: %v", false, same)
	}
	one = RequestArgs{
		AcceptedDomains: []string{"sogou.com", "bing.com"},
		MaxDepth:        0,
	}
	same = one.Same(&another)
	if same {
		t.Fatalf("Inconsistent request arguments sameness, expected: %v, actual: %v", false, same)
	}
}

func TestArgsData(t *testing.T) {
	args := genDataArgs(10, 2, 1)
	if err := args.Check(); err != nil {
		t.Fatalf("Inconsistent check result, expected: %v, actual: %v", nil, err)
	}
	dataArgsList := []DataArgs{}
	for i := 0; i < 8; i++ {
		values := [8]uint32{2, 2, 2, 2, 2, 2, 2, 2}
		values[i] = 0
		dataArgsList = append(dataArgsList, DataArgs{
			ReqBufferCap:         values[0],
			ReqMaxBufferNumber:   values[1],
			RespBufferCap:        values[2],
			RespMaxBufferNumber:  values[3],
			ItemBufferCap:        values[4],
			ItemMaxBufferNumber:  values[5],
			ErrorBufferCap:       values[6],
			ErrorMaxBufferNumber: values[7],
		})
	}
	for _, dataArgs := range dataArgsList {
		if err := dataArgs.Check(); err == nil {
			t.Fatalf("No error when check data arguments (dataArgs: %#v)", dataArgs)
		}
	}
}

func genRequestArgs(acceptedDomains []string, maxDepth uint32) RequestArgs {
	return RequestArgs{
		AcceptedDomains: acceptedDomains,
		MaxDepth:        maxDepth,
	}
}

func genDataArgs(bufferCap uint32, maxBufferNumber uint32, stepLen uint32) DataArgs {
	values := [8]uint32{}
	var bufferCapStep uint32
	var maxBufferNumberStep uint32
	for i := uint32(0); i < 8; i++ {
		if i%2 == 0 {
			values[i] = bufferCap + bufferCapStep*stepLen
			bufferCapStep++
		} else {
			values[i] = maxBufferNumber + maxBufferNumberStep*stepLen
			maxBufferNumberStep++
		}
	}
	return DataArgs{
		ReqBufferCap:         values[0],
		ReqMaxBufferNumber:   values[1],
		RespBufferCap:        values[2],
		RespMaxBufferNumber:  values[3],
		ItemBufferCap:        values[4],
		ItemMaxBufferNumber:  values[5],
		ErrorBufferCap:       values[6],
		ErrorMaxBufferNumber: values[7],
	}
}

func TestArgsModule(t *testing.T) {
	moduleArgs := genSimpleModuleArgs(3, 2, 1, t)
	if err := moduleArgs.Check(); err != nil {
		t.Fatalf("Inconsistent check result, expected: %v, actual: %v", nil, err)
	}
	expectedSummary := ModuleArgsSummary{
		DownloaderListSize: 3,
		AnalyzerListSize:   2,
		PipelineListSize:   1,
	}
	summary := moduleArgs.Summary()
	if summary != expectedSummary {
		t.Fatalf("Inconsistent module args summary, expected: %#v, actual: %#v", expectedSummary, summary)
	}
	moduleArgsList := []ModuleArgs{
		genSimpleModuleArgs(0, 2, 1, t),
		genSimpleModuleArgs(3, 0, 1, t),
		genSimpleModuleArgs(3, 2, 0, t),
		{},
	}
	for _, moduleArgs := range moduleArgsList {
		if err := moduleArgs.Check(); err == nil {
			t.Fatalf("No error when check module arguments (moduleArgs: %#v)", moduleArgs)
		}
	}
}

func genSimpleModuleArgs(downloaderNumber int8, analyzerNumber int8, pipelineNumber int8, t *testing.T) ModuleArgs {
	snGen := module.NewSNGenerator(1, 0)
	return ModuleArgs{
		Downloaders: genSimpleDownloaders(downloaderNumber, false, snGen, t),
		Analyzers:   genSimpleAnalyzers(analyzerNumber, false, snGen, t),
		Pipelines:   getSimplePipelines(pipelineNumber, false, snGen, t),
	}
}

func genSimpleDownloaders(number int8, reuseMID bool, snGen module.SNGenerator, t *testing.T) []module.Downloader {
	if number < -1 {
		return []module.Downloader{}
	} else if number == -1 {
		mid := module.MID(fmt.Sprintf("A%d", snGen.Get()))
		httpClient := &http.Client{}
		d, err := downloader.New(mid, httpClient, nil)
		if err != nil {
			t.Fatalf("An error occurs when creating a downloader: %s (mid: %s, httpClient: %#v)", err, mid, httpClient)
		}
		return []module.Downloader{d}
	}
	results := make([]module.Downloader, number)
	var mid module.MID
	for i := int8(0); i < number; i++ {
		if i == 0 || !reuseMID {
			mid = module.MID(fmt.Sprintf("D%d", snGen.Get()))
		}
		httpClient := &http.Client{}
		d, err := downloader.New(mid, httpClient, nil)
		if err != nil {
			t.Fatalf("An error occurs when creating a downloader: %s (mid: %s, httpClient: %#v)", err, mid, httpClient)
		}
		results[i] = d
	}
	return results
}

func genSimpleAnalyzers(number int8, reuseMID bool, snGen module.SNGenerator, t *testing.T) []module.Analyzer {
	respParsers := []module.ParseResponse{parseATag}
	if number < -1 {
		return []module.Analyzer{}
	} else if number == -1 {
		mid := module.MID(fmt.Sprintf("P%d", snGen.Get()))
		a, err := analyzer.New(mid, respParsers, nil)
		if err != nil {
			t.Fatalf("An error occurs when creating an analyzer: %s (mid: %s, respParses: %#v)", err, mid, respParsers)
		}
		return []module.Analyzer{a}
	}
	results := make([]module.Analyzer, number)
	var mid module.MID
	for i := int8(0); i < number; i++ {
		if i == 0 || !reuseMID {
			mid = module.MID(fmt.Sprintf("A%d", snGen.Get()))
		}
		a, err := analyzer.New(mid, respParsers, nil)
		if err != nil {
			t.Fatalf("An error occurs when creating an analyzer: %s (mid: %s, respParsers: %#v)", err, mid, respParsers)
		}
		results[i] = a
	}
	return results
}

func parseATag(httpResp *http.Response, respDepth uint32) ([]module.Data, []error) {
	if httpResp.StatusCode != 200 {
		err := fmt.Errorf(fmt.Sprintf("Unsupported status code %d (httpResponse: %v)", httpResp.StatusCode, httpResp))
		return nil, []error{err}
	}
	reqURL := httpResp.Request.URL
	httpRespBody := httpResp.Body
	defer func() {
		if httpRespBody != nil {
			httpRespBody.Close()
		}
	}()
	var dataList []module.Data
	var errs []error
	doc, err := goquery.NewDocumentFromReader(httpRespBody)
	if err != nil {
		errs = append(errs, err)
		return dataList, errs
	}
	defer httpRespBody.Close()
	doc.Find("a").Each(func(index int, sel *goquery.Selection) {
		href, exists := sel.Attr("href")
		if !exists || href == "" || href == "#" || href == "/" {
			return
		}
		href = strings.TrimSpace(href)
		lowerHref := strings.ToLower(href)
		if href != "" && !strings.HasPrefix(lowerHref, "javascript") {
			aURL, err := url.Parse(href)
			if err != nil {
				logger.Warnf("An error occurs when parsing attribute %q in tag %q: %s (href: %s)", err, "href", "a", href)
				return
			}
			if !aURL.IsAbs() {
				aURL = reqURL.ResolveReference(aURL)
			}
			httpReq, err := http.NewRequest("GET", aURL.String(), nil)
			if err != nil {
				errs = append(errs, err)
			} else {
				req := module.NewRequest(httpReq, respDepth)
				dataList = append(dataList, req)
			}
		}
		text := strings.TrimSpace(sel.Text())
		var id, name string
		if v, ok := sel.Attr("id"); ok {
			id = strings.TrimSpace(v)
		}
		if v, ok := sel.Attr("name"); ok {
			name = strings.TrimSpace(v)
		}
		m := make(map[string]interface{})
		m["a.parent"] = reqURL
		m["a.id"] = id
		m["a.name"] = name
		m["a.text"] = text
		m["a.index"] = index
		item := module.Item(m)
		dataList = append(dataList, item)
		logger.Infof("Processed item: %v", m)
	})
	return dataList, errs
}

func getSimplePipelines(number int8, reuseMID bool, snGen module.SNGenerator, t *testing.T) []module.Pipleline {
	processors := []module.ProcessItem{processItem}
	if number < -1 {
		return []module.Pipleline{}
	} else if number == -1 {
		mid := module.MID(fmt.Sprintf("D%d", snGen.Get()))
		p, err := pipeline.New(mid, processors, nil)
		if err != nil {
			t.Fatalf("An error occurs when creating a pipeline: %s (mid: %s, processors: %#v)", err, mid, processors)
		}
		return []module.Pipleline{p}
	}
	results := make([]module.Pipleline, number)
	var mid module.MID
	for i := int8(0); i < number; i++ {
		if i == 0 || !reuseMID {
			mid = module.MID(fmt.Sprintf("P%d", snGen.Get()))
		}
		p, err := pipeline.New(mid, processors, nil)
		if err != nil {
			t.Fatalf("An error occurs when creating a pipeline: %s (mid: %s, processors: %#v)", err, mid, processors)
		}
		results[i] = p
	}
	return results
}

func processItem(item module.Item) (result module.Item, err error) {
	if item == nil {
		return nil, errors.New("Invalid item")
	}
	result = make(map[string]interface{})
	for k, v := range item {
		result[k] = v
	}
	if _, ok := result["number"]; !ok {
		result["number"] = len(result)
	}
	time.Sleep(10 * time.Millisecond)
	return result, nil

}
