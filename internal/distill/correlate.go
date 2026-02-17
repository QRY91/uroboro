package distill

import "time"

const CorrelationWindow = 30 * time.Minute

// Correlate finds the nearest git commit within ±CorrelationWindow
// for each uro capture and sets its CorrelatedGitHash field.
// O(n*m) is fine at milestone 1 scale (hundreds, not millions).
func Correlate(gitExtracts []GitExtract, uroExtracts []UroExtract) {
	for i := range uroExtracts {
		bestHash := ""
		bestDelta := CorrelationWindow + 1
		for _, g := range gitExtracts {
			delta := uroExtracts[i].Timestamp.Sub(g.Date)
			if delta < 0 {
				delta = -delta
			}
			if delta <= CorrelationWindow && delta < bestDelta {
				bestDelta = delta
				bestHash = g.Hash
			}
		}
		uroExtracts[i].CorrelatedGitHash = bestHash
	}
}
