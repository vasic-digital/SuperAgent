// Package voting provides weighted voting mechanisms for AI debate consensus.
// Implements the MiniMax formula: L* = argmax Σcᵢ · 𝟙[aᵢ = L]
// Where cᵢ is the confidence score and 𝟙[aᵢ = L] is indicator for vote L.
package voting

import (
	"context"
	"fmt"
	"math"
	"sort"
	"sync"
	"time"
)

// VotingConfig configures the weighted voting system.
type VotingConfig struct {
	// MinimumVotes is the minimum number of votes required
	MinimumVotes int `json:"minimum_votes"`
	// MinimumConfidence filters out low-confidence votes
	MinimumConfidence float64 `json:"minimum_confidence"`
	// EnableDiversityBonus adds bonus for diverse perspectives
	EnableDiversityBonus bool `json:"enable_diversity_bonus"`
	// DiversityWeight is the weight for diversity bonus (0-1)
	DiversityWeight float64 `json:"diversity_weight"`
	// EnableTieBreaking enables automatic tie breaking
	EnableTieBreaking bool `json:"enable_tie_breaking"`
	// TieBreakMethod specifies how to break ties
	TieBreakMethod TieBreakMethod `json:"tie_break_method"`
	// EnableHistoricalWeight considers historical accuracy
	EnableHistoricalWeight bool `json:"enable_historical_weight"`
}

// TieBreakMethod specifies how ties are broken.
type TieBreakMethod string

const (
	TieBreakByHighestConfidence TieBreakMethod = "highest_confidence"
	TieBreakByMostVotes         TieBreakMethod = "most_votes"
	TieBreakByLeaderVote        TieBreakMethod = "leader_vote"
	TieBreakByRandom            TieBreakMethod = "random"
)

// DefaultVotingConfig returns sensible defaults.
func DefaultVotingConfig() VotingConfig {
	return VotingConfig{
		MinimumVotes:           3,
		MinimumConfidence:      0.3,
		EnableDiversityBonus:   true,
		DiversityWeight:        0.1,
		EnableTieBreaking:      true,
		TieBreakMethod:         TieBreakByHighestConfidence,
		EnableHistoricalWeight: true,
	}
}

// Vote represents a single vote from an agent.
type Vote struct {
	AgentID            string                 `json:"agent_id"`
	Choice             string                 `json:"choice"`
	Confidence         float64                `json:"confidence"` // Primary weight (0-1)
	Score              float64                `json:"score"`      // LLMsVerifier score
	Specialization     string                 `json:"specialization"`
	Role               string                 `json:"role"`
	HistoricalAccuracy float64                `json:"historical_accuracy"` // Past voting accuracy
	Reasoning          string                 `json:"reasoning,omitempty"`
	Timestamp          time.Time              `json:"timestamp"`
	Metadata           map[string]interface{} `json:"metadata,omitempty"`
}

// VoteWeight holds the calculated weight components for a vote.
type VoteWeight struct {
	BaseWeight      float64 `json:"base_weight"`      // confidence
	ScoreWeight     float64 `json:"score_weight"`     // normalized score
	DiversityBonus  float64 `json:"diversity_bonus"`  // diversity contribution
	HistoricalBonus float64 `json:"historical_bonus"` // historical accuracy
	TotalWeight     float64 `json:"total_weight"`     // final weight
}

// VotingResult represents the result of a voting round.
type VotingResult struct {
	WinningChoice    string                 `json:"winning_choice"`
	WinningScore     float64                `json:"winning_score"`
	TotalVotes       int                    `json:"total_votes"`
	ValidVotes       int                    `json:"valid_votes"`
	ChoiceScores     map[string]float64     `json:"choice_scores"`
	ChoiceVoteCounts map[string]int         `json:"choice_vote_counts"`
	VoteWeights      map[string]*VoteWeight `json:"vote_weights"` // By agent ID
	Consensus        float64                `json:"consensus"`    // 0-1 agreement level
	IsTie            bool                   `json:"is_tie"`
	TieChoices       []string               `json:"tie_choices,omitempty"`
	TieBreakUsed     bool                   `json:"tie_break_used"`
	TieBreakMethod   TieBreakMethod         `json:"tie_break_method,omitempty"`
	Method           VotingMethod           `json:"method"`
	Timestamp        time.Time              `json:"timestamp"`
	Metadata         map[string]interface{} `json:"metadata,omitempty"`
}

