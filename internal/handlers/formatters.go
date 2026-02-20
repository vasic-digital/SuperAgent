package handlers

import (
	"encoding/csv"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"html"
	"strings"
	"time"

	"dev.helix.agent/internal/services"
	"gopkg.in/yaml.v3"
)

// ============================================================================
// Output Formatter Interface
// ============================================================================

// OutputFormatter defines the interface for all output formatters
type OutputFormatter interface {
	// FormatDebateTeamIntroduction formats the debate team introduction
	FormatDebateTeamIntroduction(topic string, members []*services.DebateTeamMember) string

	// FormatPhaseHeader formats a phase header
	FormatPhaseHeader(phase services.ValidationPhase, phaseNum int) string

	// FormatPhaseContent formats debate phase content
	FormatPhaseContent(content string) string

	// FormatFinalResponse formats the final consensus response
	FormatFinalResponse(content string) string

	// FormatFallbackIndicator formats a fallback indicator
	FormatFallbackIndicator(role services.DebateRole, fromProvider, fromModel, toProvider, toModel, reason string, duration time.Duration) string

	// Name returns the formatter name
	Name() string

	// ContentType returns the MIME content type for this format
	ContentType() string
}

// ============================================================================
// Extended Output Format Constants
// ============================================================================

const (
	// OutputFormatJSON formats output as structured JSON
	OutputFormatJSON OutputFormat = "json"
	// OutputFormatYAML formats output as YAML
	OutputFormatYAML OutputFormat = "yaml"
	// OutputFormatHTML formats output as HTML for web display
	OutputFormatHTML OutputFormat = "html"
	// OutputFormatXML formats output as XML
	OutputFormatXML OutputFormat = "xml"
	// OutputFormatCSV formats tabular data as CSV
	OutputFormatCSV OutputFormat = "csv"
	// OutputFormatRTF formats output as Rich Text Format
	OutputFormatRTF OutputFormat = "rtf"
	// OutputFormatTerminal formats output with enhanced terminal colors
	OutputFormatTerminal OutputFormat = "terminal"
	// OutputFormatCompact formats output with minimal whitespace
	OutputFormatCompact OutputFormat = "compact"
)

// ============================================================================
// JSON Formatter
// ============================================================================

// JSONFormatter formats output as structured JSON
type JSONFormatter struct{}

// JSONDebateIntroduction represents the JSON structure for debate introduction
type JSONDebateIntroduction struct {
	Title       string           `json:"title"`
	Description string           `json:"description"`
	Topic       string           `json:"topic"`
	Team        []JSONTeamMember `json:"team"`
	Timestamp   string           `json:"timestamp"`
}

// JSONTeamMember represents a team member in JSON format
type JSONTeamMember struct {
	Position int             `json:"position"`
	Role     string          `json:"role"`
	Model    string          `json:"model"`
	Provider string          `json:"provider"`
	Fallback *JSONTeamMember `json:"fallback,omitempty"`
}

// JSONPhaseHeader represents a phase header in JSON format
type JSONPhaseHeader struct {
	Phase     string `json:"phase"`
	PhaseNum  int    `json:"phase_num"`
	Icon      string `json:"icon"`
	Timestamp string `json:"timestamp"`
}

// JSONPhaseContent represents phase content in JSON format
type JSONPhaseContent struct {
	Content   string `json:"content"`
	Timestamp string `json:"timestamp"`
}

// JSONFinalResponse represents the final response in JSON format
type JSONFinalResponse struct {
	Type      string `json:"type"`
	Content   string `json:"content"`
	Timestamp string `json:"timestamp"`
}

// JSONFallbackIndicator represents a fallback indicator in JSON format
type JSONFallbackIndicator struct {
	Type         string `json:"type"`
	Role         string `json:"role"`
	FromProvider string `json:"from_provider"`
	FromModel    string `json:"from_model"`
	ToProvider   string `json:"to_provider"`
	ToModel      string `json:"to_model"`
	Reason       string `json:"reason"`
	Duration     string `json:"duration"`
	Timestamp    string `json:"timestamp"`
}

// NewJSONFormatter creates a new JSON formatter
func NewJSONFormatter() *JSONFormatter {
	return &JSONFormatter{}
}

func (f *JSONFormatter) Name() string {
	return "json"
}

func (f *JSONFormatter) ContentType() string {
	return "application/json"
}

func (f *JSONFormatter) FormatDebateTeamIntroduction(topic string, members []*services.DebateTeamMember) string {
	intro := JSONDebateIntroduction{
		Title:       "HelixAgent AI Debate Ensemble",
		Description: "Five AI minds deliberate to synthesize the optimal response.",
		Topic:       topic,
		Team:        make([]JSONTeamMember, 0, len(members)),
		Timestamp:   time.Now().UTC().Format(time.RFC3339),
	}

	for _, member := range members {
		if member == nil {
			continue
		}
		teamMember := JSONTeamMember{
			Position: int(member.Position),
			Role:     string(member.Role),
			Model:    member.ModelName,
			Provider: member.ProviderName,
		}
		if member.Fallback != nil {
			teamMember.Fallback = &JSONTeamMember{
				Model:    member.Fallback.ModelName,
				Provider: member.Fallback.ProviderName,
			}
		}
		intro.Team = append(intro.Team, teamMember)
	}

	data, err := json.MarshalIndent(intro, "", "  ")
	if err != nil {
		return fmt.Sprintf(`{"error": "%s"}`, err.Error())
	}
	return string(data)
}

func (f *JSONFormatter) FormatPhaseHeader(phase services.ValidationPhase, phaseNum int) string {
	icon := getPhaseIcon(phase)
	header := JSONPhaseHeader{
		Phase:     string(phase),
		PhaseNum:  phaseNum,
		Icon:      icon,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}
	data, err := json.MarshalIndent(header, "", "  ")
	if err != nil {
		return fmt.Sprintf(`{"error": "%s"}`, err.Error())
	}
	return string(data)
}

func (f *JSONFormatter) FormatPhaseContent(content string) string {
	pc := JSONPhaseContent{
		Content:   content,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}
	data, err := json.MarshalIndent(pc, "", "  ")
	if err != nil {
		return fmt.Sprintf(`{"error": "%s"}`, err.Error())
	}
	return string(data)
}

