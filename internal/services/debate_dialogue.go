package services

import (
	"fmt"
	"strings"
	"time"
)

// DialogueStyle represents the presentation style for debate dialogue
type DialogueStyle string

const (
	// StyleTheater presents dialogue like a theater script
	StyleTheater DialogueStyle = "theater"
	// StyleNovel presents dialogue like a novel
	StyleNovel DialogueStyle = "novel"
	// StyleScript presents dialogue like a screenplay
	StyleScript DialogueStyle = "screenplay"
	// StyleMinimal presents dialogue in minimal format
	StyleMinimal DialogueStyle = "minimal"
)

// DialogueCharacter represents a character in the debate dialogue
type DialogueCharacter struct {
	Name     string             `json:"name"`
	Role     DebateRole         `json:"role"`
	Position DebateTeamPosition `json:"position"`
	Provider string             `json:"provider"`
	Model    string             `json:"model"`
	Avatar   string             `json:"avatar"` // Emoji or icon
}

// DialogueLine represents a single line of dialogue
type DialogueLine struct {
	Character   *DialogueCharacter `json:"character"`
	Content     string             `json:"content"`
	Timestamp   time.Time          `json:"timestamp"`
	RoundNumber int                `json:"round_number"`
	Action      string             `json:"action,omitempty"` // Stage direction
	Emotion     string             `json:"emotion,omitempty"`
}

// DialogueAct represents an act (round) in the debate
type DialogueAct struct {
	ActNumber   int             `json:"act_number"`
	Title       string          `json:"title"`
	Description string          `json:"description"`
	Lines       []*DialogueLine `json:"lines"`
	StartTime   time.Time       `json:"start_time"`
	EndTime     time.Time       `json:"end_time"`
}

// DebateDialogue represents the complete debate as theatrical dialogue
type DebateDialogue struct {
	Title       string              `json:"title"`
	Subtitle    string              `json:"subtitle"`
	Characters  []*DialogueCharacter `json:"characters"`
	Acts        []*DialogueAct      `json:"acts"`
	Prologue    string              `json:"prologue"`
	Epilogue    string              `json:"epilogue"`
	Style       DialogueStyle       `json:"style"`
	CreatedAt   time.Time           `json:"created_at"`
	TotalRounds int                 `json:"total_rounds"`
}

// DialogueFormatter formats debate conversations as professional dialogue
type DialogueFormatter struct {
	style      DialogueStyle
	characters map[DebateTeamPosition]*DialogueCharacter
}

// NewDialogueFormatter creates a new dialogue formatter
func NewDialogueFormatter(style DialogueStyle) *DialogueFormatter {
	if style == "" {
		style = StyleTheater
	}
	return &DialogueFormatter{
		style:      style,
		characters: make(map[DebateTeamPosition]*DialogueCharacter),
	}
}

// RegisterCharacter registers a debate member as a dialogue character
func (df *DialogueFormatter) RegisterCharacter(member *DebateTeamMember) *DialogueCharacter {
	char := &DialogueCharacter{
		Name:     df.getCharacterName(member.Role),
		Role:     member.Role,
		Position: member.Position,
		Provider: member.ProviderName,
		Model:    member.ModelName,
		Avatar:   df.getCharacterAvatar(member.Role),
	}
	df.characters[member.Position] = char
	return char
}

// GetCharacter returns the character for a given position
func (df *DialogueFormatter) GetCharacter(position DebateTeamPosition) *DialogueCharacter {
	if df.characters == nil {
		return nil
	}
	return df.characters[position]
}

// GetAllCharacters returns all registered characters
func (df *DialogueFormatter) GetAllCharacters() []*DialogueCharacter {
	chars := make([]*DialogueCharacter, 0, len(df.characters))
	for _, char := range df.characters {
		chars = append(chars, char)
	}
	return chars
}

// getCharacterName returns a theatrical name for the role
func (df *DialogueFormatter) getCharacterName(role DebateRole) string {
	names := map[DebateRole]string{
		RoleAnalyst:   "THE ANALYST",
		RoleProposer:  "THE PROPOSER",
		RoleCritic:    "THE CRITIC",
		RoleSynthesis: "THE SYNTHESIZER",
		RoleMediator:  "THE MEDIATOR",
	}
	if name, ok := names[role]; ok {
		return name
	}
	return string(role)
}

