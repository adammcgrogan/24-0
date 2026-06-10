# Contributing

Contributions are very welcome. Here's how to get started:

1. Fork the repo and create a branch from `main`
2. Run `go test ./...` to make sure everything passes
3. Make your changes, add tests where appropriate
4. Open a PR — describe what you changed and why

## Running tests

```bash
go test ./...
go vet ./...
```

## Regenerating driver data

If you want to update `data/drivers.json` with the latest season:

```bash
go run ./cmd/precompute/main.go
```

This hits the [Jolpica F1 API](https://jolpica.com) and may take a minute.

## Code style

Standard `gofmt`. Run `gofmt -w .` before committing.