// VotingMethod identifies the voting method used.
type VotingMethod string

const (
	VotingMethodWeighted  VotingMethod = "weighted"  // MiniMax formula
	VotingMethodMajority  VotingMethod = "majority"  // Simple majority
	VotingMethodUnanimous VotingMethod = "unanimous" // All agree
	VotingMethodPlurality VotingMethod = "plurality" // Most votes wins
	VotingMethodBorda     VotingMethod = "borda"     // Borda count
	VotingMethodCondorcet VotingMethod = "condorcet" // Condorcet method
)

// WeightedVotingSystem implements the MiniMax weighted voting formula.
type WeightedVotingSystem struct {
	config         VotingConfig
	votes          []*Vote
	historicalData map[string]*AgentHistory
	mu             sync.RWMutex
}

// AgentHistory tracks an agent's voting history.
type AgentHistory struct {
	AgentID       string    `json:"agent_id"`
	TotalVotes    int       `json:"total_votes"`
	CorrectVotes  int       `json:"correct_votes"`
	Accuracy      float64   `json:"accuracy"`
	AvgConfidence float64   `json:"avg_confidence"`
	LastVote      time.Time `json:"last_vote"`
}

// NewWeightedVotingSystem creates a new weighted voting system.
func NewWeightedVotingSystem(config VotingConfig) *WeightedVotingSystem {
	return &WeightedVotingSystem{
		config:         config,
		votes:          make([]*Vote, 0),
		historicalData: make(map[string]*AgentHistory),
	}
}

// AddVote adds a vote to the system.
func (wvs *WeightedVotingSystem) AddVote(vote *Vote) error {
	if vote.Choice == "" {
		return fmt.Errorf("vote choice cannot be empty")
	}

	if vote.Confidence < 0 || vote.Confidence > 1 {
		return fmt.Errorf("confidence must be between 0 and 1, got %f", vote.Confidence)
	}

	if vote.Timestamp.IsZero() {
		vote.Timestamp = time.Now()
	}

	wvs.mu.Lock()
	defer wvs.mu.Unlock()

	// Check for duplicate vote from same agent
	for i, existing := range wvs.votes {
		if existing.AgentID == vote.AgentID {
			// Replace existing vote
			wvs.votes[i] = vote
			return nil
		}
	}

	wvs.votes = append(wvs.votes, vote)
	return nil
}

// Calculate calculates the voting result using the MiniMax weighted formula.
// L* = argmax Σcᵢ · 𝟙[aᵢ = L]
func (wvs *WeightedVotingSystem) Calculate(ctx context.Context) (*VotingResult, error) {
	wvs.mu.RLock()
	defer wvs.mu.RUnlock()

	if len(wvs.votes) < wvs.config.MinimumVotes {
		return nil, fmt.Errorf("insufficient votes: got %d, need %d", len(wvs.votes), wvs.config.MinimumVotes)
	}

	// Filter votes by minimum confidence
	validVotes := wvs.filterVotes()
	if len(validVotes) < wvs.config.MinimumVotes {
		return nil, fmt.Errorf("insufficient valid votes after filtering: got %d, need %d", len(validVotes), wvs.config.MinimumVotes)
	}

	// Calculate weights for each vote
	voteWeights := make(map[string]*VoteWeight)
	for _, vote := range validVotes {
		voteWeights[vote.AgentID] = wvs.calculateVoteWeight(vote, validVotes)
	}

	// Calculate choice scores: Σcᵢ · 𝟙[aᵢ = L]
	choiceScores := make(map[string]float64)
	choiceVoteCounts := make(map[string]int)

	for _, vote := range validVotes {
		weight := voteWeights[vote.AgentID].TotalWeight
		choiceScores[vote.Choice] += weight
		choiceVoteCounts[vote.Choice]++
	}

	// Find winning choice: L* = argmax
	winningChoice, winningScore, isTie, tieChoices := wvs.findWinner(choiceScores)

	// Handle tie breaking
	tieBreakUsed := false
	tieBreakMethod := TieBreakMethod("")
	if isTie && wvs.config.EnableTieBreaking {
		winningChoice, tieBreakMethod = wvs.breakTie(tieChoices, validVotes, voteWeights)
		tieBreakUsed = true
		isTie = false
	}

	// Calculate consensus level
	consensus := wvs.calculateConsensus(choiceScores, len(validVotes))

	result := &VotingResult{
		WinningChoice:    winningChoice,
		WinningScore:     winningScore,
		TotalVotes:       len(wvs.votes),
		ValidVotes:       len(validVotes),
		ChoiceScores:     choiceScores,
		ChoiceVoteCounts: choiceVoteCounts,
		VoteWeights:      voteWeights,
		Consensus:        consensus,
		IsTie:            isTie,
		TieChoices:       tieChoices,
		TieBreakUsed:     tieBreakUsed,
		TieBreakMethod:   tieBreakMethod,
		Method:           VotingMethodWeighted,
		Timestamp:        time.Now(),
	}

	return result, nil
}

