package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"
)

// ChatCompletionRequest represents an OpenAI-compatible chat completion request
type ChatCompletionRequest struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
}

// Message represents a chat message
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ChatCompletionResponse represents an OpenAI-compatible chat completion response
type ChatCompletionResponse struct {
	ID      string   `json:"id"`
	Object  string   `json:"object"`
	Created int64    `json:"created"`
	Model   string   `json:"model"`
	Choices []Choice `json:"choices"`
	Usage   Usage    `json:"usage"`
}

// Choice represents a response choice
type Choice struct {
	Index        int     `json:"index"`
	Message      Message `json:"message"`
	FinishReason string  `json:"finish_reason"`
}

// Usage represents token usage
type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// generateResponse generates a realistic response based on the prompt
func generateResponse(prompt string) string {
	prompt = strings.ToLower(prompt)

	// Code generation responses
	if strings.Contains(prompt, "factorial") && strings.Contains(prompt, "go") {
		return `Here's a Go function to calculate factorial with error handling:

` + "```go" + `
func factorial(n int) (int, error) {
    if n < 0 {
        return 0, fmt.Errorf("factorial is not defined for negative numbers: %d", n)
    }
    if n == 0 || n == 1 {
        return 1, nil
    }
    result := 1
    for i := 2; i <= n; i++ {
        result *= i
    }
    return result, nil
}
` + "```" + `

This function handles negative numbers by returning an error, and uses iterative approach for efficiency.`
	}

	if strings.Contains(prompt, "binary search") && strings.Contains(prompt, "python") {
		return `Here's a Python implementation of binary search:

` + "```python" + `
def binary_search(sorted_list, target):
    left, right = 0, len(sorted_list) - 1

    while left <= right:
        mid = (left + right) // 2
        if sorted_list[mid] == target:
            return mid
        elif sorted_list[mid] < target:
            left = mid + 1
        else:
            right = mid - 1

    return -1
` + "```" + `

This implementation has O(log n) time complexity and O(1) space complexity.`
	}

	if strings.Contains(prompt, "typescript") && strings.Contains(prompt, "class") && strings.Contains(prompt, "user") {
		return `Here's a TypeScript class for user management:

` + "```typescript" + `
class User {
    private id: string;
    private email: string;
    private name: string;

    constructor(id: string, email: string, name: string) {
        this.id = id;
        this.email = email;
        this.name = name;
    }

    validateEmail(): boolean {
        const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
        return emailRegex.test(this.email);
    }

    validateName(): boolean {
        return this.name.length >= 2 && this.name.length <= 100;
    }

    isValid(): boolean {
        return this.validateEmail() && this.validateName();
    }
}
` + "```" + `

This class includes validation methods for email and name.`
	}

	// Code review responses
	if strings.Contains(prompt, "divide") && strings.Contains(prompt, "bug") {
		return `This code has a critical bug: **Division by zero**.

**Issue:** The function does not check if ` + "`b`" + ` is zero before performing division. Dividing by zero will cause a runtime panic in Go.

**Fix:**
` + "```go" + `
func divide(a, b int) (int, error) {
    if b == 0 {
        return 0, fmt.Errorf("division by zero")
    }
    return a / b, nil
}
` + "```" + `

Always validate denominators before division operations.`
	}

	if strings.Contains(prompt, "sql") && strings.Contains(prompt, "injection") {
		return `This code has a critical **SQL Injection vulnerability**.

**Issue:** User input is directly concatenated into the SQL query without sanitization or parameterization.

**Attack Example:**
An attacker could input: ` + "`1; DROP TABLE users; --`" + `

**Fix - Use Parameterized Queries:**
` + "```go" + `
func getUser(db *sql.DB, userInput string) (*User, error) {
    query := "SELECT * FROM users WHERE id = $1"
    row := db.QueryRow(query, userInput)
    // ... scan row into User struct
}
` + "```" + `

Never concatenate user input directly into SQL queries. Always use parameterized queries or prepared statements.`
	}

	// Reasoning responses
	if strings.Contains(prompt, "sheep") && strings.Contains(prompt, "17") {
		return `**Answer: 9 sheep are left.**

Reasoning:
- The farmer starts with 17 sheep
- "All but 9 run away" means all sheep EXCEPT 9 run away
- So 9 sheep remain with the farmer

This is a classic trick question where the phrasing "all but 9" means "all except 9", not "17 minus 9".`
	}

	if strings.Contains(prompt, "syllogism") || (strings.Contains(prompt, "all a are b") && strings.Contains(prompt, "all b are c")) {
		return `**Yes, all A are also C.**

This is a classic syllogism following the transitive property:
1. Premise 1: All A are B
2. Premise 2: All B are C
3. Conclusion: All A are C

**Reasoning:**
- If every member of set A is a member of set B
- And every member of set B is a member of set C
- Then every member of set A must also be a member of set C

This is valid deductive reasoning following the Barbara syllogism form (AAA-1).`
	}

	// Quality responses
	if strings.Contains(prompt, "rest") && strings.Contains(prompt, "api") && strings.Contains(prompt, "best practices") {
		return `# REST API Best Practices

## Naming Conventions
1. **Use nouns, not verbs**: ` + "`/users`" + ` not ` + "`/getUsers`" + `
2. **Use plural nouns**: ` + "`/users`" + ` not ` + "`/user`" + `
3. **Use kebab-case**: ` + "`/user-profiles`" + ` not ` + "`/userProfiles`" + `
4. **Use lowercase**: ` + "`/orders`" + ` not ` + "`/Orders`" + `

## HTTP Methods
| Method | Usage | Idempotent |
|--------|-------|------------|
| GET | Retrieve resources | Yes |
| POST | Create resources | No |
| PUT | Replace resources | Yes |
| PATCH | Partial update | No |
| DELETE | Remove resources | Yes |

## Scalability Tips
- Use pagination for large collections
- Implement caching with ETags
- Version your API (` + "`/v1/users`" + `)
- Use rate limiting to protect resources`
	}

	if strings.Contains(prompt, "capital") && strings.Contains(prompt, "france") {
		return `The capital of France is **Paris**.

Paris has been the capital of France since the 10th century. It is the country's largest city with a population of over 2 million in the city proper and over 12 million in the metropolitan area. It is also the political, economic, and cultural center of France.`
	}

	// Consensus response
	if strings.Contains(prompt, "2 + 2") || strings.Contains(prompt, "2+2") {
		return "4"
	}

	// Default response
	return "This is a response from the mock HelixAgent API server. The prompt was processed successfully."
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "healthy"})
}

func chatCompletionsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req ChatCompletionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Get the last user message
	var prompt string
	for i := len(req.Messages) - 1; i >= 0; i-- {
		if req.Messages[i].Role == "user" {
			prompt = req.Messages[i].Content
			break
		}
	}

	response := generateResponse(prompt)

	resp := ChatCompletionResponse{
		ID:      fmt.Sprintf("chatcmpl-%d", time.Now().UnixNano()),
		Object:  "chat.completion",
		Created: time.Now().Unix(),
		Model:   req.Model,
		Choices: []Choice{
			{
				Index: 0,
				Message: Message{
					Role:    "assistant",
					Content: response,
				},
				FinishReason: "stop",
			},
		},
		Usage: Usage{
			PromptTokens:     len(prompt) / 4,
			CompletionTokens: len(response) / 4,
			TotalTokens:      (len(prompt) + len(response)) / 4,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func main() {
	http.HandleFunc("/health", healthHandler)
	http.HandleFunc("/v1/chat/completions", chatCompletionsHandler)

	port := "8080"
	log.Printf("Mock HelixAgent API server starting on port %s...", port)
	log.Printf("Endpoints:")
	log.Printf("  GET  /health")
	log.Printf("  POST /v1/chat/completions")

	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
