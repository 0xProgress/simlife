# SIMLIFE — MASTER AI CODING CONTEXT PROMPT
## Paste this at the start of every new AI coding session

---

You are building **Simlife** — a flagship Discord persistent economy and life simulation bot. This is a production-grade, portfolio-quality SaaS product, not a hobby project. Code quality, architecture decisions, and naming conventions must reflect that standard at all times.

---

## WHAT SIMLIFE IS

Simlife is a globally-shared economy simulation running inside Discord. Every player operates in the same world. Every transaction a player makes affects the economic conditions every other player operates in. Prices shift from real supply and demand. Businesses produce real goods. Labor has market value. Wealth is a consequence of strategy.

This is not a per-server economy bot. There is one world. All servers share it.

---

## THREE-SERVICE ARCHITECTURE

The project is three independent services that share one database:

**1. Go Backend (`/bot`)** — The authoritative economic engine. Handles all Discord slash commands, the financial ledger, market matching, daily settlement, the HTTP API for the Activity frontend, all application logging (zerolog, JSON structured output), and dynamic image compositing for Discord embeds. This is the only service that writes to the financial ledger.

**2. React + TypeScript Activity Frontend (`/activity`)** — A Discord Embedded App (iframe) using `@discord/embedded-app-sdk`. Renders: City View (zoomable canvas city grid), Business Dashboard, Market View with price charts, Property Manager, Economic Report. Communicates with the Go backend via authenticated HTTPS and WebSocket only. It is a rendering layer — it holds no authoritative state.

**3. Python Analytics Service (`/analytics`)** — Read-only database access. Runs once per economic day, triggered by the NATS `settlement.complete` event. Computes: Gini coefficient, price index, transaction velocity, anomaly detection (graph-based via networkx), market trend analysis. Posts results to Go via internal HTTP API with shared-secret auth. Has no logging infrastructure of its own — it reports events to Go which logs them. Never touches live player money.

---

## SHARED INFRASTRUCTURE

- **PostgreSQL 15+** — Single source of truth. All economic state. Double-entry ledger (append-only `transactions` table, no balance columns — balances computed from ledger entries).
- **Redis 7+** — Rate limiting, session tokens, hot player data cache, pricing cache. Never a source of truth.
- **NATS JetStream** — Async events between services. Key streams: `settlement_events`, `market_events`, `notifications`.
- **sqlc** — All Go database access via type-safe generated functions from raw SQL query files. No ORM.
- **golang-migrate** — Versioned, additive-only database migrations.
- **Caddy** — Reverse proxy, TLS, Discord Activity URL mappings.

---

## FINANCIAL ENGINE RULES — NEVER VIOLATE THESE

1. Every financial transaction goes through `internal/economy/ledger.go`'s `PostTransaction` function. No exceptions.
2. No code outside the ledger package may directly write to any account balance.
3. All ledger writes use PostgreSQL `SERIALIZABLE` transaction isolation.
4. Every transaction is a double-entry pair — one debit, one credit. The sum of all entries always equals zero.
5. The `transactions` table is append-only. Rows are never updated or deleted.
6. Balances are computed by summing transaction entries. There are no balance columns to update.

---

## IMAGE COMPOSITING SYSTEM

Every Discord command response includes a dynamically composited PNG image attached to the embed. The Go backend handles this in `internal/imaging/`. The flow is: command handler calls `composer.Compose(layoutName, dataStruct)` → composer loads the pre-designed background PNG from memory (loaded at startup, never read from disk at request time) → layout renderer draws data fields onto the canvas using the `gg` library → returns PNG bytes → command handler attaches to Discord embed. Background images are static pre-designed assets in `internal/imaging/assets/backgrounds/`. Fonts are in `internal/imaging/assets/fonts/`. All compositing is pure in-memory. Target: under 30ms per image.

Layout files: `balance.go`, `business.go`, `market.go`, `property.go`, `profile.go`, `economic_news.go`, `shop.go`.

---

## LOGGING

Go is the sole logging authority for the entire system. Uses zerolog. JSON output in production, pretty-print in development. Every log entry includes: timestamp, level, service name, package, and structured fields. The Python analytics service does not log independently — it reports status to Go's internal API endpoints, and Go logs them. Log levels: DEBUG (dev only), INFO (normal operations), WARN (recoverable issues), ERROR (player-affecting failures), FATAL (process exit required).

---

## DISCORD EMBED DESIGN RULES

- Every command response embed includes: title (command category), composited PNG image (1200×630px) as the embed image, footer with current economic day and time to next settlement.
- Embed accent color by category: gold for economy commands, blue for market commands, green for income events, red for alerts/losses.
- Minimal text fields in embeds — the image carries the data.
- Ephemeral responses for personal data (balance, bank, profile).
- Public responses for market events and economic news.

---

## TECHNOLOGY VERSIONS

- Go: latest stable
- DiscordGo: github.com/bwmarrin/discordgo
- Image compositing: github.com/fogleman/gg + github.com/disintegration/imaging
- Python: 3.11+
- Python analytics libraries: pandas, numpy, networkx, SQLAlchemy (read-only), APScheduler
- React: 18+ with TypeScript
- Discord Activity SDK: @discord/embedded-app-sdk
- State management: Redux Toolkit
- Data fetching: React Query (TanStack Query)
- Charts: Recharts
- Build tool: Vite

---

## PROJECT FILE STRUCTURE (TOP LEVEL)

```
simlife/
├── bot/          # Go backend
├── activity/     # React Activity frontend
├── analytics/    # Python analytics service
├── infra/        # Docker, Caddy, NATS, Postgres init
├── .env.example
└── README.md
```

---

## CODE QUALITY STANDARDS

- Go: exported functions have godoc comments. Error handling is explicit — no ignored errors. All financial functions return errors, never panic.
- TypeScript: strict mode enabled. No `any` types. Props interfaces defined for every component.
- Python: type hints on all function signatures. No bare `except` clauses.
- SQL: queries in `db/queries/*.sql` files, compiled by sqlc. No raw string SQL in Go application code.
- All secrets come from environment variables. No hardcoded credentials anywhere.
- Every new feature lives behind a feature flag in config until it is fully tested.

---

## WHAT YOU ARE BUILDING RIGHT NOW

[REPLACE THIS SECTION WITH THE SPECIFIC PART PROMPT BEFORE SENDING]
