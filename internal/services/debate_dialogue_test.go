package services

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestDialogueStyle tests dialogue style constants
func TestDialogueStyle(t *testing.T) {
	t.Run("All styles are defined", func(t *testing.T) {
		assert.Equal(t, DialogueStyle("theater"), StyleTheater)
		assert.Equal(t, DialogueStyle("novel"), StyleNovel)
		assert.Equal(t, DialogueStyle("screenplay"), StyleScript)
		assert.Equal(t, DialogueStyle("minimal"), StyleMinimal)
	})

	t.Run("Styles are unique", func(t *testing.T) {
		styles := []DialogueStyle{StyleTheater, StyleNovel, StyleScript, StyleMinimal}
		uniqueStyles := make(map[DialogueStyle]bool)
		for _, s := range styles {
			assert.False(t, uniqueStyles[s], "Style %s should be unique", s)
			uniqueStyles[s] = true
		}
	})
}

// TestDialogueEventType tests event type constants
func TestDialogueEventType(t *testing.T) {
	t.Run("All event types are defined", func(t *testing.T) {
		assert.Equal(t, DialogueEventType("prologue"), EventPrologue)
		assert.Equal(t, DialogueEventType("act_start"), EventActStart)
		assert.Equal(t, DialogueEventType("stage_action"), EventStageAction)
		assert.Equal(t, DialogueEventType("character_line"), EventCharacterLine)
		assert.Equal(t, DialogueEventType("act_end"), EventActEnd)
		assert.Equal(t, DialogueEventType("epilogue"), EventEpilogue)
		assert.Equal(t, DialogueEventType("consensus"), EventConsensus)
	})

	t.Run("Event types are unique", func(t *testing.T) {
		types := []DialogueEventType{
			EventPrologue, EventActStart, EventStageAction,
			EventCharacterLine, EventActEnd, EventEpilogue, EventConsensus,
		}
		uniqueTypes := make(map[DialogueEventType]bool)
		for _, et := range types {
			assert.False(t, uniqueTypes[et], "Event type %s should be unique", et)
			uniqueTypes[et] = true
		}
	})
}

// TestNewDialogueFormatter tests formatter creation
func TestNewDialogueFormatter(t *testing.T) {
	t.Run("Creates formatter with default style", func(t *testing.T) {
		df := NewDialogueFormatter("")
		require.NotNil(t, df)
		assert.Equal(t, StyleTheater, df.style)
		assert.NotNil(t, df.characters)
	})

	t.Run("Creates formatter with theater style", func(t *testing.T) {
		df := NewDialogueFormatter(StyleTheater)
		require.NotNil(t, df)
		assert.Equal(t, StyleTheater, df.style)
	})

	t.Run("Creates formatter with novel style", func(t *testing.T) {
		df := NewDialogueFormatter(StyleNovel)
		require.NotNil(t, df)
		assert.Equal(t, StyleNovel, df.style)
	})

	t.Run("Creates formatter with script style", func(t *testing.T) {
		df := NewDialogueFormatter(StyleScript)
		require.NotNil(t, df)
		assert.Equal(t, StyleScript, df.style)
	})

	t.Run("Creates formatter with minimal style", func(t *testing.T) {
		df := NewDialogueFormatter(StyleMinimal)
		require.NotNil(t, df)
		assert.Equal(t, StyleMinimal, df.style)
	})
}

