package pg

import (
	"context"
	"github.com/jackc/pgx/v5"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"mateo/internal/domain"
)

func (s *Store) GetMerchantByMerchantID(ctx context.Context, merchantID string) (*domain.Merchant, error) {
	const query = `SELECT id,
       COALESCE(in_limit_card, 0) as in_limit_card,
       COALESCE(in_limit_wallet, 0) as in_limit_wallet,
       COALESCE(in_limit_sbp, 0) as in_limit_sbp
	   FROM "Merchant" 
	   WHERE id = $1`

	row := s.conn.QueryRow(ctx, query, merchantID)

	var m domain.Merchant
	err := row.Scan(&m.ID, &m.InLimitCard, &m.InLimitWallet, &m.InLimitSBP)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrorMerchantNotFound
		}

		log.Error().Err(err).
			Str("merchant_id", merchantID).
			Msg("failed to find merchant")
		return nil, domain.ErrorFailedFindMerchant
	}

	return &m, nil
}