// filterVotes filters out low-confidence votes.
func (wvs *WeightedVotingSystem) filterVotes() []*Vote {
	valid := make([]*Vote, 0, len(wvs.votes))
	for _, vote := range wvs.votes {
		if vote.Confidence >= wvs.config.MinimumConfidence {
			valid = append(valid, vote)
		}
	}
	return valid
}

// calculateVoteWeight calculates the total weight for a vote.
func (wvs *WeightedVotingSystem) calculateVoteWeight(vote *Vote, allVotes []*Vote) *VoteWeight {
	weight := &VoteWeight{}

	// Base weight = confidence (the cᵢ in the formula)
	weight.BaseWeight = vote.Confidence

	// Score weight (normalized LLMsVerifier score)
	if vote.Score > 0 {
		weight.ScoreWeight = vote.Score / 10.0 // Normalize to 0-1
	}

	// Diversity bonus
	if wvs.config.EnableDiversityBonus {
		weight.DiversityBonus = wvs.calculateDiversityBonus(vote, allVotes)
	}

	// Historical accuracy bonus
	if wvs.config.EnableHistoricalWeight {
		if history, exists := wvs.historicalData[vote.AgentID]; exists {
			weight.HistoricalBonus = history.Accuracy * 0.2 // Max 0.2 bonus
		}
	}

	// Total weight calculation
	// Primary: confidence (base)
	// Secondary: score adjustment
	// Tertiary: diversity and history bonuses
	weight.TotalWeight = weight.BaseWeight * (1 + weight.ScoreWeight*0.2) *
		(1 + weight.DiversityBonus*wvs.config.DiversityWeight) *
		(1 + weight.HistoricalBonus)

	return weight
}

// calculateDiversityBonus calculates bonus for contributing to diversity.
func (wvs *WeightedVotingSystem) calculateDiversityBonus(vote *Vote, allVotes []*Vote) float64 {
	// Count how many others share this vote's characteristics
	sameSpec := 0
	sameRole := 0
	sameChoice := 0

	for _, other := range allVotes {
		if other.AgentID == vote.AgentID {
			continue
		}
		if other.Specialization == vote.Specialization {
			sameSpec++
		}
		if other.Role == vote.Role {
			sameRole++
		}
		if other.Choice == vote.Choice {
			sameChoice++
		}
	}

	// Higher bonus for unique perspectives
	uniquenessScore := 0.0
	total := float64(len(allVotes) - 1)
	if total > 0 {
		specUniqueness := 1 - float64(sameSpec)/total
		roleUniqueness := 1 - float64(sameRole)/total

		// But if choice is unique (minority), add extra bonus
		choiceBonus := 0.0
		if float64(sameChoice) < total/2 {
			choiceBonus = 0.1 // Minority viewpoint bonus
		}

		uniquenessScore = (specUniqueness*0.4 + roleUniqueness*0.4 + choiceBonus*0.2)
	}

	return uniquenessScore
}

// findWinner finds the winning choice and checks for ties.
func (wvs *WeightedVotingSystem) findWinner(choiceScores map[string]float64) (string, float64, bool, []string) {
	if len(choiceScores) == 0 {
		return "", 0, false, nil
	}

	// Sort choices by score
	type choiceScore struct {
		choice string
		score  float64
	}

	sorted := make([]choiceScore, 0, len(choiceScores))
	for choice, score := range choiceScores {
		sorted = append(sorted, choiceScore{choice, score})
	}
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].score > sorted[j].score
	})

	winningChoice := sorted[0].choice
	winningScore := sorted[0].score

	// Check for ties (within 1% of each other)
	tieThreshold := winningScore * 0.01
	tieChoices := make([]string, 0)
	for _, cs := range sorted {
		if winningScore-cs.score <= tieThreshold {
			tieChoices = append(tieChoices, cs.choice)
		}
	}

	isTie := len(tieChoices) > 1

	return winningChoice, winningScore, isTie, tieChoices
}