// TestRegisterCharacter tests character registration
func TestRegisterCharacter(t *testing.T) {
	df := NewDialogueFormatter(StyleTheater)

	t.Run("Registers analyst character", func(t *testing.T) {
		member := &DebateTeamMember{
			Position:     PositionAnalyst,
			Role:         RoleAnalyst,
			ProviderName: "claude",
			ModelName:    ClaudeModels.SonnetLatest,
		}

		char := df.RegisterCharacter(member)
		require.NotNil(t, char)
		assert.Equal(t, "THE ANALYST", char.Name)
		assert.Equal(t, RoleAnalyst, char.Role)
		assert.Equal(t, PositionAnalyst, char.Position)
		assert.Equal(t, "claude", char.Provider)
		assert.Equal(t, "[A]", char.Avatar)
	})

	t.Run("Registers proposer character", func(t *testing.T) {
		member := &DebateTeamMember{
			Position:     PositionProposer,
			Role:         RoleProposer,
			ProviderName: "claude",
			ModelName:    ClaudeModels.Opus,
		}

		char := df.RegisterCharacter(member)
		require.NotNil(t, char)
		assert.Equal(t, "THE PROPOSER", char.Name)
		assert.Equal(t, "[P]", char.Avatar)
	})

	t.Run("Registers critic character", func(t *testing.T) {
		member := &DebateTeamMember{
			Position:     PositionCritic,
			Role:         RoleCritic,
			ProviderName: "deepseek",
			ModelName:    LLMsVerifierModels.DeepSeek,
		}

		char := df.RegisterCharacter(member)
		require.NotNil(t, char)
		assert.Equal(t, "THE CRITIC", char.Name)
		assert.Equal(t, "[C]", char.Avatar)
	})

	t.Run("Registers synthesis character", func(t *testing.T) {
		member := &DebateTeamMember{
			Position:     PositionSynthesis,
			Role:         RoleSynthesis,
			ProviderName: "gemini",
			ModelName:    LLMsVerifierModels.Gemini,
		}

		char := df.RegisterCharacter(member)
		require.NotNil(t, char)
		assert.Equal(t, "THE SYNTHESIZER", char.Name)
		assert.Equal(t, "[S]", char.Avatar)
	})

	t.Run("Registers mediator character", func(t *testing.T) {
		member := &DebateTeamMember{
			Position:     PositionMediator,
			Role:         RoleMediator,
			ProviderName: "mistral",
			ModelName:    LLMsVerifierModels.Mistral,
		}

		char := df.RegisterCharacter(member)
		require.NotNil(t, char)
		assert.Equal(t, "THE MEDIATOR", char.Name)
		assert.Equal(t, "[M]", char.Avatar)
	})

	t.Run("Stores character in map", func(t *testing.T) {
		member := &DebateTeamMember{
			Position: PositionAnalyst,
			Role:     RoleAnalyst,
		}

		df.RegisterCharacter(member)
		char := df.characters[PositionAnalyst]
		require.NotNil(t, char)
		assert.Equal(t, "THE ANALYST", char.Name)
	})
}

// TestDialogueCharacter tests character struct
func TestDialogueCharacter(t *testing.T) {
	t.Run("Character with all fields", func(t *testing.T) {
		char := &DialogueCharacter{
			Name:     "THE ANALYST",
			Role:     RoleAnalyst,
			Position: PositionAnalyst,
			Provider: "claude",
			Model:    ClaudeModels.SonnetLatest,
			Avatar:   "[A]",
		}

		assert.Equal(t, "THE ANALYST", char.Name)
		assert.Equal(t, RoleAnalyst, char.Role)
		assert.Equal(t, PositionAnalyst, char.Position)
		assert.Equal(t, "claude", char.Provider)
		assert.Equal(t, ClaudeModels.SonnetLatest, char.Model)
		assert.Equal(t, "[A]", char.Avatar)
	})
}

// TestDialogueLine tests dialogue line struct
func TestDialogueLine(t *testing.T) {
	t.Run("Line with all fields", func(t *testing.T) {
		char := &DialogueCharacter{Name: "THE ANALYST", Avatar: "[A]"}
		now := time.Now()

		line := &DialogueLine{
			Character:   char,
			Content:     "This is my analysis of the situation.",
			Timestamp:   now,
			RoundNumber: 1,
			Action:      "(steps forward)",
			Emotion:     "thoughtful",
		}

		assert.Equal(t, char, line.Character)
		assert.Equal(t, "This is my analysis of the situation.", line.Content)
		assert.Equal(t, now, line.Timestamp)
		assert.Equal(t, 1, line.RoundNumber)
		assert.Equal(t, "(steps forward)", line.Action)
		assert.Equal(t, "thoughtful", line.Emotion)
	})
}

// TestDialogueAct tests dialogue act struct
func TestDialogueAct(t *testing.T) {
	t.Run("Act with all fields", func(t *testing.T) {
		now := time.Now()
		line := &DialogueLine{Content: "Test line"}

		act := &DialogueAct{
			ActNumber:   1,
			Title:       "ACT I: THE OPENING",
			Description: "The council convenes.",
			Lines:       []*DialogueLine{line},
			StartTime:   now,
			EndTime:     now.Add(time.Minute),
		}

		assert.Equal(t, 1, act.ActNumber)
		assert.Equal(t, "ACT I: THE OPENING", act.Title)
		assert.Equal(t, "The council convenes.", act.Description)
		assert.Len(t, act.Lines, 1)
		assert.Equal(t, now, act.StartTime)
	})
}

