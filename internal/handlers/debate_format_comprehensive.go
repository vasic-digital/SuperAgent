package handlers

import (
	"fmt"
	"strings"

	"dev.helix.agent/internal/debate/comprehensive"
)

// ComprehensiveFormatterConfig configures the debate formatter
type ComprehensiveFormatterConfig struct {
	TeamOrder      []comprehensive.Team
	TeamNames      map[comprehensive.Team]string
	RoleNames      map[comprehensive.Role]string
	RoleIcons      map[comprehensive.Role]string
	ShowTeamIcons  bool
	MaxTopicLength int
}

// DefaultComprehensiveFormatterConfig returns default configuration
func DefaultComprehensiveFormatterConfig() *ComprehensiveFormatterConfig {
	return &ComprehensiveFormatterConfig{
		TeamOrder:      comprehensive.AllTeams(),
		TeamNames:      make(map[comprehensive.Team]string),
		RoleNames:      make(map[comprehensive.Role]string),
		RoleIcons:      make(map[comprehensive.Role]string),
		ShowTeamIcons:  true,
		MaxTopicLength: 70,
	}
}

// FormatComprehensiveDebateIntroduction formats the comprehensive debate dynamically
func FormatComprehensiveDebateIntroduction(topic string, teams map[comprehensive.Team][]*comprehensive.Agent, config *ComprehensiveFormatterConfig) string {
	if config == nil {
		config = DefaultComprehensiveFormatterConfig()
	}

	var sb strings.Builder

	// Count total agents and active teams
	totalAgents := 0
	activeTeams := 0
	for _, agents := range teams {
		if len(agents) > 0 {
			totalAgents += len(agents)
			activeTeams++
		}
	}

	// Header
	sb.WriteString("\n")
	sb.WriteString("# HelixAgent AI Debate Ensemble\n\n")
	sb.WriteString(fmt.Sprintf("> %d AI specialist%s across %d team%s collaborate to deliver optimal solutions.\n\n",
		totalAgents,
		pluralize(totalAgents),
		activeTeams,
		pluralize(activeTeams)))

	// Topic
	topicDisplay := topic
	if config.MaxTopicLength > 0 && len(topicDisplay) > config.MaxTopicLength {
		topicDisplay = topicDisplay[:config.MaxTopicLength] + "..."
	}
	sb.WriteString(fmt.Sprintf("**Topic:** %s\n\n", topicDisplay))

	// Debate Teams
	sb.WriteString("---\n\n")
	sb.WriteString("## Debate Teams & Roles\n\n")

	// Use configured team order, or dynamically order by team type
	teamOrder := config.TeamOrder
	if len(teamOrder) == 0 {
		teamOrder = comprehensive.AllTeams()
	}

	for _, team := range teamOrder {
		agents := teams[team]
		if len(agents) == 0 {
			continue
		}

		teamName := getDynamicTeamName(team, config.TeamNames, config.ShowTeamIcons)
		sb.WriteString(fmt.Sprintf("### %s\n\n", teamName))
		sb.WriteString("| Role | Model | Provider |\n")
		sb.WriteString("|------|-------|----------|\n")

		for _, agent := range agents {
			roleDisplay := getDynamicRoleName(agent.Role, config.RoleNames)
			sb.WriteString(fmt.Sprintf("| **%s** | %s | %s |\n",
				roleDisplay, agent.Model, agent.Provider))
		}
		sb.WriteString("\n")
	}

	sb.WriteString("---\n\n")
	sb.WriteString("## The Deliberation\n\n")

	return sb.String()
}

// FormatComprehensiveTeamProgress formats team progress dynamically
func FormatComprehensiveTeamProgress(team comprehensive.Team, agents []*comprehensive.Agent, config *ComprehensiveFormatterConfig) string {
	if config == nil {
		config = DefaultComprehensiveFormatterConfig()
	}

	var sb strings.Builder

	teamName := getDynamicTeamName(team, config.TeamNames, config.ShowTeamIcons)
	sb.WriteString(fmt.Sprintf("\n### %s Engaged\n\n", teamName))
	sb.WriteString(fmt.Sprintf("Activating %d specialist%s:\n", len(agents), pluralize(len(agents))))

	for _, agent := range agents {
		roleIcon := getDynamicRoleIcon(agent.Role, config.RoleIcons)
		roleName := getDynamicRoleName(agent.Role, config.RoleNames)
		sb.WriteString(fmt.Sprintf("- %s **%s** (%s/%s)\n",
			roleIcon, roleName, agent.Provider, agent.Model))
	}
	sb.WriteString("\n")

	return sb.String()
}

