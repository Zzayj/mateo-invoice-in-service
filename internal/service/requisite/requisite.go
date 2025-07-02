package requisite

import (
	"context"
	"github.com/pkg/errors"
	"github.com/shopspring/decimal"
	"mateo/internal/domain"
	"math/rand"
)

type Store interface {
	SelectAvailableRequisites(
		ctx context.Context,
		merchantID string,
		amount decimal.Decimal,
		requisiteType domain.RequisiteType,
		bankId string,
	) ([]*domain.Requisite, error)

	GetBoostedTeamIds(ctx context.Context) ([]string, error)

	SelectAvailableRequisitesFlexible(
		ctx context.Context,
		merchantID string,
		flexibleAmountMin decimal.Decimal,
		flexibleAmountMax decimal.Decimal,
		flexibleAmountStep decimal.Decimal,
		requisiteType domain.RequisiteType,
		bankId string,
	) ([]*domain.Requisite, error)
}

const (
	flexibleAmountStep   = 5
	defaultFlexibleRange = 20
)

type Service struct {
	store Store
}

func NewService(store Store) *Service {
	return &Service{store: store}
}

func (s *Service) SelectAvailableRequisite(
	ctx context.Context,
	merchantID string,
	amount decimal.Decimal,
	requisiteType domain.RequisiteType,
	bankID string,
	flexibleRange int,
	allowFlexibleAmount bool,
) (*domain.Requisite, error) {
	requisites, err := s.store.SelectAvailableRequisites(ctx, merchantID, amount, requisiteType, bankID)
	if err != nil {
		return nil, errors.Wrap(err, "select available requisites")
	}

	if len(requisites) == 0 {
		if !allowFlexibleAmount {
			return nil, domain.ErrorNoAvailableRequisites
		}
		if flexibleRange == 0 {
			flexibleRange = defaultFlexibleRange
		}
		maxFlexibleAmount := amount.Add(decimal.NewFromInt(int64(flexibleRange)))
		// округляем до 5
		minFlexibleAmount := roundUpToFive(roundUpToFive(amount))
		requisites, err = s.store.SelectAvailableRequisitesFlexible(
			ctx,
			merchantID,
			minFlexibleAmount,
			maxFlexibleAmount,
			decimal.NewFromInt(flexibleAmountStep),
			requisiteType,
			bankID,
		)
		if err != nil {
			return nil, errors.Wrap(err, "select available flexible requisites")
		}
		if len(requisites) == 0 {
			return nil, domain.ErrorNoAvailableRequisites
		}
	}

	boostedTeamIds, err := s.store.GetBoostedTeamIds(ctx)
	if err != nil {
		return nil, domain.ErrorNoAvailableRequisites
	}

	boostedRequisites := make([]*domain.Requisite, 0)
	for _, requisite := range requisites {
		if domain.Contains(boostedTeamIds, requisite.TeamID) {
			boostedRequisites = append(boostedRequisites, requisite)
		}
	}

	if len(boostedRequisites) == 0 {
		return requisites[rand.Intn(len(requisites))], nil
	}

	return boostedRequisites[rand.Intn(len(boostedRequisites))], nil
}

func roundUpToFive(num decimal.Decimal) decimal.Decimal {
	return num.Div(decimal.NewFromInt(5)).Ceil().Mul(decimal.NewFromInt(5))
}
