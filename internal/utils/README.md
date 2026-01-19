# Utils Package

The utils package provides common utility functions for HelixAgent.

## Overview

Includes:
- String manipulation
- File operations
- JSON helpers
- Retry logic
- Error handling utilities

## Key Functions

```go
// Retry with exponential backoff
err := utils.Retry(ctx, 3, func() error {
    return doOperation()
})

// Safe JSON unmarshaling
var data MyStruct
err := utils.SafeUnmarshal(jsonBytes, &data)
```
