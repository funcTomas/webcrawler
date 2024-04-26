package module

type CalculateScore func(counts Counts) uint64

func CalculateSocreSimple(counts Counts) uint64 {
	return counts.CalledCount +
		counts.AcceptedCount<<1 +
		counts.CompletedCount<<2 +
		counts.HandlingNumber<<4
}

func SetScore(module Module) bool {
	calculator := module.ScoreCalculator()
	if calculator == nil {
		calculator = CalculateSocreSimple
	}
	newScore := calculator(module.Counts())
	if newScore == module.Score() {
		return false
	}
	module.SetScore(newScore)
	return true
}
