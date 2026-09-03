# SIMLIFE — Discord Persistent Economy & Life Simulation Bot
## Technical Build Documentation
### Architecture · File Structure · System Design · UI Specification

---

> This document is the authoritative technical reference for building, understanding, and publishing Simlife. Every file listed here is described in terms of what it does in the fully operational bot — not as a placeholder or future consideration. Read this before writing a single line of code.

---

## TABLE OF CONTENTS

1. [System Overview](#1-system-overview)
2. [Technology Stack](#2-technology-stack)
3. [Architecture Layers](#3-architecture-layers)
4. [Full Project File Structure](#4-full-project-file-structure)
5. [Backend — Go Service](#5-backend--go-service)
6. [Discord Activity — React Frontend](#6-discord-activity--react-frontend)
7. [Analytics Service — Python](#7-analytics-service--python)
8. [Dynamic Image Compositing System](#8-dynamic-image-compositing-system)
9. [Database Design](#9-database-design)
10. [Infrastructure & DevOps](#10-infrastructure--devops)
11. [UI Design System](#11-ui-design-system)
12. [Inter-Service Communication](#12-inter-service-communication)
13. [Security Architecture](#13-security-architecture)
14. [Anti-Exploit System](#14-anti-exploit-system)

---

## 1. SYSTEM OVERVIEW

Simlife is a persistent, globally-shared economy and life simulation running inside Discord. Unlike conventional Discord economy bots — which simulate a per-server scoring system where one player's actions have zero effect on others — Simlife operates a single, shared economic world. Every transaction a player makes changes the economic conditions every other player operates in. Prices shift from supply and demand. Businesses produce real goods. Labor has market value. Wealth is a consequence of strategy, not grinding cooldown timers.

The bot is composed of three independent services that work together:

- **Go Backend** — The authoritative economic engine. Handles all player commands, financial transactions, market matching, settlement, the Discord Activity API, all application logging, and dynamic image compositing for Discord embeds. Nothing in the economy happens without passing through this service.
- **React Activity Frontend** — A Discord Embedded App (iframe) that renders the complex visual layer: the City View, Business Dashboard, Market Charts, and Property Manager. Players access it directly inside Discord without leaving the app.
- **Python Analytics Service** — A read-only background process that runs daily economic statistics: Gini coefficient, price index, velocity analysis, and anomaly detection. It posts structured results back to the Go backend. It never touches live player money and has no logging or infrastructure responsibility.

These three services share a PostgreSQL database, a Redis cache, and a NATS message bus. They are deployed independently but are architecturally coupled by clearly defined API contracts.

The world runs on a **continuous 24-hour economic day**. At the end of each day, Go's settlement engine closes the books: it processes all outstanding business transactions, pays wages, collects taxes, rebalances the market price index, and distributes event outcomes. The next day begins with an updated economic state visible to all players.

---

## 2. TECHNOLOGY STACK

| Layer | Technology | Purpose |
|---|---|---|
| Bot & API Server | Go (Golang) | Command handling, HTTP API, ledger engine, settlement, logging |
| Discord Library | DiscordGo (bwmarrin/discordgo) | Discord gateway, slash commands, interactions |
| Image Compositing | Go — gg / imaging libraries | Dynamic embed image generation |
| Activity Frontend | React + TypeScript | Discord Embedded App (iframe UI) |
| Activity SDK | @discord/embedded-app-sdk | OAuth, RPC bridge, URL mapping, IAP |
| Analytics Service | Python 3.11+ | Daily economic statistics and anomaly detection only |
| Analytics Libraries | pandas, numpy, networkx | Statistical computation, graph analysis |
| Primary Database | PostgreSQL 15+ | Ledger, player state, world state, market orders |
| Query Layer | sqlc | Type-safe SQL → Go struct generation |
| Cache | Redis 7+ | Rate limiting, session tokens, hot player data |
| Message Bus | NATS JetStream | Async event delivery between services |
| Migrations | golang-migrate | Versioned schema changes |
| Python DB Access | SQLAlchemy (read-only) | Analytics database queries |
| Containerization | Docker + Docker Compose | Local development and deployment |
| Reverse Proxy | Caddy | TLS termination, URL routing |
| Structured Logging | Go — zerolog | JSON structured logs, all services reported through Go |

---

## 3. ARCHITECTURE LAYERS

The bot's feature set is divided into twelve progressive layers. Each layer is a fully working product on its own. A later layer always adds to what came before — it never replaces or breaks it. This section describes what each layer adds to the running system in terms of live functionality.

### Layer 1 — Foundation Economy
The first working version of the bot. Players can earn currency by working, hold it in a wallet and bank, spend it in a basic shop, and check their balance. Every command response arrives as a styled Discord embed with a dynamically composited image — the balance command shows the player's name, current wallet and bank figures, and net worth stamped onto a pre-designed city background. A global economic monitor tracks total money in circulation, velocity, and Gini coefficient. The settlement engine runs its first daily close. No player can interact financially with any other player yet — this layer establishes trust in the core financial engine before complexity is added.

### Layer 2 — Player-to-Player Economy
Players can send money directly to each other, list items for sale on a global peer-to-peer marketplace, and bid on active listings. Supply and demand mechanics activate: items that many players sell become cheaper; items that few players sell become expensive. The price index begins to have meaning. The settlement engine begins publishing a daily economic news bulletin to a configured Discord channel, accompanied by a composited economic report image showing the day's key figures.

### Layer 3 — City View (Read-Only Activity Preview)
The Discord Activity infrastructure is deployed and tested for the first time. The Activity iframe opens with a rendered read-only city: players can see the global city, other players marked as residents, and basic economic data as charts. No actions are available through the Activity yet — this layer exists to validate the full Activity stack (auth, proxy, URL mapping, mobile compatibility) before Layer 4 depends on it being correct.

### Layer 4 — Business Engine
Players can open, name, and operate businesses. A business has a type (Bakery, Forge, Clinic, etc.), a physical location in the city, an inventory of input materials, and an output product. Hiring other players as workers creates employer-employee relationships. Wages are set by the business owner and paid by the settlement engine at the end of each economic day. The Activity frontend becomes interactive: business owners manage their operation through the Business Dashboard panel. The business status command generates a composited image showing the business name, daily production figures, worker count, and treasury balance overlaid on a business-category background.

### Layer 5 — Property & Real Estate
Land plots in the city become purchasable. Properties have addresses, zoning (residential, commercial, industrial), and upgrade paths. Property taxes are assessed daily and collected by the settlement engine. Rent income flows from tenants to landlords automatically. The city map in the Activity now shows owned plots, their owners, and their development level. A property transfer market opens — players can list and purchase real estate.

### Layer 6 — Labor Market
Work is no longer a simple cooldown command. Players have skill ratings across multiple professions that improve through practice. Businesses post job listings with skill requirements and offered wages. Job applications, hiring decisions, and termination all generate settlement effects. Unions can be formed by groups of workers to collectively bargain. Strikes affect business output. The labor market has its own supply curve tracked by the analytics service.

### Layer 7 — Government & Civic Systems
Players vote to elect officials to city government roles. Elected officials can set municipal tax rates, propose public spending projects, and allocate the city treasury. Public infrastructure projects (roads, parks, utilities) are funded by taxes and physically appear in the city map. Political campaigns require funding. Corruption mechanics allow officials to embezzle — detectable by the analytics service's anomaly monitor.

### Layer 8 — Banking & Financial System
A player-owned central bank appears. Players can apply for business loans with interest rates. The bank sets its own lending rate based on economic conditions. Credit scores are tracked based on repayment history. Bond markets allow the government to raise capital. Players can invest in other players' businesses through equity stakes. Dividends are distributed at settlement.

### Layer 9 — Crime & Black Market
A parallel economy operates beneath the main market. Players can choose criminal career paths: fence stolen goods, run protection rackets against businesses, launder money through shell companies, or operate contraband supply chains. Law enforcement NPCs (funded by the city budget) patrol and arrest. Criminal records affect legitimate economic opportunities. Investigators (a player role) can expose criminal networks and earn bounties.

### Layer 10 — International Trade
The global world is divided into regional economic zones. Each zone has comparative advantages in different goods. Trade routes open between zones, with transport costs and tariff mechanics. Players can establish import/export businesses and speculate on cross-zone price differentials. Embargoes can be voted into place by governments. Trade wars are possible. Economic sanctions affect entire regions.

### Layer 11 — Cultural & Social Economy
Reputation is a tradeable asset. Players build social capital through endorsements, reviews, and public standing. Artists and content creators can produce virtual cultural goods (music, artwork, writing) that sell in a cultural marketplace. Fame generates passive income. Public events (concerts, exhibitions, tournaments) can be organized by players and attract audiences whose entry fees flow to organizers.

### Layer 12 — Legacy & Succession
Generational mechanics activate. Long-running players can retire their characters, passing assets to heirs. Corporate dynasties persist across player generations. Historical records are written into the world's permanent ledger — events, recessions, booms, and notable players become part of the city's documented history. The game has a past that shapes its present.

---

## 4. FULL PROJECT FILE STRUCTURE

```
simlife/
├── bot/                                  # Go backend service
│   ├── cmd/
│   │   └── simlife/
│   │       └── main.go
│   ├── internal/
│   │   ├── bot/
│   │   │   ├── bot.go
│   │   │   ├── middleware.go
│   │   │   └── router.go
│   │   ├── commands/
│   │   │   ├── registry.go
│   │   │   ├── economy/
│   │   │   │   ├── balance.go
│   │   │   │   ├── work.go
│   │   │   │   ├── bank.go
│   │   │   │   ├── pay.go
│   │   │   │   └── shop.go
│   │   │   ├── business/
│   │   │   │   ├── open.go
│   │   │   │   ├── hire.go
│   │   │   │   ├── fire.go
│   │   │   │   ├── produce.go
│   │   │   │   └── status.go
│   │   │   ├── market/
│   │   │   │   ├── list.go
│   │   │   │   ├── buy.go
│   │   │   │   ├── bid.go
│   │   │   │   └── history.go
│   │   │   ├── property/
│   │   │   │   ├── buy.go
│   │   │   │   ├── sell.go
│   │   │   │   ├── develop.go
│   │   │   │   └── rent.go
│   │   │   ├── government/
│   │   │   │   ├── vote.go
│   │   │   │   ├── propose.go
│   │   │   │   └── treasury.go
│   │   │   └── admin/
│   │   │       ├── reset.go
│   │   │       ├── grant.go
│   │   │       └── inspect.go
│   │   ├── economy/
│   │   │   ├── ledger.go
│   │   │   ├── settlement.go
│   │   │   ├── market.go
│   │   │   ├── pricing.go
│   │   │   ├── tax.go
│   │   │   └── monitor.go
│   │   ├── business/
│   │   │   ├── engine.go
│   │   │   ├── production.go
│   │   │   ├── wages.go
│   │   │   └── types.go
│   │   ├── world/
│   │   │   ├── city.go
│   │   │   ├── plot.go
│   │   │   ├── zone.go
│   │   │   └── events.go
│   │   ├── player/
│   │   │   ├── player.go
│   │   │   ├── skills.go
│   │   │   ├── reputation.go
│   │   │   └── history.go
│   │   ├── imaging/
│   │   │   ├── composer.go
│   │   │   ├── renderer.go
│   │   │   ├── layouts/
│   │   │   │   ├── balance.go
│   │   │   │   ├── business.go
│   │   │   │   ├── market.go
│   │   │   │   ├── property.go
│   │   │   │   ├── profile.go
│   │   │   │   ├── economic_news.go
│   │   │   │   └── shop.go
│   │   │   └── assets/
│   │   │       ├── backgrounds/
│   │   │       │   ├── balance_bg.png
│   │   │       │   ├── business_bg.png
│   │   │       │   ├── market_bg.png
│   │   │       │   ├── property_bg.png
│   │   │       │   ├── profile_bg.png
│   │   │       │   ├── economic_news_bg.png
│   │   │       │   └── shop_bg.png
│   │   │       └── fonts/
│   │   │           ├── primary.ttf
│   │   │           └── mono.ttf
│   │   ├── api/
│   │   │   ├── server.go
│   │   │   ├── middleware.go
│   │   │   ├── routes.go
│   │   │   ├── handlers/
│   │   │   │   ├── activity.go
│   │   │   │   ├── player.go
│   │   │   │   ├── market.go
│   │   │   │   ├── city.go
│   │   │   │   ├── business.go
│   │   │   │   └── analytics.go
│   │   │   └── auth/
│   │   │       ├── discord.go
│   │   │       └── jwt.go
│   │   ├── jobs/
│   │   │   ├── scheduler.go
│   │   │   ├── daily_settlement.go
│   │   │   ├── economic_news.go
│   │   │   ├── tax_collection.go
│   │   │   └── wage_distribution.go
│   │   ├── events/
│   │   │   ├── publisher.go
│   │   │   └── subscriber.go
│   │   ├── cache/
│   │   │   ├── redis.go
│   │   │   └── keys.go
│   │   ├── logger/
│   │   │   ├── logger.go
│   │   │   └── middleware.go
│   │   ├── anti_exploit/
│   │   │   ├── ratelimit.go
│   │   │   ├── velocity.go
│   │   │   └── flags.go
│   │   └── config/
│   │       └── config.go
│   ├── db/
│   │   ├── migrations/
│   │   │   ├── 001_init_players.sql
│   │   │   ├── 002_init_ledger.sql
│   │   │   ├── 003_init_market.sql
│   │   │   ├── 004_init_business.sql
│   │   │   ├── 005_init_world.sql
│   │   │   ├── 006_init_property.sql
│   │   │   ├── 007_init_government.sql
│   │   │   └── 008_init_analytics_snapshots.sql
│   │   ├── queries/
│   │   │   ├── players.sql
│   │   │   ├── ledger.sql
│   │   │   ├── market.sql
│   │   │   ├── business.sql
│   │   │   ├── property.sql
│   │   │   └── world.sql
│   │   └── sqlc.yaml
│   ├── go.mod
│   ├── go.sum
│   └── Dockerfile
│
├── activity/                             # React Discord Activity frontend
│   ├── public/
│   │   ├── index.html
│   │   └── assets/
│   │       ├── fonts/
│   │       └── icons/
│   ├── src/
│   │   ├── main.tsx
│   │   ├── App.tsx
│   │   ├── discord.ts
│   │   ├── api/
│   │   │   ├── client.ts
│   │   │   ├── endpoints.ts
│   │   │   └── types.ts
│   │   ├── auth/
│   │   │   ├── DiscordAuth.tsx
│   │   │   └── useAuth.ts
│   │   ├── panels/
│   │   │   ├── CityView/
│   │   │   │   ├── CityView.tsx
│   │   │   │   ├── CityMap.tsx
│   │   │   │   ├── CityLayer.tsx
│   │   │   │   ├── PlayerMarker.tsx
│   │   │   │   └── CityView.module.css
│   │   │   ├── BusinessDashboard/
│   │   │   │   ├── BusinessDashboard.tsx
│   │   │   │   ├── ProductionQueue.tsx
│   │   │   │   ├── WorkerRoster.tsx
│   │   │   │   ├── FinancialSummary.tsx
│   │   │   │   └── BusinessDashboard.module.css
│   │   │   ├── MarketView/
│   │   │   │   ├── MarketView.tsx
│   │   │   │   ├── PriceChart.tsx
│   │   │   │   ├── OrderBook.tsx
│   │   │   │   ├── TradeHistory.tsx
│   │   │   │   └── MarketView.module.css
│   │   │   ├── PropertyManager/
│   │   │   │   ├── PropertyManager.tsx
│   │   │   │   ├── PlotCard.tsx
│   │   │   │   ├── DevelopmentPanel.tsx
│   │   │   │   └── PropertyManager.module.css
│   │   │   ├── EconomicReport/
│   │   │   │   ├── EconomicReport.tsx
│   │   │   │   ├── WealthChart.tsx
│   │   │   │   ├── VelocityChart.tsx
│   │   │   │   └── EconomicReport.module.css
│   │   │   └── PlayerProfile/
│   │   │       ├── PlayerProfile.tsx
│   │   │       ├── SkillsPanel.tsx
│   │   │       ├── ReputationBadges.tsx
│   │   │       └── PlayerProfile.module.css
│   │   ├── components/
│   │   │   ├── ui/
│   │   │   │   ├── Button.tsx
│   │   │   │   ├── Card.tsx
│   │   │   │   ├── Modal.tsx
│   │   │   │   ├── Tooltip.tsx
│   │   │   │   ├── Badge.tsx
│   │   │   │   ├── Spinner.tsx
│   │   │   │   ├── Toast.tsx
│   │   │   │   └── DataTable.tsx
│   │   │   ├── charts/
│   │   │   │   ├── LineChart.tsx
│   │   │   │   ├── BarChart.tsx
│   │   │   │   └── HeatMap.tsx
│   │   │   └── layout/
│   │   │       ├── Sidebar.tsx
│   │   │       ├── TopBar.tsx
│   │   │       └── PanelContainer.tsx
│   │   ├── hooks/
│   │   │   ├── useWebSocket.ts
│   │   │   ├── useMarketData.ts
│   │   │   ├── usePlayerData.ts
│   │   │   ├── useBusinessData.ts
│   │   │   └── useCityData.ts
│   │   ├── store/
│   │   │   ├── index.ts
│   │   │   ├── playerSlice.ts
│   │   │   ├── marketSlice.ts
│   │   │   ├── citySlice.ts
│   │   │   └── businessSlice.ts
│   │   ├── styles/
│   │   │   ├── globals.css
│   │   │   ├── tokens.css
│   │   │   └── animations.css
│   │   └── utils/
│   │       ├── format.ts
│   │       ├── currency.ts
│   │       └── time.ts
│   ├── package.json
│   ├── tsconfig.json
│   ├── vite.config.ts
│   └── Dockerfile
│
├── analytics/                            # Python analytics service
│   ├── main.py
│   ├── scheduler.py
│   ├── db/
│   │   ├── connection.py
│   │   └── queries.py
│   ├── jobs/
│   │   ├── economic_snapshot.py
│   │   ├── gini.py
│   │   ├── price_index.py
│   │   ├── velocity.py
│   │   ├── anomaly_detection.py
│   │   └── trend_analysis.py
│   ├── api/
│   │   ├── client.py
│   │   └── poster.py
│   ├── models/
│   │   ├── snapshot.py
│   │   └── alert.py
│   ├── requirements.txt
│   ├── pyproject.toml
│   └── Dockerfile
│
├── infra/
│   ├── docker-compose.yml
│   ├── docker-compose.dev.yml
│   ├── caddy/
│   │   └── Caddyfile
│   ├── postgres/
│   │   └── init.sql
│   └── nats/
│       └── nats.conf
│
├── .env.example
├── .gitignore
└── README.md
```

---

## 5. BACKEND — GO SERVICE

The Go backend is the central authority of Simlife. It is the only service that reads from and writes to the live financial ledger. All other services (the Activity frontend, the analytics service, Discord itself) interface with the world through this service's APIs. It is also responsible for all application logging and for generating every dynamic image attached to Discord embeds. It runs as a single binary started from `cmd/simlife/main.go` but is internally divided into clearly separated packages.

---

### `cmd/simlife/main.go`

The program entry point. When the binary starts, it performs the following in sequence: initializes the structured logger, loads environment configuration, initializes the database connection pool (PostgreSQL via pgx), initializes Redis, initializes the NATS JetStream connection, runs all pending database migrations, pre-loads all image compositing background assets into memory, registers all slash commands with the Discord API, starts the HTTP API server on its configured port, starts the background job scheduler, and finally opens the Discord gateway connection. The program runs until it receives a shutdown signal, at which point it gracefully drains active connections, finishes in-flight requests, and exits cleanly. This file contains no business logic — its only job is to wire everything together in the correct order.

---

### `internal/config/config.go`

Reads all runtime configuration from environment variables and validates that required values are present. Exposes a single `Config` struct used throughout the application. Configuration values include: Discord bot token, Discord application ID, Discord public key (for webhook verification), PostgreSQL DSN, Redis address, NATS URL, JWT signing secret, settlement cron schedule, Activity client ID, log level, log format (JSON for production, pretty-print for development), and feature flags for enabling or disabling specific layers. This file is the single source of truth for how the bot behaves in a given deployment environment.

---

### `internal/logger/logger.go`

Initializes and exposes the application-wide structured logger using zerolog. The logger outputs JSON in production (one JSON object per line, machine-parseable by log aggregation tools) and human-readable formatted output in development. Every log entry includes: timestamp, log level, the service name ("simlife-bot"), the Go package that emitted the log, and any structured fields attached by the caller. Log levels follow standard severity: DEBUG (development diagnostics), INFO (normal operational events like settlement start/complete, player registrations), WARN (recoverable issues like cache misses or retry attempts), ERROR (failures that affected a player action), and FATAL (unrecoverable errors that require process exit). This package is imported by every other package in the bot — it is the single logging interface for the entire Go service.

---

### `internal/logger/middleware.go`

HTTP middleware that logs every incoming API request. Captures: method, path, status code, response time in milliseconds, player ID (from JWT if authenticated), and client IP. Logs at INFO level for successful requests and ERROR level for 4xx/5xx responses. Also attaches a unique request ID to each request context, which downstream handlers can include in their own log entries so that all log lines for a single request can be correlated in a log viewer. This middleware wraps every route registered on the API server.

---

### `internal/bot/bot.go`

Creates and manages the DiscordGo session. Registers the interaction handler and delegates each interaction to the command router. Manages the bot's status presence message, which displays the current economic health indicator — updated each time the settlement engine closes the day. The bot session maintains automatic reconnection with exponential backoff. All session lifecycle events (connect, disconnect, reconnect, resume) are logged by the logger package with structured fields.

---

### `internal/bot/router.go`

Receives every incoming Discord interaction and dispatches it to the correct command handler. Before dispatching, it: confirms the player exists (creating a new record if not), applies Redis rate limiting, checks whether the command's layer is enabled, and injects the player struct into the request context. Every dispatch and rejection is logged with the player ID, command name, and outcome.

---

### `internal/bot/middleware.go`

Defines the middleware chain applied to every command handler: verify Discord signature, authenticate the player, apply rate limiting, check feature availability, inject the player struct. Additional middleware can be applied to specific commands — admin commands run an additional authorization check.

---

### `internal/commands/registry.go`

Maintains the complete list of slash command definitions. At startup, the bot syncs this registry with Discord's API. The registry maps command names to their handler functions. Adding a new command requires adding its definition here and writing its handler. The registry contains no logic.

---

### `internal/commands/economy/balance.go`

Handles `/balance`. Reads the player's wallet, bank, escrow, and net worth from the database. Calls the imaging package to composite the player's username, balances, 24-hour change, and rank percentile onto the balance background image. Attaches the generated PNG to a Discord embed and sends it as an ephemeral response. The embed's color changes based on 24-hour trend: gold for positive change, muted for flat, red-tinted for loss.

---

### `internal/commands/economy/work.go`

Handles `/work`. Checks whether the player is employed by a player-owned business — if so, credits labor to the employer's production queue and accrues wages for settlement. If unemployed, applies the base city wage rate. Enforces a cooldown via Redis. Returns a styled embed with a composited image showing hours worked, wage accrued, and current employment status stamped onto the work background.

---

### `internal/commands/economy/bank.go`

Handles `/deposit` and `/withdraw`. Validates funds in the source account, executes a double-entry ledger transaction, and confirms updated balances. Both operations are atomic. Returns a composited bank confirmation image showing the transfer amount and new balances on the balance background with a bank-specific header overlay.

---

### `internal/commands/economy/pay.go`

Handles `/pay`. Validates recipient existence and sender funds, executes the atomic double-entry transfer, and notifies both parties. The sender receives a composited confirmation embed. The recipient receives a notification embed showing the amount received, sender name, and their new total balance, composited onto the balance background with an "Incoming Transfer" header.

---

### `internal/commands/economy/shop.go`

Handles `/shop` and `/buy` for global shop items. Reads current prices from the pricing engine (market-driven, not static). Returns the shop listing as a composited image showing the available items, their prices, and a "Price changes daily" indicator on the shop background. Purchase confirmations return a composited receipt image.

---

### `internal/commands/market/list.go`

Handles `/market list`. Validates ownership, moves items to escrow, creates the listing record, and returns a composited listing confirmation image showing the item, quantity, asking price, and listing expiry on the market background.

---

### `internal/commands/market/buy.go`

Handles `/market buy`. Executes the trade atomically, updates price history, notifies both parties. Both the buyer and seller receive composited trade confirmation images showing the item, quantity, final price, and their updated balances.

---

### `internal/commands/market/history.go`

Handles `/market history`. Returns a composited price history image for a specified item — a rendered mini chart of the last 30 days of trade prices stamped onto the market background, giving players a visual price trend without opening the full Activity.

---

### `internal/economy/ledger.go`

The most critical file in the codebase. Every financial transaction flows through the ledger's `PostTransaction` function regardless of what triggered it. Implements double-entry bookkeeping: every transaction consists of paired debit and credit entries, and the sum of all entries always equals zero. Uses PostgreSQL `SERIALIZABLE` isolation for all financial writes. No code outside this package may directly modify balance columns. Every transaction is logged at INFO level with structured fields: transaction UUID, type, sender account, recipient account, amount, and duration.

---

### `internal/economy/settlement.go`

The daily settlement engine. Runs at a configured time each day. Phases in order: close expired market listings, process business production cycles, pay worker wages, collect property taxes, distribute dividends, compute the economic monitor snapshot, post the snapshot to `economic_snapshots`, notify the analytics service via NATS, and publish the economic news embed with a composited daily report image to configured Discord channels. Wrapped with a hard timeout context — if it exceeds 10 minutes it cancels, rolls back what it can, and logs a FATAL event with the failure state.

---

### `internal/economy/market.go`

Manages all active market listings. Provides functions for creating listings, matching buyers, executing trades, processing auctions, and expiring stale listings. Records every completed trade price and quantity in `market_trades` for the pricing engine.

---

### `internal/economy/pricing.go`

Computes market prices for all items. Maintains a 7-day weighted moving average of completed trade prices, a supply signal (active listings count), a demand signal (buy requests in the last 24 hours), and a floor price (cost of production inputs). Publishes computed prices to Redis after each settlement.

---

### `internal/economy/tax.go`

Implements all tax mechanics. Layer 1: transaction tax on player-to-player transfers flows to city treasury. Layer 5: property tax added. Layer 7: tax rates become configurable by elected officials within system-defined bounds. All tax collections recorded as ledger transactions with type `TAX_COLLECTION`.

---

### `internal/economy/monitor.go`

Produces the economic health snapshot after settlement. Computes: total money supply, transaction velocity, top 10 wealthiest players, and an inequality ratio. Stores results in `economic_snapshots`. Also sets the base wage rate for unaffiliated workers based on the current velocity — a slow economy slightly lowers the base wage, naturally incentivizing players to seek employment.

---

### `internal/business/engine.go`

Coordinates all business operations: opening, upgrading, and closing businesses. Validates business type availability, checks capital requirements, creates business records, charges opening fees via the ledger.

---

### `internal/business/production.go`

Manages the production cycle during settlement. For each active business, checks input inventory and logged worker labor hours. If both are sufficient, consumes inputs and adds outputs to business inventory. Logs success or failure with the reason.

---

### `internal/business/wages.go`

Calculates and pays worker wages during settlement. Reads each employed worker's wage rate and hours logged, validates business account funds, posts the payment to the ledger. If a business cannot pay, sets a debt flag and notifies workers of the pending shortfall.

---

### `internal/world/city.go`

Manages global city state: plots, zone types, ownership, development levels, business locations, and infrastructure projects. Provides city state queries for the Activity frontend. City state is cached in Redis and invalidated on plot changes.

---

### `internal/player/player.go`

Manages player records. Creates new player accounts on first interaction (Discord user ID as the permanent key). Provides player lookup functions used throughout the application. Tracks last active timestamp for online presence calculations.

---

### `internal/player/skills.go`

Manages player skill ratings across professions (activated at Layer 6). Skills improve through practice — each relevant command interaction increments the corresponding skill's experience counter. When experience crosses a threshold, the skill level increases and the player is notified. Skill levels affect wage eligibility and business hire requirements.

---

### `internal/api/server.go`

Starts the HTTP API server for the Discord Activity frontend. Handles CORS for cross-origin Activity requests. Every route is authenticated by JWT middleware before reaching its handler.

---

### `internal/api/auth/discord.go`

Implements the Discord OAuth token exchange for the Activity. Receives an OAuth code from the Activity, exchanges it with Discord's API for an access token, retrieves the player's Discord user ID, and returns a signed JWT to the Activity.

---

### `internal/api/auth/jwt.go`

Issues and validates JWTs for Activity sessions. Tokens are signed with HS256, contain player Discord ID, database player ID, and a 1-hour expiry. A refresh endpoint issues new tokens before expiry. Invalid or expired tokens return 401 with no data.

---

### `internal/api/handlers/activity.go`

Entry point for the Activity's OAuth exchange. Receives the Discord OAuth code, exchanges it, finds or creates the player, and returns the JWT plus the player's current state in a single response so the Activity renders immediately on open.

---

### `internal/api/handlers/city.go`

Returns city state for the Activity's City View. Includes all plots, zone types, owners, development levels, active businesses, infrastructure projects, and online player positions. Serves cached Redis data; only queries PostgreSQL on cache miss or invalidation.

---

### `internal/api/handlers/market.go`

Returns live market data for the Activity's Market View. Provides active listings, recent trade history, current reference prices, and the player's own listings. Also accepts POST requests for market actions submitted through the Activity UI — these go through the same ledger validation as bot commands.

---

### `internal/jobs/scheduler.go`

Manages the background job system with cron-style scheduling. Registered jobs: daily settlement, economic news publishing, stale listing expiry (hourly), Redis cache warm-up (every 15 minutes), and analytics service health check (every 25 hours). Gracefully shuts down on process signal.

---

### `internal/anti_exploit/ratelimit.go`

Enforces per-player command rate limits using Redis counters with TTL windows. Cooldowns survive bot restarts. Rejections are logged with the player ID and command name.

---

### `internal/anti_exploit/velocity.go`

Monitors transaction patterns for multi-accounting, transfer rings, and automation signals. Raises flags stored in `anomaly_flags` for the analytics service's daily graph analysis.

---

### `db/migrations/`

Numbered SQL migration files managed by golang-migrate. Always additive — no destructive down migrations in production. Migrations run automatically at startup and are idempotent.

---

### `db/queries/`

Raw SQL query files compiled into type-safe Go functions by sqlc. Each file covers one domain. Developers write SQL; sqlc generates Go. No ORM abstraction obscures what runs against the financial data.

---

## 6. DISCORD ACTIVITY — REACT FRONTEND

The Activity is a React TypeScript application running inside a Discord iframe. Players open it via the Activities button or a `/city` command invite link. It communicates with the Go backend via authenticated HTTPS and WebSocket. All state is stored in the Go backend — the Activity is a rendering layer only.

---

### `src/main.tsx`

React bootstrap. Initializes the Redux store, wraps the app in React Query, mounts the root App component, and initializes the Discord Embedded App SDK, making the SDK instance available through a React context.

---

### `src/App.tsx`

Root component managing the top-level state machine: Loading → Authenticated → Error. Once authenticated, renders the main layout: top navigation bar, sidebar panel selector, and active panel content area. Active panel tracked in URL hash state.

---

### `src/discord.ts`

Initializes the DiscordSDK instance. Calls `sdk.ready()` to signal Discord that the Activity has loaded. Configures `patchUrlMappings` for all API domains. This is the single file where URL mappings are maintained.

---

### `src/auth/DiscordAuth.tsx`

Manages the OAuth flow on mount. Gets the OAuth code from the Discord SDK, POSTs it to the Go backend, stores the JWT in memory (never localStorage), and renders children once authenticated. Handles silent token refresh 5 minutes before expiry.

---

### `src/api/client.ts`

Typed HTTP client for all Go backend communication. Automatically attaches JWTs to every request. Handles 401 by triggering refresh before one retry. If retry fails, shows reconnect prompt. Single source of truth for the API base URL.

---

### `src/hooks/useWebSocket.ts`

Manages the persistent WebSocket connection to the Go backend's real-time events endpoint. Authenticates with the player JWT. Dispatches Redux actions on server-pushed events. Reconnects with exponential backoff after disconnection. Handles Discord PiP transition disconnects correctly on mobile.

---

### `src/panels/CityView/CityView.tsx`

Main City View panel. Renders a zoomable, pannable city grid populated with plot data. Each plot shows zone color, development iconography, and owner username on hover. Active businesses display type icons. Online players appear as animated markers.

---

### `src/panels/CityView/CityMap.tsx`

Canvas rendering engine for the city grid. Manages zoom (scroll/pinch), pan (drag/touch), and viewport culling (only visible tiles rendered). Updates only affected tiles on WebSocket events — not a full canvas redraw. Designed to handle a city grid large enough to feel like a real city without DOM-node performance issues.

---

### `src/panels/BusinessDashboard/BusinessDashboard.tsx`

Business management interface for business owners. Shows: production queue (inputs loaded, output scheduled, worker hours logged vs required), worker roster (names, hours, accrued wages), daily financial summary (revenue, wage costs, material costs, net profit projection), and an alert panel for settlement warnings such as low input stock.

---

### `src/panels/MarketView/MarketView.tsx`

Global marketplace interface. Displays active listings grouped by item category with reference prices. Player can switch between Listings, My Listings, Trade History, and Price Chart tabs.

---

### `src/panels/MarketView/PriceChart.tsx`

Candlestick chart built on Recharts displaying price history for a selected item. Each candle represents one economic day (open, close, high, low). Supports time range zoom (7 days, 30 days, all time). Includes a moving average overlay.

---

### `src/panels/EconomicReport/EconomicReport.tsx`

Global economic health dashboard. Displays the latest snapshot: total money supply, transaction velocity dial, wealth inequality index, top earners, and a 30-day comparison chart. Accessible to all players. Economic transparency is a design principle — players who understand the economy make better decisions, and better decisions make the economy more interesting for everyone.

---

### `src/store/`

Redux Toolkit store split into domain slices. `playerSlice` holds balance, inventory, employment status, and notifications. `marketSlice` holds listings, prices, and the player's active offers. `citySlice` holds the city grid and online player positions. `businessSlice` holds business state for owners. Updated in real time by WebSocket events. Never persisted to localStorage.

---

### `src/styles/tokens.css`

CSS custom property definitions for the design system. Defines the complete color palette, typographic scale, spacing scale, border radius, shadow tokens, and transition timings. Every component references these tokens — the entire visual theme can be changed by editing this file.

---

## 7. ANALYTICS SERVICE — PYTHON

The Python analytics service is a separate container with read-only PostgreSQL access. Its only write path is an internal HTTP POST to the Go backend's analytics endpoint. It has no logging infrastructure of its own — it reports events to the Go backend via its API, and Go logs them. Its scope is purely statistical computation: it runs once per economic day, triggered by the settlement event, and produces a structured snapshot of economic health metrics.

---

### `main.py`

Entry point. Starts the APScheduler instance, subscribes to the NATS `settlement.complete` event to trigger the computation pipeline, and starts a minimal HTTP server that accepts health check pings from the Go backend. Logs startup status by calling the Go health-check endpoint on boot, so Go can record that the analytics service came online.

---

### `scheduler.py`

Configures the computation pipeline triggered by the NATS `settlement.complete` event. Runs the pipeline jobs in sequence. If the settlement event is not received within 25 hours, posts a stale-data alert to the Go backend's internal alert endpoint so Go can log it and notify the developer.

---

### `jobs/economic_snapshot.py`

The main daily computation orchestrator. Calls each job module in sequence, assembles results into a snapshot object, and calls `api/poster.py` to submit it to Go. If any individual computation fails, substitutes the previous day's value with a staleness flag and continues. A partial snapshot is better than none.

---

### `jobs/gini.py`

Computes the Gini coefficient across all active player balances. Reads wallet plus bank balances, sorts them, and applies the standard piecewise formula. The result is a wealth inequality measure from 0 (perfect equality) to 1 (one player owns everything), published in the snapshot and displayed in the Activity's Economic Report.

---

### `jobs/price_index.py`

Computes a basket-of-goods price index. Selects a fixed basket of commonly traded goods across major categories, computes their weighted average prices relative to launch-day baselines, and produces an index number calibrated to 100 at launch. Published in the daily economic news embed.

---

### `jobs/velocity.py`

Computes transaction velocity: total transaction value in the past 24 hours divided by total money supply. A high velocity indicates a healthy, active economy. A low velocity indicates players are hoarding. The result feeds the economic monitor's base wage calculation via the snapshot.

---

### `jobs/anomaly_detection.py`

Scans transaction records from the past 24 hours using graph analysis (networkx) to identify: circular transfer patterns (multi-account laundering), new-account large-transfer events (multi-accounting signals), velocity outliers (automation signals), and transfer graph clustering (farming rings). Detected anomalies are posted to Go's internal alert endpoint, which writes them to `anomaly_flags` and logs them.

---

### `jobs/trend_analysis.py`

Produces 7-day and 30-day trend data for Activity charts. Computes percentage change and moving averages for each tracked metric. Identifies trending market items (trade volume up more than 50% vs the prior 7-day period). Trending items are highlighted in the Activity's Market View.

---

### `api/poster.py`

Posts the completed snapshot to the Go backend's internal analytics endpoint via authenticated HTTP POST (shared secret, not player JWT). Retries three times with exponential backoff on failure. On final failure, posts a minimal error report to Go's alert endpoint so Go can log and notify.

---

## 8. DYNAMIC IMAGE COMPOSITING SYSTEM

Every Discord command response in Simlife includes a dynamically composited PNG image attached to the embed. This replaces the static text-and-field embeds used by conventional bots. Each image is generated in milliseconds server-side by the Go backend using a pre-loaded background asset and a layout-specific rendering function. The result is a visually designed response that looks consistent, branded, and rich — never auto-generated.

---

### How It Works

When a command handler prepares its response, it calls the imaging package with two things: the layout name (which determines which background and which data fields are rendered), and a data struct containing the values to display (player name, balances, business stats, etc.). The imaging package selects the correct background PNG from memory, opens a canvas over it, renders the data fields at their configured positions using the specified fonts and colors, and returns the final composited PNG as a byte slice. The command handler attaches this byte slice to the Discord embed as a file upload. Discord displays it inline in the embed.

All background images are loaded into memory at bot startup by `main.go`. There is no disk read at request time — compositing is a pure in-memory operation. Response times for image generation are typically under 30 milliseconds.

---

### `internal/imaging/composer.go`

The core compositing engine. Exposes a single `Compose(layout string, data interface{}) ([]byte, error)` function. Loads the correct background from the pre-loaded asset cache, instantiates a drawing context using the `gg` library, delegates to the layout-specific renderer, and encodes the final canvas to PNG bytes. Handles errors gracefully — if compositing fails for any reason, the command handler falls back to a plain text embed so the player still receives a response.

---

### `internal/imaging/renderer.go`

Shared rendering primitives used by all layout files: `DrawText(x, y float64, text string, font Font, color Color)`, `DrawCurrencyValue(x, y float64, amount int64, color Color)` (renders the value in the game's currency format with appropriate color for positive/negative), `DrawProgressBar(x, y, width, height, percent float64, color Color)`, `DrawMiniChart(x, y float64, values []float64, color Color)` (renders a simple sparkline from an array of historical values), and `DrawAvatar(x, y float64, imageURL string)` (fetches and composites the player's Discord avatar into the image). These primitives are the building blocks that all layout files use — they are not called directly by command handlers.

---

### `internal/imaging/layouts/balance.go`

Defines the compositing layout for the `/balance`, `/deposit`, and `/withdraw` commands. The balance background is a wide-format city skyline at night with a dark gradient overlay for text legibility. The layout renders: the player's username in the top-left with their Discord avatar beside it, the wallet balance in large gold monospace type in the center-left, the bank balance beneath it in smaller type, a thin divider line, net worth in medium type, and a 24-hour change indicator (arrow icon + percentage, green or red) in the bottom-right corner. An economic day counter is rendered in small type at the bottom edge.

---

### `internal/imaging/layouts/business.go`

Defines the compositing layout for `/business status`. The business background is an industrial-style image with a factory or workshop aesthetic. The layout renders: the business name in large type at the top, the business type icon in the top-right corner, a production status bar showing today's progress (inputs loaded, hours logged vs required, production health percentage), a small worker count badge, today's projected revenue vs cost in a two-column layout, and a status indicator in the bottom-left (PRODUCING / STALLED / CLOSED) with color-coding.

---

### `internal/imaging/layouts/market.go`

Defines the compositing layout for market command responses. The market background is a trading floor or exchange hall aesthetic. For listing confirmations, the layout renders: the item name and icon, the quantity listed, the asking price per unit, and the listing expiry time. For trade confirmations, it renders: the item exchanged, the quantity, the final price, and both parties' updated balances. For the market history command, it renders the item name and a sparkline chart of the last 30 days of prices drawn using `DrawMiniChart`.

---

### `internal/imaging/layouts/economic_news.go`

Defines the compositing layout for the daily economic news bulletin. This is the most data-dense image in the system. The economic news background is a newspaper front page aesthetic. The layout renders: the economic day number as the headline, the price index with an up/down arrow versus yesterday, the velocity reading on a stylized gauge, the Gini coefficient as a horizontal bar, the top three earners of the day in a ranked list, and a bottom banner showing total money supply. This image is posted once per day to configured Discord channels by the settlement engine after the analytics snapshot is received.

---

### `internal/imaging/layouts/profile.go`

Defines the compositing layout for player profile commands. The profile background is a personal-record or character-sheet aesthetic. The layout renders: the player's Discord avatar (large, center-left), their username, the number of days they have been active, their net worth rank percentile, their top three skills with level indicators, and any active employment or business ownership badges.

---

### `internal/imaging/assets/backgrounds/`

This directory contains the pre-designed background PNG files. Each file is a 1200×630 pixel image (standard Discord embed image ratio) designed to work as a visual template. Backgrounds are designed with intentional dark regions where text will be composited — they are not full-bleed photographic images but purpose-built UI canvases. The files are: `balance_bg.png`, `business_bg.png`, `market_bg.png`, `property_bg.png`, `profile_bg.png`, `economic_news_bg.png`, and `shop_bg.png`. These files are created externally (designed by the developer) and committed to the repository. They are never generated by code.

---

### `internal/imaging/assets/fonts/`

Contains the two font files used across all composited images: `primary.ttf` (the main display font, used for all labels, names, and descriptive text) and `mono.ttf` (a monospace font used for all numerical values — balances, prices, quantities, percentages). Using a monospace font for numbers ensures they align vertically in multi-line layouts without pixel-level positioning adjustments. Both fonts are loaded into memory at startup and never read from disk at request time.

---

## 9. DATABASE DESIGN

All data lives in PostgreSQL. The database is the single source of truth for all game state. No data that matters is stored only in Redis (Redis is a cache) or only in memory (memory is ephemeral). Every economic action is durable the moment it is committed to the database.

---

### Core Tables

**`players`** — One row per player. Contains Discord user ID (unique, permanent identity key), username, display color (for city map marker), registration timestamp, last active timestamp, and a soft-delete flag. Economic state is in separate tables joined by player ID.

**`accounts`** — The double-entry ledger's account layer. Each player has exactly three accounts: wallet (liquid, spendable), bank (safe, inaccessible from commands), and escrow (holds market listing deposits). The city treasury is also an account. Account balances are computed by summing their transaction entries — not stored as columns. This prevents balance drift by construction.

**`transactions`** — Every financial event, ever. Each row is one side of a double-entry pair. Columns: transaction UUID, account ID, amount (always positive), entry type (DEBIT or CREDIT), transaction type (PLAYER_TRANSFER, WAGE, TAX, MARKET_SALE, etc.), reference ID, created timestamp. Append-only — no row is ever updated or deleted.

**`market_listings`** — Active and recently expired market listings. Contains: listing UUID, seller player ID, item type, quantity, quantity remaining, asking price, escrow reference, status (ACTIVE, SOLD, EXPIRED, CANCELLED), timestamps.

**`market_trades`** — Completed trade records. Each row: listing ID, buyer, seller, item type, quantity, price per unit, timestamp. Feeds the pricing engine and analytics trend jobs.

**`businesses`** — One row per business. Contains: UUID, owner player ID, business type, name, city plot ID, status (ACTIVE, CLOSED, SUSPENDED), opening timestamp, current inventory (JSONB), production configuration (JSONB — recipe, output item, daily capacity).

**`employment`** — Active and historical employment relationships. Employer business ID, employee player ID, wage rate, minimum daily hours, start timestamp, status. TERMINATED rows are preserved for wage dispute resolution and analytics.

**`daily_labor`** — Work hours logged per player per economic day. Upserted by the `/work` command, read by settlement for wage calculation, cleared after settlement.

**`properties`** — Land plots. Plot coordinates, zone type, owner player ID (nullable), development level, assessed value, last tax payment timestamp.

**`economic_snapshots`** — Daily economic health records. One row per settlement. Contains all monitor metrics and analytics snapshot data. Never deleted — permanent economic history.

**`anomaly_flags`** — Anomaly detection records from the analytics service, written via the Go API. Contains: flag type, implicated player IDs, evidence summary, detection timestamp, review status (OPEN, REVIEWED, DISMISSED, ACTIONED).

**`admin_audit_log`** — Append-only record of every admin command executed. Contains: admin player ID, target player ID, action type, parameters, timestamp. Cannot be modified through any admin command interface.

---

## 10. INFRASTRUCTURE & DEVOPS

---

### `infra/docker-compose.yml`

Production Docker Compose configuration. Defines five services: `postgres` (with named data volume), `redis` (with AOF persistence), `nats` (JetStream enabled), `bot` (Go binary built from `bot/Dockerfile`), and `analytics` (Python built from `analytics/Dockerfile`). The `activity` frontend is a static build served by Caddy — not a runtime container. All services share a Docker network. Environment variables sourced from `.env`, never hardcoded.

---

### `infra/docker-compose.dev.yml`

Local development override. Mounts Go source with `air` for live reload. Activity runs via Vite dev server with HMR. Analytics runs with Python `watchdog`. Exposes database and Redis ports locally for direct inspection.

---

### `infra/caddy/Caddyfile`

Caddy configuration. Handles TLS via Let's Encrypt. Routes `/api/*` to the Go backend. Serves Activity static assets. Handles Discord's proxy URL mappings for Activity traffic.

---

### `infra/postgres/init.sql`

Runs on first PostgreSQL container initialization. Creates the application database and role, grants permissions, sets transaction isolation default, timezone, and statement timeout. Does not create tables — that is the migration system's responsibility.

---

### `infra/nats/nats.conf`

NATS JetStream configuration. Defines streams: `settlement_events` (7-day retention), `market_events` (1-day retention), `notifications` (24-hour retention). Ensures messages persist if a subscriber is temporarily offline.

---

### `.env.example`

Documents every required environment variable with descriptions and example values. Never contains real credentials. The actual `.env` is gitignored. Covers: Discord tokens, application ID, public key, Activity client ID, PostgreSQL DSN, Redis address, NATS URL, JWT secret, analytics shared secret, settlement schedule, developer notification channel ID, log level, log format, and per-layer feature flags.

---

## 11. UI DESIGN SYSTEM

The Activity is the visual face of Simlife. It must feel like a premium product. The design language and the composited embed images share the same visual identity — a player should immediately recognize that the embed image and the Activity panel they open are from the same product.

---

### Visual Identity

**Name:** Simlife

**Tagline:** "One city. Every player. Real consequences."

**Design Language:** Dark urban economy. Clean and data-forward, with warmth drawn from amber and gold accents. Functional but crafted — a financial terminal with personality.

---

### Color Palette

| Token | Value | Use |
|---|---|---|
| `--bg-base` | `#0f1117` | Main application background |
| `--bg-surface` | `#181c26` | Cards, panels, sidebars |
| `--bg-elevated` | `#1f2535` | Modals, dropdowns |
| `--bg-highlight` | `#2a3045` | Hover states, selected rows |
| `--accent-gold` | `#d4a847` | Primary interactive elements, currency icons |
| `--accent-gold-dim` | `#9a7830` | Secondary gold, adjacent borders |
| `--accent-blue` | `#4a90d9` | Links, informational indicators |
| `--accent-green` | `#4caf76` | Positive values, growth, income |
| `--accent-red` | `#e05555` | Negative values, losses, alerts |
| `--accent-purple` | `#8a6fd8` | Premium features, City Pass indicators |
| `--text-primary` | `#e8eaf0` | Main body text |
| `--text-secondary` | `#8e95a8` | Labels, metadata, timestamps |
| `--text-muted` | `#4a5268` | Placeholder text, disabled states |
| `--border-subtle` | `#2a2f42` | Panel borders, table dividers |
| `--border-active` | `#3d4460` | Focused inputs, active tabs |

These same color values are used in the compositing system's renderer for text and UI elements drawn onto embed images, keeping the bot's Discord responses visually consistent with the Activity.

---

### Typography

**Primary Font:** Inter (variable) — all UI text.
**Monospace Font:** JetBrains Mono — transaction IDs, numerical values.
Both fonts are used in the Activity and in the image compositing system.

| Scale | Size | Weight | Use |
|---|---|---|---|
| `--text-xs` | 11px | 400 | Timestamps, footnotes |
| `--text-sm` | 13px | 400 | Secondary labels, table data |
| `--text-base` | 15px | 400 | Body text, descriptions |
| `--text-md` | 17px | 500 | Panel titles, section headers |
| `--text-lg` | 21px | 600 | Major section headings |
| `--text-xl` | 28px | 700 | Balance displays, key metrics |
| `--text-2xl` | 36px | 800 | Economic headline numbers |

---

### Component Specifications

**Cards** — `border-radius: 10px`, `background: var(--bg-surface)`, `border: 1px solid var(--border-subtle)`, `padding: 16px 20px`. No drop shadows — depth comes from background color elevation.

**Buttons (Primary)** — Gold fill, dark text. `background: var(--accent-gold)`, `color: #0f1117`, `border-radius: 6px`, `padding: 8px 16px`, `font-weight: 600`. Hover: `#e8b84d`. Active: `scale(0.97)`.

**Buttons (Secondary)** — Outline style. `background: transparent`, `border: 1px solid var(--border-active)`, `color: var(--text-primary)`. Hover: `background: var(--bg-highlight)`.

**Data Tables** — Alternating rows using `--bg-surface` and `--bg-highlight`. Headers: `--text-secondary`, uppercase, 11px. Numeric columns right-aligned in monospace. Positive values in `--accent-green`, negative in `--accent-red`.

**Value Displays** — Large monospace numbers with gold currency symbol prefix. Change indicators (arrow + percentage) in green or red to the right.

**Navigation Sidebar** — Fixed left, 64px icon-only, expands to 200px on hover/tap. Active panel: gold left border with subtle icon glow.

**Top Bar** — Simlife logo (left), current economic day number and time to next settlement (center), player balance summary (right), notification bell with unread badge.

---

### Discord Embed Design

Every command response embed follows a consistent structure: a title line matching the command category, the composited image as the embed image (1200×630px), a footer showing the current economic day number and time to settlement. The embed accent color (the left stripe Discord renders) matches the response category — gold for economy commands, blue for market commands, green for income events, red for alerts and losses. The embed contains minimal text fields — the image carries the data. This keeps embeds visually clean and avoids the wall-of-text appearance of conventional economy bots.

---

### City Map Visual Design

The city grid uses an isometric-style tile renderer on canvas. Zone colors: residential (warm cream), commercial (cool slate blue), industrial (dark steel), government (deep purple). Development level shown by building height and complexity on the tile. Owner names appear on hover as tooltips — never as permanent labels. Online players appear as glowing dots with slow movement animations between their owned tiles.

---

### Mobile Considerations

The Discord mobile Activity renders in a bottom sheet at approximately 60% of screen height. All panels must be usable in a 390×480px viewport. The sidebar collapses to a bottom tab bar on mobile (viewport width breakpoint). Touch targets minimum 44×44px. Charts simplified on mobile. City map defaults to a view centered on the player's owned plot on mobile.

---

## 12. INTER-SERVICE COMMUNICATION

---

### HTTP (Synchronous)

**Activity → Go Backend:** All player-facing API calls from the Activity use authenticated HTTPS with player JWTs. The Activity never contacts Python directly.

**Python Analytics → Go Backend:** The analytics service POSTs snapshots and alerts to internal Go endpoints using a shared-secret header. This is a server-to-server call, not exposed publicly. Go logs all analytics service communications.

**Go Backend → Discord API:** The Go backend calls Discord's REST API for command registration, channel messages, and OAuth token exchange using the bot token.

---

### NATS (Asynchronous)

**`settlement.start`** — Published by Go when settlement begins. Analytics subscribes to prepare.

**`settlement.complete`** — Published by Go when settlement finishes. Analytics subscribes to trigger the daily computation pipeline.

**`market.trade`** — Published by Go's market engine on each completed trade. The Go API server's WebSocket hub forwards this to connected Activity clients viewing the Market View.

**`economic.news`** — Published by Go after the analytics snapshot is received and stored. The Go bot layer subscribes and sends the daily economic news embed with its composited image to all configured Discord channels.

**`player.notification`** — Published by Go services when a player should be notified. The notification delivery service subscribes and delivers via Discord DM or channel message, with a composited notification image where appropriate.

---

## 13. SECURITY ARCHITECTURE

---

### Financial Transaction Security

All ledger transactions use PostgreSQL `SERIALIZABLE` isolation, preventing phantom reads, non-repeatable reads, and serialization anomalies under concurrent load. Any transaction that fails serialization is rolled back and retried. The ledger's `PostTransaction` function is the only code path that can modify account balances. It validates: sufficient balance in the source account, recipient existence, positive non-zero amount, and same-world account membership. If the transaction aborts, neither entry is committed and the player's balance is unchanged.

---

### API Security

The Go API server validates Discord's webhook signature on every interaction using the bot's public key and Discord's request headers. JWT tokens are validated on every Activity API request — signature, expiry, and player existence. Invalid tokens return 401 with no data. The JWT secret is a high-entropy random string stored only in environment variables.

---

### Rate Limiting

Two independent layers: Discord-level rate limiting (DiscordGo self-limits per Discord's API headers, preventing bot API bans) and game-level rate limiting (Redis-based per-player cooldowns, intentionally tighter than Discord's limits). Rejections logged at INFO level.

---

### Data Isolation

Player identity is Discord user ID — permanent and globally unique. The bot never trusts usernames or display names as identity. Guild IDs are stored where relevant but do not partition economies. A player's economic state is identical across all servers.

---

### Logging and Audit Trail

Go's structured logger produces a complete operational record: every command received, every ledger transaction posted, every settlement phase completed, every API request, every admin action. Logs are JSON-formatted in production and retained according to the deployment's log rotation policy. The `admin_audit_log` database table maintains a permanent, append-only record of all admin commands that is independent of the log files and cannot be tampered with through any admin interface.

---

## 14. ANTI-EXPLOIT SYSTEM

---

### Prevention

The ledger's design prevents balance corruption by construction — no code path can debit without crediting, and no command handler writes balances directly. Rate limiting makes automation produce money no faster than a dedicated human player. Market escrow prevents double-spending by locking listed items until the listing resolves.

---

### Detection

The velocity monitor in `internal/anti_exploit/velocity.go` flags statistical anomalies on every transaction in real time. The Python analytics service runs daily graph-based analysis on the transfer network, catching farming rings that per-transaction checks miss. Detections are posted to Go via the internal API, written to `anomaly_flags`, and logged.

---

### Response

The developer reviews `anomaly_flags`. Available admin tools: `/admin inspect @player` (full transaction history embed with composited summary image), `/admin freeze @player` (prevents all command execution pending review), `/admin confiscate @player amount` (moves funds to city treasury with ledger record), `/admin ban @player` (soft-delete flag, blocks all future interactions). All admin actions are logged in `admin_audit_log` and the structured logger simultaneously.

---

*Simlife Build Documentation — v1.1*
*Stack: Go · React · Python · PostgreSQL · Redis · NATS · Discord Activity*
*Architecture: Global world · Double-entry ledger · Dynamic image compositing · 12-layer progressive build*
