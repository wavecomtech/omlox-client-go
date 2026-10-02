# Omlox Hub Go Client Library — Copilot Instructions

## Project Overview

This is `github.com/wavecomtech/omlox-client-go`, a Go client library and CLI tool for the [omlox Hub](https://omlox.com/) — an open standard middleware for real-time indoor localization systems. The library provides a unified connector so Go applications can interact with any omlox-compliant Hub (tested against Flowcate's DeepHub®).

The omlox Hub is a locating middleware that enables interoperability across locating technologies (UWB, RFID, 5G, BLE, Wi-Fi, GPS). It transforms local coordinates into standardized global geographical coordinates (WGS84/GeoJSON) and exposes them via REST and WebSocket APIs.

## Architecture

```
┌─────────────────────────────────────────┐
│        CLI Layer (cmd/omlox-cli)        │  Cobra commands, JSON/Table output
│        Uses the public omlox package    │
└─────────────┬───────────────────────────┘
              │
┌─────────────▼───────────────────────────┐
│       Public API (omlox package)        │  Client, HTTP APIs, WebSocket, Models
│  - HTTP: New() → client.Trackables.*    │
│  - WS:   Connect() → Subscribe()       │
└─────────────┬───────────────────────────┘
              │
┌─────────────▼───────────────────────────┐
│      Internal (internal/cli/*)          │  Settings, output formatting, resource loading
└─────────────────────────────────────────┘
```

### Key Components

| File | Purpose |
|------|---------|
| `client.go` | Main `Client` struct, HTTP request infrastructure with generics |
| `client_configuration.go` | `ClientOption` functional options, default config |
| `client_websockets.go` | WebSocket connection, reconnection, subscription management |
| `client_subscription.go` | `Subscription` type, `ReceiveAs[T]` generic channel receiver |
| `websockets.go` | WebSocket protocol types: `WrapperObject`, `Event`, `Topic`, `ErrCode` |
| `provider.go` / `provider_requests.go` | `LocationProvider` model + `ProvidersAPI` CRUD methods |
| `trackable.go` / `trackable_requests.go` | `Trackable` model + `TrackablesAPI` CRUD methods |
| `location.go` | `Location` model with GeoJSON position, timestamps, orientation |
| `point.go` | GeoJSON `Point` wrapper around `tidwall/geojson` |
| `polygon.go` | GeoJSON `Polygon` wrapper around `tidwall/geojson` |
| `rules.go` | `LocatingRule` for priority-based location provider selection |
| `duration.go` | Custom `Duration` type supporting infinite semantics |
| `error.go` | `Error` type for HTTP error responses |
| `omlox_easyjson.go` | Auto-generated easyjson marshalers (DO NOT EDIT) |
| `gen.go` | `go generate` directives for easyjson, docs, copyright headers |

## Omlox Hub Domain Knowledge

### Core Entities

1. **Zone** — Represents a physical area covered by a locating system. Defines local-to-global coordinate transformation. Contains ground control points for georeferencing.

2. **Location Provider** — A positioning device that generates location data (e.g., UWB tag, Wi-Fi AP, GPS receiver, RFID reader). Identified by MAC address or similar unique ID. Has a `provider_type` (uwb, gps, wifi, rfid, ibeacon, virtual).

3. **Trackable** — A real-world asset being tracked (forklift, worker, drone). Can have multiple Location Providers attached. Has a UUID, optional geometry (Polygon), radius, and locating rules for provider priority. This is the main entity applications interact with.

4. **Fence** — A geofence defining an area of interest. Trackables entering/exiting fences trigger `fence_events`. Defined in global coordinates (GeoJSON Polygon).

5. **Location** — A position update containing GeoJSON Point coordinates, provider info, timestamps, accuracy, floor, heading, speed. Coordinates flow: Provider → Hub (transforms local→global) → Application.

### API Structure (REST — base path `/v2`)

| Resource | Key Endpoints |
|----------|---------------|
| Zones | `GET/POST/DELETE /zones`, `GET/PUT/DELETE /zones/:id`, `PUT /zones/:id/transform` |
| Trackables | `GET /trackables/summary`, `GET/POST/DELETE /trackables`, `GET/PUT/DELETE /trackables/:id`, `GET /trackables/:id/location` |
| Providers | `GET /providers/summary`, `GET/POST/DELETE /providers`, `GET/PUT/DELETE /providers/:id`, `PUT /providers/:id/location` |
| Fences | `GET/POST/DELETE /fences`, `GET/PUT/DELETE /fences/:id` |

### WebSocket API (path: `/ws/socket`)

Topic-based pub/sub protocol using JSON wrapper objects:

```json
{
  "event": "subscribe|subscribed|unsubscribe|unsubscribed|message|error",
  "topic": "location_updates|fence_events|collision_events|trackable_motions|...",
  "subscription_id": 123,
  "payload": [...],
  "params": {"crs": "EPSG:4326", "provider_id": "..."}
}
```

**Topics:**
- `location_updates` — Real-time position updates (Location objects)
- `location_updates:geojson` — Same but as GeoJSON FeatureCollections
- `fence_events` — Geofence entry/exit events (FenceEvent objects)
- `fence_events:geojson` — Same but as GeoJSON
- `collision_events` — Trackable collision events (CollisionEvent objects)
- `trackable_motions` — Trackable movement events (TrackableMotion objects)
- `provider_changes`, `trackable_changes`, `fence_changes`, `zone_changes` — CRUD change events

**Error Codes:** 10000 (unknown event), 10001 (unknown topic), 10002 (subscription failed), 10003 (unsubscribe failed), 10004 (not authorized), 10005 (invalid payload), 10006 (not authenticated), 10007 (invalid license)

### Coordinate Reference Systems

The Hub uses WGS84 (EPSG:4326) as the default CRS. Local coordinates from zones are transformed to global coordinates. Clients can request projections via the `crs` parameter (e.g., `EPSG:32632` for UTM Zone 32N, or `local` with a `zone_id`).

## Coding Conventions

### Go Style

- **Go version**: 1.21+ (uses generics, `log/slog`)
- **Module**: `github.com/wavecomtech/omlox-client-go`
- **Package name**: `omlox` (root package is the public library API)
- **Copyright header**: `// Copyright (c) Omlox Client Go Contributors` + `// SPDX-License-Identifier: MIT`
- **JSON field naming**: `snake_case` (enforced via easyjson `-snake_case` flag and struct tags)

### Patterns

- **Functional Options**: Client configuration uses `ClientOption` function type. Add new options as `WithXxx(...)` functions returning `ClientOption`.
- **Generics**: Type-parameterized functions for HTTP request/response handling (`sendRequestParseResponse[T]`, `ReceiveAs[T]`) and resource loading (`Loader[T]`).
- **Context Propagation**: All public API methods accept `context.Context` as first parameter.
- **Channel-Based Streaming**: WebSocket subscriptions deliver messages via buffered channels (size 256).
- **Structured Logging**: Use `log/slog` with `slog.LogValuer` interface on domain types.
- **Thread Safety**: `sync.RWMutex` on Client for concurrent WebSocket access, `sync.WaitGroup` for goroutine lifecycle.

### Serialization

- **easyjson** for high-performance JSON marshaling. Structs marked with `//easyjson:json` comment get generated marshalers in `omlox_easyjson.go`.
- After adding/modifying structs with JSON serialization, run `go generate ./...` to regenerate.
- **Never manually edit** `omlox_easyjson.go`.

### API Resource Pattern

When adding a new omlox entity (e.g., Zones, Fences):

1. Create `<entity>.go` with the model struct (marked `//easyjson:json`), enum types, and JSON tags (`snake_case`).
2. Create `<entity>_requests.go` with an `<Entity>API` struct wrapping `*Client`, implementing CRUD methods matching the REST endpoints.
3. Register the API in `Client` struct in `client.go`.
4. Add tests in `<entity>_test.go` using table-driven tests and the `JSONMarshalOK`/`JSONUnmarshalOK` helpers from `utils_test.go`.
5. Run `go generate ./...` to regenerate easyjson code.

### CLI Pattern

CLI uses [Cobra](https://github.com/spf13/cobra). Commands live in `cmd/omlox-cli/`:
- Resource commands follow the pattern: `<verb>_<resource>.go` (e.g., `get_trackables.go`, `create_providers.go`)
- Output formatting via `internal/cli/output/` (Table and JSON formats)
- Resource loading from stdin/file via `internal/cli/resource/loader.go` (generic `Loader[T]`)
- Environment: `OMLOX_HUB_API` env var or `--addr` flag for hub endpoint

### Testing

- **Table-driven tests** with descriptive test names
- Generic test helpers: `JSONMarshalOK[T]()` and `JSONUnmarshalOK[T]()` in `utils_test.go`
- Use `github.com/google/go-cmp/cmp` for deep equality
- Use `github.com/nsf/jsondiff` for JSON comparison debugging
- Run: `go test -v -cover -race ./...`

### Code Generation

```bash
go generate ./...   # Runs all three generators:
# 1. easyjson -snake_case -pkg          → omlox_easyjson.go
# 2. go run ./cmd/... gen docs          → docs/cli/*.md  
# 3. copywrite headers                  → copyright headers on all files
```

**Prerequisites:**
```bash
go install github.com/mailru/easyjson/...@latest
go install github.com/hashicorp/copywrite@latest
```

### Build & Release

- Build: `make build` (outputs to `bin/`)
- Test: `make test`
- Lint: `make lint` (golangci-lint)
- CI: GitHub Actions on push/PR (`.github/workflows/ci.yml`) — checks code gen, builds, tests
- Release: Tag-triggered via GoReleaser (`.github/workflows/release.yml`)

## Implementation Status

### Currently Implemented
- **Models**: Error, LocationProvider, Trackable, Location, Point, Polygon, Duration, LocatingRule, WebSocket protocol types
- **REST API**: Full CRUD for Trackables and Providers (summary, list, get, create, update, delete, locations)
- **WebSocket**: Connect, Subscribe, Publish, auto-reconnect with backoff, subscription restoration
- **CLI**: get/create/update/delete for trackables and providers, subscribe command, version, gen docs

### Not Yet Implemented (Opportunities for Contribution)
- **Models**: Collision, CollisionEvent, Fence, FenceEvent, LineString, Zone, Proximity, TrackableMotion
- **REST API**: Zones, Fences, Anchors — all CRUD endpoints
- **REST API**: Additional Provider/Trackable sub-resources (sensors, fences, motions, multiple locations)
- **CLI**: Commands for zones, fences
- **Linting**: golangci-lint CI workflow (TODO in ci.yml)

## Known Issues & Gotchas

- **Deephub subscription bug**: The DeepHub doesn't always return `subscription_id` in subsequent messages after subscription — noted in code comments. The client works around this by routing based on topic when subscription_id is missing.
- **omlox OpenAPI spec**: The official OpenAPI spec has known technical issues making auto-generation impractical. This library is hand-coded.
- **Duration type**: Custom `Duration` supports "infinite" semantics (`Inf = -1`) in addition to normal durations. Handle accordingly.
- **Optional fields with defaults**: Translating omlox optional fields with default values to Go is an ongoing design challenge (pointer types vs zero values).

## Dependencies Reference

| Package | Purpose |
|---------|---------|
| `nhooyr.io/websocket` | WebSocket client (RFC 6455 compliant) |
| `github.com/hashicorp/go-cleanhttp` | Pooled HTTP client with safe defaults |
| `github.com/mailru/easyjson` | Fast JSON marshaling via code generation |
| `github.com/tidwall/geojson` | GeoJSON geometry types (Point, Polygon) |
| `github.com/google/uuid` | UUID generation/parsing for trackable IDs |
| `golang.org/x/time` | Rate limiting (`rate.Limiter`) |
| `github.com/spf13/cobra` | CLI framework |
| `github.com/google/go-cmp` | Test assertion deep equality |
| `github.com/nsf/jsondiff` | JSON diff for test debugging |