func (f *JSONFormatter) FormatFinalResponse(content string) string {
	resp := JSONFinalResponse{
		Type:      "final_response",
		Content:   content,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}
	data, err := json.MarshalIndent(resp, "", "  ")
	if err != nil {
		return fmt.Sprintf(`{"error": "%s"}`, err.Error())
	}
	return string(data)
}

func (f *JSONFormatter) FormatFallbackIndicator(role services.DebateRole, fromProvider, fromModel, toProvider, toModel, reason string, duration time.Duration) string {
	indicator := JSONFallbackIndicator{
		Type:         "fallback",
		Role:         string(role),
		FromProvider: fromProvider,
		FromModel:    fromModel,
		ToProvider:   toProvider,
		ToModel:      toModel,
		Reason:       reason,
		Duration:     formatDuration(duration),
		Timestamp:    time.Now().UTC().Format(time.RFC3339),
	}
	data, err := json.MarshalIndent(indicator, "", "  ")
	if err != nil {
		return fmt.Sprintf(`{"error": "%s"}`, err.Error())
	}
	return string(data)
}

// ============================================================================
// YAML Formatter
// ============================================================================

// YAMLFormatter formats output as YAML
type YAMLFormatter struct{}

// YAMLDebateIntroduction represents the YAML structure for debate introduction
type YAMLDebateIntroduction struct {
	Title       string           `yaml:"title"`
	Description string           `yaml:"description"`
	Topic       string           `yaml:"topic"`
	Team        []YAMLTeamMember `yaml:"team"`
	Timestamp   string           `yaml:"timestamp"`
}

// YAMLTeamMember represents a team member in YAML format
type YAMLTeamMember struct {
	Position int             `yaml:"position"`
	Role     string          `yaml:"role"`
	Model    string          `yaml:"model"`
	Provider string          `yaml:"provider"`
	Fallback *YAMLTeamMember `yaml:"fallback,omitempty"`
}

// YAMLPhaseHeader represents a phase header in YAML format
type YAMLPhaseHeader struct {
	Phase     string `yaml:"phase"`
	PhaseNum  int    `yaml:"phase_num"`
	Icon      string `yaml:"icon"`
	Timestamp string `yaml:"timestamp"`
}

// YAMLPhaseContent represents phase content in YAML format
type YAMLPhaseContent struct {
	Content   string `yaml:"content"`
	Timestamp string `yaml:"timestamp"`
}

// YAMLFinalResponse represents the final response in YAML format
type YAMLFinalResponse struct {
	Type      string `yaml:"type"`
	Content   string `yaml:"content"`
	Timestamp string `yaml:"timestamp"`
}

// YAMLFallbackIndicator represents a fallback indicator in YAML format
type YAMLFallbackIndicator struct {
	Type         string `yaml:"type"`
	Role         string `yaml:"role"`
	FromProvider string `yaml:"from_provider"`
	FromModel    string `yaml:"from_model"`
	ToProvider   string `yaml:"to_provider"`
	ToModel      string `yaml:"to_model"`
	Reason       string `yaml:"reason"`
	Duration     string `yaml:"duration"`
	Timestamp    string `yaml:"timestamp"`
}

// NewYAMLFormatter creates a new YAML formatter
func NewYAMLFormatter() *YAMLFormatter {
	return &YAMLFormatter{}
}

func (f *YAMLFormatter) Name() string {
	return "yaml"
}

func (f *YAMLFormatter) ContentType() string {
	return "application/x-yaml"
}

func (f *YAMLFormatter) FormatDebateTeamIntroduction(topic string, members []*services.DebateTeamMember) string {
	intro := YAMLDebateIntroduction{
		Title:       "HelixAgent AI Debate Ensemble",
		Description: "Five AI minds deliberate to synthesize the optimal response.",
		Topic:       topic,
		Team:        make([]YAMLTeamMember, 0, len(members)),
		Timestamp:   time.Now().UTC().Format(time.RFC3339),
	}

	for _, member := range members {
		if member == nil {
			continue
		}
		teamMember := YAMLTeamMember{
			Position: int(member.Position),
			Role:     string(member.Role),
			Model:    member.ModelName,
			Provider: member.ProviderName,
		}
		if member.Fallback != nil {
			teamMember.Fallback = &YAMLTeamMember{
				Model:    member.Fallback.ModelName,
				Provider: member.Fallback.ProviderName,
			}
		}
		intro.Team = append(intro.Team, teamMember)
	}

	data, err := yaml.Marshal(intro)
	if err != nil {
		return fmt.Sprintf("error: %s", err.Error())
	}
	return "---\n" + string(data)
}

func (f *YAMLFormatter) FormatPhaseHeader(phase services.ValidationPhase, phaseNum int) string {
	icon := getPhaseIcon(phase)
	header := YAMLPhaseHeader{
		Phase:     string(phase),
		PhaseNum:  phaseNum,
		Icon:      icon,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}
	data, err := yaml.Marshal(header)
	if err != nil {
		return fmt.Sprintf("error: %s", err.Error())
	}
	return "---\n" + string(data)
}

func (f *YAMLFormatter) FormatPhaseContent(content string) string {
	pc := YAMLPhaseContent{
		Content:   content,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}
	data, err := yaml.Marshal(pc)
	if err != nil {
		return fmt.Sprintf("error: %s", err.Error())
	}
	return string(data)
}

func (f *YAMLFormatter) FormatFinalResponse(content string) string {
	resp := YAMLFinalResponse{
		Type:      "final_response",
		Content:   content,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}
	data, err := yaml.Marshal(resp)
	if err != nil {
		return fmt.Sprintf("error: %s", err.Error())
	}
	return "---\n" + string(data)
}

func (f *YAMLFormatter) FormatFallbackIndicator(role services.DebateRole, fromProvider, fromModel, toProvider, toModel, reason string, duration time.Duration) string {
	indicator := YAMLFallbackIndicator{
		Type:         "fallback",
		Role:         string(role),
		FromProvider: fromProvider,
		FromModel:    fromModel,
		ToProvider:   toProvider,
		ToModel:      toModel,
		Reason:       reason,
		Duration:     formatDuration(duration),
		Timestamp:    time.Now().UTC().Format(time.RFC3339),
	}
	data, err := yaml.Marshal(indicator)
	if err != nil {
		return fmt.Sprintf("error: %s", err.Error())
	}
	return "---\n" + string(data)
}