// TestCreateDialogue tests dialogue creation
func TestCreateDialogue(t *testing.T) {
	df := NewDialogueFormatter(StyleTheater)

	// Register characters
	df.RegisterCharacter(&DebateTeamMember{Position: PositionAnalyst, Role: RoleAnalyst, ProviderName: "claude"})
	df.RegisterCharacter(&DebateTeamMember{Position: PositionProposer, Role: RoleProposer, ProviderName: "claude"})

	t.Run("Creates dialogue with topic", func(t *testing.T) {
		rounds := []DebateRound{
			{
				RoundNumber: 1,
				Responses: []DebateResponse{
					{Position: PositionAnalyst, Role: "analyst", Content: "Analysis here", Timestamp: time.Now()},
				},
				StartTime: time.Now(),
				EndTime:   time.Now().Add(time.Minute),
			},
		}

		dialogue := df.CreateDialogue("Test Topic", rounds)
		require.NotNil(t, dialogue)
		assert.Equal(t, "THE GREAT DEBATE", dialogue.Title)
		assert.Equal(t, "Test Topic", dialogue.Subtitle)
		assert.Equal(t, StyleTheater, dialogue.Style)
		assert.Equal(t, 1, dialogue.TotalRounds)
	})

	t.Run("Creates prologue", func(t *testing.T) {
		dialogue := df.CreateDialogue("Test Topic", []DebateRound{})
		assert.Contains(t, dialogue.Prologue, "THE GREAT AI DEBATE")
		assert.Contains(t, dialogue.Prologue, "DRAMATIS PERSONAE")
	})

	t.Run("Creates epilogue", func(t *testing.T) {
		dialogue := df.CreateDialogue("Test Topic", []DebateRound{})
		assert.Contains(t, dialogue.Epilogue, "CONSENSUS")
	})

	t.Run("Creates acts from rounds", func(t *testing.T) {
		rounds := []DebateRound{
			{RoundNumber: 1, Responses: []DebateResponse{}, StartTime: time.Now()},
			{RoundNumber: 2, Responses: []DebateResponse{}, StartTime: time.Now()},
			{RoundNumber: 3, Responses: []DebateResponse{}, StartTime: time.Now()},
		}

		dialogue := df.CreateDialogue("Test Topic", rounds)
		assert.Len(t, dialogue.Acts, 3)
	})
}

// TestFormatAsText tests text formatting
func TestFormatAsText(t *testing.T) {
	df := NewDialogueFormatter(StyleTheater)

	df.RegisterCharacter(&DebateTeamMember{
		Position:     PositionAnalyst,
		Role:         RoleAnalyst,
		ProviderName: "claude",
		ModelName:    "claude-3-sonnet",
	})

	t.Run("Formats complete dialogue", func(t *testing.T) {
		rounds := []DebateRound{
			{
				RoundNumber: 1,
				Responses: []DebateResponse{
					{
						Position:  PositionAnalyst,
						Role:      "analyst",
						Provider:  "claude",
						Content:   "This is my analysis.",
						Timestamp: time.Now(),
					},
				},
				StartTime: time.Now(),
				EndTime:   time.Now().Add(time.Minute),
			},
		}

		dialogue := df.CreateDialogue("Test Topic", rounds)
		text := df.FormatAsText(dialogue)

		assert.Contains(t, text, "THE GREAT AI DEBATE")
		assert.Contains(t, text, "THE ANALYST")
		assert.Contains(t, text, "This is my analysis.")
	})

	t.Run("Includes character list", func(t *testing.T) {
		dialogue := df.CreateDialogue("Test", []DebateRound{})
		text := df.FormatAsText(dialogue)

		assert.Contains(t, text, "[A] THE ANALYST")
	})
}

