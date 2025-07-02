package pg

import (
	"context"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"mateo/internal/domain"
)

// CreateInvoice создает Invoice и возвращает его ID. Обратите внимание, что обновляется ID в исходном Invoice
func (s *Store) CreateInvoice(ctx context.Context, invoice *domain.Invoice) (string, error) {
	invoice.ID = uuid.New().String()
	query := `
		INSERT INTO "InvoiceIn" (
			id,
			merchant_id,
			amount,
			status,
			type,
			terminal_id,
			user_id,
			bank_id,
			traider_account_id,
			requisite_id,
			callback_url,
			callback_key,
			internal_request_id,
			time_expires,
			exchange
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
		RETURNING id`

	err := s.conn.QueryRow(ctx, query,
		invoice.ID,
		invoice.MerchantID,
		invoice.Amount,
		invoice.Status,
		invoice.Type,
		invoice.TerminalID,
		invoice.UserID,
		invoice.BankID,
		invoice.TraiderAccountID,
		invoice.RequisiteID,
		invoice.CallbackURL,
		invoice.CallbackKey,
		invoice.InternalRequestID,
		invoice.TimeExpires,
		invoice.Exchange,
	).Scan(&invoice.ID)
	if err != nil {
		log.Error().Err(err).
			Interface("invoice", invoice).
			Msg("failed to create invoice")
		return "", domain.ErrorFailedCreateInvoice
	}

	return invoice.ID, nil
}