// getCharacterAvatar returns an avatar/icon for the role
func (df *DialogueFormatter) getCharacterAvatar(role DebateRole) string {
	avatars := map[DebateRole]string{
		RoleAnalyst:   "[A]",
		RoleProposer:  "[P]",
		RoleCritic:    "[C]",
		RoleSynthesis: "[S]",
		RoleMediator:  "[M]",
	}
	if avatar, ok := avatars[role]; ok {
		return avatar
	}
	return "[?]"
}

// CreateDialogue creates a complete debate dialogue from debate results
func (df *DialogueFormatter) CreateDialogue(topic string, rounds []DebateRound) *DebateDialogue {
	dialogue := &DebateDialogue{
		Title:       "THE GREAT DEBATE",
		Subtitle:    topic,
		Characters:  df.getCharacterList(),
		Acts:        make([]*DialogueAct, 0, len(rounds)),
		Style:       df.style,
		CreatedAt:   time.Now(),
		TotalRounds: len(rounds),
	}

	// Create prologue
	dialogue.Prologue = df.createPrologue(topic)

	// Create acts from rounds
	for i, round := range rounds {
		act := df.createAct(i+1, round)
		dialogue.Acts = append(dialogue.Acts, act)
	}

	// Create epilogue
	dialogue.Epilogue = df.createEpilogue()

	return dialogue
}

// createPrologue creates the opening text
func (df *DialogueFormatter) createPrologue(topic string) string {
	return fmt.Sprintf(`
╔══════════════════════════════════════════════════════════════════╗
║                    THE GREAT AI DEBATE                            ║
╠══════════════════════════════════════════════════════════════════╣
║  Topic: %s
║                                                                    ║
║  The stage is set. Five distinguished AI minds gather to          ║
║  deliberate, challenge, and ultimately synthesize wisdom          ║
║  from their collective intelligence.                              ║
╚══════════════════════════════════════════════════════════════════╝

DRAMATIS PERSONAE:
`, df.truncateString(topic, 50))
}

// createEpilogue creates the closing text
func (df *DialogueFormatter) createEpilogue() string {
	return `
╔══════════════════════════════════════════════════════════════════╗
║                         CONSENSUS                                 ║
╠══════════════════════════════════════════════════════════════════╣
║  After careful deliberation, the council has reached              ║
║  its conclusion. The synthesis of perspectives follows.           ║
╚══════════════════════════════════════════════════════════════════╝
`
}

// createAct creates a dialogue act from a debate round
func (df *DialogueFormatter) createAct(actNumber int, round DebateRound) *DialogueAct {
	act := &DialogueAct{
		ActNumber:   actNumber,
		Title:       df.getActTitle(actNumber),
		Description: df.getActDescription(actNumber),
		Lines:       make([]*DialogueLine, 0),
		StartTime:   round.StartTime,
		EndTime:     round.EndTime,
	}

	// Convert round responses to dialogue lines
	for _, response := range round.Responses {
		char := df.characters[response.Position]
		if char == nil {
			char = &DialogueCharacter{
				Name:     df.getCharacterName(DebateRole(response.Role)),
				Role:     DebateRole(response.Role),
				Position: response.Position,
				Provider: response.Provider,
				Avatar:   df.getCharacterAvatar(DebateRole(response.Role)),
			}
		}

		line := &DialogueLine{
			Character:   char,
			Content:     response.Content,
			Timestamp:   response.Timestamp,
			RoundNumber: actNumber,
			Action:      df.getStageDirection(DebateRole(response.Role), actNumber),
		}
		act.Lines = append(act.Lines, line)
	}

	return act
}

// getActTitle returns the title for an act
func (df *DialogueFormatter) getActTitle(actNumber int) string {
	titles := map[int]string{
		1: "ACT I: THE OPENING",
		2: "ACT II: THE CHALLENGE",
		3: "ACT III: THE SYNTHESIS",
		4: "ACT IV: THE REFINEMENT",
		5: "ACT V: THE RESOLUTION",
	}
	if title, ok := titles[actNumber]; ok {
		return title
	}
	return fmt.Sprintf("ACT %d", actNumber)
}

