# Configuration

Pallas is designed to work out of the box with zero configuration, but offers flexibility when needed.

## The `docs/` Directory

Pallas automatically scans the `docs/` folder in your project root for Markdown files. These become navigable documentation pages in the generated site.

### Subdirectory Support

You can organize documentation in subdirectories:

```
docs/
├── getting_started.md      → "Getting Started"
├── configuration.md        → "Configuration"
└── guides/
    ├── comments.md         → "Guides / Comments"
    └── api_design.md       → "Guides / Api Design"
```

The directory structure is reflected in the page titles.

## README as Overview

The `README.md` file in your project root becomes the "Overview" page. This is typically the first thing users see.

## Title Auto-Detection

If you don't specify a `--title`, Pallas will:

1. Use the project directory name
2. Convert dashes/underscores to spaces
3. Apply title case

For example, `my-awesome-sdk` becomes "My Awesome Sdk".

## Module Path Detection

Pallas reads your `go.mod` file to determine the module path. This enables:

- Correct cross-linking between packages
- Detection of internal vs external imports
- Links to pkg.go.dev for external packages
