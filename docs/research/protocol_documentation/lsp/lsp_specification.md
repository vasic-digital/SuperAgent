# Language Server Protocol (LSP) - Complete Specification

**Protocol:** LSP (Language Server Protocol)  
**Version:** 3.17  
**Status:** Stable (Microsoft)  
**HelixAgent Implementation:** [internal/lsp/](../../../internal/lsp/)  
**Analysis Date:** 2026-04-03  

---

## Executive Summary

LSP is an open, JSON-RPC based protocol for communication between code editors (clients) and language servers. It enables IDE features like code completion, go-to-definition, and diagnostics across all programming languages.

**Key Benefits:**
- Language-agnostic IDE support
- Decouples language intelligence from editors
- Industry standard (VS Code, JetBrains, Neovim, etc.)

---

## Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                     LSP ARCHITECTURE                             │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│   ┌──────────────────┐         ┌──────────────────┐             │
│   │   LSP Client     │◄───────►│   LSP Server     │             │
│   │   (Editor/IDE)   │  JSON   │   (Language      │             │
│   │                  │  RPC    │    Intelligence) │             │
│   │  • VS Code       │         │                  │             │
│   │  • JetBrains     │         │  • Syntax        │             │
│   │  • Neovim        │         │    analysis      │             │
│   │  • Emacs         │         │  • Completion    │             │
│   │  • Continue      │         │  • Diagnostics   │             │
│   │  • Cline         │         │  • Refactoring   │             │
│   └──────────────────┘         └──────────────────┘             │
│                                                                  │
│   Transport: stdio (default) or TCP                             │
│                                                                  │
└─────────────────────────────────────────────────────────────────┘
```

---

## Protocol Basics

### Message Format

LSP uses JSON-RPC 2.0:

```json
// Request
{
  "jsonrpc": "2.0",
  "id": 1,
  "method": "textDocument/completion",
  "params": {
    "textDocument": {"uri": "file:///main.go"},
    "position": {"line": 10, "character": 15}
  }
}

// Response
{
  "jsonrpc": "2.0",
  "id": 1,
  "result": {
    "items": [
      {
        "label": "Println",
        "kind": 3,
        "documentation": "Prints to stdout"
      }
    ]
  }
}

// Notification (no id)
{
  "jsonrpc": "2.0",
  "method": "textDocument/publishDiagnostics",
  "params": {
    "uri": "file:///main.go",
    "diagnostics": [...]
  }
}
```

### Core Types

```typescript
// Position in document
interface Position {
  line: number;        // 0-based
  character: number;   // 0-based
}

// Range in document
interface Range {
  start: Position;
  end: Position;
}

// Document identifier
interface TextDocumentIdentifier {
  uri: string;
}

// Document item with content
interface TextDocumentItem {
  uri: string;
  languageId: string;
  version: number;
  text: string;
}

// Location (file + range)
interface Location {
  uri: string;
  range: Range;
}
```

---

## Lifecycle

### 1. Initialize

```json
// Client → Server
{
  "jsonrpc": "2.0",
  "id": 0,
  "method": "initialize",
  "params": {
    "processId": 1234,
    "clientInfo": {
      "name": "HelixAgent-IDE",
      "version": "1.0.0"
    },
    "capabilities": {
      "textDocument": {
        "synchronization": {
          "dynamicRegistration": false,
          "willSave": true,
          "willSaveWaitUntil": true,
          "didSave": true
        },
        "completion": {
          "dynamicRegistration": false,
          "completionItem": {
            "snippetSupport": true,
            "commitCharactersSupport": true,
            "documentationFormat": ["markdown", "plaintext"],
            "deprecatedSupport": true,
            "preselectSupport": true
          }
        },
        "hover": {
          "dynamicRegistration": false,
          "contentFormat": ["markdown", "plaintext"]
        },
        "definition": {
          "dynamicRegistration": false,
          "linkSupport": true
        },
        "documentSymbol": {
          "dynamicRegistration": false,
          "hierarchicalDocumentSymbolSupport": true
        },
        "codeAction": {
          "dynamicRegistration": false,
          "codeActionLiteralSupport": {
            "codeActionKind": {
              "valueSet": ["", "quickfix", "refactor", "source"]
            }
          }
        },
        "formatting": {
          "dynamicRegistration": false
        },
        "rename": {
          "dynamicRegistration": false,
          "prepareSupport": true
        }
      }
    },
    "workspaceFolders": [
      {
        "uri": "file:///home/user/project",
        "name": "my-project"
      }
    ]
  }
}

