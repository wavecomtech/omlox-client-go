---
applyTo: "cmd/omlox-cli/**"
---

# CLI Conventions

- CLI uses the [Cobra](https://github.com/spf13/cobra) framework.
- Command file naming: `<verb>_<resource>.go` (e.g., `get_trackables.go`, `create_providers.go`).
- The hub endpoint comes from `--addr` flag or `OMLOX_HUB_API` environment variable (default: `localhost:8081`).
- Output formats: JSON (`-o json`) and Table (`-o table`, default) via `internal/cli/output/`.
- Resource input from file (`-f <path>`) or stdin via the generic `Loader[T]` in `internal/cli/resource/loader.go`.
- Delete commands require `--yes` flag for confirmation or prompt the user interactively.
- Shell completion is provided for resource names (trackable IDs, provider IDs, topic names).
- Use `internal/cli/settings.go` for environment configuration (`EnvSettings` struct).
