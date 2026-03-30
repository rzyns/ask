package skill

import (
	"math"
	"testing"
)

func TestCalculateEntropyComprehensive(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected float64
	}{
		{
			name:     "empty string returns 0",
			input:    "",
			expected: 0,
		},
		{
			name:     "single character returns 0",
			input:    "a",
			expected: 0,
		},
		{
			name:     "repeated character returns 0",
			input:    "aaaa",
			expected: 0,
		},
		{
			name:     "long repeated character returns 0",
			input:    "bbbbbbbbbb",
			expected: 0,
		},
		{
			name:     "two distinct characters equal frequency",
			input:    "ab",
			expected: 1.0, // log2(2) = 1
		},
		{
			name:     "four distinct characters equal frequency",
			input:    "abcd",
			expected: 2.0, // log2(4) = 2
		},
		{
			name:     "eight distinct characters equal frequency",
			input:    "abcdefgh",
			expected: 3.0, // log2(8) = 3
		},
		{
			name:     "two characters unequal frequency aab",
			input:    "aab",
			expected: -((2.0/3)*math.Log2(2.0/3) + (1.0/3)*math.Log2(1.0/3)),
		},
		{
			name:     "high entropy random-looking string",
			input:    "aB3$xZ9!mK",
			expected: math.Log2(10), // 10 unique characters -> ~3.3219
		},
		{
			name:     "non-ASCII UTF-8 CJK characters all unique",
			input:    "\u4f60\u597d\u4e16\u754c",
			expected: 2.0, // 4 unique runes -> log2(4) = 2
		},
		{
			name:     "mixed ASCII and UTF-8",
			input:    "abc\u4f60",
			expected: 2.0, // 4 unique runes -> log2(4) = 2
		},
		{
			name:     "emoji characters all same",
			input:    "????",
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CalculateEntropy(tt.input)

			if math.Abs(got-tt.expected) > 0.0001 {
				t.Errorf("CalculateEntropy(%q) = %f, expected %f", tt.input, got, tt.expected)
			}
		})
	}
}

func TestCalculateEntropyNonNegative(t *testing.T) {
	// Entropy should always be non-negative
	inputs := []string{
		"", "a", "aa", "ab", "abc", "password", "P@ssw0rd!",
		"aaabbbccc", "abcdefghijklmnopqrstuvwxyz",
	}
	for _, input := range inputs {
		got := CalculateEntropy(input)
		if got < 0 {
			t.Errorf("CalculateEntropy(%q) = %f, expected non-negative", input, got)
		}
	}
}

func TestCalculateEntropyMonotonicity(t *testing.T) {
	// More distinct characters with uniform distribution should yield higher entropy
	e1 := CalculateEntropy("ab")       // 2 chars -> 1.0
	e2 := CalculateEntropy("abcd")     // 4 chars -> 2.0
	e3 := CalculateEntropy("abcdefgh") // 8 chars -> 3.0

	if e1 >= e2 {
		t.Errorf("Expected entropy of 'ab' (%f) < entropy of 'abcd' (%f)", e1, e2)
	}
	if e2 >= e3 {
		t.Errorf("Expected entropy of 'abcd' (%f) < entropy of 'abcdefgh' (%f)", e2, e3)
	}
}
