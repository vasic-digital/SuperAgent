package skills

import (
	"context"
	"sort"
	"strings"
	"unicode"

	"github.com/sirupsen/logrus"
)

// Matcher handles matching user input to appropriate skills.
type Matcher struct {
	registry        *Registry
	config          *SkillConfig
	semanticMatcher SemanticMatcher
	log             *logrus.Logger
}

// SemanticMatcher is an interface for semantic matching (implemented by LLM).
type SemanticMatcher interface {
	Match(ctx context.Context, input string, candidates []string) ([]float64, error)
}

// NewMatcher creates a new skill matcher.
func NewMatcher(registry *Registry, config *SkillConfig) *Matcher {
	if config == nil {
		config = DefaultSkillConfig()
	}

	return &Matcher{
		registry: registry,
		config:   config,
		log:      logrus.New(),
	}
}

// SetLogger sets the logger for the matcher.
func (m *Matcher) SetLogger(log *logrus.Logger) {
	m.log = log
}

// SetSemanticMatcher sets the semantic matcher (LLM-based).
func (m *Matcher) SetSemanticMatcher(sm SemanticMatcher) {
	m.semanticMatcher = sm
}

// Match finds skills that match the user input.
func (m *Matcher) Match(ctx context.Context, input string) ([]SkillMatch, error) {
	matches := make([]SkillMatch, 0)

	// Normalize input
	normalizedInput := normalizeQuery(input)

	// Phase 1: Exact trigger matching
	exactMatches := m.matchExact(normalizedInput)
	matches = append(matches, exactMatches...)

	// Phase 2: Partial trigger matching
	partialMatches := m.matchPartial(normalizedInput)
	matches = append(matches, partialMatches...)

	// Phase 3: Fuzzy matching
	fuzzyMatches := m.matchFuzzy(normalizedInput)
	matches = append(matches, fuzzyMatches...)

	// Phase 4: Semantic matching (if enabled and LLM available)
	if m.config.EnableSemanticMatching && m.semanticMatcher != nil {
		semanticMatches, err := m.matchSemantic(ctx, normalizedInput)
		if err != nil {
			m.log.WithError(err).Warn("Semantic matching failed")
		} else {
			matches = append(matches, semanticMatches...)
		}
	}

	// Deduplicate and sort by confidence
	matches = m.deduplicateAndSort(matches)

	// Filter by minimum confidence
	filteredMatches := make([]SkillMatch, 0)
	for _, match := range matches {
		if match.Confidence >= m.config.MinConfidence {
			filteredMatches = append(filteredMatches, match)
		}
	}

	m.log.WithFields(logrus.Fields{
		"input":            input,
		"total_matches":    len(matches),
		"filtered_matches": len(filteredMatches),
	}).Debug("Skill matching completed")

	return filteredMatches, nil
}

// matchExact finds exact trigger matches.
func (m *Matcher) matchExact(input string) []SkillMatch {
	matches := make([]SkillMatch, 0)

	triggers := m.registry.GetTriggers()
	for _, trigger := range triggers {
		if strings.EqualFold(input, trigger) || strings.Contains(strings.ToLower(input), strings.ToLower(trigger)) {
			skills := m.registry.GetByTrigger(trigger)
			for _, skill := range skills {
				matches = append(matches, SkillMatch{
					Skill:          skill,
					Confidence:     1.0,
					MatchedTrigger: trigger,
					MatchType:      MatchTypeExact,
				})
			}
		}
	}

	return matches
}

// matchPartial finds partial trigger matches.
func (m *Matcher) matchPartial(input string) []SkillMatch {
	matches := make([]SkillMatch, 0)
	inputWords := tokenize(input)

	triggers := m.registry.GetTriggers()
	for _, trigger := range triggers {
		triggerWords := tokenize(trigger)

		// Calculate word overlap
		overlap := wordOverlap(inputWords, triggerWords)
		if overlap > 0 {
			confidence := float64(overlap) / float64(max(len(inputWords), len(triggerWords)))
			if confidence >= 0.5 && confidence < 1.0 { // Only partial matches
				skills := m.registry.GetByTrigger(trigger)
				for _, skill := range skills {
					matches = append(matches, SkillMatch{
						Skill:          skill,
						Confidence:     confidence,
						MatchedTrigger: trigger,
						MatchType:      MatchTypePartial,
					})
				}
			}
		}
	}

	return matches
}

// matchFuzzy performs fuzzy string matching.
func (m *Matcher) matchFuzzy(input string) []SkillMatch {
	matches := make([]SkillMatch, 0)

	for _, skill := range m.registry.GetAll() {
		// Check skill name similarity
		nameSimilarity := similarity(input, skill.Name)
		if nameSimilarity >= 0.6 {
			matches = append(matches, SkillMatch{
				Skill:          skill,
				Confidence:     nameSimilarity * 0.8, // Reduce confidence for fuzzy matches
				MatchedTrigger: skill.Name,
				MatchType:      MatchTypeFuzzy,
			})
		}

		// Check description similarity
		descSimilarity := similarity(input, skill.Description)
		if descSimilarity >= 0.6 {
			matches = append(matches, SkillMatch{
				Skill:          skill,
				Confidence:     descSimilarity * 0.7,
				MatchedTrigger: "description",
				MatchType:      MatchTypeFuzzy,
			})
		}
	}

	return matches
}

// matchSemantic uses LLM for semantic matching.
func (m *Matcher) matchSemantic(ctx context.Context, input string) ([]SkillMatch, error) {
	matches := make([]SkillMatch, 0)

	// Get all skill descriptions
	skills := m.registry.GetAll()
	if len(skills) == 0 {
		return matches, nil
	}

	// Limit candidates for performance
	maxCandidates := 50
	if len(skills) > maxCandidates {
		skills = skills[:maxCandidates]
	}

	candidates := make([]string, len(skills))
	for i, skill := range skills {
		candidates[i] = skill.Name + ": " + skill.Description
	}

	// Get semantic similarity scores
	scores, err := m.semanticMatcher.Match(ctx, input, candidates)
	if err != nil {
		return nil, err
	}

	// Create matches for high scores
	for i, score := range scores {
		if score >= m.config.MinConfidence {
			matches = append(matches, SkillMatch{
				Skill:          skills[i],
				Confidence:     score,
				MatchedTrigger: "semantic",
				MatchType:      MatchTypeSemantic,
			})
		}
	}

	return matches, nil
}

// deduplicateAndSort removes duplicate matches and sorts by confidence.
func (m *Matcher) deduplicateAndSort(matches []SkillMatch) []SkillMatch {
	// Deduplicate by keeping highest confidence for each skill
	bySkill := make(map[string]SkillMatch)
	for _, match := range matches {
		existing, ok := bySkill[match.Skill.Name]
		if !ok || match.Confidence > existing.Confidence {
			bySkill[match.Skill.Name] = match
		}
	}

	// Convert to slice
	result := make([]SkillMatch, 0, len(bySkill))
	for _, match := range bySkill {
		result = append(result, match)
	}

	// Sort by confidence descending
	sort.Slice(result, func(i, j int) bool {
		return result[i].Confidence > result[j].Confidence
	})

	return result
}

// MatchBest returns the best matching skill.
func (m *Matcher) MatchBest(ctx context.Context, input string) (*SkillMatch, error) {
	matches, err := m.Match(ctx, input)
	if err != nil {
		return nil, err
	}

	if len(matches) == 0 {
		return nil, nil
	}

	return &matches[0], nil
}

// MatchMultiple returns the top N matching skills.
func (m *Matcher) MatchMultiple(ctx context.Context, input string, n int) ([]SkillMatch, error) {
	matches, err := m.Match(ctx, input)
	if err != nil {
		return nil, err
	}

	if len(matches) > n {
		matches = matches[:n]
	}

	return matches, nil
}

// Utility functions

// normalizeQuery normalizes a query string for matching.
func normalizeQuery(s string) string {
	// Convert to lowercase
	s = strings.ToLower(s)

	// Remove extra whitespace
	s = strings.Join(strings.Fields(s), " ")

	return s
}

// tokenize splits a string into tokens.
func tokenize(s string) []string {
	s = strings.ToLower(s)
	var tokens []string
	var current strings.Builder

	for _, r := range s {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			current.WriteRune(r)
		} else if current.Len() > 0 {
			tokens = append(tokens, current.String())
			current.Reset()
		}
	}

	if current.Len() > 0 {
		tokens = append(tokens, current.String())
	}

	return tokens
}

// wordOverlap counts the number of common words.
func wordOverlap(a, b []string) int {
	bSet := make(map[string]bool)
	for _, w := range b {
		bSet[w] = true
	}

	overlap := 0
	for _, w := range a {
		if bSet[w] {
			overlap++
		}
	}

	return overlap
}

// similarity calculates string similarity using Jaccard index.
func similarity(a, b string) float64 {
	tokensA := tokenize(a)
	tokensB := tokenize(b)

	if len(tokensA) == 0 || len(tokensB) == 0 {
		return 0
	}

	setA := make(map[string]bool)
	for _, t := range tokensA {
		setA[t] = true
	}

	setB := make(map[string]bool)
	for _, t := range tokensB {
		setB[t] = true
	}

	intersection := 0
	for t := range setA {
		if setB[t] {
			intersection++
		}
	}

	union := len(setA) + len(setB) - intersection
	if union == 0 {
		return 0
	}

	return float64(intersection) / float64(union)
}

// containsIgnoreCase checks if s contains substr (case insensitive).
func containsIgnoreCase(s, substr string) bool {
	return strings.Contains(strings.ToLower(s), strings.ToLower(substr))
}

// max returns the maximum of two integers.
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
