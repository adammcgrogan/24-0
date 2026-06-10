# 24-0

**Can your dream F1 team go 24-0?**

Spin a random F1 constructor and season, draft one driver at a time, build your ultimate 5-driver lineup, and simulate a full 24-race championship. The perfect team wins every race — but can you find it?

Inspired by [38-0.app](https://38-0.app).

## Tech stack

- **Go** — backend + server-rendered HTML
- **HTMX** — reactive UI without a JS framework
- **Tailwind CSS** — styling via CDN
- **Vercel Postgres (Neon)** — leaderboard storage
- **Vercel** — hosting

## Local development

```bash
cp .env.example .env
# Fill in DATABASE_URL with a Postgres connection string

go run ./api/index.go
# Visit http://localhost:8080
```

## Precompute driver data

The game uses pre-computed driver ratings from the [Jolpica F1 API](https://jolpica.com) (Ergast successor). Run this once to regenerate `data/drivers.json`:

```bash
go run ./cmd/precompute/main.go
```

## Deploy

```bash
vercel --prod
```

Set `DATABASE_URL` in your Vercel project environment variables.

## Contributing

PRs welcome. See [CONTRIBUTING.md](CONTRIBUTING.md).

## License

MIT
