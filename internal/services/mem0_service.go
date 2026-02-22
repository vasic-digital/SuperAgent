package services

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"dev.helix.agent/internal/config"
	"dev.helix.agent/internal/memory"
	"dev.helix.agent/internal/models"
	"github.com/sirupsen/logrus"
)

type Mem0Service struct {
	manager *memory.Manager
	store   memory.MemoryStore
	enabled bool
	config  *config.MemoryConfig
	logger  *logrus.Logger
	stats   *Mem0Stats
	mu      sync.RWMutex
}

type Mem0Stats struct {
	mu                  sync.RWMutex
	TotalMemoriesStored int64
	TotalSearches       int64
	TotalEntities       int64
	TotalRelationships  int64
	AverageSearchLatMs  float64
	LastActivity        time.Time
	ErrorCount          int64
}

func NewMem0Service(
	store memory.MemoryStore,
	extractor memory.MemoryExtractor,
	summarizer memory.MemorySummarizer,
	embedder memory.Embedder,
	cfg *config.MemoryConfig,
	logger *logrus.Logger,
) *Mem0Service {
	if logger == nil {
		logger = logrus.New()
	}
	if cfg == nil {
		cfg = &config.MemoryConfig{
			Enabled:          true,
			Provider:         "mem0",
			MaxContextLength: 4000,
			RetentionDays:    0,
		}
	}
	manager := memory.NewManager(store, extractor, summarizer, embedder, nil, logger)
	return &Mem0Service{
		manager: manager,
		store:   store,
		enabled: cfg.Enabled,
		config:  cfg,
		logger:  logger,
		stats:   &Mem0Stats{},
	}
}

func (s *Mem0Service) IsEnabled() bool { return s.enabled }

func (s *Mem0Service) AddMemory(ctx context.Context, req *Mem0Request) (*Mem0Response, error) {
	if !s.enabled {
		return nil, fmt.Errorf("Mem0 memory service is disabled")
	}
	s.stats.mu.Lock()
	s.stats.TotalMemoriesStored++
	s.stats.LastActivity = time.Now()
	s.stats.mu.Unlock()

	mem := &memory.Memory{
		UserID:     req.UserID,
		SessionID:  req.SessionID,
		Content:    req.Content,
		Type:       memoryTypeFromString(req.Type),
		Category:   req.Category,
		Metadata:   req.Metadata,
		Importance: req.Importance,
	}
	if err := s.manager.AddMemory(ctx, mem); err != nil {
		s.stats.mu.Lock()
		s.stats.ErrorCount++
		s.stats.mu.Unlock()
		return nil, fmt.Errorf("failed to add memory: %w", err)
	}
	return &Mem0Response{ID: mem.ID, CreatedAt: mem.CreatedAt, Success: true}, nil
}

func (s *Mem0Service) SearchMemory(ctx context.Context, req *Mem0SearchRequest) (*Mem0SearchResponse, error) {
	if !s.enabled {
		return nil, fmt.Errorf("Mem0 memory service is disabled")
	}
	start := time.Now()
	opts := &memory.SearchOptions{
		UserID:       req.UserID,
		SessionID:    req.SessionID,
		TopK:         req.Limit,
		MinScore:     req.MinScore,
		IncludeGraph: req.IncludeGraph,
	}
	memories, err := s.manager.Search(ctx, req.Query, opts)
	if err != nil {
		s.stats.mu.Lock()
		s.stats.ErrorCount++
		s.stats.mu.Unlock()
		return nil, fmt.Errorf("failed to search memory: %w", err)
	}
	latency := time.Since(start).Milliseconds()
	s.stats.mu.Lock()
	s.stats.TotalSearches++
	s.stats.LastActivity = time.Now()
	n := s.stats.TotalSearches
	s.stats.AverageSearchLatMs = (s.stats.AverageSearchLatMs*float64(n-1) + float64(latency)) / float64(n)
	s.stats.mu.Unlock()

	results := make([]Mem0SearchResult, len(memories))
	for i, mem := range memories {
		results[i] = Mem0SearchResult{
			ID:          mem.ID,
			Content:     mem.Content,
			Summary:     mem.Summary,
			Type:        string(mem.Type),
			Category:    mem.Category,
			Importance:  mem.Importance,
			AccessCount: mem.AccessCount,
			CreatedAt:   mem.CreatedAt,
			UpdatedAt:   mem.UpdatedAt,
		}
	}
	return &Mem0SearchResponse{Results: results, Total: len(results)}, nil
}

func (s *Mem0Service) AddFromMessages(ctx context.Context, messages []memory.Message, userID, sessionID string) ([]*memory.Memory, error) {
	if !s.enabled {
		return nil, fmt.Errorf("Mem0 memory service is disabled")
	}
	memories, err := s.manager.AddFromMessages(ctx, messages, userID, sessionID)
	if err != nil {
		s.stats.mu.Lock()
		s.stats.ErrorCount++
		s.stats.mu.Unlock()
		return nil, err
	}
	s.stats.mu.Lock()
	s.stats.TotalMemoriesStored += int64(len(memories))
	s.stats.LastActivity = time.Now()
	s.stats.mu.Unlock()
	return memories, nil
}

