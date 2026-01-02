# Custom Parsing Example

This example demonstrates how Pallas parses Go code and extracts documentation. Understanding this helps you write better comments.

## How Pallas Works

Pallas uses the [parser.EntityInfo] struct to represent each documented entity in your code. The parser extracts:

- **Functions** and their signatures
- **Structs** with field documentation
- **Interfaces** with method contracts
- **Type aliases** and constants

## The EntityInfo Structure

When Pallas parses your code, it creates an [parser.EntityInfo] for each entity:

```go
type EntityInfo struct {
    Name        string       // Entity name (e.g., "Config")
    Type        string       // Entity type ("struct", "func", etc.)
    Description string       // Extracted comment as HTML
    Package     string       // Package name
    File        string       // Source file path
    Methods     []MethodInfo // For structs/interfaces
    Fields      []FieldInfo  // For structs
}
```

## Parsing a Package

The [parser.ParseEntitiesInPackage] function scans a directory:

```go
entities, imports, err := parser.ParseEntitiesInPackage(
    projectRoot,  // Your project root
    pkgDir,       // Package directory
    relativePath, // Path relative to root
)
```

## Generator Integration

Once parsed, entities are passed to [generator.GenerateHTML]:

```go
err := generator.GenerateHTML(
    "My Project",   // Title
    "./dist",       // Output directory  
    entities,       // Parsed entities
    imports,        // Import information
    readmeContent,  // README as HTML
    "MP",           // Initials for logo
    "github.com/me/myproject", // Module path
    docPages,       // Documentation pages
    docSections,    // Doc sections
)
```

## Type Cross-Linking

Notice how [parser.EntityInfo] and [generator.GenerateHTML] are automatically linked in this document. This works because:

1. Pallas knows these types exist in the project
2. The `[Package.Type]` syntax triggers linking
3. Links point to the correct API reference page
