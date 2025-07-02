package domain

import (
	"context"
	"github.com/shopspring/decimal"
	"time"
)

type MerchantService interface {
	ValidateMerchantInvoice(
		ctx context.Context,
		merchantID string,
		amount decimal.Decimal,
		requisiteType RequisiteType,
	) error
}

type RequisiteService interface {
	SelectAvailableRequisite(
		ctx context.Context,
		merchantID string,
		amount decimal.Decimal,
		requisiteType RequisiteType,
		bankID string,
		flexibleRange int,
		allowFlexibleAmount bool,
	) (*Requisite, error)
}

type InvoiceService interface {
	CreateInvoice(
		ctx context.Context,
		amount decimal.Decimal,
		isAmountFlexible bool,
		internalRequestID string,
		callbackURL string,
		callbackKey string,
		merchantID string,
		timeExpires time.Duration,
		requisite *Requisite,
	) (*Invoice, error)
}

type App struct {
	merchant  MerchantService
	requisite RequisiteService
	invoice   InvoiceService
}

func NewApp(merchant MerchantService, requisite RequisiteService, invoice InvoiceService) *App {
	return &App{merchant: merchant, requisite: requisite, invoice: invoice}
}
