# SIMLIFE — DISCORD SERVER SETUP & BOT DIRECTORY GUIDE

---

## PART 1 — DISCORD SERVER STRUCTURE

This is the recommended server structure for Simlife's official Discord. The server serves two audiences simultaneously: players who use the bot, and developers/investors who evaluate it as a SaaS product. The layout reflects both.

---

### Recommended Server Name
**Simlife** — keep it identical to the bot name. No "Official" or "HQ" suffix. Clean.

### Server Icon
Use the same asset you design for the bot's profile picture. Consistency across the bot avatar, server icon, and embed images is what makes it look like a real product.

---

### CHANNEL STRUCTURE

```
SIMLIFE
│
├── 📋 INFORMATION
│   ├── 📌 welcome
│   ├── 📖 about-simlife
│   ├── 📜 rules
│   ├── 🔗 links                    ← bot invite, top.gg, website
│   └── 🗺️ roadmap
│
├── 📢 ANNOUNCEMENTS
│   ├── 📣 announcements            ← major updates, Discord partner announcements
│   ├── 🔄 patch-notes              ← every bot update, versioned
│   └── 📊 economic-news            ← Simlife posts daily economic report here automatically
│
├── 🤖 BOT COMMANDS
│   ├── 💰 economy                  ← /balance /work /pay /bank /shop
│   ├── 🏢 business                 ← /business open/status/hire/fire
│   ├── 🛒 market                   ← /market list/buy/bid/history
│   ├── 🏠 property                 ← /property buy/sell/develop
│   └── 🏛️ government               ← /vote /propose /treasury
│
├── 💬 COMMUNITY
│   ├── 💬 general
│   ├── 📈 economy-talk             ← players discuss market strategy, price trends
│   ├── 🤝 hiring-board             ← players post business job openings
│   ├── 🏪 player-shops             ← players advertise their businesses
│   └── 🤔 help
│
├── 🐛 FEEDBACK
│   ├── 💡 suggestions
│   ├── 🐛 bug-reports
│   └── 📊 polls
│
└── 🔒 STAFF (private)
    ├── 🛠️ developer-log
    ├── 🚨 anomaly-alerts           ← Simlife posts exploit flags here automatically
    └── 📋 admin-commands           ← /admin inspect/freeze/ban run here
```

---

### ROLE STRUCTURE

| Role | Color | Purpose |
|---|---|---|
| Developer | `#d4a847` (gold) | You. Full permissions. |
| Moderator | `#4a90d9` (blue) | Server moderation, no admin bot commands |
| Early Adopter | `#8a6fd8` (purple) | Automatically granted to first 500 players |
| Verified Player | `#4caf76` (green) | Players who have run at least one command |
| Business Owner | `#e8b84d` (amber) | Players who own an active business |
| Top Earner | `#d4a847` (gold) | Top 10 on the weekly wealth leaderboard, auto-updated |
| Citizen | `#8e95a8` (grey) | Default member role |

---

### SERVER SETTINGS

**Verification Level:** Medium (must have verified email, must be a Discord member for 5 minutes). Prevents throwaway accounts flooding your server.

**Explicit Media Content Filter:** Scan messages from all members.

**Default Notifications:** Only @mentions. Players in an active economy server get spammed otherwise.

**System Messages Channel:** Disable join messages — with a large bot, the join notifications become noise. Use a bot (like Carl-bot) to send a custom welcome DM instead.

