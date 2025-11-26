# Omlox Hub Go Client Library

<img align="right" width="450" src="https://us.profinet.com/wp-content/uploads/2020/07/omlox-tech.png">

An [Omlox](https://omlox.com/) Hub™ compatible Go client library and CLI tool.

Omlox is an open standard for precise real-time indoor localization systems.
It specifies open interfaces for an interoperable localization system that enable industry to use a single infrastructure with different applications from different providers.

> [!WARNING]  
> This library is currently being developed. Please try it out and give us feedback! Please do not use it in production or use at your own risk.

## Contents

1. [Installation](#installation)
1. [Examples](#examples)
   - [Getting Started](#getting-started)
   - [Websockets](#websockets)
     - [Subscription](#subscription)
   - [Error Handling](#error-handling)
1. [Status](#status)
   - [Schemas](#schemas)
   - [Methods](#methods)
1. [Specification](#specification)
1. [Community, discussion, contribution, and support](#community-discussion-contribution-and-support)
   - [Code of conduct](#code-of-conduct)
1. [Development](#development)
1. [Disclaimer](#disclaimer)

## Installation

```sh
go get -u github.com/wavecomtech/omlox-client-go
```

> [!NOTE]  
> For the CLI installation, follow the documentation at [./docs/omlox-cli.md](/docs/omlox-cli.md).

## Examples

### Getting Started

Here is a simple example of using the library to query the trackables in the hub.

```go
package main

import (
    "context"
	"log"
	"time"

	"github.com/wavecomtech/omlox-client-go"
)

func main() {
    client, err := omlox.New("https://localhost:7081/v2")
    if err != nil {
        log.Fatal(err)
    }

    trackables, err := client.Trackables.List(context.Background())
    if err != nil {
        log.Fatal(err)
    }

    log.Println("trackables retrieved:", trackables)
}
```

### Websockets

#### Subscription

```go
// Dials a Omlox Hub websocket interface, subscribes to
// the location_updates topic and listens to new
// location messages.

ctx, cancel := context.WithCancel(context.Background())
defer cancel()

client, err := omlox.Connect(ctx, "localhost:7081/v2")
if err != nil {
    log.Fatal(err)
}
defer client.Close()

sub, err := client.Subscribe(ctx, omlox.TopicLocationUpdates)
if err != nil {
    log.Fatal(err)
}

for location := range omlox.ReceiveAs[omlox.Location](sub) {
    _ = location // handle location update
}
```

#### Auto-Reconnection

The client supports automatic WebSocket reconnection with configurable retry policies. This is essential for production deployments where connections may be disrupted due to network issues, server restarts, or high load.

```go
// Create client with auto-reconnection enabled, unlimited retries, and backoff timing
// Dials a Omlox Hub websocket interface, subscribes to
// the location_updates topic and listens to new
// location messages.

ctx, cancel := context.WithCancel(context.Background())
defer cancel()

client, err := omlox.Connect(
    ctx,
    "localhost:7081/v2",
    omlox.WithWSAutoReconnect(true),
    omlox.WithWSMaxRetries(-1),
    omlox.WithWSRetryWait(time.Second, 30*time.Second),
)
if err != nil {
    log.Fatal(err)
}
defer client.Close()

sub, err := client.Subscribe(ctx, omlox.TopicLocationUpdates)
if err != nil {
    log.Fatal(err)
}

for location := range omlox.ReceiveAs[omlox.Location](sub) {
    _ = location // handle location update
}
```

For high-load scenarios, also configure the HTTP connection pool:

```go
client, err := omlox.New(
    "https://localhost:7081/v2",
    omlox.WithWSAutoReconnect(true),
    omlox.WithConnectionPoolSettings(200, 100, 0),  // Prevent pool exhaustion
)
```

### Error Handling

Errors are returned when Omlox Hub responds with an HTTP status code outside of the 200 to 399 range.
If a request fails due to a network error, a different error message will be returned.

In `go>=1.13` you can use the new [errors.As](https://pkg.go.dev/errors#As) method:

```go
trackables, err := client.Trackables.List(context.Background())
if err != nil {
    var e *omlox.Error
    if errors.As(err, &e) && e.Code == http.StatusNotFound {
        // special handling for 404 errors
    }
}
```

For older go versions, you can also do type assertions:

```go
trackables, err := client.Trackables.List(context.Background())
if err != nil {
    if e, ok := err.(*omlox.Error); ok && e.Code == http.StatusNotFound {
        // special handling for 404 errors
    }
}
```

## Status

This library is coded from scratch to match the specification of Omlox Hub API.
We plan to auto-generate most of it, but the OpenAPI spec is currently invalid due to some technical decisions in the spec.
For the meantime, this library will continue to be coded from scratch.

An older iteration of this library is currently being used internally in Wavecom Technologies.
We plan to migrate to the open source version earlier next year.
The CLI is currently used by the RTLS team as a support tool for our internal Omlox Hub instance.

Following is the current checklist of the implemented schemas and API methods.

### Schemas

> [!WARNING]  
> The provided schemas are prone to change.
> Optional fields with default values have been the most challenging thing to translate well in to Go.
> We are trying different options and see which feels better.

| Schema                        |  Implemented   |
| ----------------------------- | :------------: |
| Collision                     |                |
| CollisionEvent                |                |
| Error                         |       ✅       |
| Fence                         |                |
| FenceEvent                    |                |
| LineString                    |                |
| LocatingRule                  |                |
| Location                      |                |
| LocationProvider              |       ✅       |
| Point                         |       ✅       |
| Polygon                       |       ✅       |
| Proximity                     |                |
| Trackable                     |       ✅       |
| TrackableMotion               |                |
| WebsocketError                |       ✅       |
| WebsocketMessage              | API abstracted |
| WebSocketSubscriptionResponse | API abstracted |
| WebsocketSubscriptionRequest  | API abstracted |
| Zone                          |                |

### Methods

| Method | Endpoint                     | Implemented |
| ------ | ---------------------------- | :---------: |
| GET    | `/zones/summary`             |             |
| GET    | `/zones`                     |             |
| POST   | `/zones`                     |             |
| DELETE | `/zones`                     |             |
| GET    | `/zones/:zoneID`             |             |
| PUT    | `/zones/:zoneID`             |             |
| DELETE | `/zones/:zoneID`             |             |
| PUT    | `/zones/:zoneID/transform`   |             |
| GET    | `/zones/:zoneID/createfence` |             |

| Method | Endpoint                             | Implemented |
| ------ | ------------------------------------ | :---------: |
| GET    | `/trackables/summary`                |     ✅      |
| GET    | `/trackables`                        |     ✅      |
| POST   | `/trackables`                        |     ✅      |
| DELETE | `/trackables`                        |     ✅      |
| GET    | `/trackables/:trackableID`           |     ✅      |
| DELETE | `/trackables/:trackableID`           |     ✅      |
| PUT    | `/trackables/:trackableID`           |     ✅      |
| GET    | `/trackables/:trackableID/fences`    |             |
| GET    | `/trackables/:trackableID/location`  |     ✅      |
| GET    | `/trackables/:trackableID/locations` |             |
| GET    | `/trackables/:trackableID/motion`    |             |
| GET    | `/trackables/:trackableID/providers` |             |
| GET    | `/trackables/:trackableID/sensors`   |             |
| GET    | `/trackables/motions`                |             |

| Method | Endpoint                           | Implemented |
| ------ | ---------------------------------- | :---------: |
| GET    | `/providers/summary`               |     ✅      |
| GET    | `/providers`                       |     ✅      |
| POST   | `/providers`                       |     ✅      |
| DELETE | `/providers`                       |     ✅      |
| GET    | `/providers/:providerID`           |     ✅      |
| PUT    | `/providers/:providerID`           |     ✅      |
| DELETE | `/providers/:providerID`           |     ✅      |
| PUT    | `/providers/:providerID/location`  |     ✅      |
| GET    | `/providers/:providerID/location`  |             |
| DELETE | `/providers/:providerID/location`  |             |
| GET    | `/providers/:providerID/fences`    |             |
| PUT    | `/providers/:providerID/sensors`   |             |
| GET    | `/providers/:providerID/sensors`   |             |
| GET    | `/providers/locations`             |             |
| PUT    | `/providers/locations`             |             |
| DELETE | `/providers/locations`             |             |
| PUT    | `/providers/:providerID/proximity` |             |
| PUT    | `/providers/proximities`           |             |

| Method | Endpoint                     | Implemented |
| ------ | ---------------------------- | :---------: |
| GET    | `/fences/summary`            |             |
| GET    | `/fences`                    |             |
| POST   | `/fences`                    |             |
| DELETE | `/fences`                    |             |
| GET    | `/fences/:fenceID`           |             |
| PUT    | `/fences/:fenceID`           |             |
| DELETE | `/fences/:fenceID`           |             |
| GET    | `/fences/:fenceID/providers` |             |
| GET    | `/fences/:fenceID/locations` |             |

## Specification

The Hub specification can be found at:
<https://www.profibus.com/download/omlox-hub-specification-api-and-behavior>.

## Community, discussion, contribution, and support

Contributions are made to this repo via Issues and Pull Requests (PRs).
If any part of the project has a bug or documentation mistakes, please let us know by opening an issue. The project is early in its development, so bugs and mistakes may appear.

To discuss API and feature suggestions, implementation insights and other things related to the implementation of the client, you can use Github's Discussions tab.

Before creating an issue, please check that an issue reporting the same problem does not already exist. Please try to create issues that are accurate and easy to understand.
Maintainers might ask for further information to resolve the issue.

### Code of conduct

By participating and contributing to this project, you agree to uphold our [Code of Conduct](./CODE_OF_CONDUCT.md).

## Development

> [!NOTE]  
> **For an optimal developer experience, it is recommended to install [Nix](https://nixos.org/download.html) and [direnv](https://direnv.net/docs/installation.html).**

<details><summary><i>Installing Nix and direnv</i></summary><br>

**Note: These are instructions that _SHOULD_ work in most cases. Consult the links above for the official instructions for your OS.**

Install Nix:

```sh
sh <(curl -L https://nixos.org/nix/install) --daemon
```

Consult the [installation instructions](https://direnv.net/docs/installation.html) to install direnv using your package manager.

On MacOS:

```sh
brew install direnv
```

Install from binary builds:

```sh
curl -sfL https://direnv.net/install.sh | bash
```

The last step is to configure your shell to use direnv. For example for bash, add the following lines at the end of your `~/.bashrc`:

    eval "\$(direnv hook bash)"

**Then restart the shell.**

For other shells, see [https://direnv.net/docs/hook.html](https://direnv.net/docs/hook.html).

**MacOS specific instructions**

Nix may stop working after a MacOS upgrade. If it does, follow [these instructions](https://github.com/NixOS/nix/issues/3616#issuecomment-662858874).

<hr>
</details>

Otherwise, you can install the required dependencies with Go itself.

Be sure to use Go `>=` to the one defined in `go.mod` and have `$GOPATH/bin` in your `$PATH`, so you can run installed go binaries. You can see how to do that at [setup-your-shell-to-run-go-installed-binaries](./docs/omlox-cli.md#setup-your-shell-to-run-go-installed-binaries).

Install development dependencies:

```console
go install github.com/mailru/easyjson/...@latest
go install github.com/hashicorp/copywrite@latest
```

You should be good to go!
If you have any trouble getting started, reach out to us by email (see the [MAINTAINERS](./MAINTAINERS) file).

## Disclaimer

> [!NOTE]  
> The code provided by this library is not certified by Omlox or the Profibus & Profinet International.
> Solutions using this library should go through the [certification process](https://omlox.com/certification) defined by the Omlox™ consortium to be an _"omlox certified solution"_.