// getActDescription returns a description for an act
func (df *DialogueFormatter) getActDescription(actNumber int) string {
	descriptions := map[int]string{
		1: "The council convenes. Each voice presents their initial perspective.",
		2: "Challenges arise. The Critic tests the foundations of each argument.",
		3: "Wisdom emerges. The Synthesizer weaves threads into tapestry.",
		4: "Refinement begins. Details are polished, edges smoothed.",
		5: "Resolution arrives. The Mediator guides toward consensus.",
	}
	if desc, ok := descriptions[actNumber]; ok {
		return desc
	}
	return "The debate continues..."
}

// getStageDirection returns a stage direction for character entry
func (df *DialogueFormatter) getStageDirection(role DebateRole, round int) string {
	if round == 1 {
		directions := map[DebateRole]string{
			RoleAnalyst:   "(steps forward, examining the question thoughtfully)",
			RoleProposer:  "(rises with conviction, ready to present)",
			RoleCritic:    "(leans in, eyes sharp with scrutiny)",
			RoleSynthesis: "(listens intently, gathering threads)",
			RoleMediator:  "(surveys the council with measured calm)",
		}
		if dir, ok := directions[role]; ok {
			return dir
		}
	}
	return ""
}

// getCharacterList returns all registered characters
func (df *DialogueFormatter) getCharacterList() []*DialogueCharacter {
	chars := make([]*DialogueCharacter, 0, len(df.characters))
	for pos := PositionAnalyst; pos <= PositionMediator; pos++ {
		if char, ok := df.characters[pos]; ok {
			chars = append(chars, char)
		}
	}
	return chars
}

// FormatAsText formats the dialogue as plain text
func (df *DialogueFormatter) FormatAsText(dialogue *DebateDialogue) string {
	var sb strings.Builder

	// Write prologue
	sb.WriteString(dialogue.Prologue)
	sb.WriteString("\n")

	// Write character list
	for _, char := range dialogue.Characters {
		sb.WriteString(fmt.Sprintf("  %s %s (%s/%s)\n",
			char.Avatar, char.Name, char.Provider, char.Model))
	}
	sb.WriteString("\n")

	// Write acts
	for _, act := range dialogue.Acts {
		sb.WriteString(df.formatAct(act))
	}

	// Write epilogue
	sb.WriteString(dialogue.Epilogue)

	return sb.String()
}

// formatAct formats a single act
func (df *DialogueFormatter) formatAct(act *DialogueAct) string {
	var sb strings.Builder

	// Act header
	sb.WriteString(fmt.Sprintf("\n%s\n", strings.Repeat("─", 70)))
	sb.WriteString(fmt.Sprintf("%s\n", act.Title))
	sb.WriteString(fmt.Sprintf("%s\n", act.Description))
	sb.WriteString(fmt.Sprintf("%s\n\n", strings.Repeat("─", 70)))

	// Dialogue lines
	for _, line := range act.Lines {
		sb.WriteString(df.formatLine(line))
	}

	return sb.String()
}

// formatLine formats a single dialogue line based on style
func (df *DialogueFormatter) formatLine(line *DialogueLine) string {
	switch df.style {
	case StyleTheater:
		return df.formatTheaterLine(line)
	case StyleNovel:
		return df.formatNovelLine(line)
	case StyleScript:
		return df.formatScriptLine(line)
	case StyleMinimal:
		return df.formatMinimalLine(line)
	default:
		return df.formatTheaterLine(line)
	}
}

// formatTheaterLine formats a line in theater style
func (df *DialogueFormatter) formatTheaterLine(line *DialogueLine) string {
	var sb strings.Builder

	if line.Action != "" {
		sb.WriteString(fmt.Sprintf("        %s\n", line.Action))
	}

	sb.WriteString(fmt.Sprintf("%s %s:\n", line.Character.Avatar, line.Character.Name))

	// Wrap content with indentation
	wrapped := df.wrapText(line.Content, 66)
	for _, l := range strings.Split(wrapped, "\n") {
		sb.WriteString(fmt.Sprintf("    %s\n", l))
	}
	sb.WriteString("\n")

	return sb.String()
}

// formatNovelLine formats a line in novel style
func (df *DialogueFormatter) formatNovelLine(line *DialogueLine) string {
	var sb strings.Builder

	if line.Action != "" {
		sb.WriteString(fmt.Sprintf("%s %s\n\n", line.Character.Name, line.Action))
	}

	sb.WriteString(fmt.Sprintf("\"%s,\" said %s.\n\n",
		df.truncateString(line.Content, 500), line.Character.Name))

	return sb.String()
}

