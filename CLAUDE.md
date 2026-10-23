# CLAUDE.md — omlox-client-go

## Project

Go client library and CLI for the omlox Hub real-time locating standard (`github.com/wavecomtech/omlox-client-go`). Tested against Flowcate's DeepHub®. Go 1.21+, uses generics, `log/slog`, easyjson code generation.

## Commands

```bash
make build                # Build CLI to bin/
make test                 # go test -v -cover -race ./...
make lint                 # golangci-lint run ./...
go generate ./...         # Regenerate easyjson, docs, copyright headers
```

## Architecture

Root package `omlox` is the public library. `cmd/omlox-cli/` is the Cobra CLI. `internal/cli/` has CLI internals.

**Adding a new omlox entity** (e.g., Zone, Fence):
1. `<entity>.go` — model struct marked `//easyjson:json`, enum types, `snake_case` JSON tags
2. `<entity>_requests.go` — `<Entity>API` struct wrapping `*Client` with CRUD methods
3. Register API in `Client` struct in `client.go`
4. `<entity>_test.go` — table-driven tests using `JSONMarshalOK[T]`/`JSONUnmarshalOK[T]`
5. `go generate ./...` to regenerate

## Style

- Copyright: `// Copyright (c) Omlox Client Go Contributors` + `// SPDX-License-Identifier: MIT`
- JSON: `snake_case` fields (easyjson `-snake_case` flag)
- All public API methods: `context.Context` as first parameter
- Functional options: `WithXxx(...)` returning `ClientOption`
- Never edit `omlox_easyjson.go` — auto-generated
- Tests: table-driven, `go-cmp` for equality, `jsondiff` for debugging
- CLI commands: `<verb>_<resource>.go` in `cmd/omlox-cli/`

## Domain

Omlox Hub entities: **Zone** (physical area + coordinate transform), **Location Provider** (positioning device: UWB/GPS/Wi-Fi/RFID/BLE), **Trackable** (tracked asset with multiple providers), **Fence** (geofence triggering events), **Location** (GeoJSON position update).

REST API at `/v2`. WebSocket at `/ws/socket` — topic-based pub/sub: `location_updates`, `fence_events`, `collision_events`, `trackable_motions`, `*_changes`.

Default CRS: WGS84 (EPSG:4326). Local coordinates transformed to global by Hub.

## Not Yet Implemented

Models: Collision, CollisionEvent, Fence, FenceEvent, LineString, Zone, Proximity, TrackableMotion. REST: Zones, Fences, Anchors. CLI: zone/fence commands.
