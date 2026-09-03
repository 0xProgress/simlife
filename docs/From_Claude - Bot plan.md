# DISCORD PERSISTENT ECONOMY / LIFE SIMULATION BOT
## Complete Research & Product Strategy Dossier
### Layer 0 — Research, Design & Strategic Foundation

---

> **Methodological Note:** Facts are labeled [VERIFIED], [REPORTED], [INFERENCE], or [SPECULATION] per the research brief. Where data was unavailable, this is stated explicitly. This document challenges the original concept wherever honesty demands it.

---

# DOCUMENT 1 — EXECUTIVE SUMMARY

## The Verdict Up Front

**Is this worth building? Yes — with significant scope discipline.**

The Discord economy bot ecosystem in 2025–2026 is dominated by shallow, per-server entertainment systems. No existing product achieves persistent multiplayer economic simulation at meaningful depth. The market gap is real. The ambition described is achievable — but only through ruthless layering and restraint. A solo developer attempting the full vision immediately will produce nothing. A disciplined builder working through twelve progressive layers over three to five years could produce something genuinely significant.

## The Core Finding

Every current Discord economy bot fails in the same way: **they simulate the surface of an economy without simulating interdependence.** Players earn coins, buy items, and gamble — in complete isolation from every other player. There is no emergent economy. There are no stories. There is no reason to care.

The proposed product — done correctly — would be the first Discord game where **what one player does materially affects what another player can do.** That is the differentiator. That is the reason to build it.

## The Critical Challenge

Discord's UX is not designed for complex games. Players resist learning. Retention in economy bots is notoriously low. The first version must be absurdly simple and immediately rewarding. Complexity must be earned by the player over time, not front-loaded at signup.

## Summary of Recommendations

- **Build in 12+ progressive layers**, each producing a complete playable product
- **Layer 1 must be 5-command simple** — wallet, work, spend, bank, check
- **The global/persistent world** is the non-negotiable differentiator — one shared economy, not per-server silos
- **Social interdependence** is the retention mechanism — not grind, not gambling
- **Monetization = cosmetics + server tools + premium analytics** — never pay-to-win
- **First playable target: 90 days from today**, Layer 1 only

---

# DOCUMENT 2 — COMPETITOR RESEARCH

## 2.1 Dank Memer

