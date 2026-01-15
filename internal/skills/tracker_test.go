package skills

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTracker_StartAndComplete(t *testing.T) {
	tracker := NewTracker()

	skill := createTestSkill("test-skill", "test-category", []string{"test trigger"})
	match := &SkillMatch{
		Skill:          skill,
		Confidence:     0.95,
		MatchedTrigger: "test trigger",
		MatchType:      MatchTypeExact,
	}

	// Start tracking
	usage := tracker.StartTracking("request-123", skill, match)
	require.NotNil(t, usage)
	assert.Equal(t, "test-skill", usage.SkillName)
	assert.Equal(t, "test-category", usage.Category)
	assert.Equal(t, 0.95, usage.Confidence)
	assert.Equal(t, MatchTypeExact, usage.MatchType)
	assert.False(t, usage.StartedAt.IsZero())

	// Verify active
	active := tracker.GetActiveUsage("request-123")
	require.NotNil(t, active)
	assert.Equal(t, "test-skill", active.SkillName)

	// Record tool use
	tracker.RecordToolUse("request-123", "Read")
	tracker.RecordToolUse("request-123", "Write")
	active = tracker.GetActiveUsage("request-123")
	assert.Len(t, active.ToolsInvoked, 2)

	// Complete tracking
	completed := tracker.CompleteTracking("request-123", true, "")
	require.NotNil(t, completed)
	assert.True(t, completed.Success)
	assert.Empty(t, completed.Error)
	assert.False(t, completed.CompletedAt.IsZero())

	// Verify no longer active
	active = tracker.GetActiveUsage("request-123")
	assert.Nil(t, active)
}

func TestTracker_CompleteWithError(t *testing.T) {
	tracker := NewTracker()

	skill := createTestSkill("failing-skill", "test", []string{"fail"})
	match := &SkillMatch{Skill: skill, Confidence: 0.8, MatchType: MatchTypeFuzzy}

	tracker.StartTracking("fail-request", skill, match)
	completed := tracker.CompleteTracking("fail-request", false, "execution failed")

	require.NotNil(t, completed)
	assert.False(t, completed.Success)
	assert.Equal(t, "execution failed", completed.Error)
}

func TestTracker_Stats(t *testing.T) {
	tracker := NewTracker()

	// Execute multiple skills
	for i := 0; i < 10; i++ {
		skill := createTestSkill("skill-a", "devops", []string{"trigger"})
		match := &SkillMatch{Skill: skill, Confidence: 0.9, MatchType: MatchTypeExact}
		requestID := "request-a-" + string(rune('0'+i))
		tracker.StartTracking(requestID, skill, match)
		tracker.CompleteTracking(requestID, true, "")
	}

	for i := 0; i < 5; i++ {
		skill := createTestSkill("skill-b", "security", []string{"sec"})
		match := &SkillMatch{Skill: skill, Confidence: 0.8, MatchType: MatchTypePartial}
		requestID := "request-b-" + string(rune('0'+i))
		tracker.StartTracking(requestID, skill, match)
		success := i%2 == 0 // Half succeed
		tracker.CompleteTracking(requestID, success, "")
	}

	stats := tracker.GetStats()
	assert.Equal(t, int64(15), stats.TotalInvocations)
	assert.Equal(t, int64(13), stats.SuccessfulCount) // 10 + 3 (half of 5)
	assert.Equal(t, int64(2), stats.FailedCount)

	// Check by skill
	skillAStats := stats.BySkill["skill-a"]
	require.NotNil(t, skillAStats)
	assert.Equal(t, int64(10), skillAStats.InvocationCount)

	// Check by category
	devopsStats := stats.ByCategory["devops"]
	require.NotNil(t, devopsStats)
	assert.Equal(t, int64(10), devopsStats.InvocationCount)

	// Check by match type
	assert.Equal(t, int64(10), stats.ByMatchType[MatchTypeExact])
	assert.Equal(t, int64(5), stats.ByMatchType[MatchTypePartial])
}

