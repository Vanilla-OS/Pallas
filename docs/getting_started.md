# Getting Started

Pallas is a modern, aesthetically pleasing documentation generator for Go projects. It parses your Go code, extracts documentation from comments, and generates a beautiful static HTML site.

## Installation

```bash
go install github.com/vanilla-os/pallas/cmd/pallas@latest
```

## Quick Start

Navigate to your Go project root and run:

```bash
pallas
```

This generates documentation in the `dist/` directory.

## Command Line Options

| Flag | Short | Description | Default |
|------|-------|-------------|---------|
| `--dest` | `-d` | Output directory | `./dist` |
| `--title` | `-t` | Project title | Auto-detected |
| `--readme` | `-r` | README file path | `README.md` |
| `--project` | `-p` | Project root | `.` |

### Examples

```bash
# Generate docs with custom title
pallas --title "My Awesome SDK"

# Output to a different directory
pallas --dest ./documentation

# Specify a different README
pallas --readme docs/README.md
```

## Project Structure

Pallas works best with standard Go project layouts:

```
my-project/
├── cmd/           # Main applications
├── pkg/           # Library code (parsed by Pallas)
├── internal/      # Private code (also parsed)
├── docs/          # Documentation pages (NEW!)
│   ├── getting_started.md
│   └── guides/
│       └── comments.md
├── go.mod
└── README.md      # Used as the Overview page
```

## Features

- **Automatic Package Discovery**: Recursively finds all Go packages
- **Cross-Reference Linking**: Types are automatically linked
- **Syntax Highlighting**: Code blocks with full highlight.js support
- **Dark Mode**: Respects system preferences
- **Search**: Client-side full-text search
- **docs/ Support**: Markdown files in `docs/` become documentation pages
