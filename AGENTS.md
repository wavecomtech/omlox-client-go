# Agents

## General Guidelines

This is `github.com/wavecomtech/omlox-client-go` — a Go client library and CLI for the omlox Hub real-time locating standard. It targets Go 1.21+ and uses generics, `log/slog`, easyjson code generation, and the Cobra CLI framework.

### Quick Reference

- **Module**: `github.com/wavecomtech/omlox-client-go`
- **Package**: `omlox` (root package is the public library)
- **Build**: `make build`
- **Test**: `make test` or `go test -v -cover -race ./...`
- **Generate**: `go generate ./...` (requires `easyjson` and `copywrite` installed)
- **Lint**: `make lint` (golangci-lint)

### Project Structure

```
omlox-client-go/
├── *.go                   # Public library: models, client, WebSocket, serialization
├── *_requests.go          # REST API methods per entity (e.g., trackable_requests.go)
├── *_test.go              # Table-driven tests per module
├── omlox_easyjson.go      # AUTO-GENERATED — never edit manually
├── gen.go                 # go:generate directives
├── cmd/omlox-cli/         # CLI commands (Cobra)
├── internal/cli/          # CLI internals: settings, output formatters, resource loaders
├── docs/                  # Generated CLI documentation
└── .github/workflows/     # CI and release pipelines
```

### Critical Rules

1. **Never edit `omlox_easyjson.go`** — it is auto-generated. Run `go generate ./...` after modifying structs.
2. **All public API methods must accept `context.Context` as the first parameter.**
3. **JSON field names use `snake_case`** — enforced by easyjson `-snake_case` flag and struct tags.
4. **Copyright header is mandatory** on all `.go` files: `// Copyright (c) Omlox Client Go Contributors` + `// SPDX-License-Identifier: MIT`.
5. **New model structs must be marked with `//easyjson:json`** comment for code generation.
6. **Follow the API resource pattern**: `<entity>.go` (model) + `<entity>_requests.go` (API methods) + `<entity>_test.go` (tests).
7. **CLI commands** follow `<verb>_<resource>.go` naming in `cmd/omlox-cli/`.

### Domain Context

This library wraps the **omlox Hub** API — a locating middleware standard. The core entities are:
- **Zone**: Physical area with coordinate transformation (local → global)
- **Location Provider**: A positioning device (UWB tag, GPS, Wi-Fi, RFID, BLE)
- **Trackable**: A real-world asset being tracked (can have multiple providers)
- **Fence**: A geofence area triggering entry/exit events
- **Location**: A position update with GeoJSON Point, timestamps, accuracy

The Hub provides REST (`/v2`) and WebSocket (`/ws/socket`) APIs. WebSocket uses a topic-based pub/sub model with topics like `location_updates`, `fence_events`, `collision_events`, `trackable_motions`.

### Patterns to Follow

| Pattern | Description |
|---------|-------------|
| **Functional Options** | `WithXxx(...)` functions returning `ClientOption` for client configuration |
| **Generics** | `sendRequestParseResponse[T]`, `ReceiveAs[T]`, `Loader[T]` for type-safe operations |
| **Table-driven tests** | Use `JSONMarshalOK[T]` and `JSONUnmarshalOK[T]` helpers from `utils_test.go` |
| **Channel streaming** | WebSocket subscriptions deliver via buffered channels (256 capacity) |
| **Structured logging** | Implement `slog.LogValuer` on domain types |
| **Thread safety** | `sync.RWMutex` for concurrent WebSocket access |

### Dependencies

| Package | Purpose |
|---------|---------|
| `nhooyr.io/websocket` | WebSocket client |
| `github.com/hashicorp/go-cleanhttp` | HTTP client with safe defaults |
| `github.com/mailru/easyjson` | Fast JSON via code generation |
| `github.com/tidwall/geojson` | GeoJSON geometry types |
| `github.com/google/uuid` | UUID generation/parsing |
| `golang.org/x/time` | Rate limiting |
| `github.com/spf13/cobra` | CLI framework |
| `github.com/google/go-cmp` | Test deep equality |
| `github.com/nsf/jsondiff` | JSON diff for test debugging |

### Known Issues

- **DeepHub subscription_id bug**: The DeepHub may omit `subscription_id` in subsequent WebSocket messages. The client routes by topic as fallback.
- **omlox OpenAPI spec**: The official spec has issues; this library is hand-coded rather than auto-generated.
- **Duration type**: Supports infinite semantics (`Inf = -1`). Don't assume standard `time.Duration` behavior.
- **Optional fields**: Translating omlox optional fields with defaults to Go is an ongoing design challenge (pointer types vs zero values).

### What's Not Yet Implemented

Models: Collision, CollisionEvent, Fence, FenceEvent, LineString, Zone, Proximity, TrackableMotion. REST: Zones, Fences, Anchors CRUD. CLI: zone/fence commands. CI: golangci-lint workflow.
