package pg

import (
	"context"
	"fmt"
	"github.com/rs/zerolog/log"
	"github.com/shopspring/decimal"
	"mateo/internal/domain"
	"time"
)

// requisiteFieldMaps defines the field names for each requisite type
type requisiteFields struct {
	minField    string
	isWorkField string
}

var requisiteTypeToFields = map[domain.RequisiteType]requisiteFields{
	domain.RequisiteTypeSBP:    {"min_invoice_amount_sbp", "is_work_on_sbp_pay_in"},
	domain.RequisiteTypeCard:   {"min_invoice_amount_card", "is_work_on_card_pay_in"},
	domain.RequisiteTypeWallet: {"min_invoice_amount_wallet", "is_work_on_wallet_pay_in"},
}

func (s *Store) SelectAvailableRequisites(
	ctx context.Context,
	merchantID string,
	amount decimal.Decimal,
	requisiteType domain.RequisiteType,
	bankID string,
) ([]*domain.Requisite, error) {
	fields, ok := requisiteTypeToFields[requisiteType]
	if !ok {
		return nil, fmt.Errorf("unsupported requisite type: %s", requisiteType)
	}

	minField := fields.minField
	isWorkField := fields.isWorkField
	startOfToday := time.Now().UTC().Truncate(24 * time.Hour)

	// Формируем SQL запрос с динамическими полями
	query := fmt.Sprintf(`
	WITH today_invoices AS (
		SELECT 
			terminal_id,
			requisite_id,
			traider_account_id,
			amount,
			status,
			created_at
		FROM "InvoiceIn"
		WHERE created_at >= $1
			AND status IN ('CREATED','SUCCESS','SUCCESS_HAND','SUCCESS_APPEAL')
	),
	account_aggregates AS (
		SELECT 
			traider_account_id,
			COUNT(*) AS active_count,
			SUM(amount) FILTER (WHERE status = 'CREATED') AS active_sum
		FROM today_invoices
		GROUP BY traider_account_id
	),
	terminal_aggregates AS (
		SELECT 
			terminal_id,
			COUNT(*) AS created_count,
			COUNT(*) FILTER (WHERE status != 'CREATED') AS success_count,
			SUM(amount) AS total_amount,
			MAX(created_at) AS last_invoice_time
		FROM today_invoices
		GROUP BY terminal_id
	),
	requisite_aggregates AS (
		SELECT 
			requisite_id,
			COUNT(*)  AS created_req_count,
			COUNT(*) FILTER (WHERE status != 'CREATED') AS success_req_count,
			BOOL_OR(amount = $2 AND status = 'CREATED') AS has_active_same_amount,
			MAX(created_at) AS last_invoice_time
		FROM today_invoices
		GROUP BY requisite_id
	)
	SELECT
		ta.user_id,
		r.id AS requisite_id,
		COALESCE(r.name, '') AS recipient_name,
		COALESCE(r.phone_number, '') AS phone_number,
		COALESCE(r.card_number, '') AS card_number,
		COALESCE(r.wallet_number, '') AS wallet_number,
		b.name AS bank_name,
		r.bank_id,
		t.id AS terminal_id,
		ta.id AS traider_account_id,
		COALESCE(ta.team_id, '') as team_id
	FROM "TraiderAccount" ta
	JOIN "Wallet" w ON ta.wallet_id = w.id
	JOIN "Terminal" t ON t.traider_account_id = ta.id
	JOIN "Requisite" r ON r.terminal_id = t.id
	JOIN "Bank" b ON r.bank_id = b.id
	LEFT JOIN account_aggregates aa ON aa.traider_account_id = ta.id
	LEFT JOIN terminal_aggregates ta_agg ON ta_agg.terminal_id = t.id
	LEFT JOIN requisite_aggregates ra ON ra.requisite_id = r.id
	WHERE
		ta.is_can_work = TRUE
		AND ta.is_blocked = FALSE
		AND ta.%s <= $2
		AND ta.%s = TRUE
		AND ta.max_invoice_amount_int >= $2
		AND EXISTS (
			SELECT 1 FROM "MerchantInvoicesInOnTraiderAccount" mta 
			WHERE mta.traider_account_id = ta.id AND mta.merchant_id = $3
		)
		AND w.pay_in_balance > COALESCE(aa.active_sum, 0)
		AND t.is_can_work = TRUE
		AND t.is_blocked = FALSE
		AND t.min_invoice_amount <= $2
		AND t.max_invoice_amount >= $2
		AND (t.daily_limit_money IS NULL OR COALESCE(ta_agg.total_amount, 0) + $2 <= t.daily_limit_money)
		AND (ta_agg.last_invoice_time IS NULL OR NOW() - ta_agg.last_invoice_time >= (t.invoice_interval * INTERVAL '1 minute'))
		AND (t.max_active_invoice IS NULL OR COALESCE(ta_agg.created_count, 0) < t.max_active_invoice)
		AND (t.daily_limit_invoices IS NULL OR COALESCE(ta_agg.created_count, 0) < t.daily_limit_invoices)

		AND r.type = $4
		AND r.is_can_work = TRUE
		AND r.is_blocked = FALSE
		AND r.min_invoice_amount <= $2
		AND r.max_invoice_amount >= $2
		AND (r.max_active_invoice IS NULL OR COALESCE(ra.created_req_count, 0) < r.max_active_invoice)
		AND COALESCE(ra.has_active_same_amount, FALSE) = FALSE
		AND (ra.last_invoice_time IS NULL OR NOW() - ra.last_invoice_time >= (r.invoice_interval * INTERVAL '1 minute'))
		AND (r.daily_limit_invoices IS NULL OR COALESCE(ra.created_req_count, 0) < r.daily_limit_invoices)
		AND ($5 = '' OR r.bank_id = $5)
		
		AND (ta.max_active_invoices_in IS NULL OR COALESCE(aa.active_count, 0) < ta.max_active_invoices_in)
	`, minField, isWorkField)

	rows, err := s.conn.Query(
		ctx,
		query,
		startOfToday,
		amount,
		merchantID,
		requisiteType,
		bankID,
	)
	if err != nil {
		log.Error().Err(err).
			Str("merchant_id", merchantID).
			Str("requisite_type", string(requisiteType)).
			Str("amount", amount.String()).
			Msg("query failed")
		return nil, fmt.Errorf("query failed: %w", err)
	}
	defer rows.Close()

	var requisites []*domain.Requisite

	for rows.Next() {
		r := &domain.Requisite{
			Type: requisiteType,
		}
		err := rows.Scan(
			&r.UserID,
			&r.ID,
			&r.RecipientName,
			&r.PhoneNumber,
			&r.CardNumber,
			&r.WalletNumber,
			&r.BankName,
			&r.BankID,
			&r.TerminalID,
			&r.TraiderAccountID,
			&r.TeamID,
		)
		if err != nil {
			return nil, fmt.Errorf("scan failed: %w", err)
		}
		requisites = append(requisites, r)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return requisites, nil
}

func (s *Store) SelectAvailableRequisitesFlexible(
	ctx context.Context,
	merchantID string,
	flexibleAmountMin decimal.Decimal,
	flexibleAmountMax decimal.Decimal,
	flexibleAmountStep decimal.Decimal,
	requisiteType domain.RequisiteType,
	bankId string,
) ([]*domain.Requisite, error) {
	if flexibleAmountStep.LessThanOrEqual(decimal.Zero) {
		return nil, fmt.Errorf("flexibleAmountStep must be positive")
	}
	if flexibleAmountMin.GreaterThan(flexibleAmountMax) {
		return nil, fmt.Errorf("flexibleAmountMin cannot be greater than flexibleAmountMax")
	}

	fields, ok := requisiteTypeToFields[requisiteType]
	if !ok {
		return nil, fmt.Errorf("unsupported requisite type: %s", requisiteType)
	}

	minField := fields.minField
	isWorkField := fields.isWorkField
	startOfToday := time.Now().UTC().Truncate(24 * time.Hour)

	query := fmt.Sprintf(`
		WITH today_invoices AS (
			SELECT 
				terminal_id,
				requisite_id,
				traider_account_id,
				amount,
				status,
				created_at
			FROM "InvoiceIn"
			WHERE created_at >= $1
				AND status IN ('CREATED','SUCCESS','SUCCESS_HAND','SUCCESS_APPEAL')
		),
		account_aggregates AS (
			SELECT 
				traider_account_id,
				COUNT(*) AS active_count,
				SUM(amount) FILTER (WHERE status = 'CREATED') AS active_sum
			FROM today_invoices
			GROUP BY traider_account_id
		),
		terminal_aggregates AS (
			SELECT 
				terminal_id,
				COUNT(*) AS created_count,
				COUNT(*) FILTER (WHERE status != 'CREATED') AS success_count,
				SUM(amount) AS total_amount,
				MAX(created_at) AS last_invoice_time
			FROM today_invoices
			GROUP BY terminal_id
		),
		amounts AS (
			SELECT generate_series($2::numeric, $3::numeric, $4::numeric) AS amount_val
		)
		SELECT DISTINCT ON (r.id)
			ta.user_id,
			r.id AS requisite_id,
			COALESCE(r.name, '') AS recipient_name,
			COALESCE(r.phone_number, '') AS phone_number,
			COALESCE(r.card_number, '') AS card_number,
			COALESCE(r.wallet_number, '') AS wallet_number,
			b.name AS bank_name,
			r.bank_id,
			t.id AS terminal_id,
			ta.id AS traider_account_id,
			COALESCE(ta.team_id, '') as team_id,
			amount_val AS selected_amount
		FROM "TraiderAccount" ta
		JOIN "Wallet" w ON ta.wallet_id = w.id
		JOIN "Terminal" t ON t.traider_account_id = ta.id
		JOIN "Requisite" r ON r.terminal_id = t.id
		JOIN "Bank" b ON r.bank_id = b.id
		LEFT JOIN account_aggregates aa ON aa.traider_account_id = ta.id
		LEFT JOIN terminal_aggregates ta_agg ON ta_agg.terminal_id = t.id
		CROSS JOIN amounts
		WHERE
			ta.is_can_work = TRUE
			AND ta.is_blocked = FALSE
			AND ta.%s <= amount_val
			AND ta.%s = TRUE
			AND ta.max_invoice_amount_int >= amount_val
			AND EXISTS (
				SELECT 1 FROM "MerchantInvoicesInOnTraiderAccount" mta 
				WHERE mta.traider_account_id = ta.id AND mta.merchant_id = $5
			)
			AND w.pay_in_balance > COALESCE(aa.active_sum, 0)
			AND t.is_can_work = TRUE
			AND t.is_blocked = FALSE
			AND t.min_invoice_amount <= amount_val
			AND t.max_invoice_amount >= amount_val
			AND (t.daily_limit_money IS NULL OR COALESCE(ta_agg.total_amount, 0) + amount_val <= t.daily_limit_money)
			AND (ta_agg.last_invoice_time IS NULL OR NOW() - ta_agg.last_invoice_time >= (t.invoice_interval * INTERVAL '1 minute'))
			AND (t.max_active_invoice IS NULL OR COALESCE(ta_agg.created_count, 0) < t.max_active_invoice)
			AND (t.daily_limit_invoices IS NULL OR COALESCE(ta_agg.created_count, 0) < t.daily_limit_invoices)
		
			AND r.type = $6
			AND r.is_can_work = TRUE
			AND r.is_blocked = FALSE
			AND r.min_invoice_amount <= amount_val
			AND r.max_invoice_amount >= amount_val
			AND (r.max_active_invoice IS NULL OR (
				SELECT COUNT(*) 
				FROM today_invoices ti 
				WHERE ti.requisite_id = r.id
			) < r.max_active_invoice)
			AND NOT EXISTS (
				SELECT 1 FROM today_invoices ti
				WHERE ti.requisite_id = r.id
					AND ti.amount = amount_val
					AND ti.status = 'CREATED'
			)
			AND (r.daily_limit_invoices IS NULL OR (
				SELECT COUNT(*) 
				FROM today_invoices ti 
				WHERE ti.requisite_id = r.id
			) < r.daily_limit_invoices)
			AND ($7 = '' OR r.bank_id = $7)
			
			AND (ta.max_active_invoices_in IS NULL OR COALESCE(aa.active_count, 0) < ta.max_active_invoices_in)
		ORDER BY r.id, amount_val
		`, minField, isWorkField)

	rows, err := s.conn.Query(
		ctx,
		query,
		startOfToday,
		flexibleAmountMin,
		flexibleAmountMax,
		flexibleAmountStep,
		merchantID,
		requisiteType,
		bankId,
	)
	if err != nil {
		log.Error().Err(err).
			Str("merchant_id", merchantID).
			Str("requisite_type", string(requisiteType)).
			Str("flexibleAmountMin", flexibleAmountMin.String()).
			Str("flexibleAmountMax", flexibleAmountMax.String()).
			Str("flexibleAmountStep", flexibleAmountStep.String()).
			Msg("query failed")
		return nil, fmt.Errorf("query failed: %w", err)
	}
	defer rows.Close()

	var requisites []*domain.Requisite
	for rows.Next() {
		r := &domain.Requisite{
			Type: requisiteType,
		}
		err := rows.Scan(
			&r.UserID,
			&r.ID,
			&r.RecipientName,
			&r.PhoneNumber,
			&r.CardNumber,
			&r.WalletNumber,
			&r.BankName,
			&r.BankID,
			&r.TerminalID,
			&r.TraiderAccountID,
			&r.TeamID,
			&r.FlexibleSelectedAmount,
		)
		if err != nil {
			return nil, fmt.Errorf("scan failed: %w", err)
		}
		requisites = append(requisites, r)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return requisites, nil
}