func TestTracker_GetTopSkills(t *testing.T) {
	tracker := NewTracker()

	// Create skills with different invocation counts
	skillA := createTestSkill("popular-skill", "cat", []string{"t"})
	skillB := createTestSkill("less-popular", "cat", []string{"t2"})

	for i := 0; i < 10; i++ {
		match := &SkillMatch{Skill: skillA, Confidence: 0.9}
		requestID := "pop-" + string(rune('0'+i))
		tracker.StartTracking(requestID, skillA, match)
		tracker.CompleteTracking(requestID, true, "")
	}

	for i := 0; i < 3; i++ {
		match := &SkillMatch{Skill: skillB, Confidence: 0.8}
		requestID := "less-" + string(rune('0'+i))
		tracker.StartTracking(requestID, skillB, match)
		tracker.CompleteTracking(requestID, true, "")
	}

	top := tracker.GetTopSkills(2)
	require.Len(t, top, 2)
	assert.Equal(t, "popular-skill", top[0].Name)
	assert.Equal(t, int64(10), top[0].InvocationCount)
}

func TestTracker_History(t *testing.T) {
	tracker := NewTracker()

	// Execute some skills
	for i := 0; i < 20; i++ {
		skill := createTestSkill("history-skill", "cat", []string{"t"})
		match := &SkillMatch{Skill: skill, Confidence: 0.9}
		requestID := "hist-" + string(rune('0'+i%10)) + string(rune('a'+i/10))
		tracker.StartTracking(requestID, skill, match)
		time.Sleep(time.Millisecond) // Ensure ordering
		tracker.CompleteTracking(requestID, true, "")
	}

	// Get recent history
	history := tracker.GetHistory(5)
	assert.Len(t, history, 5)

	// Get all history
	allHistory := tracker.GetHistory(0)
	assert.Len(t, allHistory, 20)
}

func TestTracker_Reset(t *testing.T) {
	tracker := NewTracker()

	// Add some data
	skill := createTestSkill("resetable", "cat", []string{"t"})
	match := &SkillMatch{Skill: skill, Confidence: 0.9}
	tracker.StartTracking("reset-test", skill, match)
	tracker.CompleteTracking("reset-test", true, "")

	stats := tracker.GetStats()
	assert.Greater(t, stats.TotalInvocations, int64(0))

	// Reset
	tracker.Reset()

	stats = tracker.GetStats()
	assert.Equal(t, int64(0), stats.TotalInvocations)
}

func TestTracker_ActiveUsages(t *testing.T) {
	tracker := NewTracker()

	// Start multiple active executions
	for i := 0; i < 5; i++ {
		skill := createTestSkill("active-skill-"+string(rune('a'+i)), "cat", []string{"t"})
		match := &SkillMatch{Skill: skill, Confidence: 0.9}
		tracker.StartTracking("active-"+string(rune('0'+i)), skill, match)
	}

	active := tracker.GetActiveUsages()
	assert.Len(t, active, 5)

	// Complete some
	tracker.CompleteTracking("active-0", true, "")
	tracker.CompleteTracking("active-1", true, "")

	active = tracker.GetActiveUsages()
	assert.Len(t, active, 3)
}

func TestTracker_SkillStats(t *testing.T) {
	tracker := NewTracker()

	skill := createTestSkill("stats-skill", "cat", []string{"trigger"})

	// Multiple executions with different triggers
	triggers := []string{"trigger", "other-trigger", "trigger"}
	for i, trigger := range triggers {
		match := &SkillMatch{
			Skill:          skill,
			Confidence:     0.8 + float64(i)*0.05,
			MatchedTrigger: trigger,
		}
		requestID := "stats-" + string(rune('0'+i))
		tracker.StartTracking(requestID, skill, match)
		time.Sleep(10 * time.Millisecond)
		tracker.CompleteTracking(requestID, true, "")
	}

	skillStats := tracker.GetSkillStats("stats-skill")
	require.NotNil(t, skillStats)
	assert.Equal(t, int64(3), skillStats.InvocationCount)
	assert.Equal(t, int64(3), skillStats.SuccessCount)
	assert.Equal(t, int64(2), skillStats.TriggersCounted["trigger"])
	assert.Equal(t, int64(1), skillStats.TriggersCounted["other-trigger"])
}
