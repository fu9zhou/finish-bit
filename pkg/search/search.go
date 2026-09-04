// Package search ranks operation metadata without loading full capability descriptions.
package search

import (
	"math"
	"sort"
	"strings"
	"unicode"

	"github.com/fu9zhou/finish-bit/pkg/operation"
)

type Match struct {
	ID      string   `json:"id"`
	Summary string   `json:"summary"`
	Score   float64  `json:"score"`
	Tags    []string `json:"tags,omitempty"`
}

type Index struct{ definitions []operation.Definition }

func New(definitions []operation.Definition) *Index { return &Index{definitions: definitions} }

func (i *Index) Search(query string, limit int) []Match {
	if limit <= 0 {
		limit = 5
	}
	terms := tokens(query)
	if len(terms) == 0 {
		return nil
	}
	matches := make([]Match, 0)
	for _, definition := range i.definitions {
		score := scoreDefinition(definition, terms)
		if score > 0 {
			matches = append(matches, Match{ID: definition.ID, Summary: definition.Summary, Score: math.Round(score*100) / 100, Tags: definition.Tags})
		}
	}
	sort.SliceStable(matches, func(a, b int) bool {
		if matches[a].Score == matches[b].Score {
			return matches[a].ID < matches[b].ID
		}
		return matches[a].Score > matches[b].Score
	})
	if len(matches) > 1 {
		minimum := matches[0].Score * 0.35
		kept := matches[:0]
		for _, match := range matches {
			if match.Score >= minimum {
				kept = append(kept, match)
			}
		}
		matches = kept
	}
	if len(matches) > limit {
		matches = matches[:limit]
	}
	return matches
}

func scoreDefinition(def operation.Definition, terms []string) float64 {
	id := strings.ToLower(def.ID)
	summary := strings.ToLower(def.Summary)
	description := strings.ToLower(def.Description)
	aliases := strings.ToLower(strings.Join(def.Aliases, " "))
	tags := strings.ToLower(strings.Join(def.Tags, " "))
	score := 0.0
	for _, term := range terms {
		switch {
		case id == term:
			score += 20
		case strings.Contains(id, term):
			score += 10
		}
		if strings.Contains(aliases, term) {
			score += 8
		}
		if strings.Contains(tags, term) {
			score += 5
		}
		if strings.Contains(summary, term) {
			score += 4
		}
		if strings.Contains(description, term) {
			score += 1
		}
	}
	return score
}

func tokens(value string) []string {
	value = strings.ToLower(strings.TrimSpace(value))
	if value == "" {
		return nil
	}
	parts := strings.FieldsFunc(value, func(r rune) bool {
		return unicode.IsSpace(r) || strings.ContainsRune(".,;:/\\|_-()[]{}\"'", r)
	})
	if len(parts) == 1 {
		runes := []rune(parts[0])
		hasHan := false
		for _, r := range runes {
			if unicode.Is(unicode.Han, r) {
				hasHan = true
				break
			}
		}
		if hasHan && len(runes) > 2 {
			for index := 0; index < len(runes)-1; index++ {
				parts = append(parts, string(runes[index:index+2]))
			}
		}
	}
	// Keep the original phrase for CJK and exact alias matching.
	if len(parts) > 1 {
		parts = append(parts, value)
	}
	return parts
}
