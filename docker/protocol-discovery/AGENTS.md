# AGENTS.md - Protocol Discovery Service

## Module Purpose for AI Agents

This is a **standalone HTTP service** that acts as a central registry and search engine for all protocol servers (MCP, LSP, ACP, Embedding). It enables HelixAgent to discover and route tool calls to the appropriate servers.

## Agent Guidelines

### DO

✅ **Add new protocol types** when supporting new server types  
✅ **Improve search algorithms** for better tool matching  
✅ **Add metrics** for monitoring and observability  
✅ **Write tests** for registry and search functionality  
✅ **Document API endpoints** in README.md  

### DON'T

❌ **Add business logic** beyond registry and search  
❌ **Store sensitive data** in server registry  
❌ **Break API compatibility** without versioning  
❌ **Import HelixAgent internal packages** - this is standalone  
❌ **Add persistent storage** without discussing architecture first  

## Common Tasks

### Adding a New Protocol Type

1. Add constant for protocol type:
```go
const (
    ProtocolMCP       = "mcp"
    ProtocolLSP       = "lsp"
    ProtocolACP       = "acp"
    ProtocolEmbedding = "embedding"
    ProtocolNEW       = "newprotocol"  // Add here
)
```

2. Update server struct if protocol needs special fields:
```go
type ServerInfo struct {
    Protocol string
    // Add protocol-specific fields if needed
}
```

3. Add validation in registration:
```go
func validateProtocol(protocol string) error {
    validProtocols := []string{
        ProtocolMCP, ProtocolLSP, ProtocolACP, 
        ProtocolEmbedding, ProtocolNEW,
    }
    // ...
}
```

4. Update documentation

### Improving Search Relevance

Current scoring weights:
```go
semanticWeight  = 0.40
bm25Weight      = 0.35
exactWeight     = 0.15
aliasWeight     = 0.10
```

To adjust:
1. Modify weights in `calculateFinalScore()`
2. Test with sample queries
3. Benchmark performance impact
4. Document changes

### Adding a New Endpoint

Example: Adding `/api/v1/stats`:

```go
// Add handler
func handleStats(w http.ResponseWriter, r *http.Request) {
    stats := registry.GetStats()
    json.NewEncoder(w).Encode(stats)
}

// Register in main()
http.HandleFunc("/api/v1/stats", handleStats)
```

## Code Structure

### Main File Organization

```go
// main.go
// 1. Constants and types (lines 1-100)
// 2. Registry implementation (lines 100-300)
// 3. Search engine (lines 300-600)
// 4. HTTP handlers (lines 600-800)
// 5. Main function (lines 800+)
```

### Key Types

**ServerInfo**: Represents a registered server
```go
type ServerInfo struct {
    ID       string       // Unique identifier
    Protocol string       // Protocol type (mcp, lsp, etc.)
    Host     string       // Hostname or IP
    Port     int          // Port number
    Tools    []ToolSchema // Available tools
    Health   HealthStatus // Current health
}
```

**ToolSchema**: Tool definition for search
```go
type ToolSchema struct {
    Name        string            // Tool identifier
    Description string            // Natural language description
    Category    string            // Tool category
    Aliases     []string          // Alternative names
    Parameters  map[string]Param  // Parameter definitions
}
```

## Testing Patterns

### Unit Test Example

```go
func TestSearchTools(t *testing.T) {
    // Setup
    registry := NewRegistry()
    registry.Register(&ServerInfo{
        ID: "test-server",
        Tools: []ToolSchema{
            {
                Name: "read_file",
                Description: "Read file contents",
            },
        },
    })

    // Test
    results := registry.Search("read file")

    // Assert
    if len(results) != 1 {
        t.Errorf("Expected 1 result, got %d", len(results))
    }
    if results[0].Tool.Name != "read_file" {
        t.Errorf("Expected read_file, got %s", results[0].Tool.Name)
    }
}
```

### Integration Test Example

