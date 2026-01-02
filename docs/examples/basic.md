# Basic Usage Example

This example shows how to set up Pallas for a simple Go project.

## Project Structure

```
my-project/
├── cmd/
│   └── myapp/
│       └── main.go
├── pkg/
│   └── config/
│       └── config.go
├── docs/
│   └── getting_started.md
├── go.mod
└── README.md
```

## Step 1: Install Pallas

```bash
go install github.com/vanilla-os/pallas/cmd/pallas@latest
```

## Step 2: Write Your Code with Comments

```go
// Package config handles application configuration.
package config

// Config holds all configuration options for the application.
type Config struct {
    // Host is the server address to listen on.
    Host string
    
    // Port is the server port.
    Port int
    
    // Debug enables verbose logging when true.
    Debug bool
}

// New creates a Config with sensible defaults.
func New() *Config {
    return &Config{
        Host:  "localhost",
        Port:  8080,
        Debug: false,
    }
}

// Validate checks if the configuration is valid.
// It returns an error if Port is outside the valid range.
func (c *Config) Validate() error {
    if c.Port < 1 || c.Port > 65535 {
        return fmt.Errorf("port must be between 1 and 65535")
    }
    return nil
}
```

## Step 3: Generate Documentation

```bash
cd my-project
pallas --title "My Project"
```

## Step 4: View the Results

Open `dist/index.html` in your browser. You'll see:

- Your README as the overview page
- Package documentation with all types and functions
- Searchable documentation with Ctrl+K
