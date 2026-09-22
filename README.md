# poweradmin-go

A Go client library for the [Poweradmin](https://www.poweradmin.org/) DNS
management API (v2).

```go
import "github.com/contentways/poweradmin-go/v4/poweradmin"
```

## Version compatibility

| poweradmin-go | Import path                                   | Poweradmin | Go     |
| ------------- | --------------------------------------------- | ---------- | ------ |
| 4.x           | `github.com/contentways/poweradmin-go/v4/...` | 4.3.0+     | ≥ 1.26 |
| 3.x           | `github.com/contentways/poweradmin-go/v3/...` | 4.3.0+     | ≥ 1.26 |
| 1.1.x         | `contentways.dev/contentways/poweradmin-go/...` | 4.3.0+   | ≥ 1.26 |
| 1.0.x         | `contentways.dev/contentways/poweradmin-go/...` | < 4.3.0  | ≥ 1.26 |

Poweradmin 4.3.0 standardized the v2 API so every endpoint wraps its payload
under a named key (`data.zones`, `data.records`, `data.rrset`, …). Earlier
releases returned most collection and single-resource endpoints as bare
arrays/objects. Pick the client line that matches your server: use 1.0.x
against Poweradmin older than 4.3.0, and a newer line against 4.3.0 and newer.

### Upgrading from v3

- Change imports from `.../poweradmin-go/v3/poweradmin` to
  `.../poweradmin-go/v4/poweradmin`.
- `ZoneUpdateOpts.Account` was removed. The Poweradmin API cannot change a
  zone's account after creation, so the field never had an effect; set the
  account via `ZoneCreateOpts` instead.
- `ZoneUpdateOpts.Masters` now actually updates the masters, and the new
  `ZoneUpdateOpts.Name` renames a zone.
- `APIError.Message` now carries the API's message (e.g. `Zone already exists`)
  instead of the raw JSON response body.

## Installation

```sh
go get github.com/contentways/poweradmin-go/v4
```

Requires Go 1.26 or newer.

## Quickstart

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/contentways/poweradmin-go/v4/poweradmin"
)

func main() {
    client, err := poweradmin.NewClient(
        poweradmin.WithBaseURL("https://dns.example.com"),
        poweradmin.WithAPIKey("your-api-key"),
        poweradmin.WithRetry(3),
    )
    if err != nil {
        log.Fatal(err)
    }

    zones, err := client.Zone.All(context.Background())
    if err != nil {
        log.Fatal(err)
    }
    for _, z := range zones {
        fmt.Printf("[%d] %s (%s)\n", z.ID, z.Name, z.Type)
    }
}
```

A more complete CLI example lives in [`examples/main.go`](examples/main.go).

## Authentication

Two authentication methods are supported:

```go
// API key (recommended) — sent as an Authorization: Bearer token.
poweradmin.WithAPIKey("your-api-key")

// HTTP Basic auth.
poweradmin.WithBasicAuth("user", "password")
```

## Client options

| Option                  | Purpose                                              |
| ----------------------- | ---------------------------------------------------- |
| `WithBaseURL`           | Base URL of the Poweradmin instance                  |
| `WithAPIKey`            | Bearer token authentication                          |
| `WithBasicAuth`         | HTTP Basic authentication                            |
| `WithAPIVersion`        | Override API version prefix (default `v2`)           |
| `WithHTTPClient`        | Inject a custom `*http.Client`                       |
| `WithTimeout`           | HTTP timeout (default 30s)                           |
| `WithRetry`             | Retry on network errors, 429, and 5xx                |
| `WithRetryBackoff`      | Custom backoff strategy (defaults to exp + jitter)   |
| `WithDebugWriter`       | Log requests to an `io.Writer` (e.g. `os.Stderr`)    |

## Resources

The client exposes the Poweradmin API resources as services on the client:

| Service                       | Endpoint                          |
| ----------------------------- | --------------------------------- |
| `client.Zone`                 | `/v2/zones`                       |
| `client.Record`               | `/v2/zones/{id}/records`          |
| `client.RRSet`                | `/v2/zones/{id}/rrsets`           |
| `client.ZoneTemplate`         | `/v2/zone-templates`              |
| `client.User`                 | `/v2/users`                       |
| `client.Group`                | `/v2/groups`                      |
| `client.Permission`           | `/v2/permissions`                 |
| `client.PermissionTemplate`   | `/v2/permission-templates`        |

Each list endpoint provides:

- `List(ctx, opts)` — one page of results plus the raw `*Response`
- `All(ctx, ...)` — iterates all pages and returns a single slice
- `GetByName(ctx, name)` — convenience lookup (linear scan over pages)

## Pagination

List endpoints accept a `ListOpts` containing `Page` and `PerPage`. The
`*Response` value carries pagination metadata; use `All` to consume every
page transparently.

```go
zones, resp, err := client.Zone.List(ctx, poweradmin.ListOpts{
    Page:    1,
    PerPage: 100,
})
_ = resp.Pagination // *schema.Pagination
```

## Retries

`WithRetry(n)` enables automatic replays on transient failures (network
errors, HTTP 429, 5xx) with exponential backoff and jitter. `n` is the
total number of attempts including the first one; values below 2 disable
retrying.

```go
poweradmin.WithRetry(5)
poweradmin.WithRetryBackoff(func(attempt int) time.Duration {
    return time.Duration(attempt) * time.Second
})
```

## Debug logging

```go
poweradmin.WithDebugWriter(os.Stderr)
```

One line is written per HTTP request: method, path, status, and duration.

## Development

Pre-commit hooks (formatting, `go mod tidy`, `go generate`, `golangci-lint`)
are configured in [`.pre-commit-config.yaml`](.pre-commit-config.yaml):

```sh
pip install pre-commit
pre-commit install
```

Run the test suite:

```sh
go test ./...
```

Regenerate the interface stubs after changing a service:

```sh
go generate ./...
```

## License

[MIT](LICENSE.md) © contentways
