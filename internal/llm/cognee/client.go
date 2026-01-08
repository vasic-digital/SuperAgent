package llm

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os/exec"
	"time"

	"github.com/helixagent/helixagent/internal/config"
	"github.com/helixagent/helixagent/internal/models"
)

type Client struct {
	baseURL string
	apiKey  string
	client  *http.Client
}

type MemoryRequest struct {
	Content     string `json:"content"`
	DatasetName string `json:"dataset_name"`
	ContentType string `json:"content_type"`
}

type MemoryResponse struct {
	VectorID   string                 `json:"vector_id"`
	GraphNodes map[string]interface{} `json:"graph_nodes"`
}

type SearchRequest struct {
	Query       string `json:"query"`
	DatasetName string `json:"dataset_name"`
	Limit       int    `json:"limit"`
}

type SearchResponse struct {
	Results []models.MemorySource `json:"results"`
}

type CognifyRequest struct {
	Datasets []string `json:"datasets,omitempty"`
}

type CognifyResponse struct {
	Status string `json:"status"`
}

type InsightsRequest struct {
	Query    string   `json:"query"`
	Datasets []string `json:"datasets,omitempty"`
	Limit    int      `json:"limit,omitempty"`
}

type InsightsResponse struct {
	Insights []map[string]interface{} `json:"insights"`
}

type CodePipelineRequest struct {
	Code        string `json:"code"`
	DatasetName string `json:"dataset_name"`
	Language    string `json:"language,omitempty"`
}

type CodePipelineResponse struct {
	Processed bool                   `json:"processed"`
	Results   map[string]interface{} `json:"results,omitempty"`
}

type DatasetRequest struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