**Scale:** [VERIFIED] Added to over 20 million Discord servers [REPORTED]. Described as "Discord's largest in-app indie game." [REPORTED, source: bot's own top.gg listing]

**Core Loop:** Players earn "coins" through work commands, rob other players, gamble, complete heists in groups, care for virtual pets, and farm. Heavy meme integration.

**Economy Mechanics:**
- Single global currency ("coins") per-user, not per-server
- Wallet (risky, robbable) vs. bank (safe)
- No supply/demand, no price discovery
- No player-to-player meaningful economic transactions beyond robbery

**Key Features:**
- 265+ unique items, 100+ skins
- Global user-to-user marketplace
- Farming system (care for farm, sell produce)
- Virtual pets (level, breed, fight)
- Adventures with evolving storylines
- Fishing system (42 fish types, NPCs with quests)
- Heists (4-player group events)
- Gambling (blackjack, slots)
- Server event tools

**Monetization:** "Worker Premium" — faster cooldowns, exclusive items, ad-free. [VERIFIED from Dank Memer official sources]

**Strengths:**
- Enormous scale provides social validation
- Meme integration makes sharing inherently viral
- Rich item variety creates collection motivation
- Group mechanics (heists) drive friend recruitment
- Regular updates maintain content freshness

**Weaknesses:**
- Economy is fundamentally cosmetic — coins are a score, not a real currency
- No player interdependence — one player's success is disconnected from others
- Per-user isolation means no emergent world
- Extreme botting/automation problem [VERIFIED — multiple public GitHub autofarming tools exist]
- Meme focus dates badly
- No meaningful progression ceiling — "rich" players have nothing to do with wealth
- Per-server setup makes communities isolated from each other

**Retention Mechanism:** Gambling loops + new item FOMO + leaderboard competition within server
**Primary Retention Failure:** "I got rich — now what?" No answer exists.

---

## 2.2 UnbelievaBoat

**Scale:** [VERIFIED] Over 2 million servers, "hundreds of millions of commands run" [source: top.gg listing]

**Core Loop:** Per-server customizable economy. Server administrators configure currency, jobs, shops. Members earn by chatting, gambling, crime. Admins sell Discord roles for virtual currency.

**Economy Mechanics:**
- Fully custom per-server currency (name, icon)
- Wallet + bank split
- Role income (passive daily/weekly salary for role holders)
- Shop sells items AND Discord roles
- No cross-server economics

**Key Features:**
- Fully customizable shop (items, roles, categories)
- Casino games: blackjack, roulette, slots
- Animal racing, cockfight (gambling)
- Crime and rob commands
- Tax system (configurable)
- Web dashboard for admin configuration
- API access for developers
- Import from UnbelievaBoat to other bots
- Moderation integration

**Monetization:** Premium features through subscription — exact pricing unclear from public sources [INFERENCE from multiple secondary references to premium tier]

**Strengths:**
- Most flexible economy configuration in market
- Web dashboard is genuinely good
- Role-as-reward is clever and useful for server management
- API allows developer extensibility
- Long track record, stable product

**Weaknesses:**
- Per-server isolation is total — economies don't interact
- No meaningful player-to-player economics beyond admin-set shop
- "Simulation" is thin — work and crime are random text responses
- No progression depth beyond balance accumulation
- No reason for players from different servers to interact
- Economy is a server management tool, not a game

**Primary Insight:** UnbelievaBoat's strength is server management. Its weakness is that it has no game.

---

## 2.3 WabbitBot

**Scale:** [INFERENCE] Newer entrant (2025–2026), explicitly positioning as UnbelievaBoat alternative. No reliable user count found.

**Core Loop:** Free alternative to UnbelievaBoat with more features. Highly customizable per-server economy.

**Key Features:**
- 12 casino games (blackjack, slots, roulette, crash, mines, tower, high-low, dice, coinflip, RPS, keno, scratch cards)
- Custom crafting recipes (multi-tier)
- Stock market simulation with price volatility and charts
- Auction house and player-to-player marketplace
- Pet system with evolution
- XP and leveling with role rewards
- Web dashboard
- Import from UnbelievaBoat
- Economy analytics and inflation controls
- Dual currency support
- [VERIFIED] 100% free, cosmetic-only monetization

**Strengths:**
- Most feature-complete free economy bot found
- Crafting system adds production depth
- Stock market simulation adds investment layer
- Inflation controls show economic awareness
- Clean documentation

**Weaknesses:**
- Still per-server isolation
- Stock market is simulated (not player-driven)
- No life simulation or career depth
- No persistent world
- Free-forever model raises sustainability questions

**Critical Observation:** WabbitBot is technically impressive but has commoditized most economy bot features. The market now has a high-quality free option for everything shallow. This raises the bar — a new entrant must offer something WabbitBot cannot, not just "more features."

---

## 2.4 Tatsu (formerly Tatsumaki)

**Scale:** [REPORTED] Connected to over 1.35 million servers, 60+ million users [source: secondary Substack reference]. Global economy (not per-server).

**Core Loop:** Social community tool with light economy. XP earned from chatting funds credits, which buy profile cosmetics and virtual house decorations.

**Key Features:**
- Global economy (cross-server)
- Profile cards (highly customizable)
- Virtual house decorating
- Reputation system (upvoting between members)
- Virtual pets
- 2,000+ items in store (updated every two weeks)
- Trading between users
- Lottery
- Leveling with rank cards
- Server management (reaction roles, auto-roles)

**Monetization:** Freemium — premium profile customization, exclusive items

**Strengths:**
- Global economy creates cross-server social connections
- Profile customization drives engagement through self-expression
- Regular item updates sustain FOMO
- Social mechanics (reputation, trading) create relationships

**Weaknesses:**
- Economy is cosmetic — credits are "likes" not capital
- No strategic depth — no investments, businesses, or real decisions
- Social features are not economically interconnected
- Career/work simulation extremely shallow

**Primary Insight:** Tatsu demonstrates that global (not per-server) economies create more social stickiness. This is a critical lesson.

---

## 2.5 IdleRPG

**Scale:** [REPORTED] ~579 ratings on top.gg, suggesting moderate scale. Not a mass-market product.

**Core Loop:** True idle RPG — player creates character, character passively gains XP and finds items while player is away. Combat is automatic.

**Key Features:**
- Passive character progression (idle)
- Quest system
- Item discovery
- Battle between characters
- Guild/team system
- Multi-language support

**Strengths:**
- Genuinely idle — no constant attention required
- Low-friction engagement model

**Weaknesses:**
- No economy to speak of
- Minimal player interaction
- Little strategic decision-making
- No life simulation elements

---

## 2.6 Epic RPG

**Scale:** [REPORTED] One of the most popular RPG bots, broad use in gaming communities.

**Core Loop:** Traditional RPG with dungeon crawling, boss fights, quests. Currency earns through combat and quests. Equipment system.

**Key Features:**
- Dungeon crawling with tiered difficulty
- Boss fights
- Quest system
- Equipment crafting and trading
- Pet system
- Area/world exploration
- Guild system
- Time-based activities (hunt, fish, adventure)
- Leaderboards

**Strengths:**
- Clear progression ladder keeps players engaged
- Regular content updates
- Strong RPG community overlap

**Weaknesses:**
- Economy is loot-table-based, not player-driven
- No life simulation elements
- Limited player interdependence beyond guilds and trading

---

## 2.7 RP Sentry (Notable Discovery)

**Scale:** Unclear — relatively new

**Core Loop:** Roleplay-focused bot with RPG infrastructure. Admins build persistent worlds.

**Key Features:**
- Custom items, currencies, jobs, shops per server
- AI NPC/game master integration
- Law enforcement systems (911, robbery)
- CAD dispatch tools for GTA-style roleplay
- Lore-aware AI responses
- Plugin system for extended mechanics

**Significance:** This is the closest existing product to the proposed concept — but it is a roleplay tool for administrators, not a persistent automated world. Players participate in a world server admins build, not a world that runs itself.

---

## 2.8 Verdant (Non-Discord Comparison — Important)

**Scale:** [REPORTED] Web/mobile app, not Discord. Relevant as design reference.

**Description:** "A deep text-based UK life simulation. 130+ careers, 600+ businesses, 580+ properties, 1000+ vehicles, generational legacy system."

**Significance:** Verdant demonstrates that deep text-based life simulation is a viable product category. It is single-player. The proposed product would be the **multiplayer, Discord-native version** of what Verdant does — with a player-driven economy layered on top.

**Critical Lesson from Verdant:** Depth does not require graphics. A text-based system can model extraordinary complexity if the underlying data is rich enough.

---

## 2.9 Demetheria (Notable Discovery)

**Scale:** [REPORTED] Recently launched, geopolitical simulation Discord bot

**Description:** "Geopolitical, economic and war simulation. Connect your server to a persistent world map, manage resources and companies, wage strategic wars."

**Significance:** Demetheria is attempting persistent world simulation at the server level. Initial observation suggests it focuses on nation-states rather than individual player lives. This is relevant — there is a market for persistent simulation. The differentiation would be: Demetheria simulates nations; the proposed product simulates lives.

---

# DOCUMENT 3 — FEATURE COMPARISON MATRIX

## Table A — 50-Feature Comparison

| Feature | Dank Memer | UnbelievaBoat | WabbitBot | Tatsu | IdleRPG | Epic RPG | Proposed Product (Layer 1) | Proposed Product (Year 5) |
|---|---|---|---|---|---|---|---|---|
| **PLAYER** | | | | | | | | |
| Character creation | ⚠️ | ❌ | ❌ | ⚠️ | ✅ | ✅ | ✅ | ✅ |
| Skills system | ❌ | ❌ | ❌ | ❌ | ⚠️ | ✅ | ❌ | ✅ |
| Reputation score | ❌ | ❌ | ❌ | ✅ | ❌ | ❌ | ❌ | ✅ |
| Education system | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ | ✅ |
| Age/life stages | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ | ✅ |
| **ECONOMY** | | | | | | | | |
| Currency system | ✅ | ✅ | ✅ | ✅ | ⚠️ | ✅ | ✅ | ✅ |
| Inflation modeling | ❌ | ⚠️ | ⚠️ | ❌ | ❌ | ❌ | ⚠️ | ✅ |
| Supply/demand pricing | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ | ✅ |
| Player-driven prices | ❌ | ❌ | ⚠️ | ⚠️ | ❌ | ⚠️ | ❌ | ✅ |
| Taxation system | ⚠️ | ⚠️ | ❌ | ❌ | ❌ | ❌ | ⚠️ | ✅ |
| **BANKING** | | | | | | | | |
| Bank accounts | ✅ | ✅ | ✅ | ✅ | ❌ | ⚠️ | ✅ | ✅ |
| Loans | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ | ✅ |
| Interest (savings) | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ | ✅ | ✅ |
| Credit scores | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ | ✅ |
| Transaction history | ❌ | ⚠️ | ⚠️ | ❌ | ❌ | ❌ | ✅ | ✅ |
| P2P transfers | ✅ | ✅ | ✅ | ✅ | ❌ | ✅ | ✅ | ✅ |
| **EMPLOYMENT** | | | | | | | | |
| Jobs system | ⚠️ | ⚠️ | ⚠️ | ❌ | ❌ | ⚠️ | ✅ | ✅ |
| Career progression | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ | ✅ |
| Promotions | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ | ✅ |
| Qualifications | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ | ✅ |
| Player employing players | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ | ✅ |
| **BUSINESS** | | | | | | | | |
| Player-owned companies | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ | ✅ |
| Hiring employees | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ | ✅ |
| Production systems | ❌ | ❌ | ⚠️ | ❌ | ❌ | ❌ | ❌ | ✅ |
| Inventory management | ⚠️ | ⚠️ | ⚠️ | ⚠️ | ❌ | ✅ | ❌ | ✅ |
| Business bankruptcy | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ | ✅ |
| **PROPERTY** | | | | | | | | |
| Housing | ❌ | ❌ | ❌ | ⚠️ (virtual house) | ❌ | ❌ | ❌ | ✅ |
| Commercial property | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ | ✅ |
| Rent mechanics | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ | ✅ |
| Mortgages | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ | ✅ |
| Property taxes | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ | ✅ |
| **FINANCIAL MARKETS** | | | | | | | | |
| Stocks | ❌ | ❌ | ⚠️ (simulated) | ❌ | ❌ | ❌ | ❌ | ✅ |
| Bonds | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ | ✅ |
| Investment funds | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ | ✅ |
| **SOCIAL** | | | | | | | | |
| Reputation system | ❌ | ❌ | ❌ | ✅ | ❌ | ❌ | ❌ | ✅ |
| Organizations/guilds | ⚠️ | ❌ | ❌ | ❌ | ✅ | ✅ | ❌ | ✅ |
| Player-created institutions | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ | ✅ |
| **CRIME/LAW** | | | | | | | | |
| Crime mechanics | ✅ | ✅ | ✅ | ❌ | ❌ | ❌ | ⚠️ | ✅ |
| Prison system | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ | ✅ |
| Criminal records | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ | ✅ |
| **GOVERNMENT** | | | | | | | | |
| Elections | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ | ✅ |
| Tax policy control | ❌ | ⚠️ (admin) | ❌ | ❌ | ❌ | ❌ | ❌ | ✅ |
| **WORLD** | | | | | | | | |
| NPC companies | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ | ✅ (basic) | ✅ |
| Economic cycles | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ | ✅ |
| Random world events | ⚠️ | ❌ | ❌ | ❌ | ❌ | ❌ | ✅ (basic) | ✅ |
| Persistent world state | ❌ | ❌ | ❌ | ⚠️ | ❌ | ❌ | ✅ | ✅ |
| **UX/TECH** | | | | | | | | |
| Global (not per-server) | ✅ | ❌ | ❌ | ✅ | ⚠️ | ⚠️ | ✅ | ✅ |
| Web dashboard | ❌ | ✅ | ✅ | ✅ | ❌ | ❌ | ❌ | ✅ |
| Slash commands | ✅ | ✅ | ✅ | ✅ | ⚠️ | ✅ | ✅ | ✅ |
| Buttons/interactions | ✅ | ✅ | ✅ | ✅ | ❌ | ✅ | ✅ | ✅ |
| Anti-exploit systems | ⚠️ | ⚠️ | ⚠️ | ⚠️ | ❌ | ⚠️ | ✅ | ✅ |

**Legend:** ✅ = Supported | ⚠️ = Partially/Weakly | ❌ = Not supported | ? = Insufficient evidence

---

# DOCUMENT 4 — MARKET GAP ANALYSIS

## 4.1 What Nobody Does Well

**1. Persistent Global World with Player Interdependence**
Every major economy bot is per-server or uses global currency in a cosmetic way. No product creates a world where Player A's business decision affects Player B's livelihood. This is the gap.

**2. Meaningful Career Progression**
Every bot has "work command" = random text + fixed payout. No product has: choose career → study → qualify → apply → get hired → perform → get promoted → earn more → change careers. The difference between a simulation and a slot machine.

**3. Player-Driven Supply and Demand**
WabbitBot has a simulated stock market. No product has prices that move because players are actually buying and selling commodities to each other in a meaningful economic chain.

**4. Player-as-Employer**
No existing bot allows a player to hire other players, set their salary, give them tasks, and have that employment be economically meaningful to both parties.

**5. Asset Ownership with Ongoing Economics**
Tatsu has a "virtual house" for cosmetics. No product has property that generates income (rent), incurs costs (maintenance, taxes), and can be mortgaged or sold to other players at market prices.

**6. Emergent Stories Through System Interaction**
In Dank Memer, the most "story" that emerges is "someone robbed me." In the proposed product, stories like "I built a company, hired workers, took out a loan, the market crashed, I went bankrupt, a competitor bought my assets" should emerge naturally.

## 4.2 Features Nobody Combines Effectively

- Jobs + businesses + investment + property (all interconnected)
- Crime + legal consequences + criminal records affecting future employment
- Politics + economy (player politicians affecting tax rates that affect everyone)
- Social reputation + economic opportunity (trusted players get better deals)

## 4.3 Underserved Player Groups

- **Adult strategy gamers** who want economic depth without graphic-intensive games
- **Roleplayers** who want an automated economic world to inhabit
- **Educators** who want to demonstrate economic concepts
- **Entrepreneurs** who want business simulation
- **Long-term community builders** who want a game that grows with their server

## 4.4 Why Players Abandon Economy Bots

Based on community feedback and design analysis [INFERENCE from research patterns]:

1. **Hitting the wealth ceiling with nothing to do** — "I got rich, now I just flex"
2. **No social consequence** — what I do doesn't affect anyone
3. **Grind fatigue** — work command every N minutes, forever
4. **Nothing to compete over** — leaderboards with no real stakes
5. **No narrative** — sequences of commands without story
6. **New player hostility** — arriving in a server where veterans have 10x your balance
7. **Bot automation** — whales bot-farm while normal players can't compete

## 4.5 The Product Differentiation

**Why install this instead of Dank Memer?**
Because in this game, what you do *matters to others*. Your business provides jobs. Your investments fund development. Your criminal activity affects the local economy. You are not a solo player with a coin balance — you are a citizen of a world.

**Why install this instead of UnbelievaBoat?**
Because this is not a server management tool with a game bolted on. It is a game-first product with server features built around it.

**Why install this instead of a real game?**
Because it lives in Discord — where your community already exists, where your friends already are, where you already spend hours a day. Zero installation, zero context switching.

---

# DOCUMENT 5 — PRODUCT VISION

## 5.1 Definitions

**One-sentence description:**
A persistent, multiplayer economic life simulation inside Discord — where players build careers, own businesses, acquire property, and shape a real economy through their decisions.

**Short description (tweet-length):**
The first Discord game where your economic decisions actually affect other players. Build a life. Start a company. Own property. Shape the market. Everything is connected.

**Long description:**
*[Name]* is a Discord-native persistent life and economic simulation. Unlike conventional economy bots that give players isolated coin balances and random work outcomes, *[Name]* creates a shared virtual city where every player is economically connected to every other.

You start as a new arrival with nothing. You choose a career, earn a salary, save money, and gradually acquire assets. Your salary at a player-owned company depends on decisions that company's owner makes. The price of an apartment you want to buy is set by supply and demand. If you start a business, you will hire other players, set their wages, manage your cash flow, and compete with other player-owned companies.

The economy breathes because players make it breathe. There is no "winning" — there is only the story of your economic life, your relationships with other players, the businesses you built, and the mark you left on a world that existed before you arrived and will exist after you leave.

## 5.2 Player Personas

### Primary Audience: The Economic Strategist
- Age 18–30
- Plays strategy games (Civilization, Crusader Kings, Transport Tycoon)
- Wants depth and long-term planning
- Reads about economics, investing, or business
- Will spend months on this if the depth is there

### Secondary Audiences:

**The Casual Progressor**
- Wants to level up, earn money, and see numbers go up
- Doesn't need to understand all systems
- Will engage with jobs and banking without touching advanced features
- Recruited by friends, stays because of social bonds

**The Roleplayer**
- Wants a character in a living world
- Will give their character a personality and backstory
- Values immersion over optimization
- Drives community narrative and engagement

**The Entrepreneur**
- Fascinated by the business-building mechanics
- Will study the market, identify opportunities, build companies
- Competitive mindset, wants to "win" economically
- High session frequency, high engagement

**The Social Climber**
- Most interested in reputation, relationships, and organizations
- Builds social capital, creates institutions
- May become a server politician or union leader
- Retention driver for community

**The Criminal**
- Wants to operate in the shadow economy
- Risk-taker who enjoys asymmetric gameplay
- May love or hate being caught and imprisoned
- Creates narrative conflict that entertains the whole server

**The Investor**
- Primarily interested in financial markets
- Will analyze stocks, bonds, and economic trends
- Rarely interactive but deeply engaged with economic data
- Could become extremely powerful in late game

**Primary target for Layer 1:** The Casual Progressor — accessible enough to onboard, with depth hooks to retain Economic Strategists and Entrepreneurs.

---

# DOCUMENT 6 — FULLY MATURE PRODUCT (YEAR 5+ VISION)

## The World

The game takes place in a single persistent virtual city called **Aether City** — a modern metropolis with districts, businesses, properties, and a population made up entirely of real players and NPC filler characters. The city exists 24/7. Businesses open and close. Prices fluctuate. Elections happen. Economic crises emerge.

Every player who has ever joined the game has left a mark. Companies founded by players who quit still exist (managed by NPCs) until other players acquire them. Properties previously owned by players are on the market. Economic history is visible.

---

## Day 1 — A New Player Joins

**What they see:**

```
Welcome to Aether City! 🏙️

You've just arrived. You have:
  💰 $500 starting cash
  🏠 Temporary housing (free, 30 days)
  📋 No job yet

Your first steps:
  [Find a Job]  [Explore the City]  [Visit the Bank]
```

The new player is presented with Discord buttons. No commands to memorize. They tap **[Find a Job]**.

**Job Selection (interactive embed):**
```
Available Entry-Level Jobs:

🏗️ Construction Worker
  Pay: $200/day | Skill: Physical Strength
  Open positions: 14 (at various player-owned companies)

🍽️ Restaurant Worker
  Pay: $180/day | Skill: Service
  Open positions: 31 (mostly at player-owned restaurants)

📦 Delivery Driver
  Pay: $220/day | Skill: Logistics
  Open positions: 8

💻 Junior Developer
  Pay: $280/day | Skill: Technology
  Requirements: Technology Skill Level 2
  ❌ You don't qualify yet

[Apply for Construction]  [Apply for Restaurant]  [Apply for Delivery]
```

The player taps **[Apply for Restaurant]**. The bot asks them to select which employer — and shows three player-owned restaurants currently hiring. One is owned by a veteran player called "Silvio." The player joins Silvio's restaurant.

**What this means:**
- Player A is now an employee of Player B (Silvio)
- Silvio's restaurant now has one more worker, which increases his production capacity
- Player A's daily earnings come from Silvio's business revenue, not from the bot directly
- If Silvio's restaurant fails, Player A needs a new job

This is interdependence. It happens on Day 1.

---

## Month 1

Player A has been working at Silvio's restaurant for a month. Highlights:

**Banking:** Has saved $4,200. Opened a bank account. Bank pays 2% annual interest (calculated daily). Has a credit score of 680 (starting score, no history yet).

**Housing:** Temporary housing expired after 30 days. Player A paid $800/month for a basic apartment owned by another player, Diego. Diego now has rental income from Player A.

**Skills:** Player A's Service skill reached Level 3. This qualifies them for the "Senior Server" job title, with a salary bump to $240/day — Silvio pays this voluntarily because losing a trained worker costs him productivity.

**Social:** Player A joined the "Restaurant Workers Union" — a player-run organization that negotiates wages and monitors employer behavior. The union has 23 members across multiple player-owned restaurants.

**Spending:** Bought furniture for their apartment from a player-run furniture shop (owner: Marcus). Marcus's shop bought furniture from a player-run factory (owner: Elena). Elena's factory bought raw materials from a player-run mining company (owner: James).

Chain: Player A buys furniture → Marcus earns → Elena earns → James earns.

---

## Month 6

Player A has saved $28,000 and made meaningful progress.

**Career Move:** Left Silvio's restaurant after a public dispute about wage increase (Silvio was voted "low reputation employer" by the union). Moved to a larger restaurant chain owned by collective of three players. Earns $310/day.

**First Investment:** Bought 200 shares of "Aether Foods Corp" (a player-owned company) at $12/share ($2,400). Over 6 months, it rose to $18/share. Net gain: $1,200.

**Credit:** Bank approved a $15,000 personal loan at 8% annual interest. Used it to take an "Advanced Culinary Skills" course at the player-run Academy (owned by a player consortium). This raised Service skill to Level 6, qualifying for Restaurant Manager roles.

**Property:** Applied for a mortgage to buy a small property ($65,000 apartment). Down payment: $13,000 (savings + partial loan). Monthly mortgage payment: $420. Now earns $150/month renting the second bedroom to a newer player.

**Social Status:** Has a reputation score of 720 (above average). Known as a reliable employee. Two smaller restaurant owners have privately messaged asking if they'd be interested in a management role.

---

## Year 1

Player A is now a Restaurant Manager. Salary: $480/day.

**Assets:**
- Property value: $73,000 (appreciated)
- 200 shares Aether Foods Corp: $4,200 (still rising)
- Savings: $42,000
- Debt: $8,000 remaining on personal loan

**Business Activity:** Started a small catering company (registered for $2,000). Hired two newer players as workers. Takes catering contracts from player-organized events (weddings, corporate events — also player-run). Earns $1,200/week through the business on top of their manager salary.

**Social:** Joined the Restaurant Owners Association. Has relationships with suppliers (food distributors, also player-run). Has a rival: another catering company owner who undercuts on price.

**Political:** Voted in the city's quarterly economic council election. Voted for a candidate running on "lower business taxes" because their catering company's tax bill is significant.

---

## Mature Player (Year 3+)

Player A is now simultaneously:

- **CEO of Aether Catering Corp** — employs 8 players, operates in 3 districts
- **Property owner** — 3 properties (2 commercial, 1 residential portfolio)
- **Investor** — holds shares in 6 different player-owned companies
- **Lender** — has extended personal loans to 3 other players through the informal lending market
- **City Councilor** — elected with 34 votes, sits on the economic policy council
- **Union member** — senior member of the Restaurant Workers Union, though now on the employer side (creates interesting tension)

**System Interactions:**

As City Councilor, Player A voted to raise the minimum wage. This affects:
- All employers in the city (including Player A's own company — net cost increase)
- All workers in the city (including former colleagues at Silvio's restaurant)
- The price of food at every restaurant (higher wages → higher menu prices)
- Food company stock prices (higher labor costs → lower margins → stock drops)
- Player A's own stock portfolio (partially negative)

This single political vote creates ripples across every economic system — experienced differently by every player depending on their position.

**What does the highest-level player look like?**

The most advanced players have wealth in the millions, employ dozens of other players, own entire districts of the city, sit on governing bodies, control industries, and shape the economic reality that new players arrive into. They are not "winners" — they are **institutions**. Their power can be challenged. Their businesses can fail. Their political positions can be voted out. Their reputation can collapse.

The game never ends. It only gets more complex.

---

# DOCUMENT 7 — COMPLETE FEATURE ARCHITECTURE

```
AETHER CITY (Persistent World)
│
├── WORLD LAYER
│   ├── City
│   │   ├── Districts (Central Business, Residential, Industrial, Commercial)
│   │   ├── Properties (fixed inventory, player-tradeable)
│   │   ├── Infrastructure (roads, utilities — abstracted)
│   │   └── NPC Filler Population (simulates a living city)
│   ├── Time System
│   │   ├── Game Days (1 real hour = 1 game day) [RECOMMENDED]
│   │   ├── Seasons (affecting some economic activity)
│   │   └── Election Cycles (quarterly)
│   └── Events System
│       ├── Random Events (market shocks, disasters, opportunities)
│       ├── Scheduled Events (elections, fiscal periods)
│       └── Player-Triggered Events (strikes, expansions, crises)
│
├── PLAYER LAYER
│   ├── Identity
│   │   ├── Name, Avatar, Bio
│   │   ├── Background (chosen starting path)
│   │   └── Discord Integration (Discord name, server role)
│   ├── Stats
│   │   ├── Age (game time accumulation)
│   │   ├── Health (affects productivity — optional, later layers)
│   │   └── Happiness (morale modifier — optional)
│   ├── Skills
│   │   ├── Physical
│   │   ├── Service
│   │   ├── Technology
│   │   ├── Business
│   │   ├── Finance
│   │   ├── Legal
│   │   ├── Political
│   │   └── Criminal
│   ├── Education
│   │   ├── Self-study (slow, free)
│   │   ├── Player-run courses (faster, paid)
│   │   └── NPC institutions (standardized)
│   ├── Reputation
│   │   ├── General (community-wide)
│   │   ├── Employer reputation
│   │   ├── Employee reputation
│   │   ├── Business reputation
│   │   └── Criminal reputation
│   └── Relationships
│       ├── Employment relationships
│       ├── Business partnerships
│       ├── Lending/borrowing relationships
│       └── Organization membership
│
├── ECONOMY LAYER
│   ├── Currency
│   │   ├── City Credits (⊄) — primary currency
│   │   └── NO crypto (scope risk, regulatory risk)
│   ├── Price Mechanisms
│   │   ├── Fixed NPC prices (anchor)
│   │   ├── Player marketplace (supply/demand)
│   │   └── Auction system (for scarce assets)
│   ├── Money Faucets (creation)
│   │   ├── NPC employment (source of new money)
│   │   ├── Government spending (optional, later)
│   │   └── New player starting funds
│   └── Money Sinks (destruction)
│       ├── Transaction fees (% on all trades)
│       ├── Property taxes (recurring)
│       ├── Loan interest (paid to bank pool, partially destroyed)
│       ├── Business operating costs
│       ├── Skills courses (tuition)
│       └── Government taxes
│
├── BANKING LAYER
│   ├── Personal Accounts
│   │   ├── Checking (wallet/liquid)
│   │   ├── Savings (interest-bearing)
│   │   └── Transaction history
│   ├── Lending
│   │   ├── NPC bank loans (standardized)
│   │   ├── Player-to-player loans (negotiated)
│   │   └── Business credit lines
│   ├── Credit System
│   │   ├── Credit score (0–1000)
│   │   ├── Factors: payment history, debt ratio, age, defaults
│   │   └── Affects loan rates and amounts
│   └── Player-Run Banks (late game)
│       ├── Can accept deposits
│       ├── Issue loans
│       └── Subject to regulation
│
├── EMPLOYMENT LAYER
│   ├── Job Market
│   │   ├── NPC employer jobs (baseline, always available)
│   │   ├── Player employer jobs (player-run companies)
│   │   └── Job listings (visible to all players)
│   ├── Employment Relationship
│   │   ├── Contract (salary, hours, role)
│   │   ├── Performance rating
│   │   ├── Termination (firing/quitting)
│   │   └── Unemployment (with limited benefits)
│   └── Career Progression
│       ├── Entry Level → Junior → Mid → Senior → Manager → Director → C-Suite
│       ├── Horizontal (change industry)
│       └── Vertical (same industry, higher level)
│
├── BUSINESS LAYER
│   ├── Company Formation
│   │   ├── Registration cost
│   │   ├── Business type selection
│   │   └── Initial capital requirement
│   ├── Operations
│   │   ├── Production (inputs → outputs)
│   │   ├── Inventory management
│   │   ├── Supplier relationships
│   │   ├── Customer relationships
│   │   └── Pricing strategy
│   ├── Finance
│   │   ├── Revenue, expenses, profit
│   │   ├── Payroll
│   │   ├── Business loans
│   │   ├── Corporate taxes
│   │   └── Bankruptcy proceedings
│   ├── Ownership
│   │   ├── Sole proprietorship
│   │   ├── Partnership
│   │   ├── Corporation (shares, shareholders)
│   │   └── Acquisition/merger (player-to-player)
│   └── NPC Competitors
│       ├── Always present, set market baseline
│       ├── Respond to player market pressure
│       └── Never dominate (players can always win)
│
├── PROPERTY LAYER
│   ├── Residential
│   │   ├── Temporary housing (starter)
│   │   ├── Apartments (varying quality/price)
│   │   └── Houses (high value)
│   ├── Commercial
│   │   ├── Shop spaces
│   │   ├── Office buildings
│   │   └── Industrial facilities
│   ├── Ownership Mechanics
│   │   ├── Purchase (cash or mortgage)
│   │   ├── Rent (tenant pays landlord)
│   │   ├── Lease (business premises)
│   │   └── Property taxes (money sink)
│   └── Market
│       ├── Property listings
│       ├── Price history
│       ├── Mortgage calculator
│       └── Valuation model
│
├── FINANCIAL MARKETS LAYER
│   ├── Stock Exchange
│   │   ├── Player company IPOs
│   │   ├── Share trading (player-to-player)
│   │   ├── Price discovery (real supply/demand)
│   │   └── Market data and charts
│   ├── Bonds
│   │   ├── Government bonds (NPC-issued)
│   │   └── Corporate bonds (player-issued)
│   └── Commodities
│       ├── Raw materials pricing
│       └── Futures (limited, late game)
│
├── SOCIAL LAYER
│   ├── Organizations
│   │   ├── Trade Unions (workers)
│   │   ├── Business Associations (employers)
│   │   ├── Political Parties
│   │   ├── Criminal Organizations
│   │   └── Social Clubs
│   ├── Communication
│   │   ├── Public channels (Discord integration)
│   │   ├── Organization channels
│   │   └── Business communications
│   └── Reputation
│       ├── Community voting/rating
│       ├── Transaction history (public record)
│       └── Role-based (employer/employee/citizen)
│
├── GOVERNMENT LAYER
│   ├── Elections
│   │   ├── Economic Council (tax policy)
│   │   ├── City Mayor (symbolic, prestige)
│   │   └── Voting mechanics
│   ├── Policy
│   │   ├── Tax rates (affects all businesses)
│   │   ├── Minimum wage
│   │   ├── Business regulations
│   │   └── Public spending
│   └── Treasury
│       ├── Tax revenue
│       ├── Public expenditure (infrastructure, services)
│       └── Emergency interventions
│
└── CRIME/LAW LAYER
    ├── Criminal Activities
    │   ├── Theft (property crime)
    │   ├── Fraud (financial crime)
    │   ├── Black market (illegal goods)
    │   └── Organized crime (gang activities)
    ├── Law Enforcement
    │   ├── Investigation mechanics
    │   ├── Arrest probability
    │   └── Police force (NPC + player-run)
    └── Legal System
        ├── Charges, trials
        ├── Fines (money sink)
        ├── Prison (time-based, limits activity)
        └── Criminal record (affects opportunities)
```

---

# DOCUMENT 8 — LAYERED DEVELOPMENT ROADMAP

## Meta-Principle

Each layer must produce a **playable, enjoyable product**. A player who joins during Layer 2 must have a complete experience without needing Layer 5 to exist. Layers add depth, they don't complete the game.

---

## LAYER 0 — Research & Design (CURRENT)
**Duration:** 4–8 weeks (solo) | 2–4 weeks (small team)
**Purpose:** Don't build yet. Understand what to build.

**Deliverables (this document is Layer 0):**
- Game Design Document ✓ (this dossier)
- Competitor analysis ✓
- Technical architecture plan
- Economic model specification
- Development roadmap ✓
- Layer 1 feature specification

**Exit Criteria:** Developer can describe every Layer 1 feature in detail from memory.

---

## LAYER 1 — FOUNDATION: CORE ECONOMY
**Duration:** 6–12 weeks (solo) | 3–6 weeks (2-person team)
**Difficulty:** Medium | **Economic Risk:** Low | **Player Value:** Medium

### Purpose
Build the irreducible minimum. Prove the economic loop is fun. Get 100 real players engaged before building anything else.

### Included Features

**1. Player Profile**
- Name (inherited from Discord username)
- Wallet balance (starting: ⊄2,500)
- Bank account (savings, earns 1.5% annual interest daily)
- Credit score (starting: 650)
- Transaction history (last 20 transactions)
- Job status

**2. NPC Jobs (5 types)**
Each has: daily wage, skill requirement, brief description, cooldown
- Construction Worker: ⊄180/day | No requirements
- Service Worker: ⊄165/day | No requirements  
- Delivery Driver: ⊄200/day | No requirements
- Office Clerk: ⊄220/day | Basic Business skill
- Technical Operator: ⊄240/day | Basic Technology skill

Jobs are worked once per game day. The bot runs "end-of-day" processing once every real-world hour.

**3. Basic Banking**
- Deposit/withdraw between wallet and savings account
- Savings interest (1.5% annually, applied daily)
- P2P transfers (with 2% fee — money sink)
- Balance check
- Transaction log

**4. NPC Shop (Basic Needs)**
- Housing (rent: ⊄400/game-month for basic apartment, ⊄600 for mid-tier)
- Food (optional flat daily cost ⊄20 — improves morale modifier)
- Transportation (optional flat cost for better income multiplier)

Housing is a money sink and a baseline asset mechanic.

**5. Basic Skills**
- 3 skills at Layer 1: Physical, Service, Technology
- Skill increases through work (slow accumulation)
- Skill unlocks better job options

**6. World Events (Simple)**
- Daily random event: economic modifier (+/- 5–15% on all wages for that game day)
- Public announcement in Discord channel
- Events are simple narrative text ("Construction boom! All physical workers earn 15% more today.")

**7. Leaderboard**
- Wealth leaderboard (wallet + savings)
- Income leaderboard
- Refreshed each game day

### Excluded From Layer 1
❌ Player businesses (Layer 4)
❌ Property ownership (Layer 3)
❌ Loans (Layer 2)
❌ Player-to-player employment (Layer 4)
❌ Markets (Layer 5)
❌ Crime (Layer 9)
❌ Organizations (Layer 7)
❌ Government (Layer 8)
❌ Stocks (Layer 6)

### Discord UX at Layer 1

All interactions use slash commands + button menus. No text commands.

```
/start → Welcome flow with buttons
/profile → Shows stats, balance, job, skill summary
/work → Works current job (or selects job if none)
/bank → Banking menu (deposit, withdraw, transfer, history)
/shop → NPC shop (housing, food)
/city → Daily world news and event
/leaderboard → Top players
/help → Feature guide
```

### Success Criteria
- ≥100 players with a second session within 7 days
- ≥50% of players work their job 3+ consecutive game days
- ≥30% of players deposit money in savings
- Average session: 3+ commands
- Zero critical economy exploits in first 2 weeks

### Failure Criteria
- <20% 7-day retention
- Any money duplication exploit going undetected for more than 24 hours
- Player complaints primarily about boredom (not bugs) — indicates loop is broken

---

## LAYER 2 — PERSONAL FINANCE & CAREER PROGRESSION
**Duration:** 4–8 weeks | **Difficulty:** Medium | **Player Value:** High

### Purpose
Give players something to plan toward. Replace "grind forever" with "save for a specific goal."

### New Features

**1. Loan System**
- NPC bank offers loans at 8–15% annual interest
- Loan amount based on credit score (max 10x monthly income)
- Repayment schedule (automatic deductions from salary)
- Default consequences: credit score drop, potential wage garnishment

**2. Credit Score System**
- Factors: payment history (40%), debt level (30%), account age (20%), financial behavior (10%)
- Visible to all players (transparency creates trust mechanics)
- Affects loan rates and amounts

**3. Career Tiers**
- Each NPC job gets 3 tiers: Entry → Mid → Senior
- Promotion requires: time at job + skill level + optional performance metric
- Salary increases: 20% per tier

**4. Education System (Basic)**
- Self-study (free but slow: takes 3 game weeks per skill level)
- NPC Academy courses (paid ⊄500–⊄2,000, faster: 1 game week)
- Academy fees are a money sink

**5. Financial Goals**
- Player can set a savings goal ("Saving for apartment deposit: ⊄15,000")
- Dashboard shows progress
- Simple motivational mechanic

**6. Expense Tracking**
- Visible breakdown: housing costs, food costs, loan repayments, savings
- Net income calculation

### Success Criteria
- ≥40% of players take at least one course
- ≥30% of players take a loan and repay it
- Average tenure with game: 14+ days

---

## LAYER 3 — PROPERTY & OWNERSHIP
**Duration:** 8–12 weeks | **Difficulty:** Medium-High | **Economic Risk:** Medium

### Purpose
Create tangible wealth that exists in the world — not just a number.

### New Features

**1. Property System**
- Fixed inventory of properties per district (e.g., 200 apartments, 50 houses, 30 commercial units)
- Properties have values set by supply/demand of player purchases
- Monthly property taxes (⊄ → game treasury/sink)
- Maintenance costs (auto-deducted)

**2. Buying Property**
- Cash purchase (available from Layer 3)
- Mortgage (30-year game equivalent, monthly payment)
- Property listed on personal profile

**3. Renting Property**
- Player-landlords set monthly rent
- Tenant players pay automatically per game month
- Late payment: credit score impact
- Eviction: legal process (simple)

**4. Property Market**
- Property listings visible to all
- Price history (graph in web dashboard)
- Properties can be re-sold at market rate

**5. Basic Valuation**
- Property value based on: base value + demand modifier + upgrade value
- If many players want apartments in a district, prices rise
- Visible "hot market" / "buyer's market" indicators

### Economic Design Note
Properties are a crucial money sink — taxes, maintenance, and mortgage interest all remove currency from the economy. They also create the first real wealth inequality between players, which is essential for competition and motivation.

### Risk Management
- Maximum properties per player (Layer 3): 2 (prevents monopolization early)
- Anti-squatting: properties left empty for 60+ game days lose value and eventually become available
- Insurance system (optional, small monthly cost): protects against random damage events

### Success Criteria
- ≥30% of active players own property
- Property market shows genuine price movement (not static)
- At least one player-to-player property sale occurs within 2 weeks of launch

---

## LAYER 4 — PLAYER BUSINESSES
**Duration:** 10–16 weeks | **Difficulty:** High | **Economic Risk:** High | **Player Value:** Very High

### Purpose
This is the layer that creates interdependence. When players can employ other players, the game becomes a real economy.

### New Features

**1. Business Registration**
- Cost: ⊄5,000 registration fee (money sink)
- Business types:
  - Service Business (restaurant, shop, cleaning service)
  - Production Business (manufacturing, farming)
  - Professional Service (accounting, legal, consulting)
  - Transport & Logistics
  - Real Estate Business (property management)
- Each type has different mechanics, costs, and revenue models

**2. Employment System**
- Business owners post job listings
- Players can apply to player-run jobs
- Owner sets salary (must be ≥ minimum wage set by government/council)
- Employment contract: role, salary, probationary period
- Firing: requires 3-day notice (game days), or instant with severance

**3. Business Operations**
- Daily revenue determined by: type of business + number of employees + skill levels of employees + market conditions
- Revenue is not automatic — requires owner to "open" each game day (button click or scheduled task)
- Expenses: payroll, rent (if using commercial property), supplies, taxes

**4. Simple Supply Chain**
- Production businesses buy inputs from NPC suppliers at fixed prices
- Service businesses require nothing (pure labor)
- Later layers add player-to-player supply chains

**5. Business Finances**
- Separate business account from personal account
- Business credit score (separate from personal)
- Profit/loss statement (visible to owner)
- Corporate tax (15% of profits, monthly)

**6. Bankruptcy**
- If business account goes negative and cannot be rescued within 7 game days:
  - Business assets (property, inventory) liquidated
  - Employees receive 3-day severance and enter job market
  - Owner receives public bankruptcy notice (reputation impact)
  - Owner barred from starting new business for 30 game days
  - Personal assets protected unless personal guarantee given

**7. Business Reputation**
- Employers rated by employees (1–5 stars after 30 game days)
- Public business directory with ratings
- High-rated employers attract better talent

### Economic Impact
The introduction of player businesses fundamentally changes the economy:
- Wages now flow from player to player (not just bot to player)
- Business success/failure creates genuine stakes
- Economic stories emerge naturally
- Wealth inequality increases — which is correct and necessary

### Risk: Economic Capture
**Problem:** One or two players could potentially dominate the economy, making it feel unfair.

**Solutions:**
- Progressive business tax (higher profit = higher rate)
- Maximum employee count per business at Layer 4 (cap: 10)
- Anti-monopoly rule: no player may control >20% of any industry's employment
- Regular "economic audits" that are public (transparency)

### Success Criteria
- ≥10 player-run businesses active within 2 weeks
- ≥30% of employed players working at player-run businesses
- At least 3 businesses survive for 30+ game days
- At least 1 bankruptcy event (validates risk/reward)
- Player-to-player economic activity exceeds NPC-based activity

---

## LAYER 5 — PLAYER-DRIVEN MARKET
**Duration:** 8–12 weeks | **Difficulty:** High | **Economic Risk:** Very High

### Purpose
Replace fixed NPC prices with a real player market. This is where supply and demand become real.

### New Features

**1. Player Marketplace**
- Players and businesses post sell orders
- Other players post buy orders
- Matching engine clears orders at agreed price
- Transaction fee: 3% (money sink, scales with volume)

**2. Commodity Market**
- Raw materials (food ingredients, building materials, tech components)
- Prices float based on supply/demand
- Production businesses sell output here
- Service businesses buy inputs here
- NPC "floor sellers" prevent price collapse; NPC "ceiling buyers" prevent hyperinflation

**3. Price Discovery**
- Price history for every commodity (last 30 game days)
- Volume data
- Moving average lines
- Supply/demand imbalance indicator

**4. Contracts**
- Players can enter supply contracts (guaranteed price, guaranteed volume)
- Breaking a contract: penalties + reputation damage
- Contracts create stability for businesses

**5. Market Manipulation Safeguards**
- Individual player may not own more than 30% of any commodity's daily volume
- Large sudden price swings trigger a "circuit breaker" (halt trading for 4 game hours)
- Anomaly detection (see Anti-Exploit section)
- Public order book (reduces information asymmetry)

### Economic Design Note
The moment player markets replace NPC prices is the moment the game becomes real. A shortage of food ingredients will raise restaurant costs, raise menu prices, lower customer spending, reduce restaurant profits, potentially cause layoffs, increase unemployment, reduce consumer spending elsewhere. This cascade is emergent — not scripted.

### Success Criteria
- Market volume replaces NPC shop as primary source of commodity exchange
- At least 3 commodity categories show genuine price movement
- No hyperinflation event within first 60 game days
- At least one arbitrage opportunity detected (players making money from price differences)

---

## LAYERS 6–12 (SUMMARY — DETAILED SPEC IN LATER DOCUMENTS)

**LAYER 6 — FINANCIAL MARKETS**
- Player company IPOs and share trading
- Bond market (government and corporate)
- Investment portfolios
- Dividends, price discovery
- Market index tracking

**LAYER 7 — SOCIAL WORLD**
- Player organizations (unions, associations, clubs)
- Organization finances, membership
- Inter-organization relationships
- Reputation system formalization
- Alliances and rivalries
- Organization events

**LAYER 8 — GOVERNMENT & POLITICS**
- Economic Council elections (players vote on policy)
- Tax rate control (income, corporate, property)
- Minimum wage setting
- Public spending allocation
- City treasury management
- Policy debate channels

**LAYER 9 — CRIME & LAW**
- Theft, fraud, black market
- Investigation mechanics (risk, discovery probability)
- Police force (NPC + optional player police)
- Prison (time-based, limits commands)
- Criminal record (affects opportunities, loan rates)
- Criminal organizations (Layer 11 territory)

**LAYER 10 — ADVANCED WORLD SIMULATION**
- Economic cycles (boom/recession driven by data)
- NPC population behaviors
- Industry sector health metrics
- Migration patterns (new areas/districts unlock)
- Major world events (stock market crash, infrastructure boom)

**LAYER 11 — ADVANCED INSTITUTIONS**
- Player-run banks (regulated by Council)
- Player-run universities
- Player-run media companies (news, reputation influence)
- Player-run insurance companies
- Criminal organizations (structured, hierarchical)
- Corporate mergers and acquisitions

**LAYER 12 — MATURE LIVE SERVICE**
- Seasonal events
- New districts and industries
- Prestige system (generational wealth)
- Player-created content initiatives
- Long-term legacy systems
- Economic history archive

---

# DOCUMENT 9 — ECONOMIC MODEL

## 9.1 Core Economic Philosophy

The economy is designed to be **stable at scale** and **interesting at the individual level**. Inflation must be controlled, inequality must be meaningful but not crushing, and new players must always have a path to meaningful participation.

## 9.2 Money Creation (Faucets)

| Source | Type | Monthly Volume (estimate) | Notes |
|---|---|---|---|
| NPC employer wages | Primary | High | Main source of new money. Scales with active players. |
| Government spending | Secondary | Medium | Enters via public contracts/services. Layer 8+ only. |
| New player starting funds | Tertiary | Low | ⊄2,500 per new player. Trivial at scale. |
| Interest income | Minor | Low | Bank pays interest; must be funded from sinks. |

**CRITICAL:** NPC wages are the primary faucet. The bot pays players for NPC work. This money must be drained as fast as it enters, or inflation occurs.

## 9.3 Money Destruction (Sinks)

| Sink | Rate | Type | Notes |
|---|---|---|---|
| P2P Transfer Fee | 2% of transfer | Transaction | Scales with economy activity. |
| Marketplace Fee | 3% of sale | Transaction | Primary volume sink. |
| Property Tax | 0.1% of property value/game month | Recurring | Scales with wealth held. |
| Housing Rent (to NPC) | ⊄400–800/game month | Recurring | Only NPC landlord rent is a sink; player-to-player is a transfer. |
| Business Operating Costs | 10–20% of revenue | Recurring | Partially NPC-directed, partially sink. |
| Corporate Tax | 15% of profits | Recurring | Significant at scale. |
| Loan Interest (NPC bank) | 8–15% annual | Recurring | Interest paid to NPC bank = destroyed. |
| Education Courses (NPC) | ⊄500–2,000 | One-time | NPC course fees are sinks. |
| Business Registration | ⊄5,000 | One-time | |
| Crime Fines | Variable | Event-based | |
| Prison Maintenance | ⊄50/game day | Recurring | While imprisoned. |
| Inflation Tax | Optional — automatic wealth tax above ⊄1M | Emergency | Safety valve; explain as "wealth management fee." |

**Design Rule:** For every ⊄1,000 created by NPC wages, approximately ⊄700–900 should be destroyed by sinks, with the remainder circulating between players. This ensures currency has value without severe deflation.

## 9.4 Inflation Control System

```
ECONOMIC HEALTH MONITOR (runs every real-world day):

Metrics Tracked:
- Total currency in circulation (sum of all wallets + accounts)
- Average wealth per player
- Price index (basket of 10 common goods)
- Velocity (transactions per player per day)
- Gini coefficient (wealth inequality)

Target Ranges:
- Monthly price inflation: 0–3%
- Player wealth growth: 5–15% per game month
- Gini coefficient: 0.35–0.55 (meaningful inequality, not extreme)

Automatic Levers (no human intervention needed):
- If inflation > 5%: Increase transaction fees by 1%, reduce NPC wages by 5%
- If deflation < -2%: Decrease transaction fees by 1%, trigger economic stimulus event
- If Gini > 0.7: Trigger wealth tax event (government decision point)
- If new player wealth gap > 100x average: New player subsidy event
```

## 9.5 Wealth Inequality Design

**Should inequality exist?** Yes. Without it, there is nothing to aspire to and nothing to compete for.

**How much?** A healthy inequality range creates motivation without making new players feel the game is rigged. The Gini coefficient target of 0.35–0.55 is roughly comparable to real-world moderate inequality (similar to the EU average, well below extremes).

**Rich Player Advantage Management:**
- Wealthy players have more to lose (property taxes, maintenance, employee costs)
- Anti-monopoly caps prevent total capture of any industry
- Expensive assets appreciate, but also carry ongoing costs
- Political systems can redistribute wealth through taxes
- Criminal vulnerability: rich players are more visible targets

**New Player Catch-Up:**
- Starting bonus creates meaningful early traction
- Skill progression is available to all (time, not money, is the bottleneck)
- NPC employment always available (never locked out of income)
- Mentorship mechanic: established players can mentor new players (bonus for both)
- Housing subsidy for first 30 days (eliminates housing pressure early on)

## 9.6 Currency Stability Rules

- **No crypto or volatile assets in Layer 1–6**
- Stable coin price: 1 City Credit always = fixed purchasing power at NPC level
- Player market prices float; NPC floor prevents complete commodity collapse
- Government can intervene: economic council votes to adjust tax rates, releasing or absorbing currency

## 9.7 Recommended Economic Model: Hybrid

| Component | Real Simulation | Abstraction | Why |
|---|---|---|---|
| Currency | Yes — fixed denominations, trackable | No fantasy gold | Players should understand the numbers |
| Supply/demand | Yes — real player orders | Price formula with player influence | Perfect player-only markets need many active players to work; NPC liquidity fills gaps |
| Employment | Abstracted daily paycheck | Not per-hour granularity | Per-hour is too granular for Discord |
| Property valuation | Simplified hedonic model | Not full appraisal | Good enough to be believable |
| Business revenue | Formula-based with many inputs | Not full P&L simulation | Complexity vs. gameplay value tradeoff |
| Financial markets | Real player trading | Simulated NPC background traders | Provides liquidity in early game |
| Economic cycles | Triggered by real aggregate data | Not autonomous AI | Safer, more controllable |

---

# DOCUMENT 10 — TECHNICAL ARCHITECTURE

## 10.1 Technology Selection

**Critical Context:**
- Discord requires sharding at 2,500+ guilds [VERIFIED from Discord documentation]
- The system must support a global (not per-server) world
- Solo developer must be able to ship Layer 1 in weeks, not months
- Architecture must survive from 100 users to 1,000,000 users without complete rewrites

## 10.2 Recommended Stack (Progressive)

### Phase 1 (Layer 1–2, 0–500 players)

| Layer | Technology | Reason |
|---|---|---|
| Language | TypeScript / Node.js | Best Discord.js ecosystem, solo-developer productivity |
| Discord Library | Discord.js v14 | Most mature, most maintained [VERIFIED — 14.26.4 as of 2026] |
| Database | PostgreSQL | Relational data fits perfectly; scales to millions of players |
| ORM | Prisma | Type-safe, excellent migrations, works solo |
| Hosting | Railway or Fly.io | Zero-ops, always-on, easy deploys |
| Cache | In-memory (bot cache) | Redis not needed at this scale |
| Queue | None initially | Add BullMQ when background jobs needed |
| Logging | Pino | Fast structured logging |

**Cost estimate Phase 1:** ⊄$10–30/month

### Phase 2 (Layer 3–5, 500–10,000 players)

| Layer | Technology | Change |
|---|---|---|
| Cache | Redis | Needed for shared state across shards |
| Queue | BullMQ + Redis | Daily end-of-day job processing |
| Sharding | Discord.js internal sharding | Required at 2,500+ guilds |
| Database | PostgreSQL + read replicas | Write-primary, read-replica pattern |
| Monitoring | Grafana + basic alerts | Know when economy breaks |
| Web Dashboard | Next.js | Player-facing web interface |

**Cost estimate Phase 2:** $100–400/month

### Phase 3 (Layer 6+, 10,000–1,000,000 players)

| Layer | Technology | Change |
|---|---|---|
| Architecture | Microservices (split economy, bot, market) | Scale independently |
| Message Queue | NATS JetStream | [REPORTED — 200k-400k messages/sec, 1-5ms latency] |
| Database | Sharded PostgreSQL + analytics DB | Separate OLTP and analytics |
| Hosting | Kubernetes / Docker Swarm | Orchestration required |
| CDN | Cloudflare | Bot assets, web dashboard |
| Analytics | ClickHouse or BigQuery | Economic analytics at scale |

**Cost estimate Phase 3:** $1,000–10,000+/month

## 10.3 Database Schema (Conceptual)

### Core Entities

```
Player
  - id (Discord user ID)
  - username
  - wallet_balance
  - savings_balance
  - credit_score
  - reputation_score
  - created_at
  - last_active_at

Employment
  - player_id
  - employer_id (Company or NPC)
  - role
  - salary_per_game_day
  - start_date
  - end_date
  - status (active/terminated/resigned)

Company
  - id
  - owner_player_id
  - name
  - type (service/production/professional/transport/realestate)
  - status (active/bankrupt/dormant)
  - checking_balance
  - credit_score
  - registered_at

Property
  - id
  - type (residential/commercial/industrial)
  - district
  - base_value
  - current_owner_id
  - purchase_price
  - purchase_date
  - monthly_tax_rate
  - maintenance_cost

Transaction
  - id
  - from_entity_id (player or company)
  - to_entity_id
  - amount
  - type (transfer/salary/tax/purchase/loan_payment/fee)
  - created_at
  - game_day

Loan
  - id
  - borrower_id
  - lender_id (NPC bank or player)
  - principal
  - interest_rate
  - outstanding_balance
  - monthly_payment
  - status (current/overdue/defaulted/paid)

MarketOrder
  - id
  - player_id
  - commodity_type
  - order_type (buy/sell)
  - quantity
  - price_per_unit
  - filled_quantity
  - status (open/filled/cancelled)
  - created_at

Skill
  - player_id
  - skill_type (physical/service/technology/business/finance/legal/political/criminal)
  - level (0–100)
  - xp_accumulated

GameDay
  - id
  - game_day_number
  - real_world_timestamp
  - daily_event_id
  - economic_modifier

EconomicSnapshot
  - game_day_id
  - total_currency_supply
  - average_player_wealth
  - gini_coefficient
  - price_index
  - transaction_volume
  - employment_rate
```

## 10.4 Multi-Server Architecture Decision

### Options Compared

**Option A — Each server has its own economy**
- Pro: Isolation, community control
- Con: No global world, not a real city, no cross-server stories
- **Verdict: Rejected.** This is what every existing bot does. It's not special.

**Option B — Shared global economy, all servers**
- Pro: True persistent world, cross-server interactions, global stories
- Con: Server administrators lose control, privacy concerns, one server's community can't "own" their economy
- **Verdict: Recommended for global game economy.**

**Option C — Per-server hub with global trade layer**
- Too complex for early development. Could be considered at Layer 10+.

**Option D — Hybrid (Recommended)**
- One global economy for all game mechanics
- Each Discord server gets a "city district" identity (flavoring only)
- Server admins can see how their members are performing
- Server-specific leaderboards AND global leaderboards
- Players belong to the global world, but their server is their "neighborhood"

**Architecture: Global World, Local Neighborhood**

This gives server admins a reason to install the bot ("our members are the #3 wealthiest neighborhood in the city") while maintaining the global interconnected economy that is the product's differentiator.

---

# DOCUMENT 11 — UX SPECIFICATION

## 11.1 Design Principles

1. **Button-first.** Players should never need to memorize commands. Every major action is reachable through buttons after one slash command.
2. **One embed, one decision.** Each embed presents one clear set of choices. No wall of text.
3. **Progress is always visible.** Every interaction shows the player how they're doing.
4. **Private by default, public when meaningful.** Balance, skills, job — private. Reputation, wealth ranking, business directory — public.

## 11.2 Key Interface Examples

### New Player Onboarding
```
/start

┌─────────────────────────────────────────────┐
│ 🏙️ Welcome to AETHER CITY                   │
│                                              │
│ You've arrived in a living city of real      │
│ players. Your decisions will shape this      │
│ world — and others will shape yours.         │
│                                              │
│ Starting Balance: ⊄2,500                    │
│ Housing: Free for 30 days                   │
│                                              │
│ First, choose a starting path:               │
│                                              │
│ [🔨 Worker]  [💼 Professional]  [🚀 Hustler] │
└─────────────────────────────────────────────┘

Worker: Higher starting wage, physical/service skills
Professional: Lower starting wage, qualifications bonus, faster career progression
Hustler: Average wage, starting bonus in business skill, minor crime unlock
```

### Profile Dashboard
```
/profile

┌─────────────────────────────────────────────┐
│ 👤 MARCUS CHEN  •  City Citizen             │
│ ─────────────────────────────────────────── │
│ 💰 Wallet:    ⊄4,280                        │
│ 🏦 Savings:   ⊄18,500 (+⊄0.76 today)       │
│ 📊 Net Worth: ⊄22,780                       │
│ ⭐ Credit:    742 (Good)                     │
│ 🌟 Rep:       680 (Respected)                │
│ ─────────────────────────────────────────── │
│ 💼 JOB: Senior Service Worker               │
│    Employer: Silvio's Kitchen               │
│    Daily Pay: ⊄285 (next: 4h 22m)           │
│ ─────────────────────────────────────────── │
│ 🎯 SKILLS                                   │
│    Service:     Lv.5 ██████░░ 68/100 XP     │
│    Technology:  Lv.2 ███░░░░░ 31/100 XP     │
│ ─────────────────────────────────────────── │
│ [💼 Job]  [🏦 Bank]  [📈 Portfolio]  [⚙️ More] │
└─────────────────────────────────────────────┘
```

### Bank Interface
```
/bank

┌─────────────────────────────────────────────┐
│ 🏦 AETHER CITY NATIONAL BANK                │
│ ─────────────────────────────────────────── │
│ Checking:  ⊄4,280  [Deposit] [Withdraw]     │
│ Savings:   ⊄18,500  APY: 1.5%              │
│ ─────────────────────────────────────────── │
│ 💳 CREDIT                                   │
│    Score: 742 (Good)                         │
│    Max Loan: ⊄28,000                        │
│    Current Debt: None                        │
│ ─────────────────────────────────────────── │
│ 📋 RECENT TRANSACTIONS                      │
│ + ⊄285  Salary (Silvio's Kitchen)  2h ago   │
│ - ⊄400  Rent payment              1d ago    │
│ + ⊄285  Salary (Silvio's Kitchen)  1d ago   │
│ - ⊄50   Transfer to @alex        2d ago    │
│ ─────────────────────────────────────────── │
│ [Transfer] [Apply for Loan] [Full History]  │
└─────────────────────────────────────────────┘
```

### Business Owner Dashboard
```
/business

┌─────────────────────────────────────────────┐
│ 🏢 CHEN'S CATERING CO.                      │
│ Est. Game Day 342 • Rating: ⭐⭐⭐⭐½         │
│ ─────────────────────────────────────────── │
│ 💰 Today's Revenue: ⊄1,840                  │
│ 💸 Payroll:        ⊄980                     │
│ 📦 Supplies:       ⊄280                     │
│ 🏛️ Taxes:          ⊄105 (est.)              │
│ ─────────────────────────────────────────── │
│ NET PROFIT TODAY:  ⊄475                     │
│ ─────────────────────────────────────────── │
│ 👥 EMPLOYEES (5/10)                         │
│   Alex K.    Senior Chef      ⊄350/day      │
│   Sam R.     Cook             ⊄220/day      │
│   Priya T.   Cook             ⊄210/day      │
│   Jordan L.  Server           ⊄195/day      │
│   Kim P.     Server           ⊄205/day      │
│ ─────────────────────────────────────────── │
│ [Post Job] [Manage Staff] [Finances] [More] │
└─────────────────────────────────────────────┘
```

### Market Interface
```
/market commodity:food_ingredients

┌─────────────────────────────────────────────┐
│ 📊 FOOD INGREDIENTS MARKET                  │
│ ─────────────────────────────────────────── │
│ Current Price:  ⊄24.50/unit                 │
│ 24h Change:     ▲ +1.2% (⊄0.29)            │
│ 7d High:        ⊄26.80                      │
│ 7d Low:         ⊄22.10                      │
│ ─────────────────────────────────────────── │
│ TOP SELLERS                                  │
│  Green Valley Farm    @farmer  ⊄24.00 x500  │
│  NPC Wholesale Co.   NPC      ⊄25.50 x∞     │
│  Riverside Produce   @river   ⊄24.50 x200   │
│ ─────────────────────────────────────────── │
│ YOUR ORDERS: None                            │
│ ─────────────────────────────────────────── │
│ [Buy] [Sell] [Price History] [Set Alert]    │
└─────────────────────────────────────────────┘
```

## 11.3 When Web Dashboard is Better Than Discord UI

Use Discord UI for:
- Real-time actions (work, bank, buy, sell)
- Quick profile checks
- Notifications and alerts
- Social interactions

Use Web Dashboard for:
- Complex charts (price history, economic trends)
- Business analytics (revenue over time, employee performance)
- Property portfolio management
- Investment portfolio overview
- Economic research (city statistics)
- Administrative tools

---

# DOCUMENT 12 — TESTING STRATEGY

## 12.1 Layer-by-Layer Testing Protocol

### Layer 1 Testing

**Internal Testing (Week 1–2):**
- Developer plays solo with test accounts
- Test all 7 core commands
- Test end-of-day processing job
- Verify money creation/destruction balance
- Test transfer fee, banking interest
- Simulate 30-day game period with 5 test accounts

**Closed Alpha (Week 3–4, 10–25 testers):**
- Invite trusted community members or developer friends
- Provide structured feedback form (Google Form or Discord form)
- Key questions:
  - "What did you do in your first 5 minutes?"
  - "Did you understand what to do next without reading documentation?"
  - "What confused you?"
  - "What was the most satisfying moment?"
  - "Did you feel like doing anything that the bot didn't support?"
  - "Would you come back tomorrow? Why or why not?"
- Monitor: command frequency, error rates, session duration

**Open Beta (Week 5–6, 50–200 testers):**
- Open invite in 2–3 Discord communities
- Track: DAU, 7-day retention, most-used commands, error patterns
- Run economic sanity checks daily (total currency supply, average wealth)

**Exploit Testing:**
- Hire or recruit 2–3 "red team" testers specifically to break the economy
- Brief: "Find any way to generate money that isn't intended"
- Reward: Significant in-game wealth for the first exploit found and reported

**Launch Criteria for Layer 1:**
✅ Zero confirmed exploit allowing money generation > 2x intended rate
✅ 7-day retention ≥ 25%
✅ At least 50 players with 5+ sessions in beta
✅ End-of-day economic processing completing reliably (< 1% failure rate)
✅ All 7 core commands functional with < 500ms response time

## 12.2 Behavioral Measurement vs. Opinion Collection

**Trust behavior, not opinions:**
- A player saying "the game is great" while never returning is a bad sign
- A player saying "the game is confusing" while returning every day is a good sign
- Measure what players DO, not what they SAY

**Key Behavioral Metrics:**

| Metric | Layer 1 Target | Layer 4 Target | Year 2 Target |
|---|---|---|---|
| Day-1 Retention | 60% | 65% | 70% |
| Day-7 Retention | 25% | 35% | 45% |
| Day-30 Retention | 10% | 20% | 30% |
| Commands/Session | 3+ | 5+ | 8+ |
| Sessions/Week | 3+ | 4+ | 5+ |
| % Players Banking | 30%+ | 50%+ | 65%+ |
| % Players Employed | 70%+ | 80%+ | 80%+ |
| P2P Transactions | 10%+ | 30%+ | 50%+ |

---

# DOCUMENT 13 — MONETIZATION

## 13.1 Core Principle: Never Pay-to-Win

Any player who pays money must NOT have an economic advantage over a player who doesn't. Violation of this principle destroys the economy and community trust simultaneously.

## 13.2 Revenue Streams

### Tier 1: Cosmetics (Available from Layer 1)

**Personal Cosmetics (Non-Pay-to-Win):**
- Profile themes and custom profile cards: $2.99–4.99 one-time
- City Credit currency display style (changes how numbers look, not values)
- Custom profile bio and title frames
- Animated profile badges

**Estimate:** 5% of active players × $3.99 average = low but growing

### Tier 2: Convenience Features (No Economic Advantage)
- "City Pass" subscription: $4.99/month
  - Shorter work command cooldown (visual only — still 1 game day)
  - Priority customer service
  - Early access to new features (non-economic)
  - Extended transaction history (last 50 vs. last 20)
  - Business analytics dashboard access (premium view)
  - NOT: more money, faster skill gain, or better market access

**Estimate:** 2% of active players × $4.99/month = meaningful recurring revenue

### Tier 3: Server Subscriptions

Server administrators pay for enhanced integration:
- **District Pro**: $9.99/server/month
  - Custom neighborhood name and branding
  - Server-specific economic leaderboard
  - Neighborhood economic stats in server
  - Priority bot response for server members
  - Server event hosting tools

- **District Enterprise**: $24.99/server/month
  - White-label economy (custom city name for server)
  - Advanced analytics dashboard for admins
  - Priority support
  - Custom district events

**Target:** Large Discord communities (10,000+ members) are the primary target for server subscriptions.

### Tier 4: Cosmetic Business Features

For business owners:
- Business "Verified" badge: $1.99/month (display only)
- Custom business banner/logo in bot interface: $2.99/month
- Premium job listing placement (visual only): $0.99/week

### Tier 5: Supporter Packages

One-time purchases:
- "Founding Citizen" package ($19.99): Exclusive founding badge, historical record in game lore, cosmetic city plaque
- "City Builder" ($49.99): All cosmetics, founding badge, premium subscription 1 year

### What NOT to Sell
❌ Faster skill progression
❌ More starting money
❌ Better loan rates
❌ Market information advantage
❌ Immunity from game mechanics (crime, taxes, unemployment)
❌ Exclusive economic features

## 13.3 Revenue Projections (SPECULATION — labeled as estimates)

| Player Stage | Players | Est. Revenue/Month |
|---|---|---|
| 1,000 active | 1,000 | $500–2,000 |
| 10,000 active | 10,000 | $5,000–20,000 |
| 100,000 active | 100,000 | $50,000–200,000 |
| 1,000,000 active | 1,000,000 | $500,000–2,000,000 |

Infrastructure costs at scale will be significant. Assume 40–60% infrastructure/operations cost ratio at scale. [SPECULATION — varies enormously based on architecture efficiency]

## 13.4 Discord Monetization Context

Discord takes 30% of subscription revenue run through their native subscription system. [VERIFIED from Discord official developer documentation]. Third-party payment processors (Stripe, etc.) charge ~3%. Direct payments through Stripe outside Discord's system keep more revenue but have less discoverability. Recommended: use Discord's system initially for discoverability, move to hybrid model at scale.

---

# DOCUMENT 14 — GROWTH STRATEGY

## 14.1 Stages of Growth

### Stage 1: 10–100 Players (Founding Community)

**Goal:** Prove the core loop works with real humans.

**Method:**
- Personal outreach to Discord communities you already belong to
- 3–5 "founding servers" who get early access and can shape the product
- Regular developer updates in a public Discord server
- Transparent development (share what you're building, why, what's next)

**Avoid:** Posting in random Discord servers, posting in server-list channels without permission, any behavior that looks like spam.

**Retention focus:** These 100 players are your product council. Talk to them constantly. They will tell you what's broken before your metrics do.

---

### Stage 2: 100–1,000 Players

**Goal:** Organic growth through player referral.

**Mechanism:**
- **Referral mechanic:** Existing players can generate an invite link. When a new player joins through their link and reaches Day 7, both get a cosmetic reward.
- **Discord bot directories:** Submit to top.gg, discordbotlist.com, bots.ondiscord.xyz — high-traffic discovery channels. [VERIFIED these exist and drive installs]
- **Reddit:** r/discordbots, r/Discord, economy game subreddits
- **Content creation:** Developer writes detailed "how it works" blog posts, YouTube explainer (even basic quality)
- **Community server:** Maintain an active developer Discord where players congregate

**Social Virality Hook:** The game must create shareable moments. When player's business goes bankrupt, when a market crash happens, when someone gets elected Mayor — these are shareable story moments. Design for them.

---

### Stage 3: 1,000–10,000 Players

**Goal:** Self-sustaining community growth.

**Mechanism:**
- Large Discord communities (50k+ member servers) installing the bot
- TikTok/YouTube content showing interesting game moments
- Discord "Activities" promotion (if applicable)
- Community managers from the player base (volunteers with special roles)
- Regular world events that generate content
- Player spotlight posts (the richest player this month, the biggest bankruptcy, the new Mayor)

**Server Admin Strategy:** Target Discord servers with 5,000–50,000 members in categories: gaming communities, economics/finance communities, roleplaying communities, university servers. These have the player density to make a shared economy interesting.

---

### Stage 4: 10,000–100,000 Players

**Goal:** Viral flywheel.

**Mechanism:**
- Creator partnerships: gaming YouTubers/TikTokers who do "economy bot" or "Discord game" content
- Regular world events that generate news within the community
- Competitive leaderboards with public recognition
- Major economic events that get community buzz ("The Great Crash of Season 3")
- Player-created content: encourage players to write about their stories
- Sponsorship of existing Discord communities (pay for installation + promotion)

**Infrastructure Note:** At 50,000+ players, solo development becomes untenable. By this stage, at least one co-developer or employee should be in place.

---

### Stage 5: 100,000–1,000,000 Players

**Goal:** Market leadership, institutional scale.

**Mechanism:**
- Press/media coverage of unusual stories from the game
- Academic interest (economics educators using the game)
- Formal creator program with revenue sharing
- International community growth (translations)
- Potential partnership with Discord itself (featured in server discovery)

---

## 14.2 Anti-Growth Mistakes to Avoid

❌ **Paying for fake reviews** — destroys trust, violates platform terms
❌ **Spamming Discord servers** — banned immediately, reputation destroyed
❌ **Pay-to-win acceleration** — short-term revenue, long-term community death
❌ **Over-promising features** — announce what exists, not what you plan
❌ **Growing faster than infrastructure** — downtime during growth = permanent player loss

---

# DOCUMENT 15 — DEVELOPMENT PLAN

## 15.1 Solo Developer Path

**Assumption:** 20–30 hours/week of focused development time.

| Phase | Duration | Focus | Milestone |
|---|---|---|---|
| Layer 0 | 4–6 weeks | Research, design, architecture | This document + full Layer 1 spec |
| Layer 1 | 8–12 weeks | Core 5 commands, economy engine, daily processing | 50 real testers engaged |
| Layer 1 Polish | 2–4 weeks | Bug fixes, UX refinement, anti-exploit | Public launch |
| Layer 2 | 6–8 weeks | Loans, career tiers, education | 200 players |
| Layer 3 | 8–12 weeks | Property system | 500 players |
| Layer 4 | 12–16 weeks | Player businesses — this is the hardest layer | 1,000 players |
| Layer 5 | 8–10 weeks | Player marketplace | 2,000 players |

**Solo Developer Total to Layer 5:** ~18–24 months [ESTIMATE]

### When to Hire

| Trigger | First Hire |
|---|---|
| 1,000 active daily players | Community manager (part-time/volunteer) |
| 5,000 active daily players | Second developer |
| 10,000 active daily players | Full-time community manager |
| 50,000 active daily players | DevOps/infrastructure specialist |

## 15.2 Small Team Path (2–3 developers)

| Phase | Duration | Focus |
|---|---|---|
| Layer 0 | 2–3 weeks | Architecture, design |
| Layers 1–2 | 8–10 weeks | Parallel development |
| Layer 3 | 4–6 weeks | Property |
| Layer 4 | 6–8 weeks | Businesses (most complex) |
| Layer 5 | 4–6 weeks | Markets |

**Small Team Total to Layer 5:** ~9–13 months [ESTIMATE]

## 15.3 Key Technical Risks

1. **End-of-day processing complexity** — runs for all players simultaneously, must be transaction-safe
2. **Economy drift** — inflation/deflation going undetected
3. **Discord API rate limits** — sharding and pagination must be correct
4. **Data integrity** — money duplication from race conditions is catastrophic
5. **Rollback capability** — must be able to reverse any 24-hour period of economic activity

---

# DOCUMENT 16 — RISK REGISTER

## 16.1 Critical Risks

| Risk | Probability | Impact | Mitigation |
|---|---|---|---|
| Money duplication exploit | High | Catastrophic | All financial transactions in atomic database transactions; comprehensive audit logging; anomaly detection; 48h rollback capability |
| Botting / automation | High | High | Rate limits, CAPTCHA events, behavior anomaly detection, account age requirements for advanced features |
| Multi-accounting | Medium | High | Discord account linking, cross-account transfer detection, IP logging (privacy-compliant) |
| Runaway inflation | Medium | High | Automated economic monitor; daily sink/faucet balance check; emergency government vote mechanisms |
| Economy stagnation | Medium | Medium | NPC liquidity provision; developer-triggered world events; regular content additions |
| Community toxicity | Medium | Medium | Clear code of conduct, moderation tools, mute/ban integration |
| Developer burnout (solo) | Very High | Catastrophic | Hard scope discipline; one layer at a time; public roadmap accountability; find co-developer early |
| Discord policy change | Low | High | Build web presence and direct player relationships independent of Discord from Day 1 |
| Competitor launches similar product | Medium | Medium | First-mover advantage; community depth; player investment too high to abandon |
| Player data privacy breach | Low | Catastrophic | Minimal personal data collected (Discord ID only); encrypt sensitive data; privacy-first architecture |

## 16.2 Economy-Specific Risks

**Hyperinflation:**
- Cause: Too many faucets, too few sinks, rapid player growth
- Early warning: Price index rising >5% per game month
- Response: Automatic fee increase + emergency government vote to raise taxes
- Rollback plan: Manual monetary policy adjustment by developer

**Oligopoly Formation:**
- Cause: 1–2 players accumulate most of the economy
- Early warning: Gini coefficient >0.70
- Response: Progressive taxation + anti-monopoly caps
- Nuclear option: Wealth redistribution event (framed as in-game economic crisis)

**Player Exodus Event:**
- Cause: Catastrophic exploit, unfair developer decision, competitor launch
- Response: Transparent communication, rollback if possible, emergency content event
- Prevention: Never make major economic policy changes without player Council vote

---

# DOCUMENT 17 — FEATURE PRIORITY MATRIX

## 17.1 Prioritization Framework

Each feature scored 1–5 on: Player Value (PV), Retention Value (RV), Social Value (SV), Revenue Potential (RP), Development Effort (DE, inverted — 5 = easy, 1 = very hard), Technical Risk (TR, inverted — 5 = low risk, 1 = very risky).

**Priority Score = (PV + RV + SV + RP + DE + TR) / 6**

| Feature | PV | RV | SV | RP | DE | TR | Score | Priority |
|---|---|---|---|---|---|---|---|---|
| Player wallet/bank | 5 | 5 | 3 | 3 | 5 | 5 | 4.3 | MUST |
| NPC jobs | 5 | 5 | 2 | 2 | 4 | 4 | 3.7 | MUST |
| P2P transfers | 4 | 4 | 5 | 2 | 5 | 4 | 4.0 | MUST |
| Loan system | 4 | 4 | 2 | 3 | 4 | 3 | 3.3 | SHOULD |
| Career progression | 4 | 5 | 2 | 2 | 3 | 4 | 3.3 | SHOULD |
| Property system | 5 | 5 | 4 | 4 | 2 | 3 | 3.8 | SHOULD |
| Player businesses | 5 | 5 | 5 | 4 | 1 | 2 | 3.7 | SHOULD |
| Player marketplace | 4 | 5 | 4 | 3 | 2 | 2 | 3.3 | SHOULD |
| Stock market | 4 | 4 | 3 | 3 | 2 | 2 | 3.0 | COULD |
| Organizations | 4 | 4 | 5 | 2 | 2 | 3 | 3.3 | COULD |
| Crime system | 3 | 3 | 4 | 2 | 2 | 2 | 2.7 | COULD |
| Government/elections | 3 | 4 | 5 | 2 | 1 | 2 | 2.8 | COULD |
| NPC world simulation | 3 | 4 | 2 | 2 | 1 | 2 | 2.3 | LATER |
| International trade | 2 | 2 | 3 | 2 | 1 | 1 | 1.8 | AVOID NOW |
| Financial derivatives | 2 | 2 | 2 | 2 | 1 | 1 | 1.7 | AVOID NOW |
| Complex health system | 2 | 2 | 1 | 1 | 2 | 3 | 1.8 | AVOID NOW |
| Vehicle system | 2 | 2 | 1 | 2 | 2 | 3 | 2.0 | AVOID NOW |

## 17.2 Do Not Build Yet List (Mandatory)

These features sound compelling but are premature. Building them early will kill the project.

**1. Political system (before Layer 7 is stable)**
*Why:* Requires large player base to be meaningful. Empty elections with 50 players are embarrassing. Adds enormous complexity. Build social organizations first.

**2. Crime system (before Layer 4)**
*Why:* Crime needs an economy worth robbing. Before businesses and property exist, crime is just "random negative event."

**3. Vehicles**
*Why:* High development cost, low strategic value, zero economic interdependence. Adds inventory complexity with minimal gameplay depth.

**4. Complex health systems**
*Why:* Potentially frustrating mechanic. Punishing players for playing less is anti-retention. Abstract at most.

**5. International trade / multiple cities**
*Why:* You need one city to work first. Adding regions before the core is stable is scope explosion.

**6. Crypto / blockchain integration**
*Why:* Regulatory risk, demographic narrowing, technical complexity, reputation risk among mainstream players. The game currency should be fictional and closed-loop.

**7. Real-money stock market integration**
*Why:* Legal/regulatory nightmare. Keep financial markets in-game only.

**8. Full weather/environmental systems**
*Why:* Cosmetic complexity masquerading as gameplay. Abstract to "economic conditions."

**9. Large crafting trees (Layer 1–3)**
*Why:* Before a player market exists, crafting is just "wait and receive reward." Add after markets.

**10. Relationship/dating simulation**
*Why:* Scope explosion with social risk and moderation complexity. The economy is the story; personal relationships are emergent from it, not designed.

---

# DOCUMENT 18 — FINAL RECOMMENDATION

## 18.1 Is This Worth Building?

**Yes. With significant caveats.**

The market gap is real and verified. No existing Discord economy bot creates player interdependence at meaningful depth. The concept of a persistent, player-driven economic world in Discord is genuinely novel and genuinely appealing to a real audience.

However, the vision described in the brief is a 5–7 year project. Attempting to build it in 12 months will produce an unfinished mess. The only path to success is strict layered development — building the simple thing first, proving it works, and adding complexity when the foundation justifies it.

## 18.2 The Strongest Version of the Idea

**One shared world. Real interdependence. Stories that emerge.**

Not: "a Discord bot with 50 economy features."
Not: "a more complex version of Dank Memer."
Not: "a Discord game with business simulation."

The correct framing: **"A Discord-native persistent multiplayer city where players are economically connected to each other."**

The killer feature is not any individual mechanic. The killer feature is that what Player A does matters to Player B. This must be true from Layer 4 onward, and the entire architecture must be built to make it true.

## 18.3 What Should Be Removed From the Vision

- Real-world market data integration (use fully fictional economy)
- Cryptocurrency/blockchain (legal risk, complexity, audience narrowing)
- Complex health/happiness simulation (abstraction is better)
- Vehicle system (no meaningful economic role)
- Dating/relationship simulation (scope, moderation, risk)
- Per-server economy mode (defeats the purpose)
- Excessive crafting complexity before markets exist

## 18.4 What Should Be Emphasized

- **Global shared economy** — the non-negotiable differentiator
- **Player-as-employer mechanics** — this is what creates stories
- **Reputation systems** — this is what creates communities
- **Economic transparency** — all economic data visible, creates engagement
- **Emergent narrative** — the economy tells stories that players share

## 18.5 What Should Layer 1 Contain?

Exactly 7 commands. No more.

```
/start     — Onboarding flow, creates player, starting funds
/profile   — Dashboard showing balance, job, skills, net worth
/work      — Collect daily wages from current job; see job options if unemployed  
/bank      — Deposit, withdraw, transfer, view history
/shop      — NPC shop for housing and basic needs
/city      — Today's world news and economic event
/leaderboard — Top 10 wealthiest players
```

Plus a background job:
- End-of-day processor: applies salaries, interest, housing costs, market events

## 18.6 Commercial Viability

**The product is commercially viable IF:**
- The global economy creates enough player retention to reach 50,000+ active daily players
- Monetization stays cosmetic/convenience only
- Infrastructure costs are managed through the layered approach (small scale = small cost)
- At least one sustainable revenue stream reaches $10,000+/month before major infrastructure investment

**The product fails commercially IF:**
- Layer 4 (player businesses) is never reached — everything before it is similar to existing products
- Player count stalls below 5,000 active users
- Economy instability destroys community trust
- Developer burns out before reaching Layer 4

**The critical milestone: Layer 4 with 1,000 active players.** At this point, player businesses, player employment, and economic interdependence exist. The product is genuinely differentiated. Everything before that point is foundation.

## 18.7 Single Biggest Competitive Advantage

**Player economic interdependence.** The moment Player A's business employs Player B, and Player B's wage comes from Player C who buys from Player A's supplier (Player D), the game has created something no other Discord bot has: **a reason to care about other players for economic reasons.**

## 18.8 Single Biggest Risk

**Scope collapse before Layer 4.** The game is not differentiated until businesses and employment exist. Everything before is similar to existing products. If development stalls at Layer 2 or 3, the project becomes another economy bot.

**Mitigation:** Make a public commitment to Layer 4. Set a public date. Build in public. Have external accountability.

## 18.9 Why Players Will Love It

They will love it because they will have a **story.** Not a score. Not a leaderboard position. A story about the company they built, the worker they promoted, the bankruptcy they survived, the investment that paid off, the election they won.

In every other economy bot, when someone asks "what did you do today?" the answer is "I pressed work and gambled."

In this game, the answer is: "I just hired my third employee, the market for raw materials crashed so my costs went up, and now I'm trying to decide whether to cut wages or raise prices before my competitor undercuts me."

**That is the reason players will stay for years.**

## 18.10 Why Players Will Leave

The most likely abandonment reason is **"nothing to do with my wealth."** This must be proactively designed against. Every layer should create new ways to deploy wealth strategically — property, businesses, loans to others, political influence, organizational funding. Wealth without purpose is just a number.

---

# THE ONE-PAGE MASTER PLAN

```
╔══════════════════════════════════════════════════════════════════╗
║          AETHER CITY — ONE PAGE MASTER PLAN                      ║
╠══════════════════════════════════════════════════════════════════╣
║ PRODUCT                                                          ║
║ A persistent, multiplayer Discord city where players' economic   ║
║ decisions affect each other. One global world. Real stakes.      ║
╠══════════════════════════════════════════════════════════════════╣
║ TARGET AUDIENCE                                                   ║
║ Primary: Economic strategists (18–30), Discord-native gamers     ║
║ Secondary: Roleplayers, Entrepreneurs, Social builders           ║
╠══════════════════════════════════════════════════════════════════╣
║ COMPETITIVE ADVANTAGE                                            ║
║ Player economic interdependence. Player A employs Player B.      ║
║ Player A's business failure affects Player B's livelihood.       ║
║ No existing Discord bot does this.                               ║
╠══════════════════════════════════════════════════════════════════╣
║ CORE GAMEPLAY LOOP                                               ║
║ Earn → Save → Invest → Build → Employ → Influence → Compete     ║
╠══════════════════════════════════════════════════════════════════╣
║ LAYER 1 (Months 1–3): FOUNDATION                                 ║
║ 7 commands. Wallet, bank, work, shop, city news, leaderboard.   ║
║ One global economy. Daily economic events. Basic skills.         ║
║ Goal: 200 players, 25% 7-day retention.                         ║
╠══════════════════════════════════════════════════════════════════╣
║ LAYER 2 (Months 4–5): PERSONAL FINANCE                          ║
║ Loans, credit scores, career tiers, education.                   ║
║ Goal: Players have something to plan toward.                     ║
╠══════════════════════════════════════════════════════════════════╣
║ LAYER 3 (Months 6–8): PROPERTY                                   ║
║ Buy, rent, mortgage. Real wealth accumulation.                   ║
║ Goal: Tangible assets, landlord/tenant relationships.            ║
╠══════════════════════════════════════════════════════════════════╣
║ LAYER 4 (Months 9–13): PLAYER BUSINESSES ← CRITICAL             ║
║ Start companies. Hire players. Set wages. Go bankrupt.           ║
║ This is the differentiator. This is where the game becomes real. ║
║ Goal: 1,000 players. 30% in player-to-player employment.        ║
╠══════════════════════════════════════════════════════════════════╣
║ LAYERS 5–12: MARKET → FINANCIAL MARKETS → SOCIAL WORLD →        ║
║ GOVERNMENT → CRIME → WORLD SIMULATION → INSTITUTIONS →          ║
║ LIVE SERVICE (Years 2–5+)                                        ║
╠══════════════════════════════════════════════════════════════════╣
║ TECHNICAL DIRECTION                                              ║
║ TypeScript + Discord.js → PostgreSQL → Redis → Microservices.   ║
║ Global world (not per-server). Web dashboard from Layer 3.       ║
║ Economic monitor runs daily. Atomic transactions always.         ║
╠══════════════════════════════════════════════════════════════════╣
║ MONETIZATION                                                     ║
║ Cosmetics only (never pay-to-win).                               ║
║ City Pass: $4.99/month. Server District: $9.99/month.           ║
║ First revenue milestone: $2,000/month at 10,000 players.        ║
╠══════════════════════════════════════════════════════════════════╣
║ GROWTH STRATEGY                                                  ║
║ L1: Personal outreach + Discord bot directories.                 ║
║ L2: Referral mechanic + Reddit/community posts.                  ║
║ L3: TikTok/YouTube + server admin outreach.                     ║
║ L4: Creator partnerships + viral moments (bankruptcies, elections║
╠══════════════════════════════════════════════════════════════════╣
║ BIGGEST RISK                                                     ║
║ Stalling at Layer 2 or 3 before businesses are built.           ║
║ Mitigation: Public commitment to Layer 4. Hard deadlines.        ║
╠══════════════════════════════════════════════════════════════════╣
║ MOST IMPORTANT METRIC                                            ║
║ % of players with at least one economic interaction with         ║
║ another player in their first 14 days.                          ║
║ Target Layer 4: 50%+                                            ║
╠══════════════════════════════════════════════════════════════════╣
║ FIRST THING TO DO TOMORROW                                       ║
║ Write the Layer 1 technical specification: define every database ║
║ table, every command, every end-of-day job. Do not write code   ║
║ until this is complete and reviewed.                             ║
╚══════════════════════════════════════════════════════════════════╝
```

---

# DOCUMENT 19 — PLAYER JOURNEYS (SELECTED)

## Journey 1 — The Beginner Who Becomes Stable

**Day 1:** Joins, confused. Presses buttons. Gets construction job. Earns ⊄180. Deposits half.
**Week 1:** Works every day. Explores bank features. Discovers savings interest. Sets a goal: save ⊄5,000.
**Month 1:** Has ⊄4,800 saved. Takes a course to improve service skill. Gets promotion to Senior Worker (⊄240/day).
**Month 3:** Has ⊄18,000. Takes a loan for apartment deposit. Owns first property. Earns rent from a new player. Feels like a real city resident.
**Moment of investment:** Receives first rent payment. Feels ownership. Now has a reason to stay.

## Journey 2 — The Entrepreneur

**Day 1:** Reads every command description. Studies leaderboard. Asks in city chat: "What businesses are most profitable?"
**Week 2:** Has mapped the economic opportunity. Food ingredients are expensive. Starts saving aggressively.
**Month 2:** Registers a food production company. Buys supplies at NPC prices. Sells food ingredients on player marketplace at 15% markup. First three customers are restaurant owners.
**Month 4:** Hires two players as production workers. Has a supply contract with the city's largest restaurant chain.
**Month 6:** Revenue exceeds personal salary by 3x. Takes business loan to expand. Faces first competitor (another player undercuts on price). Price war begins.
**Month 8:** Price war resolved through supply contract with a major buyer (mutual assured pricing). Business stabilizes.
**Year 1:** 8 employees, dominant in food production district. IPO planned for Layer 6. Company worth ⊄200,000+.

## Journey 5 — The Criminal (Layer 9+)

**Background:** Mid-level player who got bored of legitimate work.
**Month 3:** Discovers crime commands. Attempts theft from wealthy player. Fails. Pays fine. Criminal record created.
**Month 4:** Joins a small criminal organization (3 other players). More sophisticated schemes: false invoices submitted to player companies (fraud). Success rate: 60%.
**Month 5:** Gets caught. Criminal record now includes fraud. This causes:
  - Loan application denied (credit score impact)
  - Two employers refuse to consider hiring
  - Reputation drops to "Suspicious"
**Month 6:** Goes underground. Works legitimate job by day. Criminal org by night.
**Year 1:** Manages a medium criminal network. Some players know, others don't. This is a secret reputation layer — some players prefer to hire criminal players for "gray market" supply chains.
**The story:** This player's journey creates narrative conflict for the entire server. Other players speculate, investigate, cooperate with law enforcement, or do business with them. This is emergent content no developer scripted.

## Journey 10 — The Multi-System Advanced Player

**Year 2 Player — "Elena":**

On any given game day, Elena is simultaneously:
- **Corporate CEO:** Checking her tech company's daily revenue (⊄4,200), approving payroll for 12 employees, reviewing a business loan application from a competitor who wants to partner.
- **Property magnate:** 4 properties generating ⊄2,800/month in rent. Reviewing maintenance costs. Considering selling one commercial unit at peak market value.
- **Investor:** Monitoring her portfolio of 6 company stocks. One is down 8% after the city's economic council raised corporate taxes. Deciding whether to sell or hold.
- **City Councilor:** A vote is open on minimum wage increase. She knows it will raise her company's costs by ⊄400/week. She also knows her lowest-paid employees will benefit. She votes yes — because her reputation score is her most valuable asset and she knows it.
- **Lender:** Three personal loans outstanding to other players totaling ⊄45,000. One is 3 game days overdue. She sends a gentle reminder through the bot.
- **Organization leader:** Chairs the Aether City Tech Workers Association. Next meeting: debate on whether to support the minimum wage increase (as a tech company CEO, this creates internal conflict with her own council vote).

Elena has not "beaten" the game. She has become part of the game's infrastructure. Newer players take loans from her. Her employees go on to start their own companies. Her political votes shape the world newer players arrive in. This is the endgame — not a number, but a legacy.

---

# DOCUMENT 20 — FULL WORLD SIMULATION (ONE DAY, 20 PLAYERS)

**Game Day 847. A Tuesday morning in Aether City.**

The end-of-day job just ran. Here's what happened across 20 simultaneous player stories:

**PLAYER A (Chen's Catering CEO):** Revenue processed: ⊄1,840. Payroll deducted: ⊄980. Supplies from Player B's farm deducted: ⊄280. Corporate tax: ⊄105. Net profit: ⊄475.

**PLAYER B (Green Valley Farm owner):** Player A bought ⊄280 of food ingredients. Player C's restaurant chain bought ⊄560 more. Total revenue: ⊄1,820. Had 2 agricultural workers (Players M and N) to pay: ⊄410. Profits: ⊄780.

**PLAYER C (City's largest restaurant chain, 3 player co-owners):** Bought ⊄560 from Player B. Revenue from 30+ NPC and player customers: ⊄3,200. 6 employees (including Player D). Payroll: ⊄1,800. Net profit split 3 ways: ⊄280 each.

**PLAYER D (Restaurant Worker at Player C's chain):** Salary paid: ⊄310/day. Just hit Service Level 7. Qualifies for Restaurant Manager position. Applied for the open manager role at Player A's catering company.

**PLAYER E (Investor):** Holds 500 shares in Player B's farm company (listed on exchange). Farm did well today — stock up 2%. E's portfolio value increased by ⊄240. Did not actively trade; checked market from web dashboard.

**PLAYER F (Real Estate Investor):** Owns commercial unit rented to Player A for ⊄800/month. Collected automated rent. Also reviewing whether to sell — property values in the Commercial District rose 3% this game month due to high business activity.

**PLAYER G (Construction Company Owner):** Economic event today: "Infrastructure project announced." His industry boosted +15%. Revenue: ⊄2,100. Hired Player P (new to city) as entry-level construction worker.

**PLAYER H (City Councilor):** Voted on minimum wage proposal. Motion failed (6-4). Workers' union (Organization led by Player I) immediately announced a public protest. H's reputation with workers dropped -20. H's reputation with business owners rose +15.

**PLAYER I (Union Leader):** Lost the minimum wage vote. Called an organization meeting. 18 union members discuss next steps. Some want a strike (mechanics to be designed in Layer 9). Others want to run their own candidate in next election.

**PLAYER J (Financial Journalist — Media Company):** Wrote a "City Report" post in the public city news channel. Covered the minimum wage defeat. Included quotes from Player H and Player I. This is the first piece of player-created journalism in game history. 34 players reacted.

**PLAYER K (Criminal):** The infrastructure project created a construction boom. K's criminal organization is planning to submit fraudulent invoices to new construction companies. Researching targets.

**PLAYER L (New Player, Day 12):** Got first full paycheck from NPC construction job: ⊄180. Deposited ⊄120 into savings. Noticed Player G's construction company is hiring (pays ⊄220/day). Applied. Player G approved. Starting tomorrow at player company.

**PLAYER M (Farm Worker at Player B):** Received ⊄200/day salary. Saving aggressively. Goal: ⊄15,000 for apartment deposit. Currently at ⊄8,400. Also has an interview tomorrow with a larger agricultural company offering ⊄260/day.

**PLAYER N (Farm Worker at Player B):** Same situation as M. If both M and N leave Player B's farm, B's production capacity drops 40% and supply to Player A and C decreases. This will raise food ingredient prices.

**PLAYER O (Bank CEO — player-run, Layer 11):** Processing loan application from Player G who wants to expand his construction company. G's credit score: 780. Business revenue solid. O's bank approves a ⊄50,000 business loan at 9% annual interest. O's bank will earn ⊄4,500/year in interest.

**PLAYER P (Brand New Player, Day 1):** Just joined through Player L's referral link. Onboarding complete. Selected "Worker" path. Got ⊄2,500. Player G just approved their job application. Starting construction job tomorrow. First real player-to-player economic connection occurs before they've finished their first day.

**PLAYER Q (Property Developer, advanced):** Owns a block of 6 apartments. All rented. Reviewing proposal to buy an empty industrial lot and convert to apartment complex. Requires ⊄120,000 investment. Has ⊄85,000 and needs a ⊄35,000 loan from Player O's bank.

**PLAYER R (Investor — stock focus):** Noticed Player N is looking for a new job. If N leaves Player B's farm, food ingredient prices rise. R bought 200 futures contracts on food ingredients (Layer 12 mechanic) betting on price rise. [Note: This is speculative behavior emerging from public information about individual players — creates interesting privacy/transparency design questions for later layers.]

**PLAYER S (Criminal Organization Boss):** Player K reports back with the fraudulent invoice scheme. S approves. Operation begins tomorrow. Risk: Player T is investigating recent fraud patterns.

**PLAYER T (City Detective — player-run law enforcement, Layer 9):** Has been tracking unusual invoice patterns across 3 companies. Compiled evidence. Will present to the City Council next session for authorization to investigate.

**PLAYER U (New Player trying to understand the game):** Confused by the complexity of what's visible. The city news channel shows a political debate, a criminal investigation hint, market data, and someone's business bankruptcy. Needs better onboarding. [This is a real UX problem to solve.]

---

**What this simulation demonstrates:**

Every player's action connects to at least 2 other players. The minimum wage vote affects employment decisions that affect supply chains that affect prices that affect every business's profitability. A farm worker considering leaving affects a restaurant's food costs affects a catering company's pricing affects their employees' job security.

This is the web of emergent economics. It isn't scripted. It emerges from the system design. And it generates stories that players want to share.

---

*End of Research Dossier — Layer 0 Complete*
*Next Step: Write Layer 1 Technical Specification*
*Estimated time to Layer 1 launch: 8–12 weeks (solo developer)*

---

**Document Metadata:**
- Research conducted: September 2026
- Sources: Discord bot directories (top.gg, discordbotlist.com, bots.ondiscord.xyz), official bot websites, GitHub repositories, academic game economy research, EVE Online economic analysis, game design literature
- Data labels: [VERIFIED] = direct source confirmation | [REPORTED] = credible secondary source | [INFERENCE] = logical conclusion from evidence | [SPECULATION] = possible but unverified
- Scope: Layer 0 only. Layer 1 technical specification is the next deliverable.
