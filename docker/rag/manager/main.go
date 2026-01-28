package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"sync"
	"time"
)

// RAGService represents a RAG service
type RAGService struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Type        string    `json:"type"`
	URL         string    `json:"url"`
	Status      string    `json:"status"`
	LastCheck   time.Time `json:"last_check"`
	Description string    `json:"description"`
}

// Manager manages RAG services
type Manager struct {
	services map[string]*RAGService
	mu       sync.RWMutex
}

func NewManager() *Manager {
	m := &Manager{
		services: make(map[string]*RAGService),
	}
	m.initializeServices()
	return m
}

func (m *Manager) initializeServices() {
	// Vector Databases
	m.services["pgvector"] = &RAGService{
		ID:          "pgvector",
		Name:        "pgvector",
		Type:        "vector_db",
		URL:         os.Getenv("PGVECTOR_URL"),
		Description: "PostgreSQL with pgvector extension",
	}
	m.services["chromadb"] = &RAGService{
		ID:          "chromadb",
		Name:        "ChromaDB",
		Type:        "vector_db",
		URL:         os.Getenv("CHROMADB_URL"),
		Description: "ChromaDB vector database",
	}
	m.services["qdrant"] = &RAGService{
		ID:          "qdrant",
		Name:        "Qdrant",
		Type:        "vector_db",
		URL:         os.Getenv("QDRANT_URL"),
		Description: "Qdrant vector database",
	}
	m.services["weaviate"] = &RAGService{
		ID:          "weaviate",
		Name:        "Weaviate",
		Type:        "vector_db",
		URL:         os.Getenv("WEAVIATE_URL"),
		Description: "Weaviate vector database",
	}

	// Embedding Services
	m.services["sentence-transformers"] = &RAGService{
		ID:          "sentence-transformers",
		Name:        "Sentence Transformers",
		Type:        "embedding",
		URL:         os.Getenv("ST_URL"),
		Description: "Sentence Transformers embedding service",
	}
	m.services["bge-m3"] = &RAGService{
		ID:          "bge-m3",
		Name:        "BGE-M3",
		Type:        "embedding",
		URL:         os.Getenv("BGE_URL"),
		Description: "BGE-M3 multilingual embeddings",
	}

	// RAG Services
	m.services["ragatouille"] = &RAGService{
		ID:          "ragatouille",
		Name:        "RAGatouille",
		Type:        "retrieval",
		URL:         os.Getenv("RAGATOUILLE_URL"),
		Description: "ColBERT-based retrieval",
	}
	m.services["hyde"] = &RAGService{
		ID:          "hyde",
		Name:        "HyDE",
		Type:        "retrieval",
		URL:         os.Getenv("HYDE_URL"),
		Description: "Hypothetical Document Embeddings",
	}
	m.services["multi-query"] = &RAGService{
		ID:          "multi-query",
		Name:        "Multi-Query",
		Type:        "retrieval",
		URL:         os.Getenv("MQ_URL"),
		Description: "Multi-query retrieval",
	}
	m.services["reranker"] = &RAGService{
		ID:          "reranker",
		Name:        "Reranker",
		Type:        "reranking",
		URL:         os.Getenv("RERANKER_URL"),
		Description: "Cross-encoder reranking",
	}

	// Set defaults
	for _, svc := range m.services {
		svc.Status = "unknown"
	}
}

func (m *Manager) checkService(svc *RAGService) {
	if svc.URL == "" {
		m.mu.Lock()
		svc.Status = "not_configured"
		m.mu.Unlock()
		return
	}

	healthURL := svc.URL + "/health"
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(healthURL)

	m.mu.Lock()
	defer m.mu.Unlock()

	if err != nil {
		svc.Status = "offline"
	} else {
		resp.Body.Close()
		if resp.StatusCode == 200 {
			svc.Status = "online"
		} else {
			svc.Status = "degraded"
		}
	}
	svc.LastCheck = time.Now()
}

func (m *Manager) startHealthChecker() {
	ticker := time.NewTicker(30 * time.Second)
	go func() {
		for range ticker.C {
			m.mu.RLock()
			services := make([]*RAGService, 0, len(m.services))
			for _, s := range m.services {
				services = append(services, s)
			}
			m.mu.RUnlock()

			for _, svc := range services {
				go m.checkService(svc)
			}
		}
	}()

	// Initial check
	m.mu.RLock()
	for _, svc := range m.services {
		go m.checkService(svc)
	}
	m.mu.RUnlock()
}

func (m *Manager) handleHealth(w http.ResponseWriter, r *http.Request) {
	m.mu.RLock()
	online := 0
	for _, s := range m.services {
		if s.Status == "online" {
			online++
		}
	}
	m.mu.RUnlock()

	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":    "healthy",
		"service":   "rag-manager",
		"services":  len(m.services),
		"online":    online,
		"timestamp": time.Now().Unix(),
	})
}

func (m *Manager) handleServices(w http.ResponseWriter, r *http.Request) {
	serviceType := r.URL.Query().Get("type")

	m.mu.RLock()
	services := make([]*RAGService, 0)
	for _, s := range m.services {
		if serviceType == "" || s.Type == serviceType {
			services = append(services, s)
		}
	}
	m.mu.RUnlock()

	json.NewEncoder(w).Encode(map[string]interface{}{
		"services": services,
		"count":    len(services),
	})
}

func (m *Manager) handleServiceStatus(w http.ResponseWriter, r *http.Request) {
	serviceID := r.URL.Query().Get("id")
	if serviceID == "" {
		http.Error(w, "service id required", http.StatusBadRequest)
		return
	}

	m.mu.RLock()
	svc, exists := m.services[serviceID]
	m.mu.RUnlock()

	if !exists {
		http.Error(w, "service not found", http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(svc)
}

func (m *Manager) handlePipeline(w http.ResponseWriter, r *http.Request) {
	// Return available pipeline configurations
	pipelines := []map[string]interface{}{
		{
			"name":        "basic",
			"description": "Basic embedding + vector search",
			"components":  []string{"sentence-transformers", "qdrant"},
		},
		{
			"name":        "hyde",
			"description": "HyDE-enhanced retrieval",
			"components":  []string{"hyde", "sentence-transformers", "qdrant"},
		},
		{
			"name":        "multi-query",
			"description": "Multi-query expansion with fusion",
			"components":  []string{"multi-query", "sentence-transformers", "qdrant"},
		},
		{
			"name":        "colbert",
			"description": "ColBERT late interaction retrieval",
			"components":  []string{"ragatouille"},
		},
		{
			"name":        "full",
			"description": "Full pipeline with reranking",
			"components":  []string{"multi-query", "sentence-transformers", "qdrant", "reranker"},
		},
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"pipelines": pipelines,
	})
}

func main() {
	port := os.Getenv("RAG_MANAGER_PORT")
	if port == "" {
		port = "8030"
	}

	manager := NewManager()
	manager.startHealthChecker()

	http.HandleFunc("/health", manager.handleHealth)
	http.HandleFunc("/services", manager.handleServices)
	http.HandleFunc("/service", manager.handleServiceStatus)
	http.HandleFunc("/pipelines", manager.handlePipeline)

	log.Printf("RAG Manager starting on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