// TestFormatTheaterLine tests theater style formatting
func TestFormatTheaterLine(t *testing.T) {
	df := NewDialogueFormatter(StyleTheater)

	t.Run("Formats line with action", func(t *testing.T) {
		line := &DialogueLine{
			Character: &DialogueCharacter{Name: "THE ANALYST", Avatar: "[A]"},
			Content:   "This is my point.",
			Action:    "(steps forward)",
		}

		result := df.formatTheaterLine(line)
		assert.Contains(t, result, "(steps forward)")
		assert.Contains(t, result, "[A] THE ANALYST:")
		assert.Contains(t, result, "This is my point.")
	})

	t.Run("Formats line without action", func(t *testing.T) {
		line := &DialogueLine{
			Character: &DialogueCharacter{Name: "THE CRITIC", Avatar: "[C]"},
			Content:   "I disagree.",
		}

		result := df.formatTheaterLine(line)
		assert.NotContains(t, result, "(")
		assert.Contains(t, result, "[C] THE CRITIC:")
	})
}

// TestFormatNovelLine tests novel style formatting
func TestFormatNovelLine(t *testing.T) {
	df := NewDialogueFormatter(StyleNovel)

	t.Run("Formats line in novel style", func(t *testing.T) {
		line := &DialogueLine{
			Character: &DialogueCharacter{Name: "THE PROPOSER"},
			Content:   "I propose we consider this",
		}

		result := df.formatNovelLine(line)
		assert.Contains(t, result, "said THE PROPOSER")
	})
}

// TestFormatScriptLine tests screenplay style formatting
func TestFormatScriptLine(t *testing.T) {
	df := NewDialogueFormatter(StyleScript)

	t.Run("Formats line in screenplay style", func(t *testing.T) {
		line := &DialogueLine{
			Character: &DialogueCharacter{Name: "THE MEDIATOR"},
			Content:   "Let us find common ground.",
			Action:    "(calmly)",
		}

		result := df.formatScriptLine(line)
		assert.Contains(t, result, "THE MEDIATOR")
		assert.Contains(t, result, "(calmly)")
	})
}

// TestFormatMinimalLine tests minimal style formatting
func TestFormatMinimalLine(t *testing.T) {
	df := NewDialogueFormatter(StyleMinimal)

	t.Run("Formats line in minimal style", func(t *testing.T) {
		line := &DialogueLine{
			Character: &DialogueCharacter{Name: "THE ANALYST"},
			Content:   "Simple point.",
		}

		result := df.formatMinimalLine(line)
		assert.Equal(t, "[THE ANALYST] Simple point.\n\n", result)
	})
}

// TestFormatForStreaming tests streaming format
func TestFormatForStreaming(t *testing.T) {
	df := NewDialogueFormatter(StyleTheater)

	t.Run("Returns multiple lines for streaming", func(t *testing.T) {
		line := &DialogueLine{
			Character: &DialogueCharacter{Name: "THE ANALYST", Avatar: "[A]"},
			Content:   "This is a longer piece of content that should be split into multiple lines for streaming effect.",
			Action:    "(thoughtfully)",
		}

		lines := df.FormatForStreaming(line)
		assert.Greater(t, len(lines), 1)
		assert.Contains(t, lines[0], "(thoughtfully)")
	})

	t.Run("Includes character header", func(t *testing.T) {
		line := &DialogueLine{
			Character: &DialogueCharacter{Name: "THE CRITIC", Avatar: "[C]"},
			Content:   "Short.",
		}

		lines := df.FormatForStreaming(line)
		found := false
		for _, l := range lines {
			if strings.Contains(l, "[C] THE CRITIC:") {
				found = true
				break
			}
		}
		assert.True(t, found, "Should contain character header")
	})
}

// TestWrapText tests text wrapping
func TestWrapText(t *testing.T) {
	df := NewDialogueFormatter(StyleTheater)

	t.Run("Wraps long text", func(t *testing.T) {
		text := "This is a very long sentence that should definitely be wrapped because it exceeds the maximum width"
		result := df.wrapText(text, 30)

		lines := strings.Split(result, "\n")
		assert.Greater(t, len(lines), 1)
		for _, line := range lines {
			assert.LessOrEqual(t, len(line), 30)
		}
	})

	t.Run("Does not wrap short text", func(t *testing.T) {
		text := "Short text"
		result := df.wrapText(text, 50)
		assert.Equal(t, text, result)
	})

	t.Run("Handles empty text", func(t *testing.T) {
		result := df.wrapText("", 50)
		assert.Equal(t, "", result)
	})
}