// FormatComprehensiveAgentResponse formats an agent response
func FormatComprehensiveAgentResponse(agent *comprehensive.Agent, content string, config *ComprehensiveFormatterConfig) string {
	if config == nil {
		config = DefaultComprehensiveFormatterConfig()
	}

	roleIcon := getDynamicRoleIcon(agent.Role, config.RoleIcons)
	roleName := getDynamicRoleName(agent.Role, config.RoleNames)

	return fmt.Sprintf("\n%s **[%s]** (%s)\n\n%s\n\n",
		roleIcon, roleName, agent.Provider, content)
}

// FormatComprehensivePhaseStart formats phase start
func FormatComprehensivePhaseStart(phase string, description string) string {
	return fmt.Sprintf("\n---\n\n### %s\n\n%s\n\n", phase, description)
}

// FormatComprehensiveConsensus formats the final consensus
func FormatComprehensiveConsensus(content string, qualityScore float64, agentCount int) string {
	var sb strings.Builder

	sb.WriteString("\n---\n\n")
	sb.WriteString("## Consensus Achieved\n\n")
	sb.WriteString(fmt.Sprintf("**Quality Score:** %.1f%%\n\n", qualityScore*100))
	sb.WriteString(content)
	sb.WriteString("\n\n")
	sb.WriteString("---\n\n")
	sb.WriteString(fmt.Sprintf("### ✨ Powered by HelixAgent Comprehensive Debate\n"))
	sb.WriteString(fmt.Sprintf("*Synthesized from %d AI specialist%s for optimal results*\n",
		agentCount, pluralize(agentCount)))
	sb.WriteString("---\n")

	return sb.String()
}

// FormatComprehensiveStreamEvent formats a stream event for display
func FormatComprehensiveStreamEvent(event *comprehensive.StreamEvent, config *ComprehensiveFormatterConfig) string {
	if config == nil {
		config = DefaultComprehensiveFormatterConfig()
	}

	var sb strings.Builder

	switch event.Type {
	case comprehensive.StreamEventDebateStart:
		sb.WriteString("\n🎬 **Starting Comprehensive Debate**\n\n")
		if event.Content != "" {
			sb.WriteString(event.Content + "\n")
		}

	case comprehensive.StreamEventTeamStart:
		if event.Team != "" {
			teamName := getDynamicTeamName(event.Team, config.TeamNames, config.ShowTeamIcons)
			sb.WriteString(fmt.Sprintf("\n🏁 **%s** activating...\n", teamName))
		}

	case comprehensive.StreamEventAgentResponse:
		if event.Agent != nil {
			icon := getDynamicRoleIcon(event.Agent.Role, config.RoleIcons)
			roleName := getDynamicRoleName(event.Agent.Role, config.RoleNames)
			sb.WriteString(fmt.Sprintf("\n%s **[%s]** contributing...\n", icon, roleName))
			if event.Content != "" {
				sb.WriteString(fmt.Sprintf("\n%s\n", event.Content))
			}
		}

	case comprehensive.StreamEventConsensusReached:
		sb.WriteString("\n✨ **Consensus Achieved**\n")
		if event.Content != "" {
			sb.WriteString(fmt.Sprintf("\n%s\n", event.Content))
		}

	case comprehensive.StreamEventDebateComplete:
		sb.WriteString("\n✅ **Debate Complete**\n")
		if event.Metadata != nil {
			if duration, ok := event.Metadata["duration"]; ok {
				sb.WriteString(fmt.Sprintf("Duration: %v\n", duration))
			}
		}
	}

	return sb.String()
}

// getDynamicTeamName returns team name from config or uses team ID
func getDynamicTeamName(team comprehensive.Team, teamNames map[comprehensive.Team]string, showIcons bool) string {
	if name, ok := teamNames[team]; ok && name != "" {
		return name
	}

	if showIcons {
		switch team {
		case comprehensive.TeamDesign:
			return "🏗️ Design Team"
		case comprehensive.TeamImplementation:
			return "💻 Implementation Team"
		case comprehensive.TeamQuality:
			return "🔍 Quality Assurance Team"
		case comprehensive.TeamRedTeam:
			return "🔴 Red Team"
		case comprehensive.TeamRefactoring:
			return "🔄 Refactoring Team"
		}
	}

	return string(team)
}

// getDynamicRoleName returns role name from config or uses role ID
func getDynamicRoleName(role comprehensive.Role, roleNames map[comprehensive.Role]string) string {
	if name, ok := roleNames[role]; ok && name != "" {
		return name
	}

	// Use role ID as fallback (no hardcoded names)
	return string(role)
}

// getDynamicRoleIcon returns role icon from config
func getDynamicRoleIcon(role comprehensive.Role, roleIcons map[comprehensive.Role]string) string {
	if icon, ok := roleIcons[role]; ok && icon != "" {
		return icon
	}

	// Default icon only if not configured
	return "🤖"
}

// pluralize returns 's' if count != 1
func pluralize(count int) string {
	if count == 1 {
		return ""
	}
	return "s"
}
