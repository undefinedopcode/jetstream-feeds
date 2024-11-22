package main

import (
	"strings"
	"unicode"
)

// SentimentMatch represents a matched pattern with context
type SentimentMatch struct {
	Pattern         string
	Context         string
	StartIndex      int
	ConfidenceScore float64
}

// TextAnalyzer provides methods to analyze text content
type TextAnalyzer struct {
	// Configurable thresholds and settings
	MinContextLength int
	MaxContextLength int
	Patterns         map[string]float64
	Threshold        float64
	Triggers         []string
}

// NewTextAnalyzer creates a new analyzer with default settings
func NewTextAnalyzer(triggers []string, patterns map[string]float64, threshold float64) *TextAnalyzer {
	return &TextAnalyzer{
		MinContextLength: 60,
		MaxContextLength: 300,
		Patterns:         patterns,
		Threshold:        threshold,
		Triggers:         triggers,
	}
}

// preprocessText normalizes input text for analysis
func (a *TextAnalyzer) preprocessText(text string) string {
	// Convert to lowercase
	text = strings.ToLower(text)

	// Replace multiple spaces with single space
	text = strings.Join(strings.Fields(text), " ")

	return text
}

// getContext extracts surrounding text for a match
func (a *TextAnalyzer) getContext(text string, start, end int) string {
	// Find context boundaries
	contextStart := start - a.MinContextLength
	if contextStart < 0 {
		contextStart = 0
	}

	contextEnd := end + a.MinContextLength
	if contextEnd > len(text) {
		contextEnd = len(text)
	}

	// Find word boundaries
	for contextStart > 0 && !unicode.IsSpace(rune(text[contextStart-1])) {
		contextStart--
	}

	for contextEnd < len(text) && !unicode.IsSpace(rune(text[contextEnd])) {
		contextEnd++
	}

	return text[contextStart:contextEnd]
}

// AnalyzeText examines text for patterns and returns matches with context
func (a *TextAnalyzer) AnalyzeText(text string) []SentimentMatch {
	text = a.preprocessText(text)
	var matches []SentimentMatch

	// Search for patterns and collect matches with context
	for pattern, baseConfidence := range a.Patterns {
		index := 0
		for {
			oldIndex := index
			index = strings.Index(text[index:], pattern)
			if index == -1 || index < oldIndex {
				break
			}

			// Get surrounding context
			context := a.getContext(text, index, index+len(pattern))

			// Create match entry
			match := SentimentMatch{
				Pattern:         pattern,
				Context:         context,
				StartIndex:      index,
				ConfidenceScore: baseConfidence,
			}

			matches = append(matches, match)
			index += len(pattern)
		}
	}

	return matches
}

func (a *TextAnalyzer) HasTriggers(text string) bool {
	text = strings.ToLower(text)
	if len(a.Triggers) == 0 {
		return true
	}
	for _, word := range a.Triggers {
		if !strings.Contains(text, word) {
			return false
		}
	}
	return true
}

func (a *TextAnalyzer) Score(text string) (float64, bool) {
	if !a.HasTriggers(text) {
		return 0, false
	}

	matches := a.AnalyzeText(text)

	// Process matches as needed
	var total float64
	for _, match := range matches {
		//log.Println("Pattern:", match.Pattern)
		//log.Println("Context:", match.Context)
		//log.Println("Confidence:", match.ConfidenceScore)
		//log.Println("---")
		total += match.ConfidenceScore
	}
	//log.Println("Total: ", total)

	return total, total >= a.Threshold
}
