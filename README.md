# bom-dagger

[![Build and Release](https://github.com/nprimmer/bom-dagger/actions/workflows/build.yml/badge.svg)](https://github.com/nprimmer/bom-dagger/actions/workflows/build.yml)
[![Test and Lint](https://github.com/nprimmer/bom-dagger/actions/workflows/test.yml/badge.svg)](https://github.com/nprimmer/bom-dagger/actions/workflows/test.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/nprimmer/bom-dagger)](https://goreportcard.com/report/github.com/nprimmer/bom-dagger)
[![License](https://img.shields.io/github/license/nprimmer/bom-dagger)](LICENSE)
[![Release](https://img.shields.io/github/v/release/nprimmer/bom-dagger)](https://github.com/nprimmer/bom-dagger/releases/latest)

Creates a DAG for deployment order from a CycloneDX SBOM. This tool analyzes dependencies in a Software Bill of Materials and generates an optimal deployment sequence.

## Features

- Parse CycloneDX JSON SBOMs
- Build a Directed Acyclic Graph (DAG) from component dependencies
- Generate deployment order using topological sort
- Show parallel deployment groups (components that can be deployed simultaneously)
- Generate reverse order for teardown sequences
- Export to DOT format for visualization
- Detect circular dependencies

## Installation

### Download Pre-built Binaries

Download the latest release from the [GitHub Releases page](https://github.com/nprimmer/bom-dagger/releases/latest).

#### macOS/Linux
```bash
# Download the latest release (replace OS and ARCH as needed)
curl -LO https://github.com/nprimmer/bom-dagger/releases/latest/download/bom-dagger-$(uname -s | tr '[:upper:]' '[:lower:]')-$(uname -m).tar.gz

# Extract the binary
tar -xzf bom-dagger-*.tar.gz

# Make it executable
chmod +x bom-dagger

# Move to PATH (optional)
sudo mv bom-dagger /usr/local/bin/
```

#### Windows
Download the Windows executable from the [releases page](https://github.com/nprimmer/bom-dagger/releases/latest) and add it to your PATH.

### Build from Source

```bash
# Clone the repository
git clone https://github.com/nprimmer/bom-dagger.git
cd bom-dagger

# Build using Make
make build

# Or build directly with Go
go build -o bom-dagger cmd/bom-dagger/main.go

# Install to /usr/local/bin (Unix-like systems)
make install
```

### Using Go Install

```bash
go install github.com/nprimmer/bom-dagger/cmd/bom-dagger@latest
```

## Usage

```bash
bom-dagger -i <sbom-file> [options]
```

### Options

- `-i, --input <file>` - Path to CycloneDX SBOM file (JSON)
- `-o, --output <mode>` - Output mode: order (default), groups, dot
- `-r, --reverse` - Show reverse order (teardown sequence)
- `-g, --groups` - Show deployment groups (parallel deployment)
- `-s, --stats` - Show graph statistics
- `-h, --help` - Show help message

### Examples

Show deployment order:
```bash
./bom-dagger -i example-sbom.json
```

Show teardown order:
```bash
./bom-dagger -i example-sbom.json -r
```

Show parallel deployment groups:
```bash
./bom-dagger -i example-sbom.json -g
```

Generate DOT format for visualization:
```bash
./bom-dagger -i example-sbom.json -o dot > graph.dot
dot -Tpng graph.dot -o graph.png
```

## Testing

Run the unit tests:
```bash
go test ./...
```

## Example Output

Given an SBOM with web application components, the tool generates:

```
=== Deployment Order ===
Deploy components in this sequence:

Step 1:
  - PostgreSQL Database (ref: postgres-db)
  - Redis Cache (ref: redis-cache)
  - RabbitMQ (ref: rabbitmq)

Step 2:
  - Authentication Service (ref: auth-service)
  - Order Service (ref: order-service)

Step 3:
  - API Gateway (ref: api-gateway)
```

Components in the same step can be deployed in parallel as they have no interdependencies.

## Development

### Prerequisites
- Go 1.24 or later
- Make (optional, for using Makefile commands)

### Building
```bash
# Build for current platform
make build

# Build for all platforms
make build-all

# Run tests
make test

# Generate test coverage
make coverage

# Clean build artifacts
make clean
```

### CI/CD

This project uses GitHub Actions for continuous integration and deployment:

- **Push to main**: Builds binaries for multiple platforms and uploads as artifacts
- **Pull requests**: Runs tests and linting
- **Tags**: Creates GitHub releases with pre-built binaries

To create a new release:
```bash
# Tag the release
git tag -a v1.0.0 -m "Release version 1.0.0"

# Push the tag
git push origin v1.0.0
```

The GitHub Action will automatically build and create a release with binaries for:
- Linux (amd64, arm64)
- macOS (arm64)

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This project is open source and available under the [MIT License](LICENSE).
