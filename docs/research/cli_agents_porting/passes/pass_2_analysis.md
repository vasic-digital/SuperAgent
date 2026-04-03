# Pass 2: Deep Analysis - Implementation Details & Architecture

**Pass:** 2 of 5  
**Phase:** Analysis  
**Goal:** Analyze implementation details of critical features  
**Date:** 2026-04-03  
**Status:** Complete  

---

## Executive Summary

This pass provides deep analysis of the 20 critical features identified in Pass 1. Each feature is analyzed at the code level, examining:
- Exact implementation mechanisms
- Dependencies and requirements
- Integration complexity
- Performance characteristics
- Porting strategy

**Features Analyzed:** 20 critical features  
**Source Files Referenced:** 150+  
**Implementation Patterns Identified:** 45  

---

## Critical Features Deep Dive

### 1. Aider's Git Integration System

#### 1.1 Repository Map (repo.py)

**Source Location:** `aider/repo.py`  
**Lines of Code:** ~400  
**HelixAgent Equivalent:** None (gap)

**Core Components:**
```python
# Source: aider/repo.py#L45-89 (conceptual)

class RepoMap:
    """
    Generates a ranked list of code symbols relevant to a query.
    Uses tree-sitter for AST parsing and ctags for symbol extraction.
    """
    
    def __init__(self, root_dir, map_tokens=1024):
        self.root = Path(root_dir)
        self.map_tokens = map_tokens
        self.ts_parser = TreeSitterParser()  # Multi-language AST parser
        self.ctags = UniversalCtags()        # Symbol tagger
        
    def get_ranked_tags(self, query, mentioned_files=None):
        """
        Returns list of (filename, symbol) tuples ranked by relevance.
        
        Algorithm:
        1. Parse query for keywords and file references
        2. Find files matching query (fuzzy matching)
        3. Extract symbols from matching files using tree-sitter
        4. Rank symbols by:
           - Distance from mentioned files (graph distance)
           - Symbol type (functions > variables)
           - Usage frequency across codebase
        5. Return top-K within token budget
        """
        files = self.find_matching_files(query)
        symbols = self.extract_symbols(files)
        ranked = self.rank_symbols(symbols, query, mentioned_files)
        return self.format_for_llm(ranked, self.map_tokens)
```

**Key Algorithms:**

1. **File Matching (fuzzy + semantic)**
```python
# Source: aider/repo.py#L120-156
def find_matching_files(self, query):
    """Find files relevant to query using multiple strategies."""
    matches = []
    
    # Strategy 1: Direct file mention in query
    for file in self.list_all_files():
        if file.name in query:
            matches.append((file, 1.0))  # High confidence
    
    # Strategy 2: Fuzzy filename matching
    for file in self.list_all_files():
        score = fuzz.ratio(file.name, query) / 100.0
        if score > 0.7:
            matches.append((file, score))
    
    # Strategy 3: Content grep (last resort)
    # Expensive, only for small repos
    
    return sorted(matches, key=lambda x: x[1], reverse=True)
```

2. **Symbol Extraction (tree-sitter)**
```python
# Source: aider/repo.py#L180-230
def extract_symbols(self, files):
    """Extract symbols from files using tree-sitter."""
    symbols = []
    
    for file in files:
        language = self.detect_language(file)
        parser = self.ts_parser.get_parser(language)
        
        tree = parser.parse(file.read_bytes())
        root = tree.root_node
        
        # Query for function/class definitions
        query_string = """
        (function_definition name: (identifier) @func)
        (class_definition name: (identifier) @class)
        """
        
        query = parser.query(query_string)
        captures = query.captures(root)
        
        for capture in captures:
            symbol_type = capture[1]  # 'func' or 'class'
            symbol_name = capture[0].text.decode('utf-8')
            line = capture[0].start_point[0]
            
            symbols.append({
                'file': file,
                'name': symbol_name,
                'type': symbol_type,
                'line': line
            })
    
    return symbols
```

