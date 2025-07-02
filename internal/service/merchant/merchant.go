package merchant

import (
	"context"
	"github.com/pkg/errors"
	"github.com/shopspring/decimal"
	"mateo/internal/domain"
)

type Store interface {
	GetMerchantByMerchantID(
		ctx context.Context,
		merchantID string,
	) (*domain.Merchant, error)
}

type Service struct {
	store Store
}

func NewService(store Store) *Service {
	return &Service{store: store}
}

func (s *Service) ValidateMerchantInvoice(
	ctx context.Context,
	merchantID string,
	amount decimal.Decimal,
	requisiteType domain.RequisiteType,
) error {
	merchant, err := s.store.GetMerchantByMerchantID(ctx, merchantID)
	if err != nil {
		return errors.Wrap(err, "get merchant by id")
	}

	switch requisiteType {
	case domain.RequisiteTypeCard:
		if amount.LessThan(merchant.InLimitCard) {
			return domain.ErrorAmountLessThanLimit
		}
	case domain.RequisiteTypeWallet:
		if amount.LessThan(merchant.InLimitWallet) {
			return domain.ErrorAmountLessThanLimit
		}
	case domain.RequisiteTypeSBP:
		if amount.LessThan(merchant.InLimitSBP) {
			return domain.ErrorAmountLessThanLimit
		}
	default:
		return domain.ErrorUnknownRequisiteType
	}

	return nil
}
