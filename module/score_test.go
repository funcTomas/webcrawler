package module

import "testing"

func TestCalculateScoreSimple(t *testing.T) {
	counts := Counts{
		CalledCount:    100000,
		AcceptedCount:  99900,
		CompletedCount: 99500,
		HandlingNumber: 200,
	}
	expectedScore := counts.CalledCount +
		counts.AcceptedCount<<1 +
		counts.CompletedCount<<2 +
		counts.HandlingNumber<<4
	score := CalculateScoreSimple(counts)
	if score != expectedScore {
		t.Fatalf("Inconsistent score, expected: %d, actual: %d", expectedScore, score)
	}
	t.Logf("The score is %d", score)
}

func TestSetScore(t *testing.T) {
}
