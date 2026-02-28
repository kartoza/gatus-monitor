# Contributing to Gatus Monitor

Thank you for your interest in contributing to Gatus Monitor! This document provides guidelines and instructions for contributing.

## Code of Conduct

This project adheres to a [Code of Conduct](CODE_OF_CONDUCT.md). By participating, you are expected to uphold this code.

## How to Contribute

### Reporting Bugs

Before creating a bug report, please check existing issues to avoid duplicates.

When creating a bug report, include:
- A clear and descriptive title
- Steps to reproduce the issue
- Expected vs. actual behavior
- Screenshots (if applicable)
- Environment details (OS, Go version, etc.)

### Suggesting Enhancements

Enhancement suggestions are tracked as GitHub issues. When creating an enhancement suggestion, include:
- A clear and descriptive title
- Detailed description of the proposed feature
- Rationale for why this enhancement would be useful
- Possible implementation approach (if applicable)

### Pull Requests

1. **Fork the repository** and create your branch from `main`
2. **Make your changes** following our coding standards
3. **Add tests** for any new functionality
4. **Ensure all tests pass**: `go test ./...`
5. **Run linters**: `golangci-lint run`
6. **Update documentation** as needed
7. **Commit your changes** with clear, descriptive messages
8. **Push to your fork** and submit a pull request

## Development Setup

### Prerequisites

- Go 1.21 or later
- Nix with flakes enabled (recommended)

### Setup with Nix

```bash
git clone https://github.com/kartoza/gatus-monitor.git
cd gatus-monitor
nix develop
```

This will provide all necessary dependencies automatically.

### Setup without Nix

```bash
git clone https://github.com/kartoza/gatus-monitor.git
cd gatus-monitor
go mod download
```

You'll also need to install system dependencies:
- **Linux**: `libgl1-mesa-dev`, `xorg-dev`
- **macOS**: Xcode Command Line Tools
- **Windows**: MinGW-w64

### Pre-commit Hooks

Install pre-commit hooks to ensure code quality:

```bash
pre-commit install
```

This will automatically run formatters, linters, and tests before each commit.

## Coding Standards

### Go Code

- Follow standard Go conventions and idioms
- Use `gofmt` for formatting (pre-commit hook does this)
- Write clear, descriptive variable and function names
- Add comments for exported functions, types, and packages
- Keep functions focused and concise
- Prefer composition over inheritance
- Handle errors explicitly; don't ignore them

### Code Organization

- Place new packages in `internal/` directory
- Follow the existing package structure
- Keep packages focused on a single responsibility
- Avoid circular dependencies

### Testing

- Write tests for all new functionality
- Aim for >80% code coverage
- Use table-driven tests where appropriate
- Mock external dependencies
- Test both success and error cases

Example test structure:
```go
func TestFunction(t *testing.T) {
    tests := []struct {
        name     string
        input    InputType
        expected ExpectedType
        wantErr  bool
    }{
        // Test cases
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Test logic
        })
    }
}
```

### Documentation

- Update README.md for user-facing changes
- Update SPECIFICATION.md for architectural changes
- Update PACKAGES.md when adding new packages
- Add/update MkDocs documentation as appropriate
- Include godoc comments for all exported items

### Commit Messages

Follow conventional commit format:

```
<type>(<scope>): <subject>

<body>

<footer>
```

Types:
- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation changes
- `style`: Code style changes (formatting, etc.)
- `refactor`: Code refactoring
- `test`: Adding or updating tests
- `chore`: Maintenance tasks

Example:
```
feat(monitor): add support for custom error thresholds

Allow users to configure custom thresholds for orange and red
status indicators instead of using hardcoded values.

Closes #123
```

## Pull Request Process

1. **Update VERSION file** if applicable (bug fix = patch, new feature = minor, breaking change = major)
2. **Update CHANGELOG** with your changes
3. **Ensure CI passes** - all tests, lints, and builds must succeed
4. **Request review** from maintainers
5. **Address feedback** promptly and professionally
6. **Squash commits** if requested before merging

## Review Process

- Maintainers will review PRs within 1-2 weeks
- Reviewers may request changes or improvements
- Once approved, a maintainer will merge your PR
- Releases are made periodically from the `main` branch

## Release Process

Releases are handled by maintainers:

1. Update VERSION file
2. Update CHANGELOG
3. Create and push tag: `git tag -a v0.2.0 -m "Release v0.2.0"`
4. GitHub Actions will build and publish release artifacts
5. Documentation is automatically deployed

## Getting Help

- **Questions**: Open a [GitHub Discussion](https://github.com/kartoza/gatus-monitor/discussions)
- **Chat**: Join our community chat (link TBD)
- **Email**: Contact maintainers at dev@kartoza.com

## Recognition

Contributors will be recognized in:
- README.md credits section
- Release notes
- Documentation contributors page

## License

By contributing, you agree that your contributions will be licensed under the MIT License.

---

Thank you for contributing to Gatus Monitor!

Made with 💗 by [Kartoza](https://kartoza.com)