// ============================================================================
// HTML Formatter
// ============================================================================

// HTMLFormatter formats output as HTML for web display
type HTMLFormatter struct {
	// IncludeStyles determines if inline CSS styles are included
	IncludeStyles bool
}

// NewHTMLFormatter creates a new HTML formatter
func NewHTMLFormatter() *HTMLFormatter {
	return &HTMLFormatter{IncludeStyles: true}
}

func (f *HTMLFormatter) Name() string {
	return "html"
}

func (f *HTMLFormatter) ContentType() string {
	return "text/html"
}

func (f *HTMLFormatter) getStyles() string {
	if !f.IncludeStyles {
		return ""
	}
	return `<style>
.debate-container { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; max-width: 900px; margin: 0 auto; padding: 20px; }
.debate-header { background: linear-gradient(135deg, #667eea 0%, #764ba2 100%); color: white; padding: 30px; border-radius: 12px; margin-bottom: 20px; text-align: center; }
.debate-header h1 { margin: 0 0 10px 0; font-size: 28px; }
.debate-header p { margin: 0; opacity: 0.9; }
.debate-topic { background: #f8f9fa; border-left: 4px solid #667eea; padding: 15px 20px; margin-bottom: 20px; border-radius: 0 8px 8px 0; }
.team-table { width: 100%; border-collapse: collapse; margin-bottom: 20px; }
.team-table th { background: #667eea; color: white; padding: 12px; text-align: left; }
.team-table td { padding: 12px; border-bottom: 1px solid #eee; }
.team-table tr:hover { background: #f8f9fa; }
.fallback-row { font-size: 0.9em; color: #666; padding-left: 30px; }
.phase-header { background: #f0f4ff; border-left: 4px solid #667eea; padding: 15px 20px; margin: 20px 0; border-radius: 0 8px 8px 0; }
.phase-header h3 { margin: 0; color: #333; }
.phase-content { background: #fafafa; padding: 20px; border-radius: 8px; margin-bottom: 20px; white-space: pre-wrap; }
.final-response { background: linear-gradient(135deg, #11998e 0%, #38ef7d 100%); color: white; padding: 25px; border-radius: 12px; margin-top: 20px; }
.final-response h2 { margin: 0 0 15px 0; }
.fallback-indicator { background: #fff3cd; border-left: 4px solid #ffc107; padding: 12px 15px; margin: 10px 0; border-radius: 0 8px 8px 0; font-size: 0.9em; }
.role-analyst { color: #0891b2; }
.role-proposer { color: #059669; }
.role-critic { color: #d97706; }
.role-synthesis { color: #7c3aed; }
.role-mediator { color: #2563eb; }
</style>`
}

func (f *HTMLFormatter) FormatDebateTeamIntroduction(topic string, members []*services.DebateTeamMember) string {
	var sb strings.Builder

	sb.WriteString(f.getStyles())
	sb.WriteString(`<div class="debate-container">`)
	sb.WriteString(`<div class="debate-header">`)
	sb.WriteString(`<h1>HelixAgent AI Debate Ensemble</h1>`)
	sb.WriteString(`<p>Five AI minds deliberate to synthesize the optimal response.</p>`)
	sb.WriteString(`</div>`)

	// Topic
	topicDisplay := topic
	if len(topicDisplay) > 100 {
		topicDisplay = topicDisplay[:100] + "..."
	}
	sb.WriteString(fmt.Sprintf(`<div class="debate-topic"><strong>Topic:</strong> %s</div>`, html.EscapeString(topicDisplay)))

	// Team table
	sb.WriteString(`<table class="team-table">`)
	sb.WriteString(`<thead><tr><th>Role</th><th>Model</th><th>Provider</th></tr></thead>`)
	sb.WriteString(`<tbody>`)

	for _, member := range members {
		if member == nil {
			continue
		}
		roleClass := fmt.Sprintf("role-%s", member.Role)
		roleName := getRoleName(member.Role)
		sb.WriteString(fmt.Sprintf(`<tr><td class="%s"><strong>%s</strong></td><td>%s</td><td>%s</td></tr>`,
			roleClass, html.EscapeString(roleName), html.EscapeString(member.ModelName), html.EscapeString(member.ProviderName)))

		if member.Fallback != nil {
			sb.WriteString(fmt.Sprintf(`<tr class="fallback-row"><td>Fallback</td><td>%s</td><td>%s</td></tr>`,
				html.EscapeString(member.Fallback.ModelName), html.EscapeString(member.Fallback.ProviderName)))
		}
	}

	sb.WriteString(`</tbody></table>`)
	sb.WriteString(`</div>`)

	return sb.String()
}

func (f *HTMLFormatter) FormatPhaseHeader(phase services.ValidationPhase, phaseNum int) string {
	icon := getPhaseIcon(phase)
	phaseName := getPhaseDisplayName(phase)

	return fmt.Sprintf(`<div class="phase-header"><h3>%s Phase %d: %s</h3></div>`,
		html.EscapeString(icon), phaseNum, html.EscapeString(phaseName))
}

func (f *HTMLFormatter) FormatPhaseContent(content string) string {
	return fmt.Sprintf(`<div class="phase-content">%s</div>`, html.EscapeString(content))
}

func (f *HTMLFormatter) FormatFinalResponse(content string) string {
	return fmt.Sprintf(`<div class="final-response"><h2>Final Answer</h2><div>%s</div></div>`,
		html.EscapeString(content))
}

func (f *HTMLFormatter) FormatFallbackIndicator(role services.DebateRole, fromProvider, fromModel, toProvider, toModel, reason string, duration time.Duration) string {
	roleName := getRoleName(role)
	return fmt.Sprintf(`<div class="fallback-indicator"><strong>[%s]</strong> Fallback from %s to %s - %s (%s)</div>`,
		html.EscapeString(roleName),
		html.EscapeString(formatModelRef(fromProvider, fromModel)),
		html.EscapeString(formatModelRef(toProvider, toModel)),
		html.EscapeString(reason), formatDuration(duration))
}

// ============================================================================
// XML Formatter
// ============================================================================