**Community Server:** Enable Community features in Server Settings → Enable Community. This unlocks Server Discovery (free discoverability) and Announcement channels (which let members follow your #announcements to their own servers — free distribution).

**Announcement Channels:** Set `#announcements` and `#economic-news` as Announcement type channels. Members and other servers can follow them. Every time you post to these channels, it cross-posts to all followers.

---

### AUTOMOD RULES

Set up Discord's built-in AutoMod (Server Settings → AutoMod):

- Block common spam keywords
- Block invite links in all channels except `#links` and `#player-shops`
- Mention spam limit: 5 mentions per message
- Flag messages with harmful links

---

### BOTS TO ADD TO YOUR SERVER

**Carl-bot** — Welcome DMs, reaction roles, logging, custom commands. Free tier is sufficient.

Set up Carl-bot to:
- DM every new member with a welcome message that explains Simlife and links to `#about-simlife`
- Post a join message in `#general` (optional, disable if it gets noisy)
- Give the `Citizen` role automatically on join

**Statbot** (optional) — Server analytics. Useful for showing growth metrics to potential investors or partners.

---

### CHANNEL PERMISSIONS TEMPLATE

`#economic-news` — Send Messages: Simlife bot only. Everyone else: read only.
`#anomaly-alerts` — Visible to Developer role only.
`#admin-commands` — Visible to Developer role only.
`#patch-notes` — Send Messages: Developer only. Everyone else: read only.
`#announcements` — Send Messages: Developer only. Everyone else: read only.
All `BOT COMMANDS` channels — Send Messages: Everyone. But configure Simlife to only respond to commands, so it stays clean.

---

---

## PART 2 — BOT DIRECTORIES: WHERE TO LIST AND HOW TO GET VOTES

Bot directories are websites where Discord users browse and discover bots. Getting listed and getting votes does two things: drives organic installs, and acts as social proof when server admins evaluate whether to add your bot. A bot with 10,000 votes on top.gg is a bot that server admins trust.

---

### THE DIRECTORIES — RANKED BY IMPACT

---

#### 1. TOP.GG (top.gg)
**The most important directory. Non-negotiable.**

Top.gg is the largest Discord bot list. It drives the most organic discovery traffic. Every serious bot is listed here. Being absent from top.gg makes the bot look unfinished regardless of how good it is.

**How to list:**
- Go to top.gg → Add a Bot
- Requires the bot to be in at least one server (your own counts)
- Fill in the bot description (short: one punchy sentence, long: full feature overview)
- Set the bot's category tags: Economy, Utility, Fun, Games
- Add your bot's invite link and website link
- Submit for review (usually approved within 24 hours)

**How voting works on top.gg:**
- Each Discord user can vote for your bot once every 12 hours
- Top.gg's discovery algorithm heavily weights recent votes — a bot that got 50 votes today ranks above one with 10,000 total votes but no activity
- Top.gg provides a webhook: when someone votes, it sends a POST to your configured URL. You hook this into Simlife to reward voters (e.g., a small currency bonus)

**How to drive votes:**
- Add a `/vote` command to Simlife that sends the player the top.gg voting link with a message explaining the reward
- Post in `#announcements` every time you hit a vote milestone (500 votes, 1,000 votes, etc.)
- Run vote-for-reward events: "First 100 people to vote this weekend get double wages for 24 hours"
- Add the top.gg vote link to your bot's `/help` response and every embed footer

**Vote reward implementation:**
Top.gg sends a webhook POST to your configured endpoint when a user votes. In the Go backend, add an endpoint (e.g., `/webhooks/topgg`) that: validates the top.gg webhook authorization header, looks up the player by the Discord user ID in the vote payload, grants a configured currency bonus via `ledger.PostTransaction` with type `VOTE_REWARD`, and notifies the player via DM.

---

#### 2. DISCORDBOTLIST.COM (discordbotlist.com)
**Second largest. Worth listing immediately after top.gg.**

Same process as top.gg — submit the bot, write a description, add tags. Voting works identically with a webhook system. Smaller user base than top.gg but still meaningful.

---

#### 3. DISCORD.BOT (discord.bot)
**Rising directory. Good for supplementary exposure.**

Newer and growing. Listing here takes 10 minutes and the audience is smaller but tends to be more engaged with newer/less mainstream bots — which Simlife is.

---

#### 4. DISCORDS.COM (discords.com)
**Server and bot discovery combined.**

Lists both Discord servers and Discord bots. Listing your server here plus the bot doubles your surface area. Particularly useful because it drives server joins in addition to bot installs.

---

#### 5. DISCORD SERVER DISCOVERY (Built into Discord)
**Free and built-in. Activate it first.**

Once you enable Community features on your server and reach 500 members, your server is eligible for Discord's built-in Server Discovery — the Explore tab that Discord users browse inside the app. This is zero-cost organic reach inside Discord itself.

Requirements: Community server enabled, 500+ members, complete server guidelines, good standing (no recent violations), active community (measured by Discord internally).

---

### DIRECTORIES TO SKIP (FOR NOW)

**Bots on Discord (bots.ondiscord.xyz)** — Low traffic. Not worth the time at launch.

**Discord Extreme List** — Very small audience. Revisit if the others are fully optimized.

**Infinity Bot List** — Small community, opt in later if you want to maximize coverage.

---

### VOTE REWARD SYSTEM — RECOMMENDED DESIGN

This is what keeps votes coming in after launch week excitement dies down.

**Reward for voting:** 250 Simlife credits (the in-game currency) deposited directly to the player's wallet. Small enough to not be economically significant, meaningful enough to feel worth clicking.

**Vote streak bonus:** 3 consecutive days voting → 750 credits. 7 consecutive days → 2,000 credits + a "Supporter" badge displayed on their player profile image.

**Server-wide vote milestone events:** When the bot crosses a vote milestone on top.gg (e.g., 1,000 votes, 5,000 votes), announce it and trigger a one-time global economy event (e.g., "Vote Surge: Market prices drop 10% for 24 hours" or "Economy Boost: Base wages +20% for 48 hours"). This makes every vote feel like it benefits everyone, not just the voter.

**How to remind players to vote:**
- Add a subtle reminder to the `/balance` embed footer: "💛 Vote for Simlife on top.gg to earn 250 credits"
- Run a `/vote` command that sends the link and the current reward
- When a player's streak resets (hasn't voted in 13+ hours), DM them a reminder (opt-in, not default — default opt-in off to avoid being spammy)

---

### PARTNER SERVERS — PARALLEL STRATEGY

While votes build organically, reach out to large Discord servers whose communities overlap with Simlife's audience:

**Target server types:**
- Large economy bot servers (DiscordRPG, UnbelievaBoat, etc.) — these are players already into economy games
- Gaming community servers with 10,000+ members
- Entrepreneur/business community servers
- Coding/developer servers (Simlife as a portfolio piece resonates here)

**What to send:**
A short DM to the server owner or partnership channel: "Hey — I built Simlife, a persistent economy simulation for Discord. Players share one global economy — every trade affects every other player's prices. Looking to partner with communities that might enjoy it. Happy to add a mutual shoutout or set up a server-exclusive economy event." Keep it short. No walls of text.

---

### TIMELINE FOR DIRECTORY STRATEGY

**Before launch (while building):**
- Set up your Discord server completely
- Join top.gg and discordbotlist.com's developer waitlists or pre-register

**At launch (bot is public and working):**
- List on top.gg immediately
- List on discordbotlist.com the same day
- List your server on discords.com

**Week 1 post-launch:**
- Run the first vote milestone event when you hit 100 votes
- List on discord.bot
- Begin reaching out to 5 partner servers per week

**Month 1:**
- Top.gg listing should have at least 500 votes if vote rewards are implemented
- Apply for Discord Server Discovery when you approach 500 members
- Begin collecting testimonials from active players to add to bot descriptions

---

*Simlife Server & Directory Guide — v1.0*
