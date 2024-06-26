package pipeline

import (
	"errors"
	"fmt"
	"testing"
	"webcrawler/module"
	"webcrawler/module/stub"
)

func TestNew(t *testing.T) {
	mid := module.MID("P1|127.0.0.1:8080")
	processorNumber := 10
	processors := make([]module.ProcessItem, processorNumber)
	for i := 0; i < processorNumber; i++ {
		processors[i] = genTestingItemProcessor(false)
	}
	p, err := New(mid, processors, nil)
	if err != nil {
		t.Fatalf("An error occurs when creating a pipeline: %s (mid: %s, processors: %#v)", err, mid, processors)
	}
	if p == nil {
		t.Fatalf("Could not create pipeline")
	}
	if p.ID() != mid {
		t.Fatalf("Inconsistent MID for pipeline, expected: %s, actual: %s", mid, p.ID())
	}
	if len(p.ItemProcessors()) != len(processors) {
		t.Fatalf("Inconsistent item processor number for pipeline, expected: %d, actual: %d", len(p.ItemProcessors()), len(processors))
	}
	mid = module.MID("P1|127.0.0.1:8080")
	processors = append(processors, nil)
	_, err = New(mid, processors, nil)
	if err == nil {
		t.Fatalf("No error when create a pipeline with nil processor")
	}
	processorList := [][]module.ProcessItem{
		nil,
		{},
		{genTestingItemProcessor(false), nil},
	}
	for _, processors := range processorList {
		_, err = New(mid, processors, nil)
		if err == nil {
			t.Fatalf("No error when create a pipeline with nil processors")
		}
	}
}

func TestSend(t *testing.T) {
	mid := module.MID("P1|127.0.0.1:8080")
	processorNumber := 12
	processors := make([]module.ProcessItem, processorNumber)
	var expectedErrs int
	for i := 0; i < processorNumber; i++ {
		processors[i] = genTestingItemProcessor(false)
	}
	p, err := New(mid, processors, nil)
	if err != nil {
		t.Fatalf("An error occurs when creating a pipeline, %s (mid: %s, processors: %#v)", err, mid, processors)
	}
	errs := p.Send(nil)
	if len(errs) != 1 {
		t.Fatalf("Inconsistent error number after Send(), expected: 1, actual: %d", len(errs))
	}
	item := module.Item(map[string]interface{}{"number": 0})
	errs = p.Send(item)
	number := item["number"].(int)
	if number != processorNumber {
		t.Fatalf("Inconsistent number in item after Send(), expected: %d, actual: %d", 1, len(errs))
	}
	if len(errs) != expectedErrs {
		t.Fatalf("Inconsistent error number after Send(), expected: %d, actual: %d", expectedErrs, len(errs))
	}
	expectedErrs = 0
	for i := 0; i < processorNumber; i++ {
		if i%3 == 0 {
			processors[i] = genTestingItemProcessor(true)
			expectedErrs++
		} else {
			processors[i] = genTestingItemProcessor(false)
		}
	}
	p, err = New(mid, processors, nil)
	if err != nil {
		t.Fatalf("An error occurs when creating a pipeline: %s (mid: %s, processors: %#v)", err, mid, processors)
	}
	item = module.Item(map[string]interface{}{"number": 0})
	errs = p.Send(item)
	if len(errs) != expectedErrs {
		t.Fatalf("Inconsistent error number after Send(), expected: %d, actual: %d", expectedErrs, errs)
	}
	p.SetFailFast(true)
	errs = p.Send(item)
	if len(errs) != 1 {
		t.Fatalf("Inconsistent error number after Send(), expected: 1, actual: %d", len(errs))
	}
	p.SetFailFast(false)
	errs = p.Send(item)
	if len(errs) != expectedErrs {
		t.Fatalf("Inconsistent error number after Send(), expected: %d, actual: %d", expectedErrs, len(errs))
	}
}

func TestFailFast(t *testing.T) {
	mid := module.MID("P1|127.0.0.1:8080")
	processors := []module.ProcessItem{genTestingItemProcessor(false)}
	p, err := New(mid, processors, nil)
	if err != nil {
		t.Fatalf("An error occurs when creating a pipeline: %s (mid: %s, processors: %#v)", err, mid, processors)
	}
	if p.FailFast() != false {
		t.Fatalf("Inconsistent fail fast sign for pipeline, expected: %v, actual: %v", false, p.FailFast())
	}
	p.SetFailFast(true)
	if p.FailFast() != true {
		t.Fatalf("Inconsistent fail fast sign for pipeline, expected: %v, actual: %v", true, p.FailFast())
	}
}

