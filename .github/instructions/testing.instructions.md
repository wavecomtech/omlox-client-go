---
applyTo: "**/*_test.go"
---

# Testing Conventions

- Use **table-driven tests** with descriptive test case names.
- Use the generic helpers `JSONMarshalOK[T]()` and `JSONUnmarshalOK[T]()` from `utils_test.go` for JSON serialization tests.
- Use `github.com/google/go-cmp/cmp` for deep equality assertions.
- Use `github.com/nsf/jsondiff` for debugging JSON comparison failures.
- Test both marshaling and unmarshaling for every model struct.
- Test edge cases: zero values, nil pointers, empty slices, infinite Duration.
- Run with race detector: `go test -v -cover -race ./...`
