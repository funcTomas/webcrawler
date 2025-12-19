package internal

import (
	"webcrawler/module"
	"webcrawler/module/local/analyzer"
	"webcrawler/module/local/downloader"
	"webcrawler/module/local/pipeline"
)

var snGen = module.NewSNGenerator(1, 0)

func GetDownloaders(number uint8) ([]module.Downloader, error) {
	downloaders := []module.Downloader{}
	if number == 0 {
		return downloaders, nil
	}
	for range number {
		mid, err := module.GenMID(module.TYPE_DOWNLOADER, snGen.Get(), nil)
		if err != nil {
			return downloaders, err
		}
		d, err := downloader.New(mid, genHTTPClient(), module.CalculateScoreSimple)
		if err != nil {
			return downloaders, err
		}
		downloaders = append(downloaders, d)

	}
	return downloaders, nil
}

func GetAnalyzers(number uint8) ([]module.Analyzer, error) {
	analyzers := []module.Analyzer{}
	if number == 0 {
		return analyzers, nil
	}
	for range number {
		mid, err := module.GenMID(module.TYPE_ANALYZER, snGen.Get(), nil)
		if err != nil {
			return analyzers, nil
		}
		a, err := analyzer.New(mid, genResponseParsersV2(), module.CalculateScoreSimple)
		if err != nil {
			return analyzers, err
		}
		analyzers = append(analyzers, a)
	}
	return analyzers, nil
}

func GetPipelines(number uint8, dirPath string) ([]module.Pipeline, error) {
	if number == 0 {
		return nil, nil
	}
	pipelines := []module.Pipeline{}
	for range number {
		mid, err := module.GenMID(module.TYPE_PIPELINE, snGen.Get(), nil)
		if err != nil {
			return pipelines, err
		}
		a, err := pipeline.New(mid, genItemProcessorsV2(dirPath), module.CalculateScoreSimple)
		if err != nil {
			return pipelines, err
		}
		a.SetFailFast(true)
		pipelines = append(pipelines, a)
	}
	return pipelines, nil
}
