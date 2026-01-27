// MCP SSE Bridge - Standalone binary
// Wraps stdio-based MCP servers and exposes them over HTTP with SSE support.
package main

import (
	"dev.helix.agent/internal/mcp/bridge"
)

func main() {
	bridge.Main()
}
