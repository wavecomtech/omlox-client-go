---
applyTo: "{*.go,!*_test.go,!omlox_easyjson.go,!cmd/**,!internal/**}"
---

# Library Code Conventions

- All files must begin with the copyright header:
  ```
  // Copyright (c) Omlox Client Go Contributors
  // SPDX-License-Identifier: MIT
  ```
- Package name is `omlox`.
- Model structs representing omlox entities must be marked with `//easyjson:json` for code generation.
- JSON struct tags use `snake_case` naming (matching the omlox API specification).
- Use pointer types for optional fields that have meaningful zero values or default values in the omlox spec.
- All public API methods accept `context.Context` as the first parameter.
- Implement `slog.LogValuer` on domain types for structured logging.
- Enum types are defined as named `string` or `int` types with constants and custom JSON marshaling.
- GeoJSON geometry types (Point, Polygon) wrap `github.com/tidwall/geojson` with custom JSON marshaling.
- `omlox_easyjson.go` is auto-generated — never edit it manually. Run `go generate ./...` after changes.
