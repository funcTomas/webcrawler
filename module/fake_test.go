package module

import "sync/atomic"

var defaultFakeDownloader = NewFakeDownloader(MID("D0"), CalculateScoreSimple)

var defaultFakeAnalyzer = NewFakeAnalyzer(MID("A1"), CalculateScoreSimple)

var defaultFakePipeline = NewFakePipeline(MID("P2"), CalculateScoreSimple)

var fakeModules = []Module{
	defaultFakeDownloader,
	defaultFakeAnalyzer,
	defaultFakePipeline,
}

var defaultFakeModuleMap = map[Type]Module{
	TYPE_DOWNLOADER: defaultFakeDownloader,
	TYPE_ANALYZER:   defaultFakeAnalyzer,
	TYPE_PIPELINE:   defaultFakePipeline,
}

var fakeModuleFuncMap = map[Type]func(mid MID) Module{
	TYPE_DOWNLOADER: func(mid MID) Module {
		return NewFakeDownloader(mid, CalculateScoreSimple)
	},
	TYPE_ANALYZER: func(mid MID) Module {
		return NewFakeAnalyzer(mid, CalculateScoreSimple)
	},
	TYPE_PIPELINE: func(mid MID) Module {
		return NewFakePipeline(mid, CalculateScoreSimple)
	},
}

type fakeModule struct {
	mid             MID
	score           uint64
	count           uint64
	scoreCalculator CalculateScore
}

func (fm *fakeModule) ID() MID {
	return fm.mid
}

func (fm *fakeModule) Addr() string {
	parts, err := SplitMID(fm.mid)
	if err == nil {
		return parts[2]
	}
	return ""
}

func (fm *fakeModule) Score() uint64 {
	return atomic.LoadUint64(&fm.score)
}

func (fm *fakeModule) SetScore(score uint64) {
	atomic.StoreUint64(&fm.score, score)
}

func (fm *fakeModule) ScoreCalculator() CalculateScore {
	return fm.scoreCalculator
}

func (fm *fakeModule) CalledCount() uint64 {
	return fm.count + 10
}

func (fm *fakeModule) AcceptedCount() uint64 {
	return fm.count + 8
}
func (fm *fakeModule) CompletedCount() uint64 {
	return fm.count + 6
}

func (fm *fakeModule) HandlingNumber() uint64 {
	return fm.count + 2
}

func (fm *fakeModule) Counts() Counts {
	return Counts{
		fm.CalledCount(),
		fm.AcceptedCount(),
		fm.CompletedCount(),
		fm.HandlingNumber(),
	}
}

func (fm *fakeModule) Summary() SummaryStruct {
	return SummaryStruct{}
}

func NewFakeAnalyzer(mid MID, scoreCalculator CalculateScore) Analyzer {
	return &fakeAnalyzer{
		fakeModule: fakeModule{
			mid:             mid,
			scoreCalculator: scoreCalculator,
		},
	}
}

type fakeAnalyzer struct {
	fakeModule
}

func (analyzer *fakeAnalyzer) RespParsers() []ParseResponse {
	return nil
}

func (analyzer *fakeAnalyzer) Analyze(resp *Response) (dataList []Data, errorList []error) {
	return
}

type fakeDownloader struct {
	fakeModule
}

func NewFakeDownloader(mid MID, scoreCalculator CalculateScore) Downloader {
	return &fakeDownloader{
		fakeModule: fakeModule{
			mid:             mid,
			scoreCalculator: scoreCalculator,
		},
	}
}

func (downloader *fakeDownloader) Download(req *Request) (*Response, error) {
	return nil, nil
}

type fakePipeline struct {
	fakeModule
	failFast bool
}

func NewFakePipeline(mid MID, scoreCalculator CalculateScore) Pipleline {
	return &fakePipeline{
		fakeModule: fakeModule{
			mid:             mid,
			scoreCalculator: scoreCalculator,
		},
	}
}

func (pipeline *fakePipeline) ItemProcessors() []ProcessItem {
	return nil
}

func (pipleline *fakePipeline) Send(item Item) []error {
	return nil
}

func (pipleline *fakePipeline) FailFast() bool {
	return pipleline.failFast
}

func (pipeline *fakePipeline) SetFailFast(failFast bool) {
	pipeline.failFast = failFast
}