// formatScriptLine formats a line in screenplay style
func (df *DialogueFormatter) formatScriptLine(line *DialogueLine) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("                    %s\n", line.Character.Name))
	if line.Action != "" {
		sb.WriteString(fmt.Sprintf("              %s\n", line.Action))
	}
	sb.WriteString(fmt.Sprintf("          %s\n\n", df.truncateString(line.Content, 500)))

	return sb.String()
}

// formatMinimalLine formats a line in minimal style
func (df *DialogueFormatter) formatMinimalLine(line *DialogueLine) string {
	return fmt.Sprintf("[%s] %s\n\n", line.Character.Name, line.Content)
}

// FormatForStreaming returns formatted lines for real-time streaming
func (df *DialogueFormatter) FormatForStreaming(line *DialogueLine) []string {
	lines := make([]string, 0)

	// Character introduction
	if line.Action != "" {
		lines = append(lines, fmt.Sprintf("\n        %s", line.Action))
	}

	lines = append(lines, fmt.Sprintf("\n%s %s:", line.Character.Avatar, line.Character.Name))

	// Content lines (chunked for streaming effect)
	words := strings.Fields(line.Content)
	currentLine := "    "
	for _, word := range words {
		if len(currentLine)+len(word)+1 > 70 {
			lines = append(lines, currentLine)
			currentLine = "    " + word
		} else {
			if currentLine == "    " {
				currentLine += word
			} else {
				currentLine += " " + word
			}
		}
	}
	if currentLine != "    " {
		lines = append(lines, currentLine)
	}
	lines = append(lines, "")

	return lines
}

// wrapText wraps text to specified width
func (df *DialogueFormatter) wrapText(text string, width int) string {
	words := strings.Fields(text)
	if len(words) == 0 {
		return ""
	}

	var lines []string
	currentLine := words[0]

	for _, word := range words[1:] {
		if len(currentLine)+len(word)+1 <= width {
			currentLine += " " + word
		} else {
			lines = append(lines, currentLine)
			currentLine = word
		}
	}
	lines = append(lines, currentLine)

	return strings.Join(lines, "\n")
}

// truncateString truncates a string to specified length
func (df *DialogueFormatter) truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

// DebateRound represents a single round of debate (for dialogue creation)
type DebateRound struct {
	RoundNumber int               `json:"round_number"`
	Responses   []DebateResponse  `json:"responses"`
	StartTime   time.Time         `json:"start_time"`
	EndTime     time.Time         `json:"end_time"`
}

// DebateResponse represents a single response in a debate round
type DebateResponse struct {
	Position   DebateTeamPosition `json:"position"`
	Role       string             `json:"role"`
	Provider   string             `json:"provider"`
	Model      string             `json:"model"`
	Content    string             `json:"content"`
	Timestamp  time.Time          `json:"timestamp"`
	Confidence float64            `json:"confidence"`
}

// StreamingDialogueConfig configures streaming dialogue output
type StreamingDialogueConfig struct {
	Style             DialogueStyle `json:"style"`
	ShowPrologue      bool          `json:"show_prologue"`
	ShowActHeaders    bool          `json:"show_act_headers"`
	ShowStageActions  bool          `json:"show_stage_actions"`
	ShowCharacterInfo bool          `json:"show_character_info"`
	ChunkDelay        int           `json:"chunk_delay_ms"` // Delay between chunks for effect
}

// DefaultStreamingConfig returns default streaming configuration
func DefaultStreamingConfig() *StreamingDialogueConfig {
	return &StreamingDialogueConfig{
		Style:             StyleTheater,
		ShowPrologue:      true,
		ShowActHeaders:    true,
		ShowStageActions:  true,
		ShowCharacterInfo: true,
		ChunkDelay:        50,
	}
}

// DialogueEvent represents an event in the streaming dialogue
type DialogueEvent struct {
	Type      DialogueEventType `json:"type"`
	Content   string            `json:"content"`
	Character *DialogueCharacter `json:"character,omitempty"`
	ActNumber int               `json:"act_number,omitempty"`
	Timestamp time.Time         `json:"timestamp"`
}

// DialogueEventType represents the type of dialogue event
type DialogueEventType string