func TestCount(t *testing.T) {
	mid := module.MID("P1|127.0.0.1:8080")
	processors := []module.ProcessItem{genTestingItemProcessor(false)}
	p, err := New(mid, processors, nil)
	if err != nil {
		t.Fatalf("An error occurs when creating a pipeline: %s (mid: %s, processors: %#v)", err, mid, processors)
	}
	pi, ok := p.(stub.ModuleInternal)
	if !ok {
		t.Fatalf("Could not convert the type of pipeline instance to stub.ModuleInternal")
	}
	if pi.CalledCount() != 0 {
		t.Fatalf("Inconsistent called count for internal module, expected: %d, actual: %d", 0, pi.CalledCount())
	}
	if pi.AcceptedCount() != 0 {
		t.Fatalf("Inconsistent accepted count for internal module, expected: %d, actual: %d", 0, pi.AcceptedCount())
	}
	if pi.CompletedCount() != 0 {
		t.Fatalf("Inconsistent completed count for internal module, expected: %d, actual: %d", 0, pi.CompletedCount())
	}
	if pi.HandlingNumber() != 0 {
		t.Fatalf("Inconsistent handling number for internal module, expected: %d, actual: %d", 0, pi.HandlingNumber())
	}

	processors = []module.ProcessItem{genTestingItemProcessor(true)}
	p, err = New(mid, processors, nil)
	if err != nil {
		t.Fatalf("An error occurs when creating a pipeline: %s (mid: %s, processors: %#v)", err, mid, processors)
	}
	pi, ok = p.(stub.ModuleInternal)
	if !ok {
		t.Fatalf("Could not convert the type of pipeline instance to stub.ModuleInternal")
	}
	item := module.Item(map[string]interface{}{"number": 0})
	p.Send(item)
	if pi.CalledCount() != 1 {
		t.Fatalf("Inconsistent called count for internal module, expected: %d, actual: %d", 1, pi.CalledCount())
	}
	if pi.AcceptedCount() != 1 {
		t.Fatalf("Inconsistent accepted count for internal module, expected: %d, actual: %d", 1, pi.AcceptedCount())
	}
	if pi.CompletedCount() != 0 {
		t.Fatalf("Inconsistent completed count for internal module, expected: %d, actual: %d", 0, pi.CompletedCount())
	}
	if pi.HandlingNumber() != 0 {
		t.Fatalf("Inconsistent handling number for internal module, expected: %d, actual: %d", 0, pi.HandlingNumber())
	}

	processors = []module.ProcessItem{genTestingItemProcessor(false)}
	p, err = New(mid, processors, nil)
	if err != nil {
		t.Fatalf("An error occurs when creating a pipeline: %s (mid: %s, processors: %#v)", err, mid, processors)
	}
	pi, ok = p.(stub.ModuleInternal)
	if !ok {
		t.Fatalf("Could not convert the type of pipeline instance to stub.ModuleInternal")
	}
	p.Send(nil)
	if pi.CalledCount() != 1 {
		t.Fatalf("Inconsistent called count for internal module, expected: %d, actual: %d", 1, pi.CalledCount())
	}
	if pi.AcceptedCount() != 0 {
		t.Fatalf("Inconsistent accepted count for internal module, expected: %d, actual: %d", 0, pi.AcceptedCount())
	}
	if pi.CompletedCount() != 0 {
		t.Fatalf("Inconsistent completed count for internal module, expected: %d, actual: %d", 0, pi.CompletedCount())
	}
	if pi.HandlingNumber() != 0 {
		t.Fatalf("Inconsistent handling number for internal module, expected: %d, actual: %d", 0, pi.HandlingNumber())
	}

	processors = []module.ProcessItem{genTestingItemProcessor(false)}
	p, err = New(mid, processors, nil)
	if err != nil {
		t.Fatalf("An error occurs when creating a pipeline: %s (mid: %s, processors: %#v)", err, mid, processors)
	}
	pi, ok = p.(stub.ModuleInternal)
	if !ok {
		t.Fatalf("Could not convert the type of pipeline instance to stub.ModuleInternal")
	}
	item = module.Item(map[string]interface{}{"number": 0})
	p.Send(item)
	if pi.CalledCount() != 1 {
		t.Fatalf("Inconsistent called count for internal module, expected: %d, actual: %d", 1, pi.CalledCount())
	}
	if pi.AcceptedCount() != 1 {
		t.Fatalf("Inconsistent accepted count for internal module, expected: %d, actual: %d", 1, pi.AcceptedCount())
	}
	if pi.CompletedCount() != 1 {
		t.Fatalf("Inconsistent completed count for internal module, expected: %d, actual: %d", 1, pi.CompletedCount())
	}
	if pi.HandlingNumber() != 0 {
		t.Fatalf("Inconsistent handling number for internal module, expected: %d, actual: %d", 0, pi.HandlingNumber())
	}
}

func TestSummary(t *testing.T) {
	mid := module.MID("P1|127.0.0.1:8080")
	processors := []module.ProcessItem{genTestingItemProcessor(false)}
	p, err := New(mid, processors, nil)
	if err != nil {
		t.Fatalf("An error occurs when creating a pipeline: %s (mid: %s, processors: %#v)", err, mid, processors)
	}
	pi, ok := p.(*myPipeline)
	if !ok {
		t.Fatalf("Could not convert a pipeline instance to myPipeline")
	}
	number := uint64(10000)
	for i := uint64(1); i < number; i++ {
		pi.IncrCalledCount()
		pi.IncrAcceptedCount()
		pi.IncrCompletedCount()
		pi.IncrHandlingNumber()
		if i%17 == 0 {
			pi.Clear()
		}
		counts := pi.Counts()
		expectedSummary := module.SummaryStruct{
			ID:        pi.ID(),
			Called:    counts.CalledCount,
			Accepted:  counts.AcceptedCount,
			Completed: counts.CompletedCount,
			Handling:  counts.HandlingNumber,
			Extra: extraSummaryStruct{
				FailFast:        pi.FailFast(),
				ProcessorNumber: len(pi.itemProcessors),
			},
		}
		summary := pi.Summary()
		if summary != expectedSummary {
			t.Fatalf("Inconsistent summary for internal module, expected: %#v, actual: %#v", expectedSummary, summary)
		}
	}
}

func genTestingItemProcessor(fail bool) module.ProcessItem {
	if fail {
		return func(item module.Item) (result module.Item, err error) {
			return nil, fmt.Errorf("Fail! (item: %#v)", item)
		}
	}
	return func(item module.Item) (result module.Item, err error) {
		num, ok := item["number"]
		if !ok {
			return nil, errors.New("not found the number")
		}
		numInt, ok := num.(int)
		if !ok {
			return nil, fmt.Errorf("non-integer number %v", num)
		}
		item["number"] = numInt + 1
		return item, nil
	}
}
