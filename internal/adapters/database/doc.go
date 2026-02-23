// Package database provides adapters that bridge HelixAgent's database operations
// with the extracted digital.vasic.database module.
//
// # Overview
//
// This package provides two main integration paths:
//
//  1. Direct replacement: For code that can work directly with the database module's
//     interfaces (Database, Tx, Row, Rows), import digital.vasic.database directly.
//
//  2. Compatibility layer: For code that relies on HelixAgent's existing database
//     types (PostgresDB, MemoryDB), the internal/database package continues to
//     provide these types but now implements them using the extracted module.
//
// # Migration Path
//
// Existing code that uses "dev.helix.agent/internal/database" should continue
// working without changes. The internal/database package now uses the extracted
// module under the hood.
//
// New code can choose to:
//   - Continue using internal/database for HelixAgent-specific types
//   - Use digital.vasic.database directly for the generic interfaces
//   - Use this adapters package for bridging between the two
//
// # Example Usage
//
// Using the compatibility layer (no code changes needed):
//
//	import "dev.helix.agent/internal/database"
//
//	db, err := database.NewPostgresDB(cfg)
//	pool := db.GetPool() // Still works
//
// Using the extracted module directly:
//
//	import (
//	    dbmod "digital.vasic.database/pkg/database"
//	    "digital.vasic.database/pkg/postgres"
//	)
//
//	client := postgres.New(cfg)
//	client.Connect(ctx)
//	// Use client.Exec, client.Query, etc.
//
// Using the adapter for advanced scenarios:
//
//	import adapter "dev.helix.agent/internal/adapters/database"
//
//	client, err := adapter.NewClient(cfg)
//	db := client.Database() // Returns database.Database interface
//	pool := client.Pool()   // Returns *pgxpool.Pool
//
// # Repository Pattern
//
// The existing repositories in internal/database (UserRepository, ModelMetadataRepository,
// etc.) continue to use *pgxpool.Pool directly for optimal performance with pgx-specific
// features. The PostgresDB.GetPool() method provides access to this pool.
//
// For new repositories, consider using the generic repository pattern from
// digital.vasic.database/pkg/repository:
//
//	import "digital.vasic.database/pkg/repository"
//
//	type UserMapper struct{}
//	// Implement repository.EntityMapper[User]
//
//	repo := repository.NewGenericRepository(db, &UserMapper{})
//	user, err := repo.GetByID(ctx, id)
package database