// TestTruncateString tests string truncation
func TestTruncateString(t *testing.T) {
	df := NewDialogueFormatter(StyleTheater)

	t.Run("Truncates long string", func(t *testing.T) {
		text := "This is a very long string that needs truncation"
		result := df.truncateString(text, 20)
		assert.Equal(t, "This is a very lo...", result)
		assert.Len(t, result, 20)
	})

	t.Run("Does not truncate short string", func(t *testing.T) {
		text := "Short"
		result := df.truncateString(text, 20)
		assert.Equal(t, text, result)
	})

	t.Run("Handles exact length", func(t *testing.T) {
		text := "Exactly twenty char!"
		result := df.truncateString(text, 20)
		assert.Equal(t, text, result)
	})
}

// TestGetActTitle tests act title generation
func TestGetActTitle(t *testing.T) {
	df := NewDialogueFormatter(StyleTheater)

	testCases := []struct {
		actNumber int
		expected  string
	}{
		{1, "ACT I: THE OPENING"},
		{2, "ACT II: THE CHALLENGE"},
		{3, "ACT III: THE SYNTHESIS"},
		{4, "ACT IV: THE REFINEMENT"},
		{5, "ACT V: THE RESOLUTION"},
		{6, "ACT 6"}, // Fallback
	}

	for _, tc := range testCases {
		t.Run(tc.expected, func(t *testing.T) {
			result := df.getActTitle(tc.actNumber)
			assert.Equal(t, tc.expected, result)
		})
	}
}

// TestGetActDescription tests act description generation
func TestGetActDescription(t *testing.T) {
	df := NewDialogueFormatter(StyleTheater)

	t.Run("Returns description for act 1", func(t *testing.T) {
		desc := df.getActDescription(1)
		assert.Contains(t, desc, "convenes")
	})

	t.Run("Returns description for act 2", func(t *testing.T) {
		desc := df.getActDescription(2)
		assert.Contains(t, desc, "Critic")
	})

	t.Run("Returns fallback for unknown act", func(t *testing.T) {
		desc := df.getActDescription(10)
		assert.Contains(t, desc, "continues")
	})
}

// TestDefaultStreamingConfig tests default config
func TestDefaultStreamingConfig(t *testing.T) {
	config := DefaultStreamingConfig()

	t.Run("Has correct defaults", func(t *testing.T) {
		assert.Equal(t, StyleTheater, config.Style)
		assert.True(t, config.ShowPrologue)
		assert.True(t, config.ShowActHeaders)
		assert.True(t, config.ShowStageActions)
		assert.True(t, config.ShowCharacterInfo)
		assert.Equal(t, 50, config.ChunkDelay)
	})
}

// TestDialogueStream tests streaming functionality
func TestDialogueStream(t *testing.T) {
	t.Run("Creates stream with default config", func(t *testing.T) {
		stream := NewDialogueStream(nil)
		require.NotNil(t, stream)
		assert.NotNil(t, stream.Events())
		assert.NotNil(t, stream.Done())
	})

	t.Run("Creates stream with custom config", func(t *testing.T) {
		config := &StreamingDialogueConfig{Style: StyleMinimal}
		stream := NewDialogueStream(config)
		require.NotNil(t, stream)
		assert.Equal(t, StyleMinimal, stream.config.Style)
	})

	t.Run("Sends and receives events", func(t *testing.T) {
		stream := NewDialogueStream(nil)

		go func() {
			stream.SendEvent(&DialogueEvent{
				Type:    EventPrologue,
				Content: "Test prologue",
			})
			stream.Close()
		}()

		event := <-stream.Events()
		assert.Equal(t, EventPrologue, event.Type)
		assert.Equal(t, "Test prologue", event.Content)
	})
}