type DatasetResponse struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description,omitempty"`
	CreatedAt   string                 `json:"created_at"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

type DatasetsResponse struct {
	Datasets []DatasetResponse `json:"datasets"`
	Total    int               `json:"total"`
}

type VisualizeRequest struct {
	DatasetName string `json:"dataset_name,omitempty"`
	Format      string `json:"format,omitempty"` // "json", "graphml", etc.
}

type VisualizeResponse struct {
	Graph map[string]interface{} `json:"graph"`
}

func NewClient(cfg *config.Config) *Client {
	return &Client{
		baseURL: cfg.Cognee.BaseURL,
		apiKey:  cfg.Cognee.APIKey,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (c *Client) AddMemory(req *MemoryRequest) (*MemoryResponse, error) {
	url := fmt.Sprintf("%s/api/memory", c.baseURL)

	data, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	httpReq, err := http.NewRequest("POST", url, bytes.NewBuffer(data))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	if c.apiKey != "" {
		httpReq.Header.Set("Authorization", "Bearer "+c.apiKey)
	}

	resp, err := c.client.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("cognee API error: %s", resp.Status)
	}

	var response MemoryResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, err
	}

	return &response, nil
}

func (c *Client) SearchMemory(req *SearchRequest) (*SearchResponse, error) {
	url := fmt.Sprintf("%s/api/search", c.baseURL)

	data, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	httpReq, err := http.NewRequest("POST", url, bytes.NewBuffer(data))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	if c.apiKey != "" {
		httpReq.Header.Set("Authorization", "Bearer "+c.apiKey)
	}

	resp, err := c.client.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("cognee API error: %s", resp.Status)
	}

	var response SearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, err
	}

	return &response, nil
}

// Cognify processes data into knowledge graphs
func (c *Client) Cognify(req *CognifyRequest) (*CognifyResponse, error) {
	url := fmt.Sprintf("%s/api/cognify", c.baseURL)

	data, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	httpReq, err := http.NewRequest("POST", url, bytes.NewBuffer(data))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	if c.apiKey != "" {
		httpReq.Header.Set("Authorization", "Bearer "+c.apiKey)
	}

	resp, err := c.client.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("cognee API error: %s", resp.Status)
	}

	var response CognifyResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, err
	}

	return &response, nil
}

// SearchInsights performs insight-based search using graph reasoning
func (c *Client) SearchInsights(req *InsightsRequest) (*InsightsResponse, error) {
	url := fmt.Sprintf("%s/api/search", c.baseURL)

	// Add search type for insights
	insightsReq := map[string]interface{}{
		"query":       req.Query,
		"datasets":    req.Datasets,
		"limit":       req.Limit,
		"search_type": "INSIGHTS",
	}

	data, err := json.Marshal(insightsReq)
	if err != nil {
		return nil, err
	}

	httpReq, err := http.NewRequest("POST", url, bytes.NewBuffer(data))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	if c.apiKey != "" {
		httpReq.Header.Set("Authorization", "Bearer "+c.apiKey)
	}

	resp, err := c.client.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("cognee API error: %s", resp.Status)
	}

	var response InsightsResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, err
	}

	return &response, nil
}

// SearchGraphCompletion performs LLM-powered graph completion search
func (c *Client) SearchGraphCompletion(query string, datasets []string, limit int) (*SearchResponse, error) {
	url := fmt.Sprintf("%s/api/search", c.baseURL)

	searchReq := map[string]interface{}{
		"query":       query,
		"datasets":    datasets,
		"limit":       limit,
		"search_type": "GRAPH_COMPLETION",
	}

	data, err := json.Marshal(searchReq)
	if err != nil {
		return nil, err
	}

	httpReq, err := http.NewRequest("POST", url, bytes.NewBuffer(data))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	if c.apiKey != "" {
		httpReq.Header.Set("Authorization", "Bearer "+c.apiKey)
	}

	resp, err := c.client.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("cognee API error: %s", resp.Status)
	}

	var response SearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, err
	}

	return &response, nil
}

// ProcessCodePipeline processes code through Cognee's code understanding pipeline
func (c *Client) ProcessCodePipeline(req *CodePipelineRequest) (*CodePipelineResponse, error) {
	url := fmt.Sprintf("%s/api/code-pipeline/index", c.baseURL)

	data, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	httpReq, err := http.NewRequest("POST", url, bytes.NewBuffer(data))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	if c.apiKey != "" {
		httpReq.Header.Set("Authorization", "Bearer "+c.apiKey)
	}

	resp, err := c.client.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("cognee API error: %s", resp.Status)
	}

	var response CodePipelineResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, err
	}

	return &response, nil
}

// CreateDataset creates a new dataset
func (c *Client) CreateDataset(req *DatasetRequest) (*DatasetResponse, error) {
	url := fmt.Sprintf("%s/api/datasets", c.baseURL)

	data, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	httpReq, err := http.NewRequest("POST", url, bytes.NewBuffer(data))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	if c.apiKey != "" {
		httpReq.Header.Set("Authorization", "Bearer "+c.apiKey)
	}

	resp, err := c.client.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("cognee API error: %s", resp.Status)
	}

	var response DatasetResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, err
	}

	return &response, nil
}

// ListDatasets retrieves all datasets
func (c *Client) ListDatasets() (*DatasetsResponse, error) {
	url := fmt.Sprintf("%s/api/datasets", c.baseURL)

	httpReq, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	if c.apiKey != "" {
		httpReq.Header.Set("Authorization", "Bearer "+c.apiKey)
	}

	resp, err := c.client.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("cognee API error: %s", resp.Status)
	}

	var response DatasetsResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, err
	}

	return &response, nil
}

// VisualizeGraph retrieves graph visualization data
func (c *Client) VisualizeGraph(req *VisualizeRequest) (*VisualizeResponse, error) {
	url := fmt.Sprintf("%s/api/visualize", c.baseURL)

	data, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	httpReq, err := http.NewRequest("GET", url, bytes.NewBuffer(data))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	if c.apiKey != "" {
		httpReq.Header.Set("Authorization", "Bearer "+c.apiKey)
	}

	resp, err := c.client.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("cognee API error: %s", resp.Status)
	}

	var response VisualizeResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, err
	}

	return &response, nil
}

// DeleteData removes data from a dataset
func (c *Client) DeleteData(datasetName string, dataIDs []string) error {
	url := fmt.Sprintf("%s/api/delete", c.baseURL)

	deleteReq := map[string]interface{}{
		"dataset_name": datasetName,
		"data_ids":     dataIDs,
	}

	data, err := json.Marshal(deleteReq)
	if err != nil {
		return err
	}

	httpReq, err := http.NewRequest("DELETE", url, bytes.NewBuffer(data))
	if err != nil {
		return err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	if c.apiKey != "" {
		httpReq.Header.Set("Authorization", "Bearer "+c.apiKey)
	}

	resp, err := c.client.Do(httpReq)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("cognee API error: %s", resp.Status)
	}

	return nil
}

// GetBaseURL returns the base URL of the client
func (c *Client) GetBaseURL() string {
	return c.baseURL
}

// AutoContainerize starts Cognee in a container if not running
func (c *Client) AutoContainerize() error {
	// Check if Cognee is already running by testing connection
	if c.testConnection() {
		return nil // Already running
	}

	// Try to start Cognee container
	return c.startCogneeContainer()
}

// testConnection checks if Cognee is already running
func (c *Client) testConnection() bool {
	url := fmt.Sprintf("%s/health", c.baseURL)

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	return resp.StatusCode == http.StatusOK
}

// startCogneeContainer starts Cognee using Docker Compose
func (c *Client) startCogneeContainer() error {
	// Check if docker compose is available
	if _, err := exec.LookPath("docker"); err != nil {
		return fmt.Errorf("docker not found in PATH: %w", err)
	}

	// Try docker compose first (newer syntax), fall back to docker-compose
	var cmd *exec.Cmd
	if _, err := exec.LookPath("docker compose"); err == nil {
		cmd = exec.Command("docker", "compose", "up", "-d", "cognee", "chromadb")
	} else if _, err := exec.LookPath("docker-compose"); err == nil {
		cmd = exec.Command("docker-compose", "up", "-d", "cognee", "chromadb")
	} else {
		return fmt.Errorf("neither 'docker compose' nor 'docker-compose' found in PATH")
	}

	// Set working directory to project root (assuming it's where docker-compose.yml is)
	cmd.Dir = "/media/milosvasic/DATA4TB/Projects/HelixAgent"

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to start Cognee containers: %w, output: %s", err, string(output))
	}

	// Wait for services to be healthy
	time.Sleep(10 * time.Second)

	// Verify the service is now running
	if c.testConnection() {
		return nil
	}

	return fmt.Errorf("Cognee containers started but service is not responding. Check docker compose logs")
}
