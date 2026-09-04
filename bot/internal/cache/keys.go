package cache

import "fmt"

// ============================================================================
// RATE LIMITING KEYS
// ============================================================================

// CommandCooldownKey returns the Redis key for a player's command cooldown.
// Format: ratelimit:cmd:<player_id>:<command>
func CommandCooldownKey(playerID, command string) string {
	return fmt.Sprintf("ratelimit:cmd:%s:%s", playerID, command)
}

// WorkDailyKey returns the Redis key for a player's daily work cooldown.
// Format: ratelimit:work:<player_id>
func WorkDailyKey(playerID string) string {
	return fmt.Sprintf("ratelimit:work:%s", playerID)
}

// ============================================================================
// MARKET KEYS
// ============================================================================

// MarketSupplyKey returns the Redis key for an item's supply signal.
// Format: market:supply:<item_type>
func MarketSupplyKey(itemType string) string {
	return fmt.Sprintf("market:supply:%s", itemType)
}

// MarketDemandKey returns the Redis key for an item's demand signal.
// Format: market:demand:<item_type>
func MarketDemandKey(itemType string) string {
	return fmt.Sprintf("market:demand:%s", itemType)
}

// MarketPriceKey returns the Redis key for an item's current market price.
// Format: market:price:<item_type>
func MarketPriceKey(itemType string) string {
	return fmt.Sprintf("market:price:%s", itemType)
}

// ============================================================================
// ECONOMY KEYS
// ============================================================================

// BaseWageRateKey is the Redis key for the current base city wage rate.
const BaseWageRateKey = "economy:base_wage_rate"

// LastSettlementKey is the Redis key for the timestamp of the last successful settlement.
// Used by the analytics health check to detect stale data.
const LastSettlementKey = "analytics:last_settlement"

// ============================================================================
// CITY / WORLD KEYS
// ============================================================================

// CityStateKey is the Redis key for the cached city state (all plots, businesses, etc.).
const CityStateKey = "city:state"

// CityPlotsKey is the Redis key for the cached plot list.
const CityPlotsKey = "city:plots"

// ActiveWorldEventKey is the Redis key for the currently active world event.
const ActiveWorldEventKey = "world:active_event"

// ============================================================================
// TAX KEYS
// ============================================================================

// HourlyTaxKey returns the Redis key for hourly tax aggregation.
// Format: tax:hourly:<YYYY-MM-DD-HH>
func HourlyTaxKey(hour string) string {
	return fmt.Sprintf("tax:hourly:%s", hour)
}

// RollingTaxKey is the Redis key for the rolling 24-hour tax total (hash).
const RollingTaxKey = "tax:rolling_24h"

// ============================================================================
// ANTI-EXPLOIT KEYS
// ============================================================================

// TransactionVelocityKey returns the Redis key for a player's transaction velocity counter.
// Format: velocity:tx:<player_id>
func TransactionVelocityKey(playerID string) string {
	return fmt.Sprintf("velocity:tx:%s", playerID)
}

// TransferRingKey returns the Redis key for tracking a player's recent transfer partners.
// Format: velocity:ring:<player_id>
func TransferRingKey(playerID string) string {
	return fmt.Sprintf("velocity:ring:%s", playerID)
}

// ============================================================================
// SESSION / AUTH KEYS
// ============================================================================

// PlayerSessionKey returns the Redis key for a player's active API session.
// Format: session:player:<player_id>
func PlayerSessionKey(playerID string) string {
	return fmt.Sprintf("session:player:%s", playerID)
}