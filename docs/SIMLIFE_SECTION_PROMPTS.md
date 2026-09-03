# SIMLIFE — PER-SECTION BUILD PROMPTS
## How to use: Paste the Master Prompt first, then paste the relevant section prompt below it, then paste the relevant doc excerpts from SIMLIFE_BUILD_DOCUMENTATION.md

---

## HOW TO STRUCTURE EACH SESSION

Every AI coding session follows this structure:

```
[MASTER PROMPT]
+
[SECTION PROMPT from this file]
+
[Paste the relevant file section from SIMLIFE_BUILD_DOCUMENTATION.md]
+
[Paste the relevant layer description from the Architecture Layers section]
```

The master prompt gives the AI the full project picture. The section prompt tells it exactly what it is building right now. The doc excerpts give it the specific file and layer detail. You never start a session without all three.

---

---

## SECTION 1 — Go Bot Foundation
### Use when building: main.go, config, bot session, router, middleware, registry

---

You are building the Go backend foundation for Simlife. This session covers the program entry point, configuration loading, the DiscordGo session, the command router, and the command registry.

**Your job in this session:**
Build the files that wire the entire Go service together. When this session is complete, the bot should start up, connect to Discord, register slash commands, and route incoming interactions to placeholder handlers. No financial logic yet — this is the skeleton the rest of the service is built on.

**Files you are building:**
- `bot/cmd/simlife/main.go`
- `bot/internal/config/config.go`
- `bot/internal/logger/logger.go`
- `bot/internal/logger/middleware.go`
- `bot/internal/bot/bot.go`
- `bot/internal/bot/router.go`
- `bot/internal/bot/middleware.go`
- `bot/internal/commands/registry.go`

**Key requirements:**
- Logger initializes first, before anything else. All subsequent startup steps log their outcome.
- Config validation fails fast with a clear error if any required variable is missing.
- Background assets (image compositing backgrounds and fonts) are loaded into memory during startup in main.go.
- The router checks feature flags before dispatching any command.
- The registry is a pure data structure — no logic, just command definitions and handler mappings.

**[PASTE RELEVANT BUILD DOC SECTIONS HERE]**

---

---

## SECTION 2 — Database Schema and Migrations
### Use when building: migration files, sqlc queries, sqlc.yaml

---

You are building the database layer for Simlife. This session covers all PostgreSQL migration files and the sqlc query files for each domain.

**Your job in this session:**
Write the complete database schema as numbered migration files, and write the SQL query files that sqlc will compile into type-safe Go functions. When this session is complete, running the migrations against a fresh PostgreSQL database produces the complete schema, and running sqlc generates all the Go database access code the application needs.

**Files you are building:**
- `bot/db/migrations/001_init_players.sql` through `008_init_analytics_snapshots.sql`
- `bot/db/queries/players.sql`, `ledger.sql`, `market.sql`, `business.sql`, `property.sql`, `world.sql`
- `bot/db/sqlc.yaml`

**Key requirements:**
- The `transactions` table is append-only. No UPDATE or DELETE queries should exist for it in the query files.
- There are no balance columns on any table. Balances are always computed by summing transaction entries.
- Account types for the `accounts` table: WALLET, BANK, ESCROW, TREASURY.
- All foreign keys have appropriate ON DELETE behavior (soft deletes on players, not cascades that destroy economic history).
- JSONB columns for business inventory and production configuration — flexible enough to add new item types without migrations.
- sqlc.yaml must be configured to generate Go code in `bot/db/sqlc/` with the correct package name.

**[PASTE RELEVANT BUILD DOC SECTIONS HERE — Database Design section + relevant layer]**

---

---

## SECTION 3 — Financial Ledger
### Use when building: ledger.go, accounts logic, transaction posting

---

You are building the financial ledger for Simlife. This is the most critical file in the entire codebase.

**Your job in this session:**
Build `internal/economy/ledger.go` — the double-entry bookkeeping engine that every financial transaction in the system flows through. When this session is complete, the ledger can post transactions atomically, validate all inputs, enforce double-entry balance, use SERIALIZABLE isolation, and log every transaction with structured fields.

**Files you are building:**
- `bot/internal/economy/ledger.go`

**Key requirements:**
- `PostTransaction` is the only exported function that modifies account state.
- Every call to `PostTransaction` inserts exactly two rows into `transactions` — one DEBIT and one CREDIT — in a single database transaction.
- SERIALIZABLE isolation on every write. Serialization failures are returned as errors and retried by callers.
- Input validation: amount must be positive and non-zero, source account must have sufficient balance (computed from sum of prior entries), both accounts must exist, both accounts must belong to the same world.
- Structured log entry on every successful post: transaction UUID, type, source account ID, destination account ID, amount, duration.
- No panics. All error paths return explicit errors with context.

**[PASTE RELEVANT BUILD DOC SECTIONS HERE — ledger.go description + Layer 1 description]**

---

---

## SECTION 4 — Economy Commands (Layer 1)
### Use when building: balance, work, bank, pay, shop command handlers

---

You are building the Layer 1 economy commands for Simlife. These are the first player-facing commands and the first integration of the image compositing system with the command layer.

**Your job in this session:**
Build all five Layer 1 command handlers. Each handler fetches player data, calls the imaging composer to generate a PNG, and responds to the Discord interaction with an embed containing the image. When this session is complete, players can run `/balance`, `/work`, `/deposit`, `/withdraw`, `/pay`, and `/shop` and receive styled image responses.

**Files you are building:**
- `bot/internal/commands/economy/balance.go`
- `bot/internal/commands/economy/work.go`
- `bot/internal/commands/economy/bank.go`
- `bot/internal/commands/economy/pay.go`
- `bot/internal/commands/economy/shop.go`

**Key requirements:**
- Every handler calls `composer.Compose(layoutName, data)` and attaches the PNG bytes to the embed as a file.
- Balance responses are ephemeral.
- Pay responses are public (both sender confirmation and recipient notification).
- Work enforces a Redis cooldown. The cooldown duration comes from config, not a hardcoded value.
- All financial operations go through `ledger.PostTransaction`. No direct database writes to balances.
- Embed accent colors: gold for economy commands.

**[PASTE RELEVANT BUILD DOC SECTIONS HERE — economy command file descriptions + Layer 1 description]**

---

---

## SECTION 5 — Image Compositing System
### Use when building: composer, renderer, all layout files, asset loading

---

You are building the dynamic image compositing system for Simlife. This system generates every PNG image attached to Discord embed responses.

**Your job in this session:**
Build the complete imaging package — the composer, the shared renderer primitives, and all seven layout files. When this session is complete, any command handler can call `composer.Compose(layoutName, dataStruct)` and receive a PNG byte slice ready to attach to a Discord embed, generated in under 30 milliseconds.

**Files you are building:**
- `bot/internal/imaging/composer.go`
- `bot/internal/imaging/renderer.go`
- `bot/internal/imaging/layouts/balance.go`
- `bot/internal/imaging/layouts/business.go`
- `bot/internal/imaging/layouts/market.go`
- `bot/internal/imaging/layouts/property.go`
- `bot/internal/imaging/layouts/profile.go`
- `bot/internal/imaging/layouts/economic_news.go`
- `bot/internal/imaging/layouts/shop.go`

**Key requirements:**
- Background PNGs are loaded into memory at startup (passed in via the asset cache, not read from disk in Compose).
- Compose must not read from disk at request time under any circumstance.
- If Compose fails, it returns an error — the command handler falls back to a plain text embed. Compositing failure must never crash the command.
- Renderer primitives to implement: DrawText, DrawCurrencyValue (with color for positive/negative), DrawProgressBar, DrawMiniChart (sparkline from []float64), DrawAvatar (fetch Discord avatar URL and composite).
- Canvas size: 1200×630px for all layouts (standard Discord embed image ratio).
- Use the `gg` library for canvas drawing. Use the `imaging` library for image loading and resizing.
- Monospace font (mono.ttf) for all numerical values. Primary font (primary.ttf) for all other text.

**[PASTE RELEVANT BUILD DOC SECTIONS HERE — Section 8: Dynamic Image Compositing System in full]**

---

---

## SECTION 6 — Settlement Engine
### Use when building: settlement.go, monitor.go, tax.go, scheduler, settlement jobs

---

You are building the daily settlement engine for Simlife. This is the economic heartbeat of the bot — it runs once per day and closes the economic books.

**Your job in this session:**
Build the settlement engine and all its supporting components. When this session is complete, the settlement job runs on schedule, processes all outstanding economic obligations in the correct sequence, computes the economic snapshot, and publishes the daily economic news embed with a composited image to configured Discord channels.

**Files you are building:**
- `bot/internal/economy/settlement.go`
- `bot/internal/economy/monitor.go`
- `bot/internal/economy/tax.go`
- `bot/internal/jobs/scheduler.go`
- `bot/internal/jobs/daily_settlement.go`
- `bot/internal/jobs/economic_news.go`
- `bot/internal/jobs/tax_collection.go`
- `bot/internal/jobs/wage_distribution.go`

**Key requirements:**
- Settlement phases run in this exact order: expire listings → process production → pay wages → collect taxes → distribute dividends → compute snapshot → notify analytics via NATS → publish economic news.
- The entire settlement job has a hard 10-minute timeout context. On timeout: cancel, rollback where possible, log FATAL, notify developer Discord channel.
- Each settlement phase is individually logged at INFO with duration.
- The economic news embed uses the `economic_news` layout from the compositing system.
- Monitor computes: total money supply, velocity (24h transaction volume / money supply), top 10 wealthiest, inequality ratio (wealth of top 10% / wealth of bottom 50%).
- Base wage rate is set by monitor based on velocity — stored in Redis for work command to read.
- Tax collection at Layer 1 is transaction tax only. Architecture must support adding property tax (Layer 5) and configurable rates (Layer 7) without restructuring.

**[PASTE RELEVANT BUILD DOC SECTIONS HERE — settlement.go, monitor.go, tax.go descriptions + Layer 1 and Layer 2 descriptions]**

---

---

## SECTION 7 — Market System (Layer 2)
### Use when building: market.go, pricing.go, market command handlers

---

You are building the global player-to-player marketplace for Simlife. This is the first system where one player's actions directly affect another player's economic conditions.

**Your job in this session:**
Build the market engine, pricing engine, and all market command handlers. When this session is complete, players can list items for sale, buy from listings, and bid on auctions. Market prices respond to supply and demand. Price history is tracked and feeds the composited market history image.

**Files you are building:**
- `bot/internal/economy/market.go`
- `bot/internal/economy/pricing.go`
- `bot/internal/commands/market/list.go`
- `bot/internal/commands/market/buy.go`
- `bot/internal/commands/market/bid.go`
- `bot/internal/commands/market/history.go`

**Key requirements:**
- Market listings move items into escrow at creation time. Escrow uses the player's ESCROW account in the ledger — treated as a real ledger account.
- Every completed trade is recorded in `market_trades` with price, quantity, buyer, seller, and timestamp.
- Pricing engine reads the last 7 days of `market_trades` for a weighted moving average. More recent trades weighted higher.
- Supply signal: count of active listings for an item type. Demand signal: buy-request events in the last 24 hours (tracked in Redis, not database).
- Computed prices published to Redis after settlement. Market command handlers read from Redis cache, not computing on every command.
- The `/market history` command generates a composited mini sparkline chart image using `DrawMiniChart`.

**[PASTE RELEVANT BUILD DOC SECTIONS HERE — market.go, pricing.go, market command descriptions + Layer 2 description]**

---

---

## SECTION 8 — Business Engine (Layer 4)
### Use when building: business engine, production, wages, business commands

---

You are building the business engine for Simlife. This is the first system where players can employ other players and participate in a production economy.

**Your job in this session:**
Build the business engine and all business commands. When this session is complete, players can open businesses, hire workers, log labor, and see production results after settlement. Business owners manage their operation through both Discord commands and the Activity Business Dashboard.

**Files you are building:**
- `bot/internal/business/engine.go`
- `bot/internal/business/production.go`
- `bot/internal/business/wages.go`
- `bot/internal/business/types.go`
- `bot/internal/commands/business/open.go`
- `bot/internal/commands/business/hire.go`
- `bot/internal/commands/business/fire.go`
- `bot/internal/commands/business/produce.go`
- `bot/internal/commands/business/status.go`

**Key requirements:**
- Business types and their production recipes (inputs, outputs, required labor hours, daily capacity) are defined in `types.go` as a typed registry, not hardcoded in the engine.
- Opening a business charges the opening fee through `ledger.PostTransaction` with type `BUSINESS_OPEN`.
- Wage payments during settlement use `ledger.PostTransaction` with type `WAGE`. If the business account cannot cover wages, a debt record is created and the shortfall carries to next settlement with penalty.
- Production failure (insufficient inputs or labor) is logged against the business with a reason code. Business owners see this in the status command and Activity dashboard.
- `/business status` generates a composited image using the business layout.
- The business status data is also served via the Go API to the React Activity Business Dashboard.

**[PASTE RELEVANT BUILD DOC SECTIONS HERE — business engine file descriptions + Layer 4 description]**

---

---

## SECTION 9 — Discord Activity (React Frontend)
### Use when building: Activity setup, auth flow, panels, WebSocket, store

---

You are building the Discord Activity (iframe) frontend for Simlife. This is a React TypeScript application running inside Discord using the Embedded App SDK.

**Your job in this session:**
Build the complete Activity frontend. When this session is complete, players can open the Activity inside Discord, authenticate via OAuth, view the City View, Business Dashboard, Market View, Economic Report, and Player Profile panels, and interact with market and business actions through the Activity UI.

**Files you are building:**
- All files under `activity/src/`

**Key requirements:**
- The Discord SDK must be initialized before any API calls are made. `sdk.ready()` must be called before the Activity is shown.
- JWT is stored in memory only. Never localStorage. Never sessionStorage.
- The API client in `api/client.ts` is the only file that knows the API base URL and handles auth headers.
- The city map uses canvas rendering (not DOM nodes per tile) for performance.
- WebSocket connection authenticates with the player JWT and reconnects with exponential backoff.
- Redux store is the single client-side state source. WebSocket events dispatch Redux actions directly.
- Tailwind is not used — the design system uses CSS custom properties from `styles/tokens.css`.
- Color tokens, font tokens, and spacing tokens from the design system must be used everywhere. No hardcoded color values in component files.
- The Activity must be fully usable on mobile in a 390×480px viewport.

**[PASTE RELEVANT BUILD DOC SECTIONS HERE — Section 6: Discord Activity in full + Section 11: UI Design System in full]**

---

---

## SECTION 10 — Python Analytics Service
### Use when building: all analytics service files

---

You are building the Python analytics service for Simlife. This service is a pure statistical computation engine — it has no authority over player money and no independent logging.

**Your job in this session:**
Build the complete analytics service. When this session is complete, the service subscribes to the NATS `settlement.complete` event, runs the full daily computation pipeline, and posts a structured snapshot to the Go backend's internal analytics endpoint. The Go backend logs all events the analytics service reports.

**Files you are building:**
- All files under `analytics/`

**Key requirements:**
- The service has read-only database access. SQLAlchemy engine must be configured with `execution_options(readonly=True)` or equivalent enforcement.
- All status reporting goes through the Go backend's internal HTTP API (`/internal/analytics/report`). The analytics service does not write to the database directly and does not log to stdout in production (Go captures and logs it).
- Anomaly detection uses networkx for graph analysis of the transfer network. Build a directed graph where nodes are player accounts and edges are transactions. Flag: circular paths, isolated clusters, outlier degree nodes.
- If a computation job fails, substitute the previous day's value with `stale: true` in the snapshot and continue. Never abort the entire snapshot for a single job failure.
- The poster retries three times with exponential backoff before posting a minimal error report to Go's alert endpoint.
- All function signatures have type hints. No bare `except` clauses — always catch specific exception types.

**[PASTE RELEVANT BUILD DOC SECTIONS HERE — Section 7: Analytics Service in full + relevant layer descriptions]**

---

---

## SECTION 11 — Infrastructure
### Use when building: Docker files, Caddy, NATS config, compose files, init.sql

---

You are building the infrastructure configuration for Simlife.

**Your job in this session:**
Write all infrastructure files. When this session is complete, running `docker compose up` in the `infra/` directory starts a fully connected local development environment with all five services running, the database initialized and migrated, and the Activity dev server accessible.

**Files you are building:**
- `infra/docker-compose.yml`
- `infra/docker-compose.dev.yml`
- `infra/caddy/Caddyfile`
- `infra/postgres/init.sql`
- `infra/nats/nats.conf`
- `bot/Dockerfile`
- `analytics/Dockerfile`
- `activity/Dockerfile`
- `.env.example`

**Key requirements:**
- PostgreSQL data volume is named and persists across container restarts.
- Redis has AOF persistence enabled — the cache survives restarts.
- NATS JetStream is enabled. Streams `settlement_events`, `market_events`, and `notifications` are defined in nats.conf.
- The Go bot container waits for PostgreSQL to be healthy before starting (use healthcheck + depends_on condition).
- The analytics container waits for the Go bot to be healthy (it depends on Go's internal API being available).
- The dev compose override uses `air` for Go live reload and Vite dev server for the Activity.
- The Caddyfile handles the Discord Activity proxy URL mappings. The Activity's domain and the Go API domain are both configured.
- `.env.example` documents every variable with an inline comment explaining what it is and what format it expects.

**[PASTE RELEVANT BUILD DOC SECTIONS HERE — Section 10: Infrastructure & DevOps in full]**

---

---

## SECTION 12 — Anti-Exploit System
### Use when building: ratelimit, velocity monitor, anomaly flags, admin commands

---

You are building the anti-exploit system for Simlife.

**Your job in this session:**
Build the complete anti-exploit layer and all admin command handlers. When this session is complete, rate limiting is enforced on every command via Redis, velocity monitoring flags suspicious transaction patterns in real time, and admin commands give the developer tools to inspect, freeze, confiscate from, and ban players.

**Files you are building:**
- `bot/internal/anti_exploit/ratelimit.go`
- `bot/internal/anti_exploit/velocity.go`
- `bot/internal/anti_exploit/flags.go`
- `bot/internal/commands/admin/reset.go`
- `bot/internal/commands/admin/grant.go`
- `bot/internal/commands/admin/inspect.go`

**Key requirements:**
- Rate limit state lives in Redis with TTL-based windows. Redis key format: `ratelimit:{playerID}:{commandName}`.
- Velocity flags are written to `anomaly_flags` via the database, not just logged — they must persist for the analytics service's daily graph analysis.
- Admin commands require a specific Discord role ID configured in the environment, not a hardcoded user ID.
- Every admin action writes to the `admin_audit_log` table via the ledger package's audit function (not a separate direct write).
- `/admin inspect @player` generates a composited summary image using the profile layout.
- Freeze flag is checked by the router middleware before dispatching any command — a frozen player receives a specific "account under review" embed on any command attempt.

**[PASTE RELEVANT BUILD DOC SECTIONS HERE — Section 14: Anti-Exploit System in full + anti_exploit file descriptions]**

---
