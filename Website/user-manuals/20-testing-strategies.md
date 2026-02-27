# User Manual 20: Testing Strategies

## Overview
Comprehensive testing approach for HelixAgent.

## Test Pyramid
1. Unit tests (70%)
2. Integration tests (20%)
3. E2E tests (10%)

## Unit Testing
```go
func TestFeature(t *testing.T) {
    // Arrange
    input := "test"
    
    // Act
    result := process(input)
    
    // Assert
    assert.Equal(t, expected, result)
}
```

## Integration Testing
```go
func TestDatabase(t *testing.T) {
    db := setupTestDB()
    defer teardown(db)
    
    // Test with real database
}
```

## Running Tests
```bash
make test          # All tests
make test-unit     # Unit only
make test-race     # With race detector
```