// XMLFormatter formats output as XML
type XMLFormatter struct{}

// XMLDebateIntroduction represents the XML structure for debate introduction
type XMLDebateIntroduction struct {
	XMLName     xml.Name `xml:"debate_introduction"`
	Title       string   `xml:"title"`
	Description string   `xml:"description"`
	Topic       string   `xml:"topic"`
	Team        XMLTeam  `xml:"team"`
	Timestamp   string   `xml:"timestamp"`
}

// XMLTeam represents the team in XML format
type XMLTeam struct {
	Members []XMLTeamMember `xml:"member"`
}

// XMLTeamMember represents a team member in XML format
type XMLTeamMember struct {
	Position int            `xml:"position,attr"`
	Role     string         `xml:"role"`
	Model    string         `xml:"model"`
	Provider string         `xml:"provider"`
	Fallback *XMLTeamMember `xml:"fallback,omitempty"`
}

// XMLPhaseHeader represents a phase header in XML format
type XMLPhaseHeader struct {
	XMLName   xml.Name `xml:"phase_header"`
	Phase     string   `xml:"phase"`
	PhaseNum  int      `xml:"phase_num"`
	Icon      string   `xml:"icon"`
	Timestamp string   `xml:"timestamp"`
}

// XMLPhaseContent represents phase content in XML format
type XMLPhaseContent struct {
	XMLName   xml.Name `xml:"phase_content"`
	Content   string   `xml:"content"`
	Timestamp string   `xml:"timestamp"`
}

// XMLFinalResponse represents the final response in XML format
type XMLFinalResponse struct {
	XMLName   xml.Name `xml:"final_response"`
	Content   string   `xml:"content"`
	Timestamp string   `xml:"timestamp"`
}

// XMLFallbackIndicator represents a fallback indicator in XML format
type XMLFallbackIndicator struct {
	XMLName      xml.Name `xml:"fallback"`
	Role         string   `xml:"role"`
	FromProvider string   `xml:"from_provider"`
	FromModel    string   `xml:"from_model"`
	ToProvider   string   `xml:"to_provider"`
	ToModel      string   `xml:"to_model"`
	Reason       string   `xml:"reason"`
	Duration     string   `xml:"duration"`
	Timestamp    string   `xml:"timestamp"`
}

// NewXMLFormatter creates a new XML formatter
func NewXMLFormatter() *XMLFormatter {
	return &XMLFormatter{}
}

func (f *XMLFormatter) Name() string {
	return "xml"
}

func (f *XMLFormatter) ContentType() string {
	return "application/xml"
}

func (f *XMLFormatter) FormatDebateTeamIntroduction(topic string, members []*services.DebateTeamMember) string {
	intro := XMLDebateIntroduction{
		Title:       "HelixAgent AI Debate Ensemble",
		Description: "Five AI minds deliberate to synthesize the optimal response.",
		Topic:       topic,
		Team:        XMLTeam{Members: make([]XMLTeamMember, 0, len(members))},
		Timestamp:   time.Now().UTC().Format(time.RFC3339),
	}

	for _, member := range members {
		if member == nil {
			continue
		}
		teamMember := XMLTeamMember{
			Position: int(member.Position),
			Role:     string(member.Role),
			Model:    member.ModelName,
			Provider: member.ProviderName,
		}
		if member.Fallback != nil {
			teamMember.Fallback = &XMLTeamMember{
				Model:    member.Fallback.ModelName,
				Provider: member.Fallback.ProviderName,
			}
		}
		intro.Team.Members = append(intro.Team.Members, teamMember)
	}

	data, err := xml.MarshalIndent(intro, "", "  ")
	if err != nil {
		return fmt.Sprintf(`<error>%s</error>`, err.Error())
	}
	return xml.Header + string(data)
}

func (f *XMLFormatter) FormatPhaseHeader(phase services.ValidationPhase, phaseNum int) string {
	icon := getPhaseIcon(phase)
	header := XMLPhaseHeader{
		Phase:     string(phase),
		PhaseNum:  phaseNum,
		Icon:      icon,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}
	data, err := xml.MarshalIndent(header, "", "  ")
	if err != nil {
		return fmt.Sprintf(`<error>%s</error>`, err.Error())
	}
	return string(data)
}

func (f *XMLFormatter) FormatPhaseContent(content string) string {
	pc := XMLPhaseContent{
		Content:   content,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}
	data, err := xml.MarshalIndent(pc, "", "  ")
	if err != nil {
		return fmt.Sprintf(`<error>%s</error>`, err.Error())
	}
	return string(data)
}

func (f *XMLFormatter) FormatFinalResponse(content string) string {
	resp := XMLFinalResponse{
		Content:   content,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}
	data, err := xml.MarshalIndent(resp, "", "  ")
	if err != nil {
		return fmt.Sprintf(`<error>%s</error>`, err.Error())
	}
	return string(data)
}

func (f *XMLFormatter) FormatFallbackIndicator(role services.DebateRole, fromProvider, fromModel, toProvider, toModel, reason string, duration time.Duration) string {
	indicator := XMLFallbackIndicator{
		Role:         string(role),
		FromProvider: fromProvider,
		FromModel:    fromModel,
		ToProvider:   toProvider,
		ToModel:      toModel,
		Reason:       reason,
		Duration:     formatDuration(duration),
		Timestamp:    time.Now().UTC().Format(time.RFC3339),
	}
	data, err := xml.MarshalIndent(indicator, "", "  ")
	if err != nil {
		return fmt.Sprintf(`<error>%s</error>`, err.Error())
	}
	return string(data)
}

// ============================================================================
// CSV Formatter
// ============================================================================

// CSVFormatter formats tabular data as CSV
type CSVFormatter struct {
	// Delimiter is the field delimiter (default: comma)
	Delimiter rune
}

// NewCSVFormatter creates a new CSV formatter
func NewCSVFormatter() *CSVFormatter {
	return &CSVFormatter{Delimiter: ','}
}

func (f *CSVFormatter) Name() string {
	return "csv"
}

func (f *CSVFormatter) ContentType() string {
	return "text/csv"
}

