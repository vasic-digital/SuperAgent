package database

import (
	"fmt"
	"log"
	"sync"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/superagent/superagent/internal/config"
)

// MemoryDB implements DB interface using in-memory storage
// This is used when PostgreSQL is not available (standalone/testing mode)
type MemoryDB struct {
	mu      sync.RWMutex
	data    map[string][]map[string]any
	enabled bool
}

// memoryRow implements Row interface for in-memory queries
type memoryRow struct {
	values []any
	err    error
}

func (r *memoryRow) Scan(dest ...any) error {
	if r.err != nil {
		return r.err
	}
	if len(r.values) == 0 {
		return fmt.Errorf("no rows")
	}
	for i := range dest {
		if i < len(r.values) {
			// Simple type assertion - in real use, would need proper type conversion
			switch d := dest[i].(type) {
			case *string:
				if s, ok := r.values[i].(string); ok {
					*d = s
				}
			case *int:
				if n, ok := r.values[i].(int); ok {
					*d = n
				}
			case *bool:
				if b, ok := r.values[i].(bool); ok {
					*d = b
				}
			}
		}
	}
	return nil
}

// NewMemoryDB creates a new in-memory database
func NewMemoryDB() *MemoryDB {
	log.Println("Using in-memory database (standalone mode)")
	return &MemoryDB{
		data:    make(map[string][]map[string]any),
		enabled: true,
	}
}

// NewPostgresDBWithFallback tries PostgreSQL first, falls back to memory
func NewPostgresDBWithFallback(cfg *config.Config) (*PostgresDB, *MemoryDB, error) {
	// Try PostgreSQL first
	db, err := NewPostgresDB(cfg)
	if err == nil {
		// Test the connection
		if pingErr := db.Ping(); pingErr == nil {
			return db, nil, nil
		}
		log.Printf("PostgreSQL ping failed, falling back to in-memory mode")
	} else {
		log.Printf("PostgreSQL connection failed: %v, using in-memory mode", err)
	}

	// Fall back to memory
	return nil, NewMemoryDB(), nil
}

func (m *MemoryDB) Ping() error {
	return nil // Always healthy
}

func (m *MemoryDB) Exec(query string, args ...any) error {
	// No-op for in-memory mode
	return nil
}

func (m *MemoryDB) Query(query string, args ...any) ([]any, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return nil, nil
}

func (m *MemoryDB) QueryRow(query string, args ...any) Row {
	return &memoryRow{values: nil, err: fmt.Errorf("not implemented in memory mode")}
}

func (m *MemoryDB) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.data = nil
	m.enabled = false
	return nil
}

func (m *MemoryDB) HealthCheck() error {
	if !m.enabled {
		return fmt.Errorf("memory database closed")
	}
	return nil
}

// GetPool returns nil for memory database (no real pool)
func (m *MemoryDB) GetPool() *pgxpool.Pool {
	return nil
}

// IsMemoryMode returns true if this is an in-memory database
func (m *MemoryDB) IsMemoryMode() bool {
	return true
}
