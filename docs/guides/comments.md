# Writing Effective Go Comments

Pallas extracts documentation from your Go comments. Writing good comments means better documentation.

## The Golden Rules

1. **Comment exported symbols**: Every exported function, type, and constant should have a comment
2. **Start with the name**: Comments should start with the symbol name
3. **Be concise but complete**: Explain *what*, not *how*

## Function Comments

```go
// ParseConfig reads and validates the configuration file at the given path.
// It returns an error if the file doesn't exist or contains invalid YAML.
func ParseConfig(path string) (*Config, error) {
    // ...
}
```

**Good:**
- Starts with the function name
- Explains what it does
- Documents error conditions

**Bad:**
```go
// This function parses config
func ParseConfig(path string) (*Config, error) {}
```

## Type Comments

```go
// Server represents an HTTP server with graceful shutdown support.
// It wraps the standard library's http.Server with additional lifecycle methods.
type Server struct {
    // Addr is the TCP address to listen on (e.g., ":8080").
    Addr string

    // ReadTimeout is the maximum duration for reading the entire request.
    ReadTimeout time.Duration
}
```

**Key points:**
- Document the type's purpose
- Document exported fields
- Include example values where helpful

## Interface Comments

```go
// Store defines the contract for data persistence.
// Implementations must be safe for concurrent use.
type Store interface {
    // Get retrieves a value by key. Returns ErrNotFound if the key doesn't exist.
    Get(ctx context.Context, key string) ([]byte, error)

    // Set stores a value with the given key. It overwrites any existing value.
    Set(ctx context.Context, key string, value []byte) error
}
```

## Package Comments

Place a comment directly before the `package` declaration:

```go
// Package config provides configuration loading and validation.
//
// It supports YAML, JSON, and TOML formats with automatic detection
// based on file extension.
package config
```

For larger packages, create a `doc.go` file with just the package comment.

## Example Sections

Pallas recognizes `Example:` sections in comments:

```go
// NewClient creates a configured HTTP client with sensible defaults.
//
// Example:
//
//     client := NewClient(WithTimeout(30 * time.Second))
//     resp, err := client.Get("https://api.example.com")
func NewClient(opts ...Option) *Client {
    // ...
}
```

## Notes and Warnings

Use a `Note:` prefix for important information:

```go
// Close releases all resources associated with the connection.
//
// Note: Close must be called exactly once. Calling it multiple times
// results in undefined behavior.
func (c *Conn) Close() error {
    // ...
}
```

## What NOT to Comment

- **Don't explain the obvious**: `// i is the loop counter` is noise
- **Don't comment internal logic**: Inline comments should explain *why*, not *what*
- **Don't leave TODO comments in public APIs**: Address them or remove them
