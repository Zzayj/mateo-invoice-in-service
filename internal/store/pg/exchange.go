package pg

import (
	"context"
	"github.com/rs/zerolog/log"
	"mateo/internal/domain"

	"github.com/shopspring/decimal"
)

func (s *Store) GetExchangeRate(ctx context.Context) (decimal.Decimal, error) {
	const query = `
		SELECT exchange_rate
		FROM "Settings"
		LIMIT 1
	`

	var rate decimal.Decimal
	err := s.conn.QueryRow(ctx, query).Scan(&rate)
	if err != nil {
		log.Error().Err(err).Msg("failed to get exchange rate")
		return decimal.Decimal{}, domain.ErrorFailedGetExchangeRate
	}

	return rate, nil
}
