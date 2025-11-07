package monitor

import (
	"context"
	"encoding/json"
	"fmt"
	"runtime"
	"time"
	"webcrawler/helper/log"
	sched "webcrawler/scheduler"
)

var logger = log.DLogger()

type summary struct {
	NumGoroutine int                 `json:"goroutine_number"`
	SchedSummary sched.SummaryStruct `json:"sched_summary"`
	EscapedTime  string              `json:"escaped_time"`
}

var msgReachMaxIdleCount = "The scheduler has been idle for a period of time" +
	" (about %s)." + " Consider to stop it now."

var msgStopScheduler = "Stop scheduler... %s."

type Record func(level uint8, content string)

func Monitor(
	scheduler sched.Scheduler,
	checkInterval time.Duration,
	summarizeInterval time.Duration,
	maxIdleCount uint,
	autoStop bool,
	record Record) <-chan uint64 {
	if scheduler == nil {
		panic("The scheduler is invalid")
	}
	if checkInterval < time.Millisecond*100 {
		checkInterval = time.Millisecond * 100
	}
	if summarizeInterval < time.Second {
		summarizeInterval = time.Second
	}
	if maxIdleCount < 10 {
		maxIdleCount = 10
	}
	logger.Infof("Monitor parameters : checkInterval: %s,"+
		" summarizeInterval: %s, maxIdleCount: %d, autoStop: %t",
		checkInterval, summarizeInterval, maxIdleCount, autoStop)
	stopNotifier, stopFunc := context.WithCancel(context.Background())
	reportError(scheduler, record, stopNotifier)
	recordSummary(scheduler, summarizeInterval, record, stopNotifier)
	checkCountChan := make(chan uint64, 2)
	checkStatus(scheduler,
		checkInterval,
		maxIdleCount,
		autoStop,
		checkCountChan,
		record,
		stopFunc,
	)
	return checkCountChan
}

func checkStatus(
	scheduler sched.Scheduler,
	checkInterval time.Duration,
	maxIdleCnt uint,
	autoStop bool,
	checkCountChan chan<- uint64,
	record Record,
	stopFunc context.CancelFunc,
) {
	go func() {
		var checkCount uint64
		defer func() {
			stopFunc()
			checkCountChan <- checkCount
		}()
		waitForSchedulerStart(scheduler)
		var idleCount uint
		var firstIdleTime time.Time
		for {
			if scheduler.Idle() {
				idleCount++
				if idleCount == 1 {
					firstIdleTime = time.Now()
				}
				if idleCount >= maxIdleCnt {
					msg := fmt.Sprintf(msgReachMaxIdleCount, time.Since(firstIdleTime).String())
					record(0, msg)
					if scheduler.Idle() {
						if autoStop {
							var result string
							if err := scheduler.Stop(); err == nil {
								result = "success"
							} else {
								result = fmt.Sprintf("failing(%s)", err)
							}
							msg = fmt.Sprintf(msgStopScheduler, result)
							record(0, msg)
						}
						break
					} else {
						if idleCount > 0 {
							idleCount = 0
						}
					}
				}
			} else {
				if idleCount > 0 {
					idleCount = 0
				}
			}
			checkCount++
			time.Sleep(checkInterval)
		}
	}()
}

func recordSummary(
	scheduler sched.Scheduler,
	summarizeInterval time.Duration,
	record Record,
	stopNotifier context.Context) {
	go func() {
		waitForSchedulerStart(scheduler)
		var (
			prevSchedSummaryStruct sched.SummaryStruct
			prevNumGoroutine       int
			recordCnt              uint64 = 1
		)
		startTime := time.Now()
		for {
			select {
			case <-stopNotifier.Done():
				return
			default:
			}
			currNumGoroutine := runtime.NumGoroutine()
			currSchedSummaryStruct := scheduler.Summary().Struct()
			if currNumGoroutine != prevNumGoroutine ||
				!currSchedSummaryStruct.Same(prevSchedSummaryStruct) {
				summary := summary{
					NumGoroutine: runtime.NumGoroutine(),
					SchedSummary: currSchedSummaryStruct,
					EscapedTime:  time.Since(startTime).String(),
				}
				b, err := json.MarshalIndent(summary, "", "    ")
				if err != nil {
					logger.Errorf("An error occurs when generatiing scheduler summary %s\n", err)
					continue
				}
				msg := fmt.Sprintf("Monitor summary[%d]:\n%s", recordCnt, b)
				record(0, msg)
				prevNumGoroutine = currNumGoroutine
				prevSchedSummaryStruct = currSchedSummaryStruct
				recordCnt++
			}
			time.Sleep(summarizeInterval)
		}
	}()
}

func reportError(
	scheduler sched.Scheduler,
	record Record,
	stopNotifier context.Context) {
	go func() {
		waitForSchedulerStart(scheduler)
		errChan := scheduler.ErrorChan()
		for {
			select {
			case <-stopNotifier.Done():
				return
			default:
			}
			err, ok := <-errChan
			if ok {
				errMsg := fmt.Sprintf("Received an error from error Channel %s", err)
				record(2, errMsg)
			}
			time.Sleep(time.Microsecond)
		}
	}()
}

func waitForSchedulerStart(scheduler sched.Scheduler) {
	for scheduler.Status() != sched.SCHED_STATUS_STARTED {
		time.Sleep(time.Microsecond)
	}
}