// TestFormatEventAsText tests event formatting
func TestFormatEventAsText(t *testing.T) {
	t.Run("Formats prologue event", func(t *testing.T) {
		event := &DialogueEvent{Type: EventPrologue, Content: "Welcome"}
		result := FormatEventAsText(event)
		assert.Equal(t, "Welcome", result)
	})

	t.Run("Formats act start event", func(t *testing.T) {
		event := &DialogueEvent{Type: EventActStart, Content: "ACT I"}
		result := FormatEventAsText(event)
		assert.Contains(t, result, "ACT I")
		assert.Contains(t, result, "─")
	})

	t.Run("Formats stage action event", func(t *testing.T) {
		event := &DialogueEvent{Type: EventStageAction, Content: "(enters stage)"}
		result := FormatEventAsText(event)
		assert.Contains(t, result, "(enters stage)")
	})

	t.Run("Formats character line event", func(t *testing.T) {
		event := &DialogueEvent{
			Type:      EventCharacterLine,
			Content:   "Hello world",
			Character: &DialogueCharacter{Name: "TEST", Avatar: "[T]"},
		}
		result := FormatEventAsText(event)
		assert.Contains(t, result, "[T] TEST:")
		assert.Contains(t, result, "Hello world")
	})

	t.Run("Formats consensus event", func(t *testing.T) {
		event := &DialogueEvent{Type: EventConsensus, Content: "Final answer"}
		result := FormatEventAsText(event)
		assert.Contains(t, result, "CONSENSUS")
		assert.Contains(t, result, "Final answer")
	})
}

// TestDebateRound tests debate round struct
func TestDebateRound(t *testing.T) {
	t.Run("Round with all fields", func(t *testing.T) {
		now := time.Now()
		round := DebateRound{
			RoundNumber: 1,
			Responses: []DebateResponse{
				{Position: PositionAnalyst, Content: "Analysis"},
			},
			StartTime: now,
			EndTime:   now.Add(time.Minute),
		}

		assert.Equal(t, 1, round.RoundNumber)
		assert.Len(t, round.Responses, 1)
		assert.Equal(t, now, round.StartTime)
	})
}

// TestDebateResponse tests debate response struct
func TestDebateResponse(t *testing.T) {
	t.Run("Response with all fields", func(t *testing.T) {
		now := time.Now()
		response := DebateResponse{
			Position:   PositionCritic,
			Role:       "critic",
			Provider:   "deepseek",
			Model:      "deepseek-chat",
			Content:    "I challenge this assertion.",
			Timestamp:  now,
			Confidence: 0.85,
		}

		assert.Equal(t, PositionCritic, response.Position)
		assert.Equal(t, "critic", response.Role)
		assert.Equal(t, "deepseek", response.Provider)
		assert.Equal(t, "deepseek-chat", response.Model)
		assert.Equal(t, "I challenge this assertion.", response.Content)
		assert.Equal(t, now, response.Timestamp)
		assert.Equal(t, 0.85, response.Confidence)
	})
}