func (f *CSVFormatter) FormatDebateTeamIntroduction(topic string, members []*services.DebateTeamMember) string {
	var sb strings.Builder
	w := csv.NewWriter(&sb)
	w.Comma = f.Delimiter

	// Header comment
	sb.WriteString("# HelixAgent AI Debate Ensemble\n")
	sb.WriteString(fmt.Sprintf("# Topic: %s\n", strings.ReplaceAll(topic, "\n", " ")))

	// CSV header
	_ = w.Write([]string{"Position", "Role", "Model", "Provider", "Fallback_Model", "Fallback_Provider"})

	for _, member := range members {
		if member == nil {
			continue
		}
		fallbackModel := ""
		fallbackProvider := ""
		if member.Fallback != nil {
			fallbackModel = member.Fallback.ModelName
			fallbackProvider = member.Fallback.ProviderName
		}
		_ = w.Write([]string{
			fmt.Sprintf("%d", member.Position),
			string(member.Role),
			member.ModelName,
			member.ProviderName,
			fallbackModel,
			fallbackProvider,
		})
	}

	w.Flush()
	return sb.String()
}

func (f *CSVFormatter) FormatPhaseHeader(phase services.ValidationPhase, phaseNum int) string {
	var sb strings.Builder
	w := csv.NewWriter(&sb)
	w.Comma = f.Delimiter

	_ = w.Write([]string{"Phase_Type", "Phase_Num", "Icon", "Timestamp"})
	_ = w.Write([]string{string(phase), fmt.Sprintf("%d", phaseNum), getPhaseIcon(phase), time.Now().UTC().Format(time.RFC3339)})

	w.Flush()
	return sb.String()
}

func (f *CSVFormatter) FormatPhaseContent(content string) string {
	var sb strings.Builder
	w := csv.NewWriter(&sb)
	w.Comma = f.Delimiter

	_ = w.Write([]string{"Content", "Timestamp"})
	_ = w.Write([]string{content, time.Now().UTC().Format(time.RFC3339)})

	w.Flush()
	return sb.String()
}

func (f *CSVFormatter) FormatFinalResponse(content string) string {
	var sb strings.Builder
	w := csv.NewWriter(&sb)
	w.Comma = f.Delimiter

	_ = w.Write([]string{"Type", "Content", "Timestamp"})
	_ = w.Write([]string{"final_response", content, time.Now().UTC().Format(time.RFC3339)})

	w.Flush()
	return sb.String()
}

func (f *CSVFormatter) FormatFallbackIndicator(role services.DebateRole, fromProvider, fromModel, toProvider, toModel, reason string, duration time.Duration) string {
	var sb strings.Builder
	w := csv.NewWriter(&sb)
	w.Comma = f.Delimiter

	_ = w.Write([]string{"Role", "From_Provider", "From_Model", "To_Provider", "To_Model", "Reason", "Duration"})
	_ = w.Write([]string{string(role), fromProvider, fromModel, toProvider, toModel, reason, formatDuration(duration)})

	w.Flush()
	return sb.String()
}

// ============================================================================
// RTF Formatter
// ============================================================================

// RTFFormatter formats output as Rich Text Format
type RTFFormatter struct{}

// NewRTFFormatter creates a new RTF formatter
func NewRTFFormatter() *RTFFormatter {
	return &RTFFormatter{}
}

func (f *RTFFormatter) Name() string {
	return "rtf"
}

func (f *RTFFormatter) ContentType() string {
	return "application/rtf"
}

// escapeRTF escapes special RTF characters
func escapeRTF(s string) string {
	s = strings.ReplaceAll(s, "\\", "\\\\")
	s = strings.ReplaceAll(s, "{", "\\{")
	s = strings.ReplaceAll(s, "}", "\\}")
	return s
}

func (f *RTFFormatter) FormatDebateTeamIntroduction(topic string, members []*services.DebateTeamMember) string {
	var sb strings.Builder

	// RTF header
	sb.WriteString("{\\rtf1\\ansi\\deff0\n")
	sb.WriteString("{\\fonttbl{\\f0\\fswiss Helvetica;}{\\f1\\fmodern Courier New;}}\n")
	sb.WriteString("{\\colortbl;\\red102\\green126\\blue234;\\red118\\green75\\blue162;\\red0\\green0\\blue0;}\n")

	// Title
	sb.WriteString("\\f0\\fs36\\b\\cf1 HelixAgent AI Debate Ensemble\\b0\\fs24\\cf3\\par\n")
	sb.WriteString("\\par\n")
	sb.WriteString("\\i Five AI minds deliberate to synthesize the optimal response.\\i0\\par\n")
	sb.WriteString("\\par\n")

	// Topic
	topicDisplay := topic
	if len(topicDisplay) > 100 {
		topicDisplay = topicDisplay[:100] + "..."
	}
	sb.WriteString(fmt.Sprintf("\\b Topic:\\b0  %s\\par\n", escapeRTF(topicDisplay)))
	sb.WriteString("\\par\n")

	// Team
	sb.WriteString("\\b\\fs28 Debate Team\\b0\\fs24\\par\n")
	sb.WriteString("\\line\n")

	for _, member := range members {
		if member == nil {
			continue
		}
		roleName := getRoleName(member.Role)
		sb.WriteString(fmt.Sprintf("\\b %s:\\b0  %s (%s)\\par\n",
			escapeRTF(roleName), escapeRTF(member.ModelName), escapeRTF(member.ProviderName)))

		if member.Fallback != nil {
			sb.WriteString(fmt.Sprintf("\\tab\\i Fallback:\\i0  %s (%s)\\par\n",
				escapeRTF(member.Fallback.ModelName), escapeRTF(member.Fallback.ProviderName)))
		}
	}

	sb.WriteString("\\par\n")
	sb.WriteString("}")

	return sb.String()
}

func (f *RTFFormatter) FormatPhaseHeader(phase services.ValidationPhase, phaseNum int) string {
	icon := getPhaseIcon(phase)
	phaseName := getPhaseDisplayName(phase)

	var sb strings.Builder
	sb.WriteString("{\\rtf1\\ansi\\deff0\n")
	sb.WriteString("{\\fonttbl{\\f0\\fswiss Helvetica;}}\n")
	sb.WriteString(fmt.Sprintf("\\f0\\fs28\\b %s Phase %d: %s\\b0\\fs24\\par\n",
		escapeRTF(icon), phaseNum, escapeRTF(phaseName)))
	sb.WriteString("\\line\\par\n")
	sb.WriteString("}")

	return sb.String()
}

