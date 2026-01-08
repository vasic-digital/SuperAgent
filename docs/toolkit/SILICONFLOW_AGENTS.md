# AGENTS.md - SiliconFlow Go Provider

## Essential Commands

### Installation & Setup
- **Add to Go project**: `go get github.com/helixagent/toolkit/SiliconFlow/providers/siliconflow`
- **Initialize module**: `go mod init github.com/helixagent/toolkit/SiliconFlow`
- **Install dependencies**: `go mod tidy`

### Testing
- **Run unit tests**: `go test ./providers/siliconflow/...`
- **Run with verbose output**: `go test -v ./providers/siliconflow/...`
- **Run integration tests**: `SILICONFLOW_API_KEY=your-key go test -tags=integration ./providers/siliconflow/...`

### Development
- **Format code**: `gofmt -w .`
- **Run linter**: `golangci-lint run`
- **Build**: `go build ./...`

## Project Structure

### Core Files
- **`providers/siliconflow/siliconflow.go`** - Main provider implementation
- **`providers/siliconflow/client.go`** - HTTP client and API interactions
- **`providers/siliconflow/discovery.go`** - Model discovery and categorization
- **`providers/siliconflow/builder.go`** - Provider factory and registration

### Documentation
- **`README.md`** - Main project documentation
- **`AGENTS.md`** - Agent development guidelines (this file)
- **`API_REFERENCE.md`** - Technical API documentation

## Dependencies and Environment

### Go Dependencies
- **Go version**: 1.23+
- **External dependencies**: None (uses only standard library)
- **Module path**: `github.com/helixagent/toolkit/SiliconFlow`

### Runtime Requirements
- **Go**: 1.23+ for building and running
- **API Key**: SiliconFlow API key for full functionality
- **Network**: HTTPS access to SiliconFlow API

### Cache Management
- Model discovery uses in-memory caching with configurable expiry
- No persistent cache files
- Cache automatically refreshed on provider restart

## Code Style Guidelines

### Go Standards
- **Go version**: 1.23+
- **Formatting**: Use `gofmt` for consistent formatting
- **Imports**: Group imports (stdlib, third-party, internal), no unused imports
- **Naming**: Use CamelCase for exports, camelCase for private, UPPER_SNAKE_CASE for constants
- **Error Handling**: Always handle errors, use structured error wrapping with context
- **Documentation**: Document all exported functions/types with examples

### File Organization
- **Single responsibility**: Each file has a clear, focused purpose
- **Interface compliance**: Implements HelixAgent Toolkit interfaces
- **Test coverage**: Comprehensive unit tests for all functionality
- **Build system**: Standard Go module and build system

### Git and Version Control
- **.gitignore**: Go-specific ignore patterns for build artifacts
- **Commit messages**: Use clear, descriptive commit messages
- **Branch naming**: Use descriptive branch names (feature/, bugfix/, etc.)

## API Coverage

### Supported Endpoints (via SiliconFlow Go Provider)
- **Chat Completions**: Full support with streaming, tools, reasoning
- **Embeddings**: Text embedding generation
- **Rerank**: Document reranking and relevance scoring
- **Images**: Text-to-image generation
- **Audio**: Speech synthesis and transcription
- **Video**: Text-to-video generation
- **Models**: Dynamic model discovery and metadata

### Model Categories
- **Chat Models**: 77+ models for conversational AI
- **Vision Models**: 12+ models for image understanding
- **Audio Models**: 3+ models for speech synthesis
- **Video Models**: 2+ models for video generation

## Configuration Compatibility

### HelixAgent Toolkit Integration
- Fully compatible with HelixAgent Toolkit `Provider` interface
- Supports all standard provider configuration options
- Intelligent default model selection based on capabilities
- Thread-safe implementation for concurrent use

## Development Workflow

### Standard Process
1. **Model Discovery**: Automatically fetch and categorize all SiliconFlow models
2. **API Integration**: Implement endpoints using Go standard library
3. **Testing**: Comprehensive unit tests with mocked HTTP responses
4. **Validation**: Ensure interface compliance and functionality
5. **Documentation**: Update relevant .md files with Go examples

