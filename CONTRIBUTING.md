# Contributing to go-mongo-platform

Thank you for your interest in contributing to `go-mongo-platform`!

## Code Style Guidelines

- Follow standard Go formatting (`go fmt`)
- Use `golangci-lint` for linting (configuration in `.golangci.yml` if present)
- Write clear, concise code comments
- Use meaningful variable and function names

## Pull Request Process

1. **Fork the repository** and create a feature branch
2. **Make your changes** following the code style guidelines
3. **Add tests** for new functionality
4. **Update documentation** if needed
5. **Ensure all tests pass** (`go test ./...`)
6. **Submit a pull request** with a clear description

### PR Description Template

- **What:** Brief description of changes
- **Why:** Motivation for the change
- **How:** Implementation approach (if non-trivial)
- **Testing:** How the change was tested

## Testing Requirements

- All new code must include tests
- Tests should be in `*_test.go` files
- Run `go test ./...` before submitting
- Ensure test coverage is maintained or improved

## Commit Message Conventions

Use clear, descriptive commit messages:

```
Short summary (50 chars or less)

More detailed explanation if needed. Wrap at 72 characters.
Explain what and why, not how.
```

## Scope Considerations

Please note the project's scope (see [NON_GOALS_AND_ANTI_AUDIENCE.md](.github/internal-docs/NON_GOALS_AND_ANTI_AUDIENCE.md)):

- **In scope:** Bug fixes, documentation improvements, stability enhancements
- **Out of scope:** Performance optimizations, new API surface, distributed features, TTL/eviction logic

## Questions?

- Open a GitHub Issue for bug reports
- Use GitHub Discussions for questions
- See [SUPPORT.md](SUPPORT.md) for commercial support options