func (f *RTFFormatter) FormatPhaseContent(content string) string {
	var sb strings.Builder
	sb.WriteString("{\\rtf1\\ansi\\deff0\n")
	sb.WriteString("{\\fonttbl{\\f0\\fswiss Helvetica;}}\n")
	sb.WriteString(fmt.Sprintf("\\f0\\fs22 %s\\par\n", escapeRTF(content)))
	sb.WriteString("}")

	return sb.String()
}

func (f *RTFFormatter) FormatFinalResponse(content string) string {
	var sb strings.Builder
	sb.WriteString("{\\rtf1\\ansi\\deff0\n")
	sb.WriteString("{\\fonttbl{\\f0\\fswiss Helvetica;}}\n")
	sb.WriteString("{\\colortbl;\\red17\\green153\\blue142;}\n")
	sb.WriteString("\\f0\\fs32\\b\\cf1 Final Answer\\b0\\cf0\\fs24\\par\n")
	sb.WriteString("\\line\n")
	sb.WriteString(fmt.Sprintf("%s\\par\n", escapeRTF(content)))
	sb.WriteString("}")

	return sb.String()
}

func (f *RTFFormatter) FormatFallbackIndicator(role services.DebateRole, fromProvider, fromModel, toProvider, toModel, reason string, duration time.Duration) string {
	roleName := getRoleName(role)

	var sb strings.Builder
	sb.WriteString("{\\rtf1\\ansi\\deff0\n")
	sb.WriteString("{\\fonttbl{\\f0\\fswiss Helvetica;}}\n")
	sb.WriteString("{\\colortbl;\\red255\\green193\\blue7;}\n")
	sb.WriteString(fmt.Sprintf("\\f0\\fs20\\cf1\\b [%s] Fallback\\b0\\cf0  from %s to %s - %s (%s)\\par\n",
		escapeRTF(roleName),
		escapeRTF(formatModelRef(fromProvider, fromModel)),
		escapeRTF(formatModelRef(toProvider, toModel)),
		escapeRTF(reason), formatDuration(duration)))
	sb.WriteString("}")

	return sb.String()
}

// ============================================================================
// Terminal Formatter (Enhanced ANSI)
// ============================================================================

// TerminalFormatter formats output with enhanced terminal colors and formatting
type TerminalFormatter struct {
	// Use256Colors enables 256-color mode
	Use256Colors bool
	// UseTrueColor enables 24-bit true color mode
	UseTrueColor bool
}

// Extended ANSI codes for 256-color mode
const (
	ANSI256FgPrefix = "\033[38;5;"
	ANSI256BgPrefix = "\033[48;5;"
	ANSITrueColorFg = "\033[38;2;"
	ANSITrueColorBg = "\033[48;2;"
)

// NewTerminalFormatter creates a new terminal formatter
func NewTerminalFormatter() *TerminalFormatter {
	return &TerminalFormatter{
		Use256Colors: true,
		UseTrueColor: false,
	}
}

func (f *TerminalFormatter) Name() string {
	return "terminal"
}

func (f *TerminalFormatter) ContentType() string {
	return "text/plain"
}

func (f *TerminalFormatter) color256(colorCode int) string {
	return fmt.Sprintf("%s%dm", ANSI256FgPrefix, colorCode)
}

func (f *TerminalFormatter) bgColor256(colorCode int) string {
	return fmt.Sprintf("%s%dm", ANSI256BgPrefix, colorCode)
}

func (f *TerminalFormatter) trueColor(r, g, b int) string {
	return fmt.Sprintf("%s%d;%d;%dm", ANSITrueColorFg, r, g, b)
}

func (f *TerminalFormatter) FormatDebateTeamIntroduction(topic string, members []*services.DebateTeamMember) string {
	var sb strings.Builder

	// Enhanced header with gradient-like effect
	sb.WriteString("\n")
	if f.Use256Colors {
		sb.WriteString(f.color256(39)) // Light blue
	} else {
		sb.WriteString(ANSIBrightCyan)
	}
	sb.WriteString(ANSIBold)
	sb.WriteString("+" + strings.Repeat("-", 76) + "+\n")
	sb.WriteString("|" + strings.Repeat(" ", 18) + "HELIXAGENT AI DEBATE ENSEMBLE" + strings.Repeat(" ", 29) + "|\n")
	sb.WriteString("+" + strings.Repeat("-", 76) + "+\n")
	sb.WriteString(ANSIReset)

	sb.WriteString(ANSIDim)
	sb.WriteString("  Five AI minds deliberate to synthesize the optimal response.\n\n")
	sb.WriteString(ANSIReset)

	// Topic with fancy border
	topicDisplay := topic
	if len(topicDisplay) > 70 {
		topicDisplay = topicDisplay[:70] + "..."
	}
	sb.WriteString(ANSIBold)
	sb.WriteString("Topic: ")
	sb.WriteString(ANSIReset)
	sb.WriteString(topicDisplay)
	sb.WriteString("\n\n")

	// Separator
	if f.Use256Colors {
		sb.WriteString(f.color256(240)) // Gray
	} else {
		sb.WriteString(ANSIBrightBlack)
	}
	sb.WriteString(strings.Repeat("-", 78))
	sb.WriteString(ANSIReset)
	sb.WriteString("\n")

	// Team members with colored roles
	roleColors := map[services.DebateRole]int{
		services.RoleAnalyst:   45,  // Cyan
		services.RoleProposer:  46,  // Green
		services.RoleCritic:    214, // Orange
		services.RoleSynthesis: 135, // Purple
		services.RoleMediator:  39,  // Blue
	}

	for _, member := range members {
		if member == nil {
			continue
		}
		roleName := getRoleName(member.Role)

		if f.Use256Colors {
			sb.WriteString(f.color256(roleColors[member.Role]))
		} else {
			sb.WriteString(getRoleColor(member.Role))
		}
		sb.WriteString(ANSIBold)
		sb.WriteString(fmt.Sprintf("  %-12s", roleName))
		sb.WriteString(ANSIReset)

		sb.WriteString(fmt.Sprintf(" | %s (%s)\n", member.ModelName, member.ProviderName))

		if member.Fallback != nil {
			sb.WriteString(ANSIDim)
			sb.WriteString(fmt.Sprintf("    Fallback: %s (%s)\n", member.Fallback.ModelName, member.Fallback.ProviderName))
			sb.WriteString(ANSIReset)
		}
	}

	sb.WriteString("\n")
	if f.Use256Colors {
		sb.WriteString(f.color256(240))
	} else {
		sb.WriteString(ANSIBrightBlack)
	}
	sb.WriteString(strings.Repeat("-", 78))
	sb.WriteString(ANSIReset)
	sb.WriteString("\n\n")

	return sb.String()
}