// TestDialogueDebate integration tests
func TestDialogueDebateIntegration(t *testing.T) {
	t.Run("Complete debate dialogue flow", func(t *testing.T) {
		df := NewDialogueFormatter(StyleTheater)

		// Register all 5 characters
		members := []*DebateTeamMember{
			{Position: PositionAnalyst, Role: RoleAnalyst, ProviderName: "claude", ModelName: ClaudeModels.SonnetLatest},
			{Position: PositionProposer, Role: RoleProposer, ProviderName: "claude", ModelName: ClaudeModels.Opus},
			{Position: PositionCritic, Role: RoleCritic, ProviderName: "deepseek", ModelName: LLMsVerifierModels.DeepSeek},
			{Position: PositionSynthesis, Role: RoleSynthesis, ProviderName: "gemini", ModelName: LLMsVerifierModels.Gemini},
			{Position: PositionMediator, Role: RoleMediator, ProviderName: "mistral", ModelName: LLMsVerifierModels.Mistral},
		}

		for _, m := range members {
			df.RegisterCharacter(m)
		}

		// Create 3 rounds of debate
		rounds := make([]DebateRound, 3)
		for i := 0; i < 3; i++ {
			responses := make([]DebateResponse, 5)
			for j, m := range members {
				responses[j] = DebateResponse{
					Position:   m.Position,
					Role:       string(m.Role),
					Provider:   m.ProviderName,
					Model:      m.ModelName,
					Content:    "Response from " + string(m.Role) + " in round " + string(rune('1'+i)),
					Timestamp:  time.Now(),
					Confidence: 0.9,
				}
			}
			rounds[i] = DebateRound{
				RoundNumber: i + 1,
				Responses:   responses,
				StartTime:   time.Now(),
				EndTime:     time.Now().Add(time.Minute),
			}
		}

		// Create dialogue
		dialogue := df.CreateDialogue("What is the meaning of AI?", rounds)

		// Verify dialogue structure
		assert.Equal(t, "THE GREAT DEBATE", dialogue.Title)
		assert.Equal(t, 3, dialogue.TotalRounds)
		assert.Len(t, dialogue.Acts, 3)
		assert.Len(t, dialogue.Characters, 5)

		// Verify each act has lines
		for _, act := range dialogue.Acts {
			assert.Greater(t, len(act.Lines), 0)
		}

		// Format as text
		text := df.FormatAsText(dialogue)
		assert.NotEmpty(t, text)
		assert.Contains(t, text, "THE ANALYST")
		assert.Contains(t, text, "THE PROPOSER")
		assert.Contains(t, text, "THE CRITIC")
		assert.Contains(t, text, "THE SYNTHESIZER")
		assert.Contains(t, text, "THE MEDIATOR")
	})

	t.Run("All 15 LLMs represented in dialogue", func(t *testing.T) {
		df := NewDialogueFormatter(StyleTheater)

		// All 15 LLMs (5 positions × 3 levels)
		allLLMs := []struct {
			Position DebateTeamPosition
			Role     DebateRole
			Provider string
			Model    string
			Level    string
		}{
			// Position 1
			{PositionAnalyst, RoleAnalyst, "claude", ClaudeModels.SonnetLatest, "primary"},
			{PositionAnalyst, RoleAnalyst, "groq", LLMsVerifierModels.Groq, "fallback1"},
			{PositionAnalyst, RoleAnalyst, "qwen", QwenModels.Max, "fallback2"},
			// Position 2
			{PositionProposer, RoleProposer, "claude", ClaudeModels.Opus, "primary"},
			{PositionProposer, RoleProposer, "cerebras", LLMsVerifierModels.Cerebras, "fallback1"},
			{PositionProposer, RoleProposer, "qwen", QwenModels.Plus, "fallback2"},
			// Position 3
			{PositionCritic, RoleCritic, "deepseek", LLMsVerifierModels.DeepSeek, "primary"},
			{PositionCritic, RoleCritic, "claude", ClaudeModels.Haiku, "fallback1"},
			{PositionCritic, RoleCritic, "qwen", QwenModels.Turbo, "fallback2"},
			// Position 4
			{PositionSynthesis, RoleSynthesis, "gemini", LLMsVerifierModels.Gemini, "primary"},
			{PositionSynthesis, RoleSynthesis, "claude", ClaudeModels.Haiku, "fallback1"},
			{PositionSynthesis, RoleSynthesis, "qwen", QwenModels.Coder, "fallback2"},
			// Position 5
			{PositionMediator, RoleMediator, "mistral", LLMsVerifierModels.Mistral, "primary"},
			{PositionMediator, RoleMediator, "claude", ClaudeModels.Haiku, "fallback1"},
			{PositionMediator, RoleMediator, "qwen", QwenModels.Long, "fallback2"},
		}

		assert.Len(t, allLLMs, 15, "Should have exactly 15 LLMs defined")

		// Register primary characters
		for _, llm := range allLLMs {
			if llm.Level == "primary" {
				df.RegisterCharacter(&DebateTeamMember{
					Position:     llm.Position,
					Role:         llm.Role,
					ProviderName: llm.Provider,
					ModelName:    llm.Model,
				})
			}
		}

		assert.Len(t, df.characters, 5, "Should have 5 primary characters registered")
	})
}

// TestDialogueStyles tests all formatting styles
func TestDialogueStyles(t *testing.T) {
	styles := []DialogueStyle{StyleTheater, StyleNovel, StyleScript, StyleMinimal}

	for _, style := range styles {
		t.Run(string(style), func(t *testing.T) {
			df := NewDialogueFormatter(style)
			df.RegisterCharacter(&DebateTeamMember{
				Position:     PositionAnalyst,
				Role:         RoleAnalyst,
				ProviderName: "claude",
			})

			line := &DialogueLine{
				Character: &DialogueCharacter{Name: "THE ANALYST", Avatar: "[A]"},
				Content:   "Test content",
				Action:    "(test action)",
			}

			result := df.formatLine(line)
			assert.NotEmpty(t, result)
		})
	}
}