func (s *Mem0Service) GetContext(ctx context.Context, query, userID string, maxTokens int) (string, error) {
	if !s.enabled {
		return "", fmt.Errorf("Mem0 memory service is disabled")
	}
	return s.manager.GetContext(ctx, query, userID, maxTokens)
}

func (s *Mem0Service) GetUserMemories(ctx context.Context, userID string, limit, offset int) ([]*memory.Memory, error) {
	if !s.enabled {
		return nil, fmt.Errorf("Mem0 memory service is disabled")
	}
	return s.manager.GetUserMemories(ctx, userID, &memory.ListOptions{Limit: limit, Offset: offset})
}

func (s *Mem0Service) DeleteMemory(ctx context.Context, id string) error {
	if !s.enabled {
		return fmt.Errorf("Mem0 memory service is disabled")
	}
	return s.manager.DeleteMemory(ctx, id)
}

func (s *Mem0Service) DeleteUserMemories(ctx context.Context, userID string) error {
	if !s.enabled {
		return fmt.Errorf("Mem0 memory service is disabled")
	}
	return s.manager.DeleteUserMemories(ctx, userID)
}

func (s *Mem0Service) GetStats() *Mem0Stats { return s.stats }

func (s *Mem0Service) GetRelatedEntities(ctx context.Context, query string, limit int) ([]*memory.Entity, error) {
	if !s.enabled {
		return nil, fmt.Errorf("Mem0 memory service is disabled")
	}
	return s.manager.GetRelatedEntities(ctx, query, limit)
}

func (s *Mem0Service) GetEntityRelationships(ctx context.Context, entityID string) ([]*memory.Relationship, error) {
	if !s.enabled {
		return nil, fmt.Errorf("Mem0 memory service is disabled")
	}
	return s.manager.GetEntityRelationships(ctx, entityID)
}

func (s *Mem0Service) EnhanceRequest(ctx context.Context, req *models.LLMRequest) error {
	if !s.enabled {
		return nil
	}
	keywords := s.extractKeywords(req)
	if keywords == "" {
		return nil
	}
	contextStr, err := s.GetContext(ctx, keywords, req.UserID, 1000)
	if err != nil {
		s.logger.WithError(err).Debug("Failed to get memory context")
		return nil
	}
	if contextStr != "" {
		if req.Memory == nil {
			req.Memory = make(map[string]string)
		}
		req.Memory["mem0_context"] = contextStr
	}
	return nil
}

func (s *Mem0Service) GetMemorySources(ctx context.Context, req *models.LLMRequest) ([]models.MemorySource, error) {
	if !s.enabled {
		return nil, fmt.Errorf("Mem0 memory service is disabled")
	}
	keywords := s.extractKeywords(req)
	memories, err := s.manager.Search(ctx, keywords, &memory.SearchOptions{UserID: req.UserID, TopK: 10})
	if err != nil {
		return nil, err
	}
	sources := make([]models.MemorySource, len(memories))
	for i, mem := range memories {
		sources[i] = models.MemorySource{
			DatasetName:    "mem0",
			Content:        mem.Content,
			RelevanceScore: mem.Importance,
			SourceType:     "mem0",
		}
	}
	return sources, nil
}

func (s *Mem0Service) extractKeywords(req *models.LLMRequest) string {
	var parts []string
	if req.Prompt != "" {
		parts = append(parts, req.Prompt)
	}
	for _, msg := range req.Messages {
		if msg.Content != "" {
			parts = append(parts, msg.Content)
		}
	}
	keywords := strings.Join(parts, " ")
	if len(keywords) > 200 {
		keywords = keywords[:200]
	}
	return keywords
}

func memoryTypeFromString(s string) memory.MemoryType {
	switch strings.ToLower(s) {
	case "episodic", "conversation", "event":
		return memory.MemoryTypeEpisodic
	case "semantic", "fact", "knowledge":
		return memory.MemoryTypeSemantic
	case "procedural", "howto", "procedure":
		return memory.MemoryTypeProcedural
	case "working", "temporary", "context":
		return memory.MemoryTypeWorking
	default:
		return memory.MemoryTypeEpisodic
	}
}

type Mem0Request struct {
	UserID     string                 `json:"user_id"`
	SessionID  string                 `json:"session_id,omitempty"`
	Content    string                 `json:"content"`
	Type       string                 `json:"type,omitempty"`
	Category   string                 `json:"category,omitempty"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
	Importance float64                `json:"importance,omitempty"`
}

type Mem0Response struct {
	ID        string    `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	Success   bool      `json:"success"`
}

type Mem0SearchRequest struct {
	Query        string  `json:"query"`
	UserID       string  `json:"user_id,omitempty"`
	SessionID    string  `json:"session_id,omitempty"`
	Limit        int     `json:"limit"`
	MinScore     float64 `json:"min_score"`
	IncludeGraph bool    `json:"include_graph"`
}

type Mem0SearchResponse struct {
	Results []Mem0SearchResult `json:"results"`
	Total   int                `json:"total"`
}

type Mem0SearchResult struct {
	ID          string                 `json:"id"`
	Content     string                 `json:"content"`
	Summary     string                 `json:"summary,omitempty"`
	Type        string                 `json:"type"`
	Category    string                 `json:"category,omitempty"`
	Importance  float64                `json:"importance"`
	AccessCount int                    `json:"access_count"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}