// Server → Client
{
  "jsonrpc": "2.0",
  "id": 0,
  "result": {
    "capabilities": {
      "textDocumentSync": {
        "openClose": true,
        "change": 2,  // Incremental
        "willSave": false,
        "willSaveWaitUntil": false,
        "save": {
          "includeText": false
        }
      },
      "completionProvider": {
        "resolveProvider": false,
        "triggerCharacters": [".", ":", ">"]
      },
      "hoverProvider": true,
      "definitionProvider": true,
      "documentSymbolProvider": true,
      "codeActionProvider": true,
      "documentFormattingProvider": true,
      "renameProvider": {
        "prepareProvider": true
      }
    },
    "serverInfo": {
      "name": "helixagent-lsp",
      "version": "1.0.0"
    }
  }
}

// Client → Server: Initialized notification
{
  "jsonrpc": "2.0",
  "method": "initialized",
  "params": {}
}
```

### 2. Text Document Synchronization

```json
// Client opens document
{
  "jsonrpc": "2.0",
  "method": "textDocument/didOpen",
  "params": {
    "textDocument": {
      "uri": "file:///main.go",
      "languageId": "go",
      "version": 1,
      "text": "package main\n\nfunc main() {}"
    }
  }
}

// Client changes document (incremental)
{
  "jsonrpc": "2.0",
  "method": "textDocument/didChange",
  "params": {
    "textDocument": {
      "uri": "file:///main.go",
      "version": 2
    },
    "contentChanges": [
      {
        "range": {
          "start": {"line": 2, "character": 13},
          "end": {"line": 2, "character": 15}
        },
        "text": "fmt.Println(\"Hello\")"
      }
    ]
  }
}

// Client saves document
{
  "jsonrpc": "2.0",
  "method": "textDocument/didSave",
  "params": {
    "textDocument": {
      "uri": "file:///main.go"
    }
  }
}

// Client closes document
{
  "jsonrpc": "2.0",
  "method": "textDocument/didClose",
  "params": {
    "textDocument": {
      "uri": "file:///main.go"
    }
  }
}
```

---

## Core Features

### 1. Completion

```json
// Request
{
  "jsonrpc": "2.0",
  "id": 1,
  "method": "textDocument/completion",
  "params": {
    "textDocument": {
      "uri": "file:///main.go"
    },
    "position": {
      "line": 5,
      "character": 10
    },
    "context": {
      "triggerKind": 1,  // Invoked
      "triggerCharacter": "."
    }
  }
}

// Response
{
  "jsonrpc": "2.0",
  "id": 1,
  "result": {
    "isIncomplete": false,
    "items": [
      {
        "label": "Println",
        "kind": 3,  // Function
        "detail": "func(a ...interface{}) (n int, err error)",
        "documentation": {
          "kind": "markdown",
          "value": "Println formats using the default formats for its operands..."
        },
        "sortText": "1",
        "filterText": "Println",
        "insertText": "Println(${1:})",
        "insertTextFormat": 2,  // Snippet
        "commitCharacters": ["("]
      },
      {
        "label": "Printf",
        "kind": 3,
        "detail": "func(format string, a ...interface{}) (n int, err error)",
        "documentation": "Printf formats according to a format specifier...",
        "sortText": "2"
      }
    ]
  }
}
```

### 2. Hover

```json
// Request
{
  "jsonrpc": "2.0",
  "id": 2,
  "method": "textDocument/hover",
  "params": {
    "textDocument": {
      "uri": "file:///main.go"
    },
    "position": {
      "line": 5,
      "character": 8
    }
  }
}