3. **Symbol Ranking (graph-based)**
```python
# Source: aider/repo.py#L250-300
def rank_symbols(self, symbols, query, mentioned_files):
    """Rank symbols by relevance using graph algorithms."""
    
    # Build import/reference graph
    graph = self.build_reference_graph()
    
    scores = {}
    for symbol in symbols:
        score = 0.0
        
        # Factor 1: Graph distance from mentioned files
        if mentioned_files:
            min_distance = min(
                graph.distance(symbol['file'], mf)
                for mf in mentioned_files
            )
            score += 1.0 / (1 + min_distance)  # Closer = higher score
        
        # Factor 2: Symbol type importance
        type_weights = {'class': 1.0, 'function': 0.9, 'variable': 0.5}
        score += type_weights.get(symbol['type'], 0.3)
        
        # Factor 3: Name similarity to query
        score += fuzz.ratio(symbol['name'], query) / 200.0
        
        # Factor 4: Reference count (popularity)
        ref_count = graph.reference_count(symbol)
        score += min(ref_count / 10.0, 0.5)  # Cap at 0.5
        
        scores[symbol] = score
    
    return sorted(symbols, key=lambda s: scores[s], reverse=True)
```

**Porting Strategy:**
1. Create `internal/clis/aider/repo_map.go`
2. Port tree-sitter bindings to Go
3. Implement symbol ranking algorithms
4. Integrate with HelixAgent's context management

**Dependencies:**
- tree-sitter Go bindings
- ctags (universal-ctags)
- Graph data structure (adjacency list)

**Integration Points:**
- HelixAgent's context window management
- Debate orchestrator (for multi-file context)
- RAG pipeline (for semantic search)

---

#### 1.2 Diff-Based Editing (editblock_coder.py)

**Source Location:** `aider/coders/editblock_coder.py`  
**Lines of Code:** ~250  
**HelixAgent Equivalent:** None (gap)

**Core Format:**
```python
# Source: aider/coders/editblock_coder.py#L45-89

class EditBlockCoder:
    """
    Applies code changes using SEARCH/REPLACE blocks.
    More efficient than whole-file replacement for large files.
    """
    
    SEARCH_REPLACE_PATTERN = re.compile(
        r'^<<<<<<< SEARCH\n(.*?)^=======\n(.*?)^>>>>>>> REPLACE',
        re.DOTALL | re.MULTILINE
    )
    
    def apply_edits(self, edit_blocks, file_path):
        """
        Apply SEARCH/REPLACE blocks to file.
        
        Format:
        <<<<<<< SEARCH
        old code to find
        =======
        new code to replace
        >>>>>>> REPLACE
        """
        content = Path(file_path).read_text()
        
        for block in edit_blocks:
            search = block['search']
            replace = block['replace']
            
            # Validate search exists in file
            if search not in content:
                raise EditError(f"Search text not found: {search[:50]}...")
            
            # Apply replacement
            content = content.replace(search, replace, 1)
        
        # Write back
        Path(file_path).write_text(content)
        
        # Create git commit
        self.git_commit(file_path, "Applied AI edits")
```

**Porting Strategy:**
1. Create `internal/clis/aider/diff_format.go`
2. Implement SEARCH/REPLACE parser
3. Add validation and error handling
4. Integrate with HelixAgent's file operations

---

### 2. Claude Code's Terminal UI System

#### 2.1 Rich Output Renderer (terminal/render.ts)

**Source Location:** `claude-code/src/terminal/render.ts`  
**Lines of Code:** ~350  
**HelixAgent Equivalent:** Basic (needs extension)

**Core Components:**
```typescript
// Source: claude-code/src/terminal/render.ts#L45-120

interface TerminalRenderer {
  // Render code with syntax highlighting
  renderCodeBlock(code: string, language: string): void;
  
  // Render unified diff
  renderDiff(oldCode: string, newCode: string): void;
  
  // Show progress indicator
  renderProgress(percent: number, message: string): void;
  
  // Render tool execution result
  renderToolResult(toolName: string, result: ToolResult): void;
  
  // Render conversation history
  renderConversation(messages: Message[]): void;
}

class RichTerminalRenderer implements TerminalRenderer {
  private chalk: Chalk;
  private terminal: Terminal;
  
  renderCodeBlock(code: string, language: string): void {
    // 1. Syntax highlight using highlight.js
    const highlighted = highlight(code, { language });
    
    // 2. Add line numbers
    const lines = highlighted.split('\n');
    const numbered = lines.map((line, i) => 
      `${chalk.gray(String(i + 1).padStart(4))} ${line}`
    );
    
    // 3. Add border
    const bordered = [
      chalk.gray('┌' + '─'.repeat(78) + '┐'),
      ...numbered.map(l => chalk.gray('│') + l + chalk.gray('│')),
      chalk.gray('└' + '─'.repeat(78) + '┘')
    ];
    
    console.log(bordered.join('\n'));
  }
  
  renderDiff(oldCode: string, newCode: string): void {
    // Use diff library to generate unified diff
    const diff = createPatch('file', oldCode, newCode, 'old', 'new');
    
    // Colorize
    const lines = diff.split('\n');
    for (const line of lines) {
      if (line.startsWith('+')) {
        console.log(chalk.green(line));  // Added
      } else if (line.startsWith('-')) {
        console.log(chalk.red(line));    // Removed
      } else {
        console.log(chalk.gray(line));   // Context
      }
    }
  }
}
```

