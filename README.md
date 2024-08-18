# Pallas

Pallas is a simple documentation generator for Go projects. It parses Go source code, extracts information about functions, types, and interfaces, and generates clean and modern HTML documentation using Tailwind CSS.

## Features

- Extracts and documents functions, types, and interfaces
- Generates a fully responsive HTML documentation with dark mode support*
- Automatically organizes and indexes packages based on their structure
- Provides a search feature to quickly find entities and packages
- Allows picking a custom title and export directory

## Installation

### Prerequisites

- [Go 1.20+](https://go.dev/dl/) installed on your machine.

### Building Pallas

To build Pallas, clone the repository and run:

```bash
git clone https://github.com/vanilla-os/pallas.git
cd pallas
go build -o pallas cmd/pallas/main.go
```

This will produce an executable named `pallas` in the project root.

## Usage

To generate documentation for your Go project, navigate to your project's root directory and run:

```bash
./pallas [options] [projectPath]
```

### Options

- `--dest <path>`: Specify a custom destination directory for the generated documentation; the default is `./dist` in the current working directory
- `--title <name>`: Specify a custom title for the documentation, if not provided, the name of the root directory of the project will be used as the title

### Examples

#### Basic Usage

To generate documentation for the current directory with default settings:

```bash
./pallas
```

This will generate documentation in the `./dist` directory with the title based on the current directory name.

#### Custom Destination Directory

To generate documentation and save it to a specific directory:

```bash
./pallas --dest /path/to/output
```

This will generate documentation in `/path/to/output`.

#### Custom Documentation Title

To generate documentation with a custom title:

```bash
./pallas --title "My Project"
```

This will set the documentation title to "My Project".

### Combining Flags

Flags can be combined to customize both the output directory and the title:

```bash
./pallas --dest /path/to/output --title "My Project"
./pallas /my/project --dest /path/to/output --title "My Project"
```

This will generate documentation for `/my/project` in `/path/to/output` with the title "My Project".

## How It Works

1. **Parsing**: Pallas scans the provided Go project's root directory recursively, looking for Go packages. It then uses the Go built-in `go/parser`, `go/token`, and `go/ast` packages to parse and analyze the source code of each package, extracting information about functions, types, and interfaces

2. **Generating HTML**: Pallas then generates a series of HTML files, one for each package, organized into groups based on their directory structure. An `index.html` file is also generated, providing an overview and easy navigation between the different packages

3. **Customization**: The generated documentation is styled using Tailwind CSS and Highlight.js for code syntax highlighting

## License

This project is licensed under the GPLv3 License. See the [LICENSE](LICENSE) file for details.

## Acknowledgments

- Go's native `go/parser` and `go/ast` packages for parsing and analyzing Go source code.
- [Tailwind CSS](https://tailwindcss.com/)
- [Highlight.js](https://highlightjs.org/)

## Why the name "Pallas"?

Pallas was the Titan god of warcraft and wisdom in Greek mythology. The name was chosen to reflect both the power of the tool and the wisdom it provides through documentation.