// Response
{
  "jsonrpc": "2.0",
  "id": 2,
  "result": {
    "contents": {
      "kind": "markdown",
      "value": "```go\nfunc Println(a ...interface{}) (n int, err error)\n```\n\nPrintln formats using the default formats..."
    },
    "range": {
      "start": {"line": 5, "character": 5},
      "end": {"line": 5, "character": 12}
    }
  }
}
```

### 3. Go to Definition

```json
// Request
{
  "jsonrpc": "2.0",
  "id": 3,
  "method": "textDocument/definition",
  "params": {
    "textDocument": {
      "uri": "file:///main.go"
    },
    "position": {
      "line": 10,
      "character": 15
    }
  }
}

// Response
{
  "jsonrpc": "2.0",
  "id": 3,
  "result": {
    "uri": "file:///utils/helper.go",
    "range": {
      "start": {"line": 20, "character": 5},
      "end": {"line": 20, "character": 15}
    }
  }
}
```

### 4. Diagnostics

```json
// Server → Client (notification)
{
  "jsonrpc": "2.0",
  "method": "textDocument/publishDiagnostics",
  "params": {
    "uri": "file:///main.go",
    "version": 2,
    "diagnostics": [
      {
        "range": {
          "start": {"line": 5, "character": 10},
          "end": {"line": 5, "character": 20}
        },
        "severity": 1,  // Error
        "code": "unused",
        "source": "helixagent-lsp",
        "message": "Unused variable 'result'",
        "relatedInformation": [
          {
            "location": {
              "uri": "file:///main.go",
              "range": {
                "start": {"line": 5, "character": 4},
                "end": {"line": 5, "character": 10}
              }
            },
            "message": "Variable declared here"
          }
        ]
      }
    ]
  }
}
```

### 5. Code Actions

```json
// Request
{
  "jsonrpc": "2.0",
  "id": 4,
  "method": "textDocument/codeAction",
  "params": {
    "textDocument": {
      "uri": "file:///main.go"
    },
    "range": {
      "start": {"line": 5, "character": 0},
      "end": {"line": 5, "character": 30}
    },
    "context": {
      "diagnostics": [
        {
          "range": {
            "start": {"line": 5, "character": 10},
            "end": {"line": 5, "character": 20}
          },
          "severity": 1,
          "message": "Unused variable"
        }
      ]
    }
  }
}

// Response
{
  "jsonrpc": "2.0",
  "id": 4,
  "result": [
    {
      "title": "Remove unused variable",
      "kind": "quickfix",
      "diagnostics": [...],
      "edit": {
        "changes": {
          "file:///main.go": [
            {
              "range": {
                "start": {"line": 5, "character": 0},
                "end": {"line": 6, "character": 0}
              },
              "newText": ""
            }
          ]
        }
      }
    },
    {
      "title": "Suppress warning",
      "kind": "quickfix",
      "command": {
        "title": "Suppress",
        "command": "helixagent.suppressDiagnostic"
      }
    }
  ]
}
```

---

## HelixAgent LSP Implementation

### Architecture

**Source:** [`internal/lsp/`](../../../internal/lsp/)

```
internal/lsp/
├── server.go               # LSP server core
├── handlers.go             # Request handlers
├── documents.go            # Document management
├── completion.go           # Completion engine
├── diagnostics.go          # Diagnostics provider
├── symbols.go              # Symbol resolution
├── hover.go                # Hover provider
├── definition.go           # Go-to-definition
├── code_actions.go         # Code actions
├── formatting.go           # Document formatting
├── rename.go               # Symbol rename
├── workspace.go            # Workspace operations
└── ai_features.go          # AI-enhanced features
```

### Server Implementation

**Source:** [`internal/lsp/server.go`](../../../internal/lsp/server.go)

```go
package lsp

// LSPServer implements Language Server Protocol
// Source: internal/lsp/server.go#L1-220

type LSPServer struct {
    documents    *DocumentManager
    completer    *AICompletionEngine
    diagnostics  *DiagnosticsProvider
    symbols      *SymbolResolver
    workspace    *WorkspaceManager
    clients      map[string]*Client
}

// Initialize handles LSP initialize request
// Source: internal/lsp/server.go#L45-95
func (s *LSPServer) Initialize(ctx context.Context, params *InitializeParams) (*InitializeResult, error) {
    // Register workspace
    for _, folder := range params.WorkspaceFolders {
        if err := s.workspace.AddFolder(ctx, folder); err != nil {
            return nil, err
        }
    }
    
    // Return server capabilities
    return &InitializeResult{
        Capabilities: ServerCapabilities{
            TextDocumentSync: &TextDocumentSyncOptions{
                OpenClose: true,
                Change:    TextDocumentSyncKindIncremental,
                Save: &SaveOptions{
                    IncludeText: false,
                },
            },
            CompletionProvider: &CompletionOptions{
                ResolveProvider:   false,
                TriggerCharacters: []string{".", ":", ">", "("},
            },
            HoverProvider:              true,
            DefinitionProvider:         true,
            DocumentSymbolProvider:     true,
            CodeActionProvider:         true,
            DocumentFormattingProvider: true,
            RenameProvider: &RenameOptions{
                PrepareProvider: true,
            },
        },
        ServerInfo: &ServerInfo{
            Name:    "helixagent-lsp",
            Version: "1.0.0",
        },
    }, nil
}

// HandleTextDocumentCompletion provides AI-enhanced completions
// Source: internal/lsp/server.go#L145-200
func (s *LSPServer) HandleTextDocumentCompletion(ctx context.Context, params *CompletionParams) (*CompletionList, error) {
    // Get document content
    doc, err := s.documents.Get(params.TextDocument.URI)
    if err != nil {
        return nil, err
    }
    
    // Get context around cursor
    context := doc.GetContext(params.Position, 50)
    
    // Query AI for completions
    completions, err := s.completer.GetCompletions(ctx, &CompletionRequest{
        Language: doc.LanguageID,
        Context:  context,
        Position: params.Position,
    })
    if err != nil {
        return nil, err
    }
    
    // Convert to LSP format
    items := make([]CompletionItem, 0, len(completions))
    for _, comp := range completions {
        items = append(items, CompletionItem{
            Label:            comp.Label,
            Kind:             comp.Kind,
            Detail:           comp.Detail,
            Documentation:    comp.Documentation,
            InsertText:       comp.InsertText,
            InsertTextFormat: InsertTextFormatSnippet,
            SortText:         comp.SortText,
        })
    }
    
    return &CompletionList{
        IsIncomplete: false,
        Items:        items,
    }, nil
}
```

### AI-Enhanced Completion

**Source:** [`internal/lsp/completion.go`](../../../internal/lsp/completion.go)

```go
package lsp

// AICompletionEngine uses LLM for intelligent completions
// Source: internal/lsp/completion.go#L1-150

type AICompletionEngine struct {
    provider  llm.Provider
    cache     *CompletionCache
    ensemble  *services.Ensemble
}

// GetCompletions generates AI-powered completions
// Source: internal/lsp/completion.go#L45-120
func (e *AICompletionEngine) GetCompletions(ctx context.Context, req *CompletionRequest) ([]AICompletion, error) {
    // Check cache first
    if cached := e.cache.Get(req); cached != nil {
        return cached, nil
    }
    
    // Build prompt for completion
    prompt := fmt.Sprintf(`
        Language: %s
        Context:
        %s
        
        Provide 5 relevant completions for the cursor position.
        Return as JSON array with label, kind, detail, and insertText.
    `, req.Language, req.Context)
    
    // Use ensemble for better completions
    resp, err := e.ensemble.Execute(ctx, &llm.CompletionRequest{
        Messages: []llm.Message{
            {Role: "user", Content: prompt},
        },
        MaxTokens: 500,
    })
    if err != nil {
        return nil, err
    }
    
    // Parse completions from response
    completions := parseCompletions(resp.Content)
    
    // Cache results
    e.cache.Set(req, completions)
    
    return completions, nil
}
```

---

## LSP Integration with CLI Agents

### Agent LSP Support Matrix

| Agent | LSP Support | Role | Integration Method |
|-------|-------------|------|-------------------|
| **Continue** | Full | Client | Native LSP |
| **Cline** | Partial | Client | VS Code API → LSP |
| **Kiro** | Full | Client | Native LSP |
| **Aider** | None | N/A | Via adapter |
| **Claude Code** | None | N/A | Via adapter |
| **Codex** | Partial | Client | VS Code extension |
| **OpenHands** | None | N/A | Via adapter |

### Continue IDE Integration

**Source:** [`internal/lsp/continue.go`](../../../internal/lsp/continue.go)

```go
package lsp