```go
func TestHTTPSearchEndpoint(t *testing.T) {
    // Start test server
    go main()
    time.Sleep(100 * time.Millisecond) // Wait for startup

    // Make request
    resp, err := http.Get("http://localhost:8765/api/v1/search?q=test")
    if err != nil {
        t.Fatal(err)
    }

    // Check response
    if resp.StatusCode != http.StatusOK {
        t.Errorf("Expected 200, got %d", resp.StatusCode)
    }
}
```

## Debugging Tips

### Search Not Finding Tools

1. Check if tool is indexed:
```go
// Add debug logging
log.Printf("Indexed %d tools", len(searchEngine.tools))
```

2. Check tokenization:
```go
tokens := tokenize(query)
log.Printf("Query tokens: %v", tokens)
```

3. Check scoring:
```go
for _, result := range results {
    log.Printf("Tool: %s, Score: %.2f", result.Tool.Name, result.Score)
}
```

### Server Registration Failing

1. Check request format:
```bash
curl -X POST http://localhost:8765/api/v1/servers \
  -H "Content-Type: application/json" \
  -d @server.json -v
```

2. Check server logs:
```bash
docker logs helixagent-protocol-discovery
```

3. Validate JSON schema:
- Must have `id`, `protocol`, `host`, `port`
- `tools` array is optional but recommended

## Configuration Reference

### Environment Variables

```go
PORT                    // HTTP port (default: 8765)
MAX_SERVERS            // Max registered servers (default: 1000)
HEALTH_CHECK_INTERVAL  // Health check frequency (default: 30s)
SEARCH_MAX_RESULTS     // Max search results (default: 10)
SEARCH_MIN_SCORE       // Min relevance score (default: 0.3)
METRICS_ENABLED        // Enable Prometheus (default: true)
LOG_LEVEL              // debug, info, warn, error (default: info)
```

### Docker Environment

```dockerfile
ENV PORT=8765
ENV MAX_SERVERS=1000
ENV HEALTH_CHECK_INTERVAL=30s
EXPOSE 8765
```

## Integration with HelixAgent

### How HelixAgent Uses This Service

1. **Startup**: HelixAgent queries `/api/v1/servers` to discover available MCP servers
2. **Tool Selection**: When a tool is needed, HelixAgent calls `/api/v1/search` to find best match
3. **Execution**: HelixAgent routes tool call to appropriate server based on search results
4. **Monitoring**: Health endpoint used for service dependency checks

### Example Integration Flow

```go
// In HelixAgent service
func (s *Service) executeTool(ctx context.Context, toolName string, params map[string]any) (any, error) {
    // 1. Discover server hosting this tool
    resp, _ := http.Get("http://protocol-discovery:8765/api/v1/search?q=" + toolName)
    var results SearchResponse
    json.NewDecoder(resp.Body).Decode(&results)
    
    if len(results.Results) == 0 {
        return nil, errors.New("tool not found")
    }
    
    // 2. Get server info
    server := results.Results[0].Server
    
    // 3. Call tool via appropriate protocol
    return s.callMCPTool(ctx, server, toolName, params)
}
```

## Performance Tuning

### High Query Load

If experiencing high latency:
1. Increase `SEARCH_MIN_SCORE` to reduce candidates
2. Implement result caching
3. Add rate limiting
4. Consider horizontal scaling

### Memory Issues

If memory usage is high:
1. Reduce `MAX_SERVERS` limit
2. Prune inactive servers more aggressively
3. Reduce tool metadata size
4. Monitor with pprof

## Troubleshooting Checklist

- [ ] Service is running and healthy
- [ ] Port 8765 is accessible
- [ ] Servers are registering successfully
- [ ] Search returns expected results
- [ ] Health checks are passing
- [ ] Metrics are being collected
- [ ] Logs show no errors

## Questions?

- Check [README.md](README.md) for usage examples
- Review main.go for implementation details
- Look at tests/ for test patterns
- See [docs/protocols/](../../docs/protocols/) for protocol specs
