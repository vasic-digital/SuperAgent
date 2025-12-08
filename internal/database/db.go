package database

type DB interface {
	Ping() error
	Exec(query string, args ...any) error
	Query(query string, args ...any) ([]any, error)
}

type MockDB struct{}

func (m *MockDB) Ping() error                                    { return nil }
func (m *MockDB) Exec(query string, args ...any) error           { return nil }
func (m *MockDB) Query(query string, args ...any) ([]any, error) { return nil, nil }

// Connect returns a mock database handle. In a full implementation this would
// establish a real Postgres connection via pgx.
func Connect() (DB, error) {
	return &MockDB{}, nil
}
