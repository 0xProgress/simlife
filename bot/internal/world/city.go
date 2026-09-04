package world

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	sqlc "github.com/0xProgress/simlife/bot/db/sqlc"
	"github.com/0xProgress/simlife/bot/internal/logger"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"
)

const (
	cityStateCacheKey   = "city:state"
	cityStateCacheTTL   = 5 * time.Minute
	cityPlotsCacheKey   = "city:plots"
	cityPlotsCacheTTL   = 2 * time.Minute
)

// CityManager manages the global city state, providing aggregated views
// for the Activity frontend and other services. City state is cached in Redis
// and invalidated when plots change.
type CityManager struct {
	queries *sqlc.Queries
	plots   *PlotManager
	redis   *redis.Client
	log     zerolog.Logger
}

// NewCityManager initializes the city manager.
func NewCityManager(q *sqlc.Queries, plots *PlotManager, r *redis.Client, log zerolog.Logger) *CityManager {
	return &CityManager{
		queries: q,
		plots:   plots,
		redis:   r,
		log:     log.With().Str("component", "world.city").Logger(),
	}
}

// CityState represents the complete city state for the Activity frontend.
type CityState struct {
	Plots              []*PlotSummary     `json:"plots"`
	Infrastructure     []Infrastructure   `json:"infrastructure"`
	OnlinePlayers      []OnlinePlayer     `json:"online_players"`
	TotalPopulation    int                `json:"total_population"`
	OccupancyRate      float64            `json:"occupancy_rate"`
	LastUpdated        time.Time          `json:"last_updated"`
}

// PlotSummary is a lightweight plot representation for the city map.
type PlotSummary struct {
	ID               string  `json:"id"`
	X                int32   `json:"x"`
	Y                int32   `json:"y"`
	Zone             string  `json:"zone"`
	ZoneColor        string  `json:"zone_color"`
	Owner            string  `json:"owner,omitempty"` // Username, empty if unowned
	DevelopmentLevel int     `json:"development_level"`
	HasBusiness      bool    `json:"has_business"`
	BusinessType     string  `json:"business_type,omitempty"`
}

// Infrastructure represents a city infrastructure project.
type Infrastructure struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Type        string `json:"type"`
	Status      string `json:"status"`
	Completion  int    `json:"completion_percent"`
}

// OnlinePlayer represents a player currently active in the city.
type OnlinePlayer struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	PlotID   string `json:"plot_id,omitempty"`
}

// GetCityState returns the complete city state, using Redis cache when available.
func (c *CityManager) GetCityState(ctx context.Context) (*CityState, error) {
	log := logger.FromContext(ctx, "world.city")

	// 1. Try cache first
	cached, err := c.redis.Get(ctx, cityStateCacheKey).Result()
	if err == nil {
		var state CityState
		if err := json.Unmarshal([]byte(cached), &state); err == nil {
			log.Debug().Msg("city state served from cache")
			return &state, nil
		}
		log.Warn().Err(err).Msg("failed to unmarshal cached city state, rebuilding")
	} else if !errors.Is(err, redis.Nil) {
		log.Warn().Err(err).Msg("redis cache read failed, rebuilding city state")
	}

	// 2. Cache miss or error — rebuild from database
	state, err := c.buildCityState(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to build city state: %w", err)
	}

	// 3. Cache the result
	stateJSON, err := json.Marshal(state)
	if err != nil {
		log.Error().Err(err).Msg("failed to marshal city state for cache")
		return state, nil // Return the state even if caching fails
	}

	if err := c.redis.Set(ctx, cityStateCacheKey, stateJSON, cityStateCacheTTL).Err(); err != nil {
		log.Warn().Err(err).Msg("failed to cache city state")
	}

	return state, nil
}

// InvalidateCityCache forces a rebuild of the city state cache on the next request.
// This should be called whenever a plot changes ownership, development level, or business status.
func (c *CityManager) InvalidateCityCache(ctx context.Context) {
	log := logger.FromContext(ctx, "world.city")

	if err := c.redis.Del(ctx, cityStateCacheKey, cityPlotsCacheKey).Err(); err != nil {
		log.Warn().Err(err).Msg("failed to invalidate city cache")
		return
	}

	log.Info().Msg("city cache invalidated")
}

