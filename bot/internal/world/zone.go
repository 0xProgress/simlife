package world

import (
	"fmt"

	"github.com/shopspring/decimal"
)

// ZoneType represents the zoning classification of a city plot.
// Each zone type has distinct economic properties, tax rates, and allowed uses.
type ZoneType string

const (
	ZoneResidential ZoneType = "RESIDENTIAL"
	ZoneCommercial  ZoneType = "COMMERCIAL"
	ZoneIndustrial  ZoneType = "INDUSTRIAL"
	ZoneGovernment  ZoneType = "GOVERNMENT"
	ZoneMixed       ZoneType = "MIXED"
)

// ZoneConfig defines the immutable properties of a zone type.
// These are loaded at startup and cached in memory — they never change at runtime.
type ZoneConfig struct {
	Type                 ZoneType
	DisplayName          string
	MaxDevelopmentLevel  int
	BaseTaxRate          decimal.Decimal // Annual property tax rate as a decimal (e.g., 0.01 = 1%)
	RentMultiplier       decimal.Decimal // Multiplier applied to base rent calculations
	AllowedBusinessTypes []string        // Business types permitted in this zone
	Color                string          // Hex color for UI rendering (matches design tokens)
	Description          string
}

// ZoneRegistry is the authoritative in-memory registry of all zone configurations.
// Initialized once at startup and never modified.
var ZoneRegistry = map[ZoneType]*ZoneConfig{
	ZoneResidential: {
		Type:                 ZoneResidential,
		DisplayName:          "Residential District",
		MaxDevelopmentLevel:  5,
		BaseTaxRate:          decimal.NewFromFloat(0.005), // 0.5% annual
		RentMultiplier:       decimal.NewFromFloat(1.0),
		AllowedBusinessTypes: []string{}, // No businesses in pure residential
		Color:                "#F5E6D3", // Warm cream (matches design tokens)
		Description:          "Housing for Aether City residents. Generates rental income.",
	},
	ZoneCommercial: {
		Type:                 ZoneCommercial,
		DisplayName:          "Commercial District",
		MaxDevelopmentLevel:  7,
		BaseTaxRate:          decimal.NewFromFloat(0.015), // 1.5% annual
		RentMultiplier:       decimal.NewFromFloat(1.5),
		AllowedBusinessTypes: []string{"SERVICE", "RETAIL", "RESTAURANT", "OFFICE"},
		Color:                "#4A90D9", // Cool slate blue
		Description:          "Retail and office space. High foot traffic, premium rents.",
	},
	ZoneIndustrial: {
		Type:                 ZoneIndustrial,
		DisplayName:          "Industrial District",
		MaxDevelopmentLevel:  6,
		BaseTaxRate:          decimal.NewFromFloat(0.012), // 1.2% annual
		RentMultiplier:       decimal.NewFromFloat(0.8),
		AllowedBusinessTypes: []string{"MANUFACTURING", "PRODUCTION", "LOGISTICS", "FORGE"},
		Color:                "#3D4460", // Dark steel
		Description:          "Manufacturing and logistics. Lower rents, high production capacity.",
	},
	ZoneGovernment: {
		Type:                 ZoneGovernment,
		DisplayName:          "Civic Center",
		MaxDevelopmentLevel:  4,
		BaseTaxRate:          decimal.Zero, // Government properties are tax-exempt
		RentMultiplier:       decimal.Zero, // Cannot be rented
		AllowedBusinessTypes: []string{},   // No private businesses
		Color:                "#8a6fd8",    // Deep purple
		Description:          "City administration. Not available for private ownership.",
	},
	ZoneMixed: {
		Type:                 ZoneMixed,
		DisplayName:          "Mixed-Use District",
		MaxDevelopmentLevel:  6,
		BaseTaxRate:          decimal.NewFromFloat(0.010), // 1.0% annual
		RentMultiplier:       decimal.NewFromFloat(1.2),
		AllowedBusinessTypes: []string{"SERVICE", "RETAIL", "RESTAURANT", "OFFICE", "CLINIC"},
		Color:                "#d4a847", // Amber gold
		Description:          "Flexible zoning. Supports both residential and commercial use.",
	},
}

// GetZoneConfig returns the configuration for a given zone type.
// Returns an error if the zone type is unknown.
func GetZoneConfig(zoneType ZoneType) (*ZoneConfig, error) {
	cfg, ok := ZoneRegistry[zoneType]
	if !ok {
		return nil, fmt.Errorf("unknown zone type: %s", zoneType)
	}
	return cfg, nil
}

// IsValidZoneType checks if a string is a valid zone type.
func IsValidZoneType(zoneType string) bool {
	_, ok := ZoneRegistry[ZoneType(zoneType)]
	return ok
}

// IsBusinessAllowedInZone checks if a business type is permitted in a given zone.
func IsBusinessAllowedInZone(zoneType ZoneType, businessType string) bool {
	cfg, ok := ZoneRegistry[zoneType]
	if !ok {
		return false
	}
	for _, allowed := range cfg.AllowedBusinessTypes {
		if allowed == businessType {
			return true
		}
	}
	return false
}

// CalculateBasePropertyValue computes the base assessed value of a plot
// based on its zone type and development level.
// Formula: base_value * (1 + (development_level * 0.25))
func CalculateBasePropertyValue(zoneType ZoneType, developmentLevel int) decimal.Decimal {
	_, ok := ZoneRegistry[zoneType]
	if !ok {
		return decimal.Zero
	}

	// Base values per zone type (in City Credits)
	baseValues := map[ZoneType]decimal.Decimal{
		ZoneResidential: decimal.NewFromInt(25000),
		ZoneCommercial:  decimal.NewFromInt(50000),
		ZoneIndustrial:  decimal.NewFromInt(35000),
		ZoneGovernment:  decimal.NewFromInt(100000),
		ZoneMixed:       decimal.NewFromInt(40000),
	}

	base, exists := baseValues[zoneType]
	if !exists {
		return decimal.Zero
	}

	// Development multiplier: each level adds 25% to base value
	devMultiplier := decimal.NewFromFloat(1.0).Add(
		decimal.NewFromInt(int64(developmentLevel)).Mul(decimal.NewFromFloat(0.25)),
	)

	return base.Mul(devMultiplier).Truncate(0)
}