// ContinueAdapter integrates Continue IDE with HelixAgent LSP
// Source: internal/lsp/continue.go#L1-120

type ContinueAdapter struct {
    lspServer *LSPServer
}

// HandleContinueRequest processes Continue-specific requests
func (a *ContinueAdapter) HandleContinueRequest(ctx context.Context, req *ContinueRequest) (*ContinueResponse, error) {
    switch req.Type {
    case "autocomplete":
        // Map to LSP completion
        lspReq := &CompletionParams{
            TextDocument: TextDocumentIdentifier{URI: req.File},
            Position:     req.Position,
        }
        return a.lspServer.HandleTextDocumentCompletion(ctx, lspReq)
        
    case "quickFix":
        // Map to LSP code action
        lspReq := &CodeActionParams{
            TextDocument: TextDocumentIdentifier{URI: req.File},
            Range:        req.Range,
            Context:      req.Context,
        }
        return a.lspServer.HandleTextDocumentCodeAction(ctx, lspReq)
        
    case "gotoDefinition":
        // Map to LSP definition
        lspReq := &DefinitionParams{
            TextDocument: TextDocumentIdentifier{URI: req.File},
            Position:     req.Position,
        }
        return a.lspServer.HandleTextDocumentDefinition(ctx, lspReq)
    }
}
```

---

## Source Code Reference

### LSP Core Files

| Component | Source File | Lines | Description |
|-----------|-------------|-------|-------------|
| Server | `internal/lsp/server.go` | 220 | LSP server core |
| Handlers | `internal/lsp/handlers.go` | 280 | Request handlers |
| Documents | `internal/lsp/documents.go` | 150 | Document manager |
| Completion | `internal/lsp/completion.go` | 150 | AI completions |
| Diagnostics | `internal/lsp/diagnostics.go` | 120 | Diagnostics |
| Symbols | `internal/lsp/symbols.go` | 180 | Symbol resolution |
| Hover | `internal/lsp/hover.go` | 90 | Hover provider |
| Definition | `internal/lsp/definition.go` | 110 | Go-to-definition |
| Code Actions | `internal/lsp/code_actions.go` | 130 | Code actions |
| Formatting | `internal/lsp/formatting.go` | 80 | Document format |
| Rename | `internal/lsp/rename.go` | 100 | Symbol rename |
| Workspace | `internal/lsp/workspace.go` | 140 | Workspace ops |
| Tests | `internal/lsp/server_test.go` | 340 | Unit tests |

---

## API Endpoints

### LSP Server Endpoints

```
stdio://                    # Standard LSP transport (default)
tcp://localhost:7062/lsp   # TCP transport
ws://localhost:7061/lsp    # WebSocket transport
```

### HelixAgent REST Endpoints for LSP

```
POST /v1/lsp/initialize          # Initialize LSP
POST /v1/lsp/completion          # Get completions
POST /v1/lsp/hover               # Get hover info
POST /v1/lsp/definition          # Go to definition
POST /v1/lsp/diagnostics         # Get diagnostics
POST /v1/lsp/codeAction          # Get code actions
POST /v1/lsp/formatting          # Format document
POST /v1/lsp/rename              # Rename symbol
```

---

## Conclusion

LSP is the **industry standard for IDE integration**. HelixAgent provides:

- ✅ Full LSP 3.17 specification implementation
- ✅ AI-enhanced completions
- ✅ Multi-language support
- ✅ Integration with Continue and other IDE agents
- ✅ REST API alternative for web-based editors

**Recommendation:** Use LSP for IDE/editor integration and REST API for web-based tools.

---

*Specification Version: LSP 3.17*  
*Last Updated: 2026-04-03*  
*HelixAgent Commit: aa960946*