func (f *TerminalFormatter) FormatPhaseHeader(phase services.ValidationPhase, phaseNum int) string {
	icon := getPhaseIcon(phase)
	phaseName := getPhaseDisplayName(phase)

	var sb strings.Builder
	sb.WriteString("\n")

	// Phase-specific colors (256-color mode)
	phaseColors := map[services.ValidationPhase]int{
		services.PhaseInitialResponse: 45,  // Cyan
		services.PhaseValidation:      46,  // Green
		services.PhasePolishImprove:   214, // Yellow/Orange
		services.PhaseFinalConclusion: 255, // White
	}

	if f.Use256Colors {
		sb.WriteString(f.color256(phaseColors[phase]))
	} else {
		if indicator, ok := PhaseIndicators[phase]; ok {
			sb.WriteString(indicator.Color)
		}
	}

	sb.WriteString(ANSIBold)
	sb.WriteString(fmt.Sprintf("=== %s Phase %d: %s ===", icon, phaseNum, phaseName))
	sb.WriteString(ANSIReset)
	sb.WriteString("\n\n")

	return sb.String()
}

func (f *TerminalFormatter) FormatPhaseContent(content string) string {
	return ANSIDim + content + ANSIReset
}

func (f *TerminalFormatter) FormatFinalResponse(content string) string {
	var sb strings.Builder

	sb.WriteString("\n")
	if f.Use256Colors {
		sb.WriteString(f.color256(46)) // Green
	} else {
		sb.WriteString(ANSIBrightGreen)
	}
	sb.WriteString(ANSIBold)
	sb.WriteString("=== FINAL ANSWER ===\n\n")
	sb.WriteString(ANSIReset)

	sb.WriteString(ANSIBrightWhite)
	sb.WriteString(content)
	sb.WriteString(ANSIReset)
	sb.WriteString("\n")

	return sb.String()
}

func (f *TerminalFormatter) FormatFallbackIndicator(role services.DebateRole, fromProvider, fromModel, toProvider, toModel, reason string, duration time.Duration) string {
	roleName := getRoleName(role)

	var sb strings.Builder
	if f.Use256Colors {
		sb.WriteString(f.color256(214)) // Orange/Yellow
	} else {
		sb.WriteString(ANSIYellow)
	}
	sb.WriteString(fmt.Sprintf("[%s] Fallback: %s -> %s (%s, %s)",
		roleName, formatModelRef(fromProvider, fromModel), formatModelRef(toProvider, toModel), reason, formatDuration(duration)))
	sb.WriteString(ANSIReset)
	sb.WriteString("\n")

	return sb.String()
}

// ============================================================================
// Compact Formatter
// ============================================================================

// CompactFormatter formats output with minimal whitespace for compact display
type CompactFormatter struct{}

// NewCompactFormatter creates a new compact formatter
func NewCompactFormatter() *CompactFormatter {
	return &CompactFormatter{}
}

func (f *CompactFormatter) Name() string {
	return "compact"
}

func (f *CompactFormatter) ContentType() string {
	return "text/plain"
}

func (f *CompactFormatter) FormatDebateTeamIntroduction(topic string, members []*services.DebateTeamMember) string {
	var sb strings.Builder

	topicDisplay := topic
	if len(topicDisplay) > 50 {
		topicDisplay = topicDisplay[:50] + "..."
	}
	sb.WriteString(fmt.Sprintf("DEBATE:%s|TEAM:", topicDisplay))

	teamParts := make([]string, 0, len(members))
	for _, member := range members {
		if member == nil {
			continue
		}
		part := fmt.Sprintf("%s=%s", string(member.Role), member.ModelName)
		if member.Fallback != nil {
			part += fmt.Sprintf("(fb:%s)", member.Fallback.ModelName)
		}
		teamParts = append(teamParts, part)
	}
	sb.WriteString(strings.Join(teamParts, ","))

	return sb.String()
}

func (f *CompactFormatter) FormatPhaseHeader(phase services.ValidationPhase, phaseNum int) string {
	return fmt.Sprintf("[P%d:%s]", phaseNum, string(phase))
}

func (f *CompactFormatter) FormatPhaseContent(content string) string {
	// Remove excessive whitespace and newlines
	content = strings.TrimSpace(content)
	// Repeatedly replace double newlines until none remain
	for strings.Contains(content, "\n\n") {
		content = strings.ReplaceAll(content, "\n\n", "\n")
	}
	// Repeatedly replace double spaces until none remain
	for strings.Contains(content, "  ") {
		content = strings.ReplaceAll(content, "  ", " ")
	}
	return content
}

func (f *CompactFormatter) FormatFinalResponse(content string) string {
	content = strings.TrimSpace(content)
	content = strings.ReplaceAll(content, "\n\n", "\n")
	return fmt.Sprintf("[FINAL]%s", content)
}

func (f *CompactFormatter) FormatFallbackIndicator(role services.DebateRole, fromProvider, fromModel, toProvider, toModel, reason string, duration time.Duration) string {
	return fmt.Sprintf("[FB:%s:%s->%s:%s]", string(role), fromModel, toModel, formatDuration(duration))
}

// ============================================================================
// Formatter Registry
// ============================================================================

// FormatterRegistry manages available formatters
type FormatterRegistry struct {
	formatters map[OutputFormat]OutputFormatter
}

