# Cross-Linking in Documentation

Pallas automatically creates links between your code entities and supports manual linking from Markdown documentation to your API.

## Automatic Type Linking

In API documentation pages, Pallas automatically links types in:

- Function signatures (parameters and return types)
- Struct field types
- Interface method signatures
- Type aliases

For example, if you have:

```go
func NewServer(config *Config) *Server
```

Both `Config` and `Server` will automatically link to their respective documentation.

## Manual Links in Markdown

When writing documentation in the `docs/` folder, use the bracket syntax to link to code types:

```markdown
Configure your application using [parser.Config].
```

**Syntax**: `[PackageName.TypeName]`

This creates a clickable link that takes users directly to the type's API documentation.

## Import Badges

On API pages, imports are tagged with colored badges:

| Badge | Color | Meaning |
|-------|-------|---------|
| **STD** | Blue | Go standard library (`fmt`, `net/http`, etc.) |
| **PKG** | Purple | External third-party package |
| **INT** | Green | Internal project package |

Clicking on external packages opens their documentation on [pkg.go.dev](https://pkg.go.dev).

## Best Practices

1. **Use qualified names**: Write `[parser.EntityInfo]` not just `[EntityInfo]` to avoid ambiguity
2. **Link on first mention**: Link types the first time they appear in a document
3. **Don't over-link**: Repeated links to the same type are noise