const (
	EventPrologue      DialogueEventType = "prologue"
	EventActStart      DialogueEventType = "act_start"
	EventStageAction   DialogueEventType = "stage_action"
	EventCharacterLine DialogueEventType = "character_line"
	EventActEnd        DialogueEventType = "act_end"
	EventEpilogue      DialogueEventType = "epilogue"
	EventConsensus     DialogueEventType = "consensus"
	// Multi-pass validation phase events
	EventPhaseStart       DialogueEventType = "phase_start"
	EventPhaseProgress    DialogueEventType = "phase_progress"
	EventPhaseEnd         DialogueEventType = "phase_end"
	EventValidationResult DialogueEventType = "validation_result"
	EventPolishResult     DialogueEventType = "polish_result"
	EventFinalSynthesis   DialogueEventType = "final_synthesis"
)

// DialogueStream provides streaming dialogue events
type DialogueStream struct {
	events chan *DialogueEvent
	done   chan struct{}
	config *StreamingDialogueConfig
}

// NewDialogueStream creates a new dialogue stream
func NewDialogueStream(config *StreamingDialogueConfig) *DialogueStream {
	if config == nil {
		config = DefaultStreamingConfig()
	}
	return &DialogueStream{
		events: make(chan *DialogueEvent, 100),
		done:   make(chan struct{}),
		config: config,
	}
}

// Events returns the event channel
func (ds *DialogueStream) Events() <-chan *DialogueEvent {
	return ds.events
}

// Done returns the done channel
func (ds *DialogueStream) Done() <-chan struct{} {
	return ds.done
}

// Close closes the stream
func (ds *DialogueStream) Close() {
	close(ds.done)
	close(ds.events)
}

// SendEvent sends an event to the stream
func (ds *DialogueStream) SendEvent(event *DialogueEvent) {
	select {
	case ds.events <- event:
	case <-ds.done:
	}
}

// FormatEventAsText formats a dialogue event as text
func FormatEventAsText(event *DialogueEvent) string {
	switch event.Type {
	case EventPrologue:
		return event.Content
	case EventActStart:
		return fmt.Sprintf("\n%s\n%s\n%s\n",
			strings.Repeat("─", 70),
			event.Content,
			strings.Repeat("─", 70))
	case EventStageAction:
		return fmt.Sprintf("        %s\n", event.Content)
	case EventCharacterLine:
		if event.Character != nil {
			return fmt.Sprintf("%s %s:\n    %s\n",
				event.Character.Avatar, event.Character.Name, event.Content)
		}
		return fmt.Sprintf("    %s\n", event.Content)
	case EventActEnd:
		return fmt.Sprintf("\n[End of Act %d]\n", event.ActNumber)
	case EventEpilogue:
		return event.Content
	case EventConsensus:
		return fmt.Sprintf("\n╔══ CONSENSUS ══╗\n%s\n╚═══════════════╝\n", event.Content)
	// Multi-pass validation phase events
	case EventPhaseStart:
		return fmt.Sprintf("\n%s\n%s\n%s\n",
			strings.Repeat("═", 70),
			event.Content,
			strings.Repeat("═", 70))
	case EventPhaseProgress:
		return fmt.Sprintf("  ▸ %s\n", event.Content)
	case EventPhaseEnd:
		return fmt.Sprintf("\n%s\n[Phase Complete]\n%s\n",
			strings.Repeat("─", 70),
			strings.Repeat("─", 70))
	case EventValidationResult:
		return fmt.Sprintf("  ✓ Validation: %s\n", event.Content)
	case EventPolishResult:
		return fmt.Sprintf("  ✨ Polish: %s\n", event.Content)
	case EventFinalSynthesis:
		return fmt.Sprintf(`
╔══════════════════════════════════════════════════════════════════════════════╗
║                           📜 FINAL CONCLUSION 📜                              ║
╠══════════════════════════════════════════════════════════════════════════════╣
%s
╚══════════════════════════════════════════════════════════════════════════════╝
`, event.Content)
	default:
		return event.Content
	}
}

// DialogueEvent extension fields for multi-pass validation
type PhaseEventData struct {
	Phase           string  `json:"phase,omitempty"`
	PhaseOrder      int     `json:"phase_order,omitempty"`
	PhaseIcon       string  `json:"phase_icon,omitempty"`
	ValidationScore float64 `json:"validation_score,omitempty"`
	PolishScore     float64 `json:"polish_score,omitempty"`
	Confidence      float64 `json:"confidence,omitempty"`
}