// NewFormatterRegistry creates a new formatter registry with all formatters registered
func NewFormatterRegistry() *FormatterRegistry {
	registry := &FormatterRegistry{
		formatters: make(map[OutputFormat]OutputFormatter),
	}

	// Register all formatters
	registry.Register(OutputFormatJSON, NewJSONFormatter())
	registry.Register(OutputFormatYAML, NewYAMLFormatter())
	registry.Register(OutputFormatHTML, NewHTMLFormatter())
	registry.Register(OutputFormatXML, NewXMLFormatter())
	registry.Register(OutputFormatCSV, NewCSVFormatter())
	registry.Register(OutputFormatRTF, NewRTFFormatter())
	registry.Register(OutputFormatTerminal, NewTerminalFormatter())
	registry.Register(OutputFormatCompact, NewCompactFormatter())

	return registry
}

// Register registers a formatter for a given format
func (r *FormatterRegistry) Register(format OutputFormat, formatter OutputFormatter) {
	r.formatters[format] = formatter
}

// Get returns the formatter for the given format, or nil if not found
func (r *FormatterRegistry) Get(format OutputFormat) OutputFormatter {
	return r.formatters[format]
}

// GetOrDefault returns the formatter for the given format, or a default formatter
func (r *FormatterRegistry) GetOrDefault(format OutputFormat, defaultFormat OutputFormat) OutputFormatter {
	if f := r.formatters[format]; f != nil {
		return f
	}
	return r.formatters[defaultFormat]
}

// List returns all available format names
func (r *FormatterRegistry) List() []string {
	names := make([]string, 0, len(r.formatters))
	for format := range r.formatters {
		names = append(names, string(format))
	}
	return names
}

// ============================================================================
// Helper Functions
// ============================================================================

// getPhaseIcon returns the icon for a given phase
func getPhaseIcon(phase services.ValidationPhase) string {
	switch phase {
	case services.PhaseInitialResponse:
		return "?"
	case services.PhaseValidation:
		return "V"
	case services.PhasePolishImprove:
		return "*"
	case services.PhaseFinalConclusion:
		return "#"
	default:
		return ">"
	}
}

// GetFormatterForFormat returns the appropriate formatter for the given format
func GetFormatterForFormat(format OutputFormat) OutputFormatter {
	registry := NewFormatterRegistry()
	return registry.Get(format)
}

// FormatDebateIntroductionForFormat formats debate introduction using the specified format
func FormatDebateIntroductionForFormat(format OutputFormat, topic string, members []*services.DebateTeamMember) string {
	// Handle existing formats
	switch format {
	case OutputFormatANSI:
		return FormatDebateTeamIntroduction(topic, members)
	case OutputFormatMarkdown:
		return FormatDebateTeamIntroductionMarkdown(topic, members)
	case OutputFormatPlain:
		return FormatDebateTeamIntroductionPlain(topic, members)
	}

	// Handle new formats via registry
	formatter := GetFormatterForFormat(format)
	if formatter != nil {
		return formatter.FormatDebateTeamIntroduction(topic, members)
	}

	// Default to Markdown
	return FormatDebateTeamIntroductionMarkdown(topic, members)
}

// FormatPhaseHeaderForAllFormats formats a phase header for any format
func FormatPhaseHeaderForAllFormats(format OutputFormat, phase services.ValidationPhase, phaseNum int) string {
	// Handle existing formats
	switch format {
	case OutputFormatANSI:
		return FormatPhaseHeader(phase, phaseNum)
	case OutputFormatMarkdown:
		return FormatPhaseHeaderMarkdown(phase, phaseNum)
	case OutputFormatPlain:
		return FormatPhaseHeaderPlain(phase, phaseNum)
	}

	// Handle new formats via registry
	formatter := GetFormatterForFormat(format)
	if formatter != nil {
		return formatter.FormatPhaseHeader(phase, phaseNum)
	}

	// Default to Markdown
	return FormatPhaseHeaderMarkdown(phase, phaseNum)
}

// FormatPhaseContentForAllFormats formats phase content for any format
func FormatPhaseContentForAllFormats(format OutputFormat, content string) string {
	// Handle existing formats
	switch format {
	case OutputFormatANSI:
		return FormatPhaseContent(content)
	case OutputFormatMarkdown:
		return FormatPhaseContentMarkdown(content)
	case OutputFormatPlain:
		return content
	}

	// Handle new formats via registry
	formatter := GetFormatterForFormat(format)
	if formatter != nil {
		return formatter.FormatPhaseContent(content)
	}

	// Default to content as-is
	return content
}

// FormatFinalResponseForAllFormats formats the final response for any format
func FormatFinalResponseForAllFormats(format OutputFormat, content string) string {
	// Handle existing formats
	switch format {
	case OutputFormatANSI:
		return FormatFinalResponse(content)
	case OutputFormatMarkdown:
		return FormatFinalResponseMarkdown(content)
	case OutputFormatPlain:
		return FormatFinalResponsePlain(content)
	}

	// Handle new formats via registry
	formatter := GetFormatterForFormat(format)
	if formatter != nil {
		return formatter.FormatFinalResponse(content)
	}

	// Default to Markdown
	return FormatFinalResponseMarkdown(content)
}

// FormatFallbackIndicatorForAllFormats formats a fallback indicator for any format
func FormatFallbackIndicatorForAllFormats(format OutputFormat, role services.DebateRole, fromProvider, fromModel, toProvider, toModel, reason string, duration time.Duration) string {
	// Handle existing formats
	switch format {
	case OutputFormatANSI:
		return FormatFallbackTriggeredMarkdown(getRoleName(role), fromProvider, fromModel, toProvider, toModel, reason, categorizeErrorString(reason), duration)
	case OutputFormatMarkdown:
		return FormatFallbackTriggeredMarkdown(getRoleName(role), fromProvider, fromModel, toProvider, toModel, reason, categorizeErrorString(reason), duration)
	case OutputFormatPlain:
		return fmt.Sprintf("[%s] Fallback: %s -> %s (%s)\n", getRoleName(role), formatModelRef(fromProvider, fromModel), formatModelRef(toProvider, toModel), reason)
	}

	// Handle new formats via registry
	formatter := GetFormatterForFormat(format)
	if formatter != nil {
		return formatter.FormatFallbackIndicator(role, fromProvider, fromModel, toProvider, toModel, reason, duration)
	}

	// Default to Markdown
	return FormatFallbackTriggeredMarkdown(getRoleName(role), fromProvider, fromModel, toProvider, toModel, reason, categorizeErrorString(reason), duration)
}
