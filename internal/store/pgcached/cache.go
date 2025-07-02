package pgcached

import (
	"context"
	"encoding/json"
	"github.com/shopspring/decimal"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"mateo/internal/store/pg"
)

type exchangeRateCache struct {
	Rate      float64   `json:"rate"`
	UpdatedAt time.Time `json:"updatedAt"`
}

const (
	boostedTeamCacheKey  = "boosted_team_ids"
	exchangeRateCacheKey = "exchangeRate"
	boostedTeamCacheTTL  = time.Minute * 5
	exchangeRateCacheTTL = time.Minute * 5
)

type CachedStore struct {
	*pg.Store
	redisClient *redis.Client
}

func NewCachedStore(store *pg.Store, redisClient *redis.Client) *CachedStore {
	return &CachedStore{
		Store:       store,
		redisClient: redisClient,
	}
}

func (c *CachedStore) GetBoostedTeamIds(ctx context.Context) ([]string, error) {
	// Try to get from cache
	cached, err := c.redisClient.Get(ctx, boostedTeamCacheKey).Result()
	if err == nil {
		var teamIds []string
		if err := json.Unmarshal([]byte(cached), &teamIds); err == nil {
			return teamIds, nil
		}
		log.Error().Err(err).Msg("failed to unmarshal team ids")
	}

	// If not in cache or error, get from database
	teamIds, err := c.Store.GetBoostedTeamIds(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "get from storage")
	}

	// Update cache
	teamIdsJSON, err := json.Marshal(teamIds)
	if err != nil {
		log.Error().Err(err).Msg("failed to marshal team ids")
		return teamIds, nil
	}

	if err := c.redisClient.Set(ctx, boostedTeamCacheKey, teamIdsJSON, boostedTeamCacheTTL).Err(); err != nil {
		log.Error().Err(err).Msg("failed to set cache in redis")
	}

	return teamIds, nil
}

func (c *CachedStore) GetExchangeRate(ctx context.Context) (decimal.Decimal, error) {
	// Try to get from cache
	cached, err := c.redisClient.Get(ctx, exchangeRateCacheKey).Result()
	if err == nil {
		var cachedRate exchangeRateCache
		if err := json.Unmarshal([]byte(cached), &cachedRate); err == nil {
			return decimal.NewFromFloat(cachedRate.Rate), nil
		}
		log.Error().Err(err).Msg("failed to unmarshal exchange rate from cache")
	}

	// If not in cache or error, get from database
	rate, err := c.Store.GetExchangeRate(ctx)
	if err != nil {
		return decimal.Zero, errors.Wrap(err, "get exchange rate from storage")
	}

	// Update cache
	rateFloat, _ := rate.Float64()
	cacheData := exchangeRateCache{
		Rate:      rateFloat,
		UpdatedAt: time.Now(),
	}

	cacheJSON, err := json.Marshal(cacheData)
	if err != nil {
		log.Error().Err(err).Msg("failed to marshal exchange rate for cache")
		return rate, nil
	}

	if err := c.redisClient.Set(ctx, exchangeRateCacheKey, cacheJSON, exchangeRateCacheTTL).Err(); err != nil {
		log.Error().Err(err).Msg("failed to set exchange rate cache in redis")
	}

	return rate, nil
}
