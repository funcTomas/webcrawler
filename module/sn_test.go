package module

import (
	"math"
	"testing"
)

func TestGenerator(t *testing.T) {
	maxmax := uint64(math.MaxUint64)
	start := uint64(1)
	max := uint64(0)
	snGen := NewSNGenerator(start, max)
	if snGen == nil {
		t.Fatalf("Counld not create SN Generator (start: %d, max: %d)", start, max)
	}
	if snGen.Start() != start {
		t.Fatalf("Inconsistent start for SN, expected: %d, acutal: %d", start, snGen.Start())
	}
	if snGen.Max() != maxmax {
		t.Fatalf("Inconsistent max for SN, expected: %d, actual: %d", maxmax, snGen.Max())
	}
	max = uint64(7)
	max = uint64(101)
	snGen = NewSNGenerator(start, max)
	if snGen == nil {
		t.Fatalf("Counld not create SN Generator (start: %d, max: %d)", start, max)
	}
	if snGen.Max() != max {
		t.Fatalf("Inconsistent max for SN, expected: %d, actual: %d", max, snGen.Max())
	}
	end := snGen.Max()*5 + 11
	expectedSN := start
	var expectedNext uint64
	var expectedCycleCount uint64
	for i := start; i < end; i++ {
		sn := snGen.Get()
		if expectedSN > snGen.Max() {
			expectedSN = start
		}
		if sn != expectedSN {
			t.Fatalf("Incosistent ID, expected: %d, acutal: %d, (index: %d)", expectedSN, sn, i)
		}
		expectedNext = expectedSN + 1
		if expectedNext > snGen.Max() {
			expectedNext = start
		}
		if snGen.Next() != expectedNext {
			t.Fatalf("Inconsistent next ID, expected: %d, actutal: %d, (sn: %d, index: %d, )",
				expectedNext, snGen.Next(), sn, i)
		}
		if sn == snGen.Max() {
			expectedCycleCount++
		}
		if snGen.CycleCount() != expectedCycleCount {
			t.Fatalf("Inconsistent cycle count, expected: %d, actual: %d (sn: %d, index: %d)",
				expectedCycleCount, snGen.CycleCount(), sn, i)
		}
		expectedSN++
	}
}
