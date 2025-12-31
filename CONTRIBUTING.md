# Contributing to Toutā DataMapper

Thank you for your interest in contributing to DataMapper! This document provides guidelines and information for contributors.

## Code of Conduct

This project adheres to a code of conduct. By participating, you are expected to uphold this code. Please report unacceptable behavior to the project maintainers.

## How to Contribute

### Reporting Issues

- Use the GitHub issue tracker
- Check if the issue already exists
- Provide detailed information:
  - Go version
  - Operating system
  - Steps to reproduce
  - Expected vs actual behavior

### Pull Requests

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Make your changes
4. Write or update tests
5. Ensure tests pass (`go test ./...`)
6. Ensure code is formatted (`go fmt ./...`)
7. Run linter (`golangci-lint run`)
8. Commit with conventional commit format
9. Push to your fork
10. Open a Pull Request

### Commit Convention

We use [Conventional Commits](https://www.conventionalcommits.org/):

```
<type>(<scope>): <subject>

<body>

<footer>
```

**Types:**
- `feat`: New feature
- `fix`: Bug fix
- `perf`: Performance improvement
- `refactor`: Code restructuring
- `test`: Test additions or modifications
- `docs`: Documentation changes
- `chore`: Build, CI, or tooling changes

**Examples:**
```
feat(adapter): add Redis adapter support
fix(config): handle nil pointer in credential loading
perf(mapper): optimize bulk insert operations
docs(readme): add usage examples
test(adapter): add filesystem adapter tests
```

## Development Setup

### Prerequisites

- Go 1.22 or higher
- Git
- golangci-lint (for linting)

### Getting Started

```bash
# Clone your fork
git clone https://github.com/YOUR_USERNAME/toutago-datamapper
cd toutago-datamapper

# Install dependencies
go mod download

# Run tests
go test ./...

# Run tests with coverage
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

## Testing Requirements

- All new code must include tests
- Maintain minimum 80% code coverage
- Tests must pass with race detector: `go test -race ./...`
- Integration tests for new adapters

## Code Quality Standards

- Follow Go best practices and idioms
- Use meaningful variable and function names
- Keep functions focused and small
- Document exported types and functions
- Pass golangci-lint without errors

## Documentation

- Update README.md for user-facing changes
- Update doc.go for API changes
- Add examples for new features
- Keep CHANGELOG.md current

## Questions?

Feel free to open an issue for questions or discussion.

## License

By contributing, you agree that your contributions will be licensed under the MIT License.
