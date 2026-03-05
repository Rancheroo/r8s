# Testing r8s

Guide to running and writing tests for r8s.

---

## Quick Start

```bash
# Run all tests
make test

# Run tests with coverage report
make coverage

# Run full CI pipeline locally (lint + test + coverage)
make ci
```

---

## Test Commands

| Command | What It Does |
|---------|-------------|
| `make test` | Run all unit tests (`go test -v ./...`) |
| `make coverage` | Generate coverage report (`coverage.out`) |
| `make lint` | Run golangci-lint |
| `make vet` | Run `go vet` |
| `make dev` | Run all dev checks (tidy, fmt, vet, lint) |
| `make ci` | Full CI pipeline (lint + test + coverage) |

---

## Writing Tests

- Use **table-driven tests** (see `CONTRIBUTING.md` for examples)
- Test files: `*_test.go` alongside the code they test
- Naming: `TestFunctionName` or `TestType_Method`
- Run a specific package: `go test -v ./internal/ai/...`

### Test Coverage

- Target: 45%+ overall (currently ~70%)
- Critical paths (AI patterns, bundle parsing): aim for 90%+
- Check coverage: `make coverage`

---

## Manual Testing

For features that touch CLI output or bundle parsing, test with real bundles:

```bash
# Build
make build

# Test core commands against a bundle
./bin/r8s validate ./path/to/bundle/
./bin/r8s analyze ./path/to/bundle/
./bin/r8s get pods ./path/to/bundle/
./bin/r8s logs ./path/to/bundle/ <pod-name>
./bin/r8s ask ./path/to/bundle/ "why is nginx crashing?"

# Verify exit codes
./bin/r8s analyze ./path/to/bundle/; echo $?
# 0 = healthy, 1 = issues found, 2 = error
```

### Before Any Release

1. `make ci` passes
2. Test with at least 2 real bundles (one healthy, one with known issues)
3. Verify all core commands produce expected output
4. Check `./bin/r8s version` shows correct version

---

## Test Plans

Version-specific manual test plans live in `docs/testing/`.
