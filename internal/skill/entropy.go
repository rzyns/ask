package skill

import (
	"math"
)

// CalculateEntropy calculates the Shannon entropy of a string.
// Higher entropy indicates more randomness, which is common in secrets/keys.
func CalculateEntropy(s string) float64 {
	if s == "" {
		return 0
	}

	freq := make(map[rune]float64)
	for _, char := range s {
		freq[char]++
	}

	var entropy float64
	length := float64(len(s))

	for _, count := range freq {
		p := count / length
		entropy -= p * math.Log2(p)
	}

	return entropy
}
