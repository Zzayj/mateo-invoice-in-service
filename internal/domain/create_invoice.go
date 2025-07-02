package domain

import (
	"context"
	"time"

	"github.com/pkg/errors"
	"github.com/shopspring/decimal"
)

func (a *App) CreateInvoice(
	ctx context.Context,
	amount decimal.Decimal,
	merchantID string,
	requisiteType RequisiteType,
	internalRequestID string,
	callbackURL string,
	callbackKey string,
	activeTime time.Duration,
	bankID string,
	flexibleRange int,
	allowFlexibleAmount bool,
) (*Invoice, *Requisite, error) {
	// Может ли Merchant принять такой Invoice?
	err := a.merchant.ValidateMerchantInvoice(
		ctx,
		merchantID,
		amount,
		requisiteType,
	)
	if err != nil {
		return nil, nil, errors.Wrap(err, "cannot create invoice for this merchant")
	}

	// Выбираем доступный реквизит
	requisite, err := a.requisite.SelectAvailableRequisite(
		ctx,
		merchantID,
		amount,
		requisiteType,
		bankID,
		flexibleRange,
		allowFlexibleAmount,
	)
	if err != nil {
		return nil, nil, errors.Wrap(err, "cannot select requisite")
	}

	if requisite.FlexibleSelectedAmount.GreaterThan(decimal.Zero) {
		amount = requisite.FlexibleSelectedAmount
	}

	// Создаем Invoice
	invoice, err := a.invoice.CreateInvoice(
		ctx,
		amount,
		requisite.FlexibleSelectedAmount.GreaterThan(decimal.Zero),
		internalRequestID,
		callbackURL,
		callbackKey,
		merchantID,
		activeTime,
		requisite,
	)
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to create invoice")
	}

	// Собираем ответ
	return invoice, requisite, nil
}