### Code Change Process
1. **Start with tests**: Write tests first for new functionality
2. **Implement changes**: Modify provider code with proper error handling
3. **Run tests**: Use `go test` to validate changes
4. **Format code**: Use `gofmt` for consistent formatting
5. **Update docs**: Update README and examples

## Testing Patterns

### Unit Tests
- Uses Go `testing` framework
- Mock-based testing of HTTP responses
- Tests all endpoints with mocked data
- Configuration validation tests

### Integration Tests
- **Requires real API key**: Set `SILICONFLOW_API_KEY` environment variable
- Tests with real SiliconFlow API
- Validates model availability and functionality
- Uses build tags for conditional compilation

### Test Command Patterns
```bash
# Unit tests (no API key needed)
go test ./providers/siliconflow/...

# Integration tests (requires key)
SILICONFLOW_API_KEY=your-key go test -tags=integration ./providers/siliconflow/...

# Run with coverage
go test -cover ./providers/siliconflow/...

# Run with race detection
go test -race ./providers/siliconflow/...
```

## Important Gotchas and Patterns

### Security Patterns
- **API keys**: Never hardcode; use environment variables or secure configs
- **HTTPS only**: All API calls use HTTPS
- **Input validation**: Comprehensive validation of all inputs
- **No secrets in logs**: API keys and sensitive data are never logged

### Performance Optimizations
- **HTTP client reuse**: Single HTTP client instance with connection pooling
- **Context support**: Proper context handling for cancellation
- **Memory efficiency**: Efficient JSON processing
- **Concurrent safety**: Thread-safe implementation

### Error Handling Patterns
- **Structured errors**: Use `fmt.Errorf` with error wrapping
- **Context cancellation**: Respect context cancellation
- **Retry logic**: Implement intelligent retry with backoff
- **Input validation**: Validate all inputs before API calls

### Platform-Specific Considerations
- **HelixAgent Toolkit**: Direct integration via Provider interface
- **Go ecosystem**: Uses standard library for HTTP and JSON
- **Cross-platform**: Works on all Go-supported platforms

## Linting and Code Quality

### Go Tools
- **Formatting**: `gofmt` for consistent code formatting
- **Linting**: `golangci-lint` for comprehensive linting
- **Vetting**: `go vet` for static analysis

### Style Enforcement
- Follow Go conventions and effective Go practices
- Use `gofmt` for all code formatting
- Run `go vet` and `golangci-lint` before commits
- Maintain high test coverage

## Useful Commands for Development

### Testing and Validation
```bash
# Run all tests
go test ./...

# Run with verbose output
go test -v ./...

# Run with coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Run linter
golangci-lint run

# Format code
gofmt -w .

# Vet code
go vet ./...
```

### Module Management
```bash
# Initialize module
go mod init github.com/helixagent/toolkit/SiliconFlow

# Add dependencies
go get package-name

# Tidy dependencies
go mod tidy

# Download dependencies
go mod download

# Verify dependencies
go mod verify
```

### Build and Install
```bash
# Build the module
go build ./...

# Install binaries
go install ./...

# Clean build cache
go clean -cache
```

## Project Architecture Notes

### Single Responsibility Principle
- Each file has a clear, single purpose
- `siliconflow.go`: Main provider interface implementation
- `client.go`: HTTP client and API interactions
- `discovery.go`: Model discovery and metadata management
- `builder.go`: Factory pattern for provider creation

### Standard Library Only
- Pure Go with no external dependencies
- Uses `net/http` for API calls
- Uses `encoding/json` for serialization
- Uses `context` for cancellation

### Interface-Driven Design
- Implements HelixAgent Toolkit `Provider` interface
- Clean separation between interface and implementation
- Easy to test and mock

### Testing Philosophy
- Unit tests for fast iteration and regression prevention
- Integration tests for confidence with real APIs
- Mock-based testing for isolation
- High coverage as a quality gate