// breakTie breaks a tie using the configured method.
func (wvs *WeightedVotingSystem) breakTie(tieChoices []string, votes []*Vote, weights map[string]*VoteWeight) (string, TieBreakMethod) {
	switch wvs.config.TieBreakMethod {
	case TieBreakByHighestConfidence:
		return wvs.breakTieByHighestConfidence(tieChoices, votes, weights)
	case TieBreakByMostVotes:
		return wvs.breakTieByMostVotes(tieChoices, votes)
	case TieBreakByLeaderVote:
		return wvs.breakTieByLeaderVote(tieChoices, votes)
	default:
		// Random (just pick first alphabetically for determinism)
		sort.Strings(tieChoices)
		return tieChoices[0], TieBreakByRandom
	}
}

// breakTieByHighestConfidence selects the choice with the highest single confidence vote.
func (wvs *WeightedVotingSystem) breakTieByHighestConfidence(tieChoices []string, votes []*Vote, weights map[string]*VoteWeight) (string, TieBreakMethod) {
	tieSet := make(map[string]bool)
	for _, c := range tieChoices {
		tieSet[c] = true
	}

	maxConfidence := 0.0
	winner := tieChoices[0]

	for _, vote := range votes {
		if tieSet[vote.Choice] && vote.Confidence > maxConfidence {
			maxConfidence = vote.Confidence
			winner = vote.Choice
		}
	}

	return winner, TieBreakByHighestConfidence
}

// breakTieByMostVotes selects the choice with the most raw votes.
func (wvs *WeightedVotingSystem) breakTieByMostVotes(tieChoices []string, votes []*Vote) (string, TieBreakMethod) {
	tieSet := make(map[string]bool)
	for _, c := range tieChoices {
		tieSet[c] = true
	}

	voteCounts := make(map[string]int)
	for _, vote := range votes {
		if tieSet[vote.Choice] {
			voteCounts[vote.Choice]++
		}
	}

	maxVotes := 0
	winner := tieChoices[0]
	for choice, count := range voteCounts {
		if count > maxVotes {
			maxVotes = count
			winner = choice
		}
	}

	return winner, TieBreakByMostVotes
}

// breakTieByLeaderVote uses the vote of the highest-scored agent.
func (wvs *WeightedVotingSystem) breakTieByLeaderVote(tieChoices []string, votes []*Vote) (string, TieBreakMethod) {
	tieSet := make(map[string]bool)
	for _, c := range tieChoices {
		tieSet[c] = true
	}

	// Find highest scored agent that voted for a tie choice
	maxScore := 0.0
	winner := tieChoices[0]

	for _, vote := range votes {
		if tieSet[vote.Choice] && vote.Score > maxScore {
			maxScore = vote.Score
			winner = vote.Choice
		}
	}

	return winner, TieBreakByLeaderVote
}

// calculateConsensus calculates the level of agreement.
func (wvs *WeightedVotingSystem) calculateConsensus(choiceScores map[string]float64, totalVotes int) float64 {
	if len(choiceScores) == 0 || totalVotes == 0 {
		return 0
	}

	// Calculate total score
	totalScore := 0.0
	maxScore := 0.0
	for _, score := range choiceScores {
		totalScore += score
		if score > maxScore {
			maxScore = score
		}
	}

	if totalScore == 0 {
		return 0
	}

	// Consensus = winning score / total score
	// 1.0 = unanimous, 0.5 = split, < 0.5 = fragmented
	return maxScore / totalScore
}

// UpdateHistory updates the historical accuracy for an agent.
func (wvs *WeightedVotingSystem) UpdateHistory(agentID string, wasCorrect bool, confidence float64) {
	wvs.mu.Lock()
	defer wvs.mu.Unlock()

	history, exists := wvs.historicalData[agentID]
	if !exists {
		history = &AgentHistory{
			AgentID: agentID,
		}
		wvs.historicalData[agentID] = history
	}

	history.TotalVotes++
	if wasCorrect {
		history.CorrectVotes++
	}
	history.Accuracy = float64(history.CorrectVotes) / float64(history.TotalVotes)
	history.AvgConfidence = (history.AvgConfidence*float64(history.TotalVotes-1) + confidence) / float64(history.TotalVotes)
	history.LastVote = time.Now()
}