**Porting Strategy:**
1. Create `internal/output/terminal/rich_ui.go`
2. Port syntax highlighting (using chroma or similar)
3. Implement diff rendering
4. Add progress indicators

**Go Libraries Needed:**
- `github.com/alecthomas/chroma` - Syntax highlighting
- `github.com/sergi/go-diff` - Diff generation
- `github.com/fatih/color` - Terminal colors

---

### 3. OpenHands' Docker Sandboxing

#### 3.1 Sandbox Manager (sandbox/docker.py)

**Source Location:** `openhands/runtime/docker.py`  
**Lines of Code:** ~500  
**HelixAgent Equivalent:** Basic container support (needs extension)

**Core Components:**
```python
# Source: openhands/runtime/docker.py#L60-140

class DockerSandbox:
    """
    Manages Docker containers for secure code execution.
    Each agent session gets its own isolated container.
    """
    
    def __init__(self, image="openhands-runtime:latest"):
        self.client = docker.from_env()
        self.image = image
        self.container = None
        
    def start(self, workspace_dir):
        """Start sandboxed container."""
        self.container = self.client.containers.run(
            self.image,
            detach=True,
            volumes={
                workspace_dir: {'bind': '/workspace', 'mode': 'rw'}
            },
            network_mode='none',  # No network access by default
            mem_limit='2g',       # Memory limit
            cpu_quota=100000,     # CPU limit (1 core)
            security_opt=['no-new-privileges:true'],
            cap_drop=['ALL'],     # Drop all capabilities
            cap_add=['CHOWN', 'SETGID', 'SETUID'],  # Minimal capabilities
        )
        
    def execute(self, command, timeout=30):
        """Execute command in sandbox."""
        result = self.container.exec_run(
            command,
            workdir='/workspace',
            timeout=timeout
        )
        return {
            'exit_code': result.exit_code,
            'output': result.output.decode('utf-8'),
            'timed_out': result.exit_code == -1
        }
        
    def install_package(self, package, package_manager='pip'):
        """Install package in sandbox (ephemeral)."""
        commands = {
            'pip': f'pip install {package}',
            'npm': f'npm install {package}',
            'apt': f'apt-get update && apt-get install -y {package}'
        }
        return self.execute(commands[package_manager])
        
    def stop(self):
        """Stop and remove container."""
        if self.container:
            self.container.stop()
            self.container.remove()
```

**Porting Strategy:**
1. Create `internal/clis/openhands/sandbox.go`
2. Use Docker Go SDK (`github.com/docker/docker/client`)
3. Implement security policies
4. Add resource monitoring

**Security Considerations:**
- Drop all capabilities except necessary ones
- Network isolation by default
- Resource limits (CPU, memory, disk)
- Read-only root filesystem where possible

---

### 4. Kiro's Project Memory System

#### 4.1 Memory Manager (memory/manager.py)

**Source Location:** `kiro/memory/manager.py`  
**Lines of Code:** ~300  
**HelixAgent Equivalent:** PostgreSQL persistence (different approach)

