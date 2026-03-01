package comprehensive

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMemoryEntry_Access(t *testing.T) {
	entry := NewMemoryEntry(MemoryTypeShortTerm, "agent-1", "Test content", 0.8)

	assert.Equal(t, 0, entry.AccessCount)

	entry.Access()
	assert.Equal(t, 1, entry.AccessCount)

	entry.Access()
	assert.Equal(t, 2, entry.AccessCount)
}

func TestMemoryEntry_RelevanceScore(t *testing.T) {
	entry := NewMemoryEntry(MemoryTypeLongTerm, "agent-1", "Important lesson", 0.9)

	score := entry.RelevanceScore()
	assert.Greater(t, score, 0.0)
	assert.LessOrEqual(t, score, 1.1) // Max is 1.0 + small access bonus
}

func TestShortTermMemory_Add(t *testing.T) {
	mem := NewShortTermMemory("agent-1", 5)

	entry := mem.Add("Message 1", nil)
	assert.NotNil(t, entry)
	assert.Equal(t, 1, mem.Size())

	// Add more entries
	mem.Add("Message 2", nil)
	mem.Add("Message 3", nil)
	assert.Equal(t, 3, mem.Size())
}

func TestShortTermMemory_MaxSize(t *testing.T) {
	mem := NewShortTermMemory("agent-1", 3)

	// Add more than max
	mem.Add("1", nil)
	mem.Add("2", nil)
	mem.Add("3", nil)
	mem.Add("4", nil)
	mem.Add("5", nil)

	// Should only keep last 3
	assert.Equal(t, 3, mem.Size())
}

func TestShortTermMemory_GetRecent(t *testing.T) {
	mem := NewShortTermMemory("agent-1", 10)

	mem.Add("First", nil)
	mem.Add("Second", nil)
	mem.Add("Third", nil)

	recent := mem.GetRecent(2)
	assert.Len(t, recent, 2)
	assert.Equal(t, "Second", recent[0].Content)
	assert.Equal(t, "Third", recent[1].Content)
}

func TestShortTermMemory_Clear(t *testing.T) {
	mem := NewShortTermMemory("agent-1", 10)

	mem.Add("Test", nil)
	assert.Equal(t, 1, mem.Size())

	mem.Clear()
	assert.Equal(t, 0, mem.Size())
}

func TestLongTermMemory_Store(t *testing.T) {
	mem := NewLongTermMemory("agent-1", 100)

	entry := mem.Store("Learned pattern A", 0.8, map[string]interface{}{"topic": "patterns"})
	assert.NotNil(t, entry)
	assert.Equal(t, "Learned pattern A", entry.Content)
	assert.Equal(t, 0.8, entry.Importance)
}

func TestLongTermMemory_Store_Duplicate(t *testing.T) {
	mem := NewLongTermMemory("agent-1", 100)

	mem.Store("Same lesson", 0.7, nil)
	mem.Store("Same lesson", 0.9, nil)

	// Should only have 1 entry with averaged importance
	all := mem.GetAll()
	assert.Len(t, all, 1)
	assert.Equal(t, 0.8, all[0].Importance) // Average of 0.7 and 0.9
}

func TestLongTermMemory_Retrieve(t *testing.T) {
	mem := NewLongTermMemory("agent-1", 100)

	mem.Store("Pattern for auth", 0.8, map[string]interface{}{"topic": "auth"})
	mem.Store("Pattern for db", 0.7, map[string]interface{}{"topic": "database"})
	mem.Store("Another auth tip", 0.9, map[string]interface{}{"topic": "auth"})

	results := mem.Retrieve("auth", 10)
	assert.Len(t, results, 2)
}

func TestLongTermMemory_Retrieve_Limit(t *testing.T) {
	mem := NewLongTermMemory("agent-1", 100)

	mem.Store("Lesson 1", 0.5, nil)
	mem.Store("Lesson 2", 0.6, nil)
	mem.Store("Lesson 3", 0.7, nil)

	results := mem.Retrieve("Lesson", 2)
	assert.Len(t, results, 2)
}

func TestLongTermMemory_Prune(t *testing.T) {
	mem := NewLongTermMemory("agent-1", 2)

	mem.Store("High importance", 0.9, nil)
	mem.Store("Medium importance", 0.5, nil)
	mem.Store("Low importance", 0.1, nil) // Should trigger prune

	all := mem.GetAll()
	assert.LessOrEqual(t, len(all), 2)
}

func TestEpisodicMemory_AddReflection(t *testing.T) {
	mem := NewEpisodicMemory("agent-1", 50)

	entry := mem.AddReflection(
		"Should validate inputs earlier",
		"Nil pointer exception",
		map[string]interface{}{"type": "runtime_error"},
	)

	assert.NotNil(t, entry)
	assert.Contains(t, entry.Content, "Should validate inputs earlier")
	assert.Contains(t, entry.Content, "Nil pointer exception")
}

func TestEpisodicMemory_GetRelevantReflections(t *testing.T) {
	mem := NewEpisodicMemory("agent-1", 50)

	mem.AddReflection("Fix error handling", "Panic occurred", nil)
	mem.AddReflection("Optimize loop", "Slow performance", nil)
	mem.AddReflection("Handle nil pointer", "Nil error", nil)

	results := mem.GetRelevantReflections("error", 10)
	assert.NotEmpty(t, results)
}

func TestMemoryManager_AddToShortTerm(t *testing.T) {
	mgr := NewMemoryManager("agent-1")

	mgr.AddToShortTerm("Recent message", nil)
	assert.Equal(t, 1, mgr.ShortTerm.Size())
}

func TestMemoryManager_StoreLesson(t *testing.T) {
	mgr := NewMemoryManager("agent-1")

	mgr.StoreLesson("Always validate input", 0.9, nil)
	assert.Equal(t, 1, len(mgr.LongTerm.GetAll()))
}

func TestMemoryManager_AddReflection(t *testing.T) {
	mgr := NewMemoryManager("agent-1")

	mgr.AddReflection("Need better error messages", "User confused", nil)
	assert.Equal(t, 1, len(mgr.Episodic.GetAll()))
}

func TestMemoryManager_GetContext(t *testing.T) {
	mgr := NewMemoryManager("agent-1")

	mgr.AddToShortTerm("Recent", nil)
	mgr.StoreLesson("Important lesson", 0.9, nil)
	mgr.AddReflection("Reflection", "Error", nil)

	ctx := mgr.GetContext("test")

	assert.NotNil(t, ctx["recent_history"])
	assert.NotNil(t, ctx["relevant_lessons"])
	assert.NotNil(t, ctx["relevant_reflections"])
}

func TestMemoryTypes(t *testing.T) {
	assert.Equal(t, MemoryType("short_term"), MemoryTypeShortTerm)
	assert.Equal(t, MemoryType("long_term"), MemoryTypeLongTerm)
	assert.Equal(t, MemoryType("episodic"), MemoryTypeEpisodic)
}
