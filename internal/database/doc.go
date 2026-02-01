// Package database provides PostgreSQL database access and repositories for HelixAgent.
//
// This package implements the data access layer using pgx/v5 for PostgreSQL
// connectivity, providing repository patterns for all persistent data.
//
// # Database Connection
//
// Connection is established using pgx connection pool:
//
//	config := &database.Config{
//	    Host:     "localhost",
//	    Port:     5432,
//	    User:     "helixagent",
//	    Password: "secret",
//	    Database: "helixagent_db",
//	}
//
//	pool, err := database.NewPool(config)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer func() { _ = pool.Close() }()
//
// # Repository Pattern
//
// Each domain entity has a corresponding repository:
//
//	type TaskRepository struct {
//	    pool *pgxpool.Pool
//	}
//
//	func (r *TaskRepository) Create(ctx context.Context, task *Task) error
//	func (r *TaskRepository) GetByID(ctx context.Context, id string) (*Task, error)
//	func (r *TaskRepository) Update(ctx context.Context, task *Task) error
//	func (r *TaskRepository) Delete(ctx context.Context, id string) error
//	func (r *TaskRepository) List(ctx context.Context, filter *TaskFilter) ([]*Task, error)
//
// # Available Repositories
//
//   - TaskRepository: Background task persistence
//   - SessionRepository: User session management
//   - AuditRepository: Audit log storage
//   - CogneeMemoryRepository: Cognee memory persistence
//   - ProviderRepository: Provider configuration storage
//
// # Transaction Support
//
// Transactions are supported for complex operations:
//
//	tx, err := pool.Begin(ctx)
//	if err != nil {
//	    return err
//	}
//	defer tx.Rollback(ctx)
//
//	// Perform operations
//	if err := repo.CreateWithTx(ctx, tx, entity); err != nil {
//	    return err
//	}
//
//	return tx.Commit(ctx)
//
// # Database Schema
//
// Key tables:
//
//	background_tasks   - Background task queue
//	sessions           - User/API sessions
//	audit_events       - Audit log
//	cognee_memories    - Cognee knowledge storage
//	provider_configs   - Provider configurations
//
// # Environment Configuration
//
// Database connection via environment variables:
//
//	DB_HOST      - PostgreSQL host (default: localhost)
//	DB_PORT      - PostgreSQL port (default: 5432)
//	DB_USER      - Database username
//	DB_PASSWORD  - Database password
//	DB_NAME      - Database name
//	DB_SSL_MODE  - SSL mode (disable, require, verify-ca, verify-full)
//
// # Connection Pooling
//
// The package uses pgxpool for connection pooling with configurable:
//
//   - Maximum connections
//   - Minimum connections
//   - Connection lifetime
//   - Health check interval
//
// # Key Files
//
//   - pool.go: Connection pool management
//   - background_task_repository.go: Task queue persistence
//   - session_repository.go: Session management
//   - cognee_memory_repository.go: Cognee integration
//   - migrations/: Database schema migrations
package database
