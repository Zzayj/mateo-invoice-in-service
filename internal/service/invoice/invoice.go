package invoice

import (
	"context"
	"github.com/pkg/errors"
	"github.com/shopspring/decimal"
	"mateo/internal/domain"
	"time"
)

const (
	defaultActiveTime = time.Minute * 15
)

type Store interface {
	// CreateInvoice создает Invoice и возвращает его ID
	CreateInvoice(
		ctx context.Context,
		invoice *domain.Invoice,
	) (string, error)

	// GetExchangeRate возвращает курс RUB/USD
	GetExchangeRate(ctx context.Context) (decimal.Decimal, error)
}

type Service struct {
	store Store
}

func NewService(store Store) *Service {
	return &Service{store: store}
}

func (s *Service) CreateInvoice(
	ctx context.Context,
	amount decimal.Decimal,
	isFlexibleAmount bool,
	internalRequestID string,
	callbackURL string,
	callbackKey string,
	merchantID string,
	activeTime time.Duration,
	requisite *domain.Requisite,
) (*domain.Invoice, error) {
	exchangeRate, err := s.store.GetExchangeRate(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "get exchange rate")
	}

	if activeTime == 0 {
		activeTime = defaultActiveTime
	}

	timeExpires := time.Now().Add(activeTime)

	invoice := &domain.Invoice{
		Amount:            amount,
		InternalRequestID: internalRequestID,
		TerminalID:        requisite.TerminalID,
		CallbackURL:       callbackURL,
		CallbackKey:       callbackKey,
		IsFlexibleAmount:  isFlexibleAmount,
		UserID:            requisite.UserID,
		MerchantID:        merchantID,
		BankID:            requisite.BankID,
		TraiderAccountID:  requisite.TraiderAccountID,
		RequisiteID:       requisite.ID,
		Type:              requisite.Type,
		Status:            domain.InvoiceStatusCreated,
		TimeExpires:       timeExpires,
		Exchange:          exchangeRate,
	}

	invoiceID, err := s.store.CreateInvoice(ctx, invoice)
	if err != nil {
		return nil, errors.Wrap(err, "create invoice")
	}

	invoice.ID = invoiceID

	return invoice, nil
}
