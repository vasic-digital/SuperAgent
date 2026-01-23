// Package skills provides a skill framework for HelixAgent.
//
// This package enables defining, matching, and executing skills that enhance
// HelixAgent's capabilities. Skills are reusable units of functionality that
// can be invoked via MCP, ACP, LSP, or direct API calls.
//
// # Skill Definition
//
// Skills are defined with instructions and triggers:
//
//	skill := &skills.Skill{
//	    Name:        "code_review",
//	    Category:    "code_review",
//	    Description: "Review code for quality and best practices",
//	    Instructions: "Analyze the provided code for:\n" +
//	        "1. Code quality\n" +
//	        "2. Best practices\n" +
//	        "3. Potential bugs\n" +
//	        "4. Security issues",
//	    Triggers: []string{"review", "analyze", "check"},
//	}
//
// # Skill Categories
//
// Built-in skill categories:
//
//   - code_generation: Generate code from descriptions
//   - code_review: Review and analyze code
//   - documentation: Generate documentation
//   - testing: Create and run tests
//   - refactoring: Improve code structure
//   - debugging: Debug and fix issues
//   - analysis: Analyze code/data
//   - search: Search codebase/web
//
// # Skill Service
//
// The service manages skill lifecycle:
//
//	service := skills.NewService(config)
//
//	// Register skill
//	if err := service.RegisterSkill(skill); err != nil {
//	    log.Fatal(err)
//	}
//
//	// Get all skills
//	allSkills := service.GetAllSkills()
//
//	// Get skill by name
//	skill, ok := service.GetSkill("code_review")
//
// # Skill Matching
//
// Match user requests to skills:
//
//	matcher := skills.NewMatcher(service)
//	match, err := matcher.Match(ctx, "Please review my code")
//
//	if match.Confidence > 0.8 {
//	    // Execute matched skill
//	}
//
// Match types:
//   - Exact: Trigger word exact match
//   - Fuzzy: Trigger word fuzzy match
//   - Semantic: Semantic similarity match
//
// # Protocol Adapters
//
// Skills are exposed via multiple protocols:
//
//	adapter := skills.NewProtocolSkillAdapter(service)
//	adapter.RegisterAllSkillsAsTools()
//
//	// MCP tools
//	mcpTools := adapter.GetMCPTools()
//
//	// ACP actions
//	acpActions := adapter.GetACPActions()
//
//	// LSP commands
//	lspCommands := adapter.GetLSPCommands()
//
// # Skill Execution
//
// Execute skills through the adapter:
//
//	// Via MCP
//	result, err := adapter.InvokeMCPTool(ctx, "skill_code_review", params)
//
//	// Via ACP
//	result, err := adapter.InvokeACPAction(ctx, "skill.code_review", params)
//
//	// Via LSP
//	result, err := adapter.InvokeLSPCommand(ctx, "helixagent.skill.code_review", args)
//
// # Skill Usage Tracking
//
// Track skill execution for analytics:
//
//	// Start tracking
//	requestID := service.StartSkillExecution(id, skill, match)
//
//	// ... execute skill ...
//
//	// Complete tracking
//	usage := service.CompleteSkillExecution(requestID, success, errorMsg)
//
// # Key Files
//
//   - service.go: Skill service and registry
//   - skill.go: Skill type definitions
//   - matcher.go: Skill matching logic
//   - protocol_adapter.go: Protocol adapters (MCP/ACP/LSP)
//   - tracker.go: Usage tracking
//
// # Example: Custom Skill
//
//	skill := &skills.Skill{
//	    Name:        "database_query",
//	    Category:    "analysis",
//	    Description: "Generate and explain SQL queries",
//	    Instructions: `Generate SQL queries based on natural language.
//	    Always explain the query and suggest optimizations.`,
//	    Triggers: []string{"sql", "query", "database", "select"},
//	    Config: map[string]interface{}{
//	        "max_results": 100,
//	        "explain":     true,
//	    },
//	}
//
//	service.RegisterSkill(skill)
package skills