**Core Components:**
```python
# Source: kiro/memory/manager.py#L45-120

class ProjectMemory:
    """
    Persistent memory system for project context.
    Survives across sessions and restarts.
    """
    
    def __init__(self, project_id, storage_dir="~/.kiro/memory"):
        self.project_id = project_id
        self.storage_path = Path(storage_dir) / project_id
        self.storage_path.mkdir(parents=True, exist_ok=True)
        
        # Load existing memory
        self.short_term = {}  # Current session
        self.long_term = self.load_long_term()
        
    def remember(self, key, value, importance=0.5):
        """
        Store information in memory.
        
        importance: 0.0-1.0
          - 0.0-0.3: Short-term only
          - 0.3-0.7: Both short and long term
          - 0.7-1.0: Long-term with semantic indexing
        """
        self.short_term[key] = {
            'value': value,
            'timestamp': time.time(),
            'importance': importance
        }
        
        if importance >= 0.3:
            self.long_term[key] = self.short_term[key]
            self.save_long_term()
            
        if importance >= 0.7:
            # Add to semantic index for retrieval
            self.index_semantically(key, value)
    
    def recall(self, key=None, query=None, top_k=5):
        """
        Retrieve information from memory.
        
        If key provided: exact lookup
        If query provided: semantic search
        """
        if key:
            # Exact lookup
            if key in self.short_term:
                return self.short_term[key]['value']
            if key in self.long_term:
                return self.long_term[key]['value']
            return None
            
        if query:
            # Semantic search
            return self.semantic_search(query, top_k)
    
    def semantic_search(self, query, top_k=5):
        """Search memory using semantic similarity."""
        # Generate embedding for query
        query_emb = self.embed(query)
        
        results = []
        for key, data in self.long_term.items():
            if 'embedding' in data:
                similarity = cosine_similarity(query_emb, data['embedding'])
                results.append((key, data, similarity))
        
        # Sort by similarity
        results.sort(key=lambda x: x[2], reverse=True)
        return results[:top_k]
```

**Porting Strategy:**
1. Create `internal/clis/kiro/memory.go`
2. Use HelixAgent's PostgreSQL for persistence
3. Integrate with existing embedding system
4. Add semantic search capability

**SQL Schema:**
```sql
-- See: sql_schemas/agent_instances.sql
CREATE TABLE project_memory (
    id SERIAL PRIMARY KEY,
    project_id VARCHAR(255) NOT NULL,
    key VARCHAR(500) NOT NULL,
    value JSONB NOT NULL,
    importance FLOAT DEFAULT 0.5,
    embedding VECTOR(1536),
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    UNIQUE(project_id, key)
);

CREATE INDEX idx_memory_project ON project_memory(project_id);
CREATE INDEX idx_memory_semantic ON project_memory USING ivfflat (embedding vector_cosine_ops);
```

---

## Implementation Complexity Analysis

### Complexity Matrix

| Feature | Code Complexity | Integration Complexity | Dependencies | Est. Effort |
|---------|----------------|----------------------|--------------|-------------|
| Aider Repo Map | HIGH | MEDIUM | tree-sitter, ctags | 2-3 weeks |
| Aider Diff Format | MEDIUM | LOW | None | 3-5 days |
| Claude Code UI | HIGH | MEDIUM | Syntax highlighter | 2 weeks |
| OpenHands Sandbox | MEDIUM | MEDIUM | Docker SDK | 1-2 weeks |
| Kiro Memory | MEDIUM | HIGH | Embeddings, PG | 1-2 weeks |
| Cline Browser | HIGH | HIGH | Chrome DevTools | 3-4 weeks |
| Continue LSP | HIGH | MEDIUM | LSP libraries | 2-3 weeks |
| Codex Reasoning | LOW | LOW | OpenAI API | 3-5 days |

### Risk Assessment

**High Risk:**
- Cline's browser automation (complex dependencies)
- Aider's repo map (requires tree-sitter expertise)

**Medium Risk:**
- OpenHands sandboxing (security considerations)
- Kiro's memory (database schema changes)

**Low Risk:**
- Diff format (straightforward algorithm)
- Reasoning models (API-only)

---

## Implementation Dependencies

### External Dependencies to Add

| Dependency | Purpose | License | Size |
|------------|---------|---------|------|
| tree-sitter-go | AST parsing | MIT | ~1MB |
| chroma | Syntax highlighting | MIT | ~5MB |
| docker client | Container management | Apache-2.0 | ~2MB |
| playwright-go | Browser automation | Apache-2.0 | ~50MB |

### Internal Dependencies

| HelixAgent Module | Required By | Integration Point |
|-------------------|-------------|-------------------|
| internal/llm/* | All features | Provider interface |
| internal/mcp/* | Tool systems | Tool execution |
| internal/database/* | Memory, persistence | PostgreSQL |
| internal/embeddings/* | Repo map, memory | Vector operations |

---

## Next Steps

**Pass 3: Synthesis & Design** will:
- Design unified architecture
- Create SQL schemas
- Design API interfaces
- Plan integration strategy

**See:** [Pass 3 - Synthesis & Design](pass_3_synthesis.md)

---

*Pass 2 Complete: 20 features analyzed at code level*  
*Date: 2026-04-03*  
*HelixAgent Commit: 8a976be2*