// buildCityState constructs the complete city state from the database.
func (c *CityManager) buildCityState(ctx context.Context) (*CityState, error) {
	// 1. Fetch all plots
	allPlots, err := c.queries.GetAllPlots(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch all plots: %w", err)
	}

	// 2. Fetch active businesses to map to plots
	businesses, err := c.queries.GetActiveBusinessesWithPlots(ctx)
	if err != nil {
		c.log.Warn().Err(err).Msg("failed to fetch businesses, continuing without business data")
		businesses = []sqlc.GetActiveBusinessesWithPlotsRow{}
	}

	// Build a map of plot_id -> business for quick lookup
	businessByPlot := make(map[string]sqlc.GetActiveBusinessesWithPlotsRow)
	for _, biz := range businesses {
		if biz.CityPlotID.Valid {
			businessByPlot[biz.CityPlotID.String] = biz
		}
	}

	// 3. Fetch online players (active in last 15 minutes)
	onlinePlayers, err := c.queries.GetOnlinePlayers(ctx)
	if err != nil {
		c.log.Warn().Err(err).Msg("failed to fetch online players")
		onlinePlayers = []sqlc.GetOnlinePlayersRow{}
	}

	// 4. Fetch infrastructure projects
	infraProjects, err := c.queries.GetInfrastructureProjects(ctx)
	if err != nil {
		c.log.Warn().Err(err).Msg("failed to fetch infrastructure projects")
		infraProjects = []sqlc.GetInfrastructureProjectsRow{}
	}

	// 5. Build plot summaries
	plotSummaries := make([]*PlotSummary, 0, len(allPlots))
	ownedCount := 0

	for _, plot := range allPlots {
		zoneCfg, _ := GetZoneConfig(ZoneType(plot.ZoneType))
		zoneColor := "#2a2f42" // Default border color
		if zoneCfg != nil {
			zoneColor = zoneCfg.Color
		}

		summary := &PlotSummary{
			ID:               plot.ID,
			X:                plot.PlotX,
			Y:                plot.PlotY,
			Zone:             plot.ZoneType,
			ZoneColor:        zoneColor,
			DevelopmentLevel: int(plot.DevelopmentLevel),
		}

		if plot.OwnerPlayerID.Valid {
			summary.Owner = plot.OwnerUsername.String
			ownedCount++

			// Check if this plot has a business
			if biz, ok := businessByPlot[plot.ID]; ok {
				summary.HasBusiness = true
				summary.BusinessType = biz.BusinessType
			}
		}

		plotSummaries = append(plotSummaries, summary)
	}

	// 6. Calculate occupancy rate
	totalPlots := len(allPlots)
	occupancyRate := 0.0
	if totalPlots > 0 {
		occupancyRate = float64(ownedCount) / float64(totalPlots) * 100
	}

	// 7. Build infrastructure list
	infraList := make([]Infrastructure, 0, len(infraProjects))
	for _, proj := range infraProjects {
		infraList = append(infraList, Infrastructure{
			ID:         proj.ID,
			Name:       proj.Name,
			Type:       proj.ProjectType,
			Status:     proj.Status,
			Completion: int(proj.CompletionPercent),
		})
	}

	// 8. Build online players list
	onlineList := make([]OnlinePlayer, 0, len(onlinePlayers))
	for _, p := range onlinePlayers {
		onlineList = append(onlineList, OnlinePlayer{
			ID:       p.ID,
			Username: p.Username,
			PlotID:   p.CurrentPlotID.String,
		})
	}

	return &CityState{
		Plots:           plotSummaries,
		Infrastructure:  infraList,
		OnlinePlayers:   onlineList,
		TotalPopulation: ownedCount,
		OccupancyRate:   occupancyRate,
		LastUpdated:     time.Now(),
	}, nil
}

// GetDistrictSummary returns aggregated statistics for a specific zone/district.
func (c *CityManager) GetDistrictSummary(ctx context.Context, zoneType ZoneType) (*DistrictSummary, error) {
	row, err := c.queries.GetDistrictSummary(ctx, string(zoneType))
	if err != nil {
		return nil, fmt.Errorf("failed to fetch district summary: %w", err)
	}

	return &DistrictSummary{
		ZoneType:        zoneType,
		TotalPlots:      int(row.TotalPlots),
		OwnedPlots:      int(row.OwnedPlots),
		AvgDevelopment:  numericToDecimal(row.AvgDevelopment).InexactFloat64(),
		TotalValue:      numericToDecimal(row.TotalValue),
		ActiveBusinesses: int(row.ActiveBusinesses),
	}, nil
}

// DistrictSummary holds aggregated statistics for a city district.
type DistrictSummary struct {
	ZoneType         ZoneType          `json:"zone_type"`
	TotalPlots       int               `json:"total_plots"`
	OwnedPlots       int               `json:"owned_plots"`
	AvgDevelopment   float64           `json:"avg_development"`
	TotalValue       decimal.Decimal   `json:"total_value"`
	ActiveBusinesses int               `json:"active_businesses"`
}