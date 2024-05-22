package scheduler

import (
	"sync"
	"testing"
)

func TestCheckStatus(t *testing.T) {
	var currentStatus, wantedStatus Status
	var currentStatusList, wantedStatusList []Status
	currentStatusList = []Status{
		SCHED_STATUS_INITIALIZING,
		SCHED_STATUS_STARTING,
		SCHED_STATUS_STOPPING,
	}
	wantedStatus = SCHED_STATUS_INITIALIZING
	for _, currentStatus := range currentStatusList {
		if err := checkStatus(currentStatus, wantedStatus, nil); err == nil {
			t.Fatalf("It can still check status with incorrect current status %q", GetStatusDescription(currentStatus))
		}
	}

	currentStatus = SCHED_STATUS_UNINITIALIZED
	wantedStatusList = []Status{
		SCHED_STATUS_UNINITIALIZED,
		SCHED_STATUS_INITIALIZED,
		SCHED_STATUS_STARTED,
		SCHED_STATUS_STOPPED,
	}
	for _, wantedStatus := range wantedStatusList {
		if err := checkStatus(currentStatus, wantedStatus, nil); err == nil {
			t.Fatalf("It can still check status with incorrect current status %q", GetStatusDescription(currentStatus))
		}
	}
	currentStatus = SCHED_STATUS_UNINITIALIZED
	wantedStatusList = []Status{
		SCHED_STATUS_STARTING,
		SCHED_STATUS_STOPPING,
	}
	for _, wantedStatus := range wantedStatusList {
		if err := checkStatus(currentStatus, wantedStatus, nil); err == nil {
			t.Fatalf("It can still check status with incorrect current status %q, wanted Status %q",
				GetStatusDescription(currentStatus), GetStatusDescription(wantedStatus))
		}
	}
	wantedStatus = SCHED_STATUS_INITIALIZING
	if err := checkStatus(currentStatus, wantedStatus, nil); err != nil {
		t.Fatalf("An error occurs when checking status, current status %q, wanted Status %q",
			GetStatusDescription(currentStatus), GetStatusDescription(wantedStatus))
	}
	currentStatus = SCHED_STATUS_STARTED
	wantedStatusList = []Status{
		SCHED_STATUS_INITIALIZING,
		SCHED_STATUS_STARTING,
	}
	for _, wantedStatus := range wantedStatusList {
		if err := checkStatus(currentStatus, wantedStatus, nil); err == nil {
			t.Fatalf("It can still check status with incorrect current status %q, wanted Status %q",
				GetStatusDescription(currentStatus), GetStatusDescription(wantedStatus))
		}
	}
	wantedStatus = SCHED_STATUS_STOPPING
	if err := checkStatus(currentStatus, wantedStatus, nil); err != nil {
		t.Fatalf("An error occurs when checking status, current status %q, wanted Status %q",
			GetStatusDescription(currentStatus), GetStatusDescription(wantedStatus))
	}
	currentStatusList = []Status{
		SCHED_STATUS_UNINITIALIZED,
		SCHED_STATUS_INITIALIZING,
		SCHED_STATUS_INITIALIZED,
		SCHED_STATUS_STARTING,
		SCHED_STATUS_STOPPING,
		SCHED_STATUS_STOPPED,
	}
	wantedStatus = SCHED_STATUS_STOPPING
	for _, currentStatus := range currentStatusList {
		if err := checkStatus(currentStatus, wantedStatus, nil); err == nil {
			t.Fatalf("It can still check status with incorrect current status %q, wanted Status %q",
				GetStatusDescription(currentStatus), GetStatusDescription(wantedStatus))
		}
	}
	currentStatus = SCHED_STATUS_STARTED
	if err := checkStatus(currentStatus, wantedStatus, nil); err != nil {
		t.Fatalf("It can still check status with incorrect current status %q, wanted Status %q",
			GetStatusDescription(currentStatus), GetStatusDescription(wantedStatus))
	}
}

func TestCheckStatusInParallel(t *testing.T) {
	number := 1000
	var lock sync.Mutex
	t.Run("Check Status in parallel 1", func(t *testing.T) {
		for i := 0; i < number; i++ {
			currentStatus := SCHED_STATUS_UNINITIALIZED
			wantedStatus := SCHED_STATUS_INITIALIZING
			if err := checkStatus(currentStatus, wantedStatus, &lock); err != nil {
				t.Fatalf("An error occurs when checking status: %s (currentStatus: %q, wantedStatus: %q)",
					err, GetStatusDescription(currentStatus), GetStatusDescription(wantedStatus))
			}
		}
	})
	t.Run("Check Status in parallel 2", func(t *testing.T) {
		for i := 0; i < number; i++ {
			currentStatus := SCHED_STATUS_INITIALIZED
			wantedStatusList := []Status{
				SCHED_STATUS_INITIALIZING,
				SCHED_STATUS_STARTING,
			}
			for _, wantedStatus := range wantedStatusList {
				if err := checkStatus(currentStatus, wantedStatus, &lock); err != nil {
					t.Fatalf("An error occurs when checking status %s (currentStatus: %q, wantedStatus: %q)",
						err, GetStatusDescription(currentStatus), GetStatusDescription(wantedStatus))
				}
			}
		}
	})
	t.Run("Check Status in parallel 3", func(t *testing.T) {
		for i := 0; i < number; i++ {
			currentStatus := SCHED_STATUS_STARTED
			wantedStatus := SCHED_STATUS_STOPPING
			if err := checkStatus(currentStatus, wantedStatus, &lock); err != nil {
				t.Fatalf("An error occurs when checking status %s (currentStatus: %q, wantedStatus: %q)",
					err, GetStatusDescription(currentStatus), GetStatusDescription(wantedStatus))
			}
		}
	})
	t.Run("Check Status in parallel 4", func(t *testing.T) {
		for i := 0; i < number; i++ {
			currentStatus := SCHED_STATUS_STOPPED
			wantedStatusList := []Status{
				SCHED_STATUS_INITIALIZING,
				SCHED_STATUS_STARTING,
			}
			for _, wantedStatus := range wantedStatusList {
				if err := checkStatus(currentStatus, wantedStatus, &lock); err != nil {
					t.Fatalf("An error occurs when checking status %s (currentStatus: %q, wantedStatus: %q)",
						err, GetStatusDescription(currentStatus), GetStatusDescription(wantedStatus))
				}
			}
		}
	})
}
