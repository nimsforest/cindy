# Contributing to Cindy

Cindy is an open protocol. Contributions are welcome.

## How to contribute

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Run tests: `cd go && go test ./...`
5. Submit a pull request

## Guidelines

- The Go package must remain dependency-free (stdlib only)
- All label transitions must be documented in SPEC.md
- Schema safety rules are non-negotiable: schemas can only be extended, never broken
- Example manifests should cover realistic scenarios
- JSON Schema must stay in sync with the Go struct definitions

## Protocol changes

Changes to the label state machine or schema safety rules are significant. Please open an issue first to discuss before submitting a PR.

## License

By contributing, you agree that your contributions will be licensed under the MIT License.
