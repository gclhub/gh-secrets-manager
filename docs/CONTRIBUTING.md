# Contributing to gh-secrets-manager

This guide explains how to set up your development environment to build and test gh-secrets-manager.

## Development Workflow

1. **Clone and Setup**:
```bash
git clone https://github.com/gclhub/gh-secrets-manager.git
cd gh-secrets-manager
go mod tidy
```

2. **Development Process**:
   - Make changes to the Go source code
   - Test changes using `go run` or rebuild the binary
   - Run tests to ensure no regressions
   - Install locally for end-to-end testing
   - Test with real GitHub API using your development environment

3. **Build the CLI**:
```bash
# Build for current platform
go build -o bin/gh-secrets-manager ./cmd/gh-secrets-manager

# Or build for specific platforms
GOOS=linux GOARCH=amd64 go build -o bin/gh-secrets-manager-linux ./cmd/gh-secrets-manager
GOOS=darwin GOARCH=amd64 go build -o bin/gh-secrets-manager-darwin ./cmd/gh-secrets-manager
GOOS=windows GOARCH=amd64 go build -o bin/gh-secrets-manager.exe ./cmd/gh-secrets-manager
```

## Testing

### Running Tests
```bash
# Run all tests
go test ./...

# Run tests with verbose output
go test ./... -v

# Run tests for specific package
go test ./pkg/api -v

# Run tests with coverage
go test ./... -cover
```

### Local Testing

You can test your changes using any of these methods:

1. **Direct Go Execution**:
```bash
go run ./cmd/gh-secrets-manager secrets list --org myorg
```

2. **Using Built Binary**:
```bash
./bin/gh-secrets-manager secrets list --org myorg
```

3. **Install as Local Extension**:
```bash
gh extension install .
```

## Managing Dependencies

Update project dependencies:
```bash
# Download and verify dependencies
go mod tidy

# Update all dependencies to latest versions
go get -u ./...
go mod tidy
```

## Debugging

Enable verbose logging for debugging:
```bash
# Using go run
go run ./cmd/gh-secrets-manager secrets list --org myorg --verbose

# Using built binary
./bin/gh-secrets-manager secrets list --org myorg --verbose
```

## Working with Auth Server

When developing with the auth server component:

1. **Start the auth server** in development mode:
```bash
cd auth-server
go run cmd/server/main.go --port 8080 --private-key-path /path/to/key.pem --team myteam --verbose
```

2. **Configure the CLI** to use your local auth server:
```bash
go run ./cmd/gh-secrets-manager config set auth-server http://localhost:8080
go run ./cmd/gh-secrets-manager config set app-id YOUR_APP_ID
go run ./cmd/gh-secrets-manager config set installation-id YOUR_INSTALLATION_ID
```

3. **Test the integration**:
```bash
go run ./cmd/gh-secrets-manager secrets list --org myorg --verbose
```

## Prerequisites

- Go 1.24.2 or later
- GitHub CLI (`gh`) installed and authenticated
- Git

## Pull Request Process

1. Ensure your code follows the existing style
2. Update the documentation with details of changes if needed
3. Add or update tests as needed
4. Create a Pull Request with a clear description of the changes