// GetHistory returns the history for an agent.
func (wvs *WeightedVotingSystem) GetHistory(agentID string) *AgentHistory {
	wvs.mu.RLock()
	defer wvs.mu.RUnlock()

	if history, exists := wvs.historicalData[agentID]; exists {
		copy := *history
		return &copy
	}
	return nil
}

// Reset clears all votes (keeps history).
func (wvs *WeightedVotingSystem) Reset() {
	wvs.mu.Lock()
	defer wvs.mu.Unlock()

	wvs.votes = make([]*Vote, 0)
}

// GetVotes returns all current votes.
func (wvs *WeightedVotingSystem) GetVotes() []*Vote {
	wvs.mu.RLock()
	defer wvs.mu.RUnlock()

	votes := make([]*Vote, len(wvs.votes))
	copy(votes, wvs.votes)
	return votes
}

// VoteCount returns the current number of votes.
func (wvs *WeightedVotingSystem) VoteCount() int {
	wvs.mu.RLock()
	defer wvs.mu.RUnlock()

	return len(wvs.votes)
}

// CalculateSimpleMajority calculates result using simple majority.
func (wvs *WeightedVotingSystem) CalculateSimpleMajority(ctx context.Context) (*VotingResult, error) {
	wvs.mu.RLock()
	defer wvs.mu.RUnlock()

	if len(wvs.votes) < wvs.config.MinimumVotes {
		return nil, fmt.Errorf("insufficient votes")
	}

	choiceCounts := make(map[string]int)
	for _, vote := range wvs.votes {
		choiceCounts[vote.Choice]++
	}

	// Find majority
	winningChoice := ""
	maxCount := 0
	for choice, count := range choiceCounts {
		if count > maxCount {
			maxCount = count
			winningChoice = choice
		}
	}

	// Convert counts to float64 for result
	choiceScores := make(map[string]float64)
	for choice, count := range choiceCounts {
		choiceScores[choice] = float64(count)
	}

	isMajority := float64(maxCount) > float64(len(wvs.votes))/2

	result := &VotingResult{
		WinningChoice:    winningChoice,
		WinningScore:     float64(maxCount),
		TotalVotes:       len(wvs.votes),
		ValidVotes:       len(wvs.votes),
		ChoiceScores:     choiceScores,
		ChoiceVoteCounts: choiceCounts,
		Consensus:        float64(maxCount) / float64(len(wvs.votes)),
		IsTie:            !isMajority && maxCount < len(wvs.votes)/2+1,
		Method:           VotingMethodMajority,
		Timestamp:        time.Now(),
	}

	return result, nil
}

// CalculateBordaCount calculates result using Borda count method.
func (wvs *WeightedVotingSystem) CalculateBordaCount(ctx context.Context, rankings map[string][]string) (*VotingResult, error) {
	if len(rankings) < wvs.config.MinimumVotes {
		return nil, fmt.Errorf("insufficient rankings")
	}

	// Count unique choices
	allChoices := make(map[string]bool)
	for _, ranking := range rankings {
		for _, choice := range ranking {
			allChoices[choice] = true
		}
	}
	n := len(allChoices)

	// Calculate Borda scores
	choiceScores := make(map[string]float64)
	for _, ranking := range rankings {
		for i, choice := range ranking {
			// Borda count: n-1 points for 1st, n-2 for 2nd, etc.
			points := float64(n - 1 - i)
			if points < 0 {
				points = 0
			}
			choiceScores[choice] += points
		}
	}

	// Find winner
	winningChoice := ""
	maxScore := 0.0
	for choice, score := range choiceScores {
		if score > maxScore {
			maxScore = score
			winningChoice = choice
		}
	}

	totalScore := 0.0
	for _, score := range choiceScores {
		totalScore += score
	}

	result := &VotingResult{
		WinningChoice:    winningChoice,
		WinningScore:     maxScore,
		TotalVotes:       len(rankings),
		ValidVotes:       len(rankings),
		ChoiceScores:     choiceScores,
		ChoiceVoteCounts: make(map[string]int),
		Consensus:        maxScore / totalScore,
		Method:           VotingMethodBorda,
		Timestamp:        time.Now(),
	}

	return result, nil
}

// SimulateProductiveChaos adds randomness to break groupthink.
// From MiniMax research on "Productive Chaos" in debates.
func (wvs *WeightedVotingSystem) SimulateProductiveChaos(chaosLevel float64) {
	if chaosLevel <= 0 || chaosLevel > 1 {
		return
	}

	wvs.mu.Lock()
	defer wvs.mu.Unlock()

	// Randomly adjust confidence values slightly
	for _, vote := range wvs.votes {
		// Add controlled noise
		noise := (math.Sin(float64(vote.Timestamp.UnixNano())) * chaosLevel * 0.1)
		vote.Confidence = math.Max(0, math.Min(1, vote.Confidence+noise))
	}
}

// GetVotingStatistics returns statistics about the voting.
type VotingStatistics struct {
	TotalVotes         int                `json:"total_votes"`
	UniqueChoices      int                `json:"unique_choices"`
	AvgConfidence      float64            `json:"avg_confidence"`
	ConfidenceStdDev   float64            `json:"confidence_std_dev"`
	ChoiceDistribution map[string]float64 `json:"choice_distribution"`
	SpecializationMix  map[string]int     `json:"specialization_mix"`
	RoleMix            map[string]int     `json:"role_mix"`
}

// GetStatistics returns voting statistics.
func (wvs *WeightedVotingSystem) GetStatistics() *VotingStatistics {
	wvs.mu.RLock()
	defer wvs.mu.RUnlock()

	stats := &VotingStatistics{
		TotalVotes:         len(wvs.votes),
		ChoiceDistribution: make(map[string]float64),
		SpecializationMix:  make(map[string]int),
		RoleMix:            make(map[string]int),
	}

	if len(wvs.votes) == 0 {
		return stats
	}

	// Calculate statistics
	choices := make(map[string]int)
	totalConfidence := 0.0

	for _, vote := range wvs.votes {
		choices[vote.Choice]++
		totalConfidence += vote.Confidence
		stats.SpecializationMix[vote.Specialization]++
		stats.RoleMix[vote.Role]++
	}

	stats.UniqueChoices = len(choices)
	stats.AvgConfidence = totalConfidence / float64(len(wvs.votes))

	// Calculate standard deviation
	varianceSum := 0.0
	for _, vote := range wvs.votes {
		diff := vote.Confidence - stats.AvgConfidence
		varianceSum += diff * diff
	}
	stats.ConfidenceStdDev = math.Sqrt(varianceSum / float64(len(wvs.votes)))

	// Calculate distribution
	for choice, count := range choices {
		stats.ChoiceDistribution[choice] = float64(count) / float64(len(wvs.votes))
	}

	return stats
}

// CondorcetMatrix holds pairwise comparison data.
type CondorcetMatrix struct {
	Candidates []string                  `json:"candidates"`
	Wins       map[string]map[string]int `json:"wins"` // wins[A][B] = voters preferring A over B
}

// CalculateCondorcet calculates result using Condorcet method.
// A Condorcet winner beats all other candidates in pairwise comparisons.
// Falls back to Borda count if no Condorcet winner exists (cycle).
func (wvs *WeightedVotingSystem) CalculateCondorcet(ctx context.Context, rankings map[string][]string) (*VotingResult, error) {
	if len(rankings) < wvs.config.MinimumVotes {
		return nil, fmt.Errorf("insufficient rankings: got %d, need %d",
			len(rankings), wvs.config.MinimumVotes)
	}

	// Build pairwise comparison matrix
	matrix := wvs.buildCondorcetMatrix(rankings)

	// Attempt to find a Condorcet winner
	winner, found := wvs.findCondorcetWinner(matrix)

	// If no Condorcet winner (cycle), fall back to Borda count
	if !found {
		bordaResult, err := wvs.CalculateBordaCount(ctx, rankings)
		if err != nil {
			return nil, fmt.Errorf("condorcet cycle detected and borda fallback failed: %w", err)
		}
		bordaResult.Method = VotingMethodCondorcet
		if bordaResult.Metadata == nil {
			bordaResult.Metadata = make(map[string]interface{})
		}
		bordaResult.Metadata["fallback_used"] = true
		bordaResult.Metadata["fallback_reason"] = "condorcet_cycle"
		return bordaResult, nil
	}

	// Calculate choice scores from pairwise wins for the result
	choiceScores := make(map[string]float64)
	choiceVoteCounts := make(map[string]int)
	for _, candidate := range matrix.Candidates {
		totalWins := 0
		for _, other := range matrix.Candidates {
			if candidate == other {
				continue
			}
			totalWins += matrix.Wins[candidate][other]
		}
		choiceScores[candidate] = float64(totalWins)
	}

	// Count first-place appearances for vote counts
	for _, ranking := range rankings {
		if len(ranking) > 0 {
			choiceVoteCounts[ranking[0]]++
		}
	}

	// Calculate consensus as winner's pairwise wins / total comparisons
	totalComparisons := 0.0
	winnerWins := 0.0
	for _, other := range matrix.Candidates {
		if winner == other {
			continue
		}
		winnerWins += float64(matrix.Wins[winner][other])
		totalComparisons += float64(matrix.Wins[winner][other] + matrix.Wins[other][winner])
	}
	consensus := 0.0
	if totalComparisons > 0 {
		consensus = winnerWins / totalComparisons
	}

	result := &VotingResult{
		WinningChoice:    winner,
		WinningScore:     choiceScores[winner],
		TotalVotes:       len(rankings),
		ValidVotes:       len(rankings),
		ChoiceScores:     choiceScores,
		ChoiceVoteCounts: choiceVoteCounts,
		Consensus:        consensus,
		Method:           VotingMethodCondorcet,
		Timestamp:        time.Now(),
	}

	return result, nil
}

// buildCondorcetMatrix builds a pairwise comparison matrix from rankings.
// For each pair (A, B), counts voters preferring A over B.
func (wvs *WeightedVotingSystem) buildCondorcetMatrix(rankings map[string][]string) *CondorcetMatrix {
	// Collect all unique candidates
	candidateSet := make(map[string]bool)
	for _, ranking := range rankings {
		for _, choice := range ranking {
			candidateSet[choice] = true
		}
	}

	candidates := make([]string, 0, len(candidateSet))
	for c := range candidateSet {
		candidates = append(candidates, c)
	}
	sort.Strings(candidates)

	// Initialize wins matrix
	wins := make(map[string]map[string]int)
	for _, c := range candidates {
		wins[c] = make(map[string]int)
	}

	// For each voter's ranking, compare every pair
	for _, ranking := range rankings {
		// Build position lookup for this voter
		position := make(map[string]int)
		for i, choice := range ranking {
			position[choice] = i
		}

		// Compare all pairs
		for i := 0; i < len(candidates); i++ {
			for j := i + 1; j < len(candidates); j++ {
				a := candidates[i]
				b := candidates[j]

				posA, aRanked := position[a]
				posB, bRanked := position[b]

				if aRanked && bRanked {
					// Both ranked: lower position index = higher preference
					if posA < posB {
						wins[a][b]++
					} else if posB < posA {
						wins[b][a]++
					}
				} else if aRanked && !bRanked {
					// A ranked, B not: A preferred over B
					wins[a][b]++
				} else if !aRanked && bRanked {
					// B ranked, A not: B preferred over A
					wins[b][a]++
				}
			}
		}
	}

	return &CondorcetMatrix{
		Candidates: candidates,
		Wins:       wins,
	}
}

// findCondorcetWinner finds a candidate that beats all others in pairwise comparison.
// Returns the winner and true if found, or empty string and false if a cycle exists.
func (wvs *WeightedVotingSystem) findCondorcetWinner(matrix *CondorcetMatrix) (string, bool) {
	for _, candidate := range matrix.Candidates {
		isWinner := true
		for _, other := range matrix.Candidates {
			if candidate == other {
				continue
			}
			// Condorcet winner must beat every other candidate head-to-head
			if matrix.Wins[candidate][other] <= matrix.Wins[other][candidate] {
				isWinner = false
				break
			}
		}
		if isWinner {
			return candidate, true
		}
	}
	return "", false
}

// CalculatePlurality calculates result using plurality (first-past-the-post).
// Winner is the choice with the most votes, regardless of majority.
func (wvs *WeightedVotingSystem) CalculatePlurality(ctx context.Context) (*VotingResult, error) {
	wvs.mu.RLock()
	defer wvs.mu.RUnlock()

	if len(wvs.votes) < wvs.config.MinimumVotes {
		return nil, fmt.Errorf("insufficient votes: got %d, need %d",
			len(wvs.votes), wvs.config.MinimumVotes)
	}

	choiceCounts := make(map[string]int)
	for _, vote := range wvs.votes {
		choiceCounts[vote.Choice]++
	}

	// Find the choice with the most votes
	winningChoice := ""
	maxCount := 0
	for choice, count := range choiceCounts {
		if count > maxCount {
			maxCount = count
			winningChoice = choice
		}
	}

	// Convert counts to float64 scores
	choiceScores := make(map[string]float64)
	for choice, count := range choiceCounts {
		choiceScores[choice] = float64(count)
	}

	// Check for tie (multiple choices with same max count)
	tieChoices := make([]string, 0)
	for choice, count := range choiceCounts {
		if count == maxCount {
			tieChoices = append(tieChoices, choice)
		}
	}
	sort.Strings(tieChoices)
	isTie := len(tieChoices) > 1

	// Handle tie breaking
	tieBreakUsed := false
	tieBreakMethod := TieBreakMethod("")
	if isTie && wvs.config.EnableTieBreaking {
		validVotes := wvs.filterVotes()
		voteWeights := make(map[string]*VoteWeight)
		for _, vote := range validVotes {
			voteWeights[vote.AgentID] = wvs.calculateVoteWeight(vote, validVotes)
		}
		winningChoice, tieBreakMethod = wvs.breakTie(tieChoices, wvs.votes, voteWeights)
		tieBreakUsed = true
		isTie = false
	}

	// Consensus = winner count / total votes
	consensus := float64(maxCount) / float64(len(wvs.votes))

	result := &VotingResult{
		WinningChoice:    winningChoice,
		WinningScore:     float64(maxCount),
		TotalVotes:       len(wvs.votes),
		ValidVotes:       len(wvs.votes),
		ChoiceScores:     choiceScores,
		ChoiceVoteCounts: choiceCounts,
		Consensus:        consensus,
		IsTie:            isTie,
		TieChoices:       tieChoices,
		TieBreakUsed:     tieBreakUsed,
		TieBreakMethod:   tieBreakMethod,
		Method:           VotingMethodPlurality,
		Timestamp:        time.Now(),
	}

	return result, nil
}

// CalculateUnanimous calculates result using unanimous agreement.
// All voters must agree on the same choice for consensus to be reached.
func (wvs *WeightedVotingSystem) CalculateUnanimous(ctx context.Context) (*VotingResult, error) {
	wvs.mu.RLock()
	defer wvs.mu.RUnlock()

	if len(wvs.votes) < wvs.config.MinimumVotes {
		return nil, fmt.Errorf("insufficient votes: got %d, need %d",
			len(wvs.votes), wvs.config.MinimumVotes)
	}

	choiceCounts := make(map[string]int)
	for _, vote := range wvs.votes {
		choiceCounts[vote.Choice]++
	}

	// Find the most-voted choice (useful for fallback even without unanimity)
	winningChoice := ""
	maxCount := 0
	for choice, count := range choiceCounts {
		if count > maxCount {
			maxCount = count
			winningChoice = choice
		}
	}

	// Convert counts to float64 scores
	choiceScores := make(map[string]float64)
	for choice, count := range choiceCounts {
		choiceScores[choice] = float64(count)
	}

	// Check for unanimity: all votes must be for the same choice
	isUnanimous := len(choiceCounts) == 1

	var consensus float64
	isTie := false
	var tieChoices []string

	if isUnanimous {
		consensus = 1.0
	} else {
		// No unanimity: consensus is the proportion of the most-voted choice
		consensus = float64(maxCount) / float64(len(wvs.votes))
		isTie = true
		tieChoices = make([]string, 0, len(choiceCounts))
		for choice := range choiceCounts {
			tieChoices = append(tieChoices, choice)
		}
		sort.Strings(tieChoices)
	}

	result := &VotingResult{
		WinningChoice:    winningChoice,
		WinningScore:     float64(maxCount),
		TotalVotes:       len(wvs.votes),
		ValidVotes:       len(wvs.votes),
		ChoiceScores:     choiceScores,
		ChoiceVoteCounts: choiceCounts,
		Consensus:        consensus,
		IsTie:            isTie,
		TieChoices:       tieChoices,
		Method:           VotingMethodUnanimous,
		Timestamp:        time.Now(),
	}

	return result, nil
}

// AutoSelectMethod selects the best voting method based on agent count and context.
func (wvs *WeightedVotingSystem) AutoSelectMethod(agentCount int) VotingMethod {
	if agentCount < 3 {
		return VotingMethodUnanimous
	}
	if agentCount <= 5 {
		return VotingMethodWeighted
	}
	return VotingMethodBorda
}
