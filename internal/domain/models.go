package domain

import (
	"github.com/shopspring/decimal"
	"time"
)

type RequisiteType string

const (
	RequisiteTypeCard   RequisiteType = "CARD"
	RequisiteTypeWallet RequisiteType = "WALLET"
	RequisiteTypeSBP    RequisiteType = "SBP"
)

type InvoiceStatus string

const (
	InvoiceStatusCreated InvoiceStatus = "CREATED"
)

type Merchant struct {
	ID string

	InLimitCard   decimal.Decimal
	InLimitWallet decimal.Decimal
	InLimitSBP    decimal.Decimal
}

type Provider struct {
	ID          string
	Name        string
	Description string
	IsActive    bool
}

type Invoice struct {
	ID                string
	MerchantID        string
	Amount            decimal.Decimal
	Status            InvoiceStatus
	Type              RequisiteType
	IsFlexibleAmount  bool
	TerminalID        string
	UserID            string
	BankID            string
	TraiderAccountID  string
	RequisiteID       string
	CallbackURL       string
	CallbackKey       string
	InternalRequestID string
	TimeExpires       time.Time
	Exchange          decimal.Decimal
}

type Team struct {
	ID        string
	Name      string
	IsBoosted bool
}

type TraderAccount struct {
	ID           string
	TeamID       string
	Balance      decimal.Decimal
	DailyLimit   decimal.Decimal
	IsActive     bool
	BlockedUntil *int64
}

type Requisite struct {
	ID                     string
	Type                   RequisiteType
	PhoneNumber            string
	CardNumber             string
	WalletNumber           string
	RecipientName          string
	TerminalID             string
	UserID                 string
	BankID                 string
	BankName               string
	TeamID                 string
	TraiderAccountID       string
	FlexibleSelectedAmount decimal.Decimal
}

type CreateInvoiceDTO struct {
	Amount            decimal.Decimal
	MerchantID        string
	RequisiteType     RequisiteType
	CallbackURL       string
	InternalRequestID string
}

func ParseRequisiteType(t string) (RequisiteType, error) {
	switch t {
	case "CARD":
		return RequisiteTypeCard, nil
	case "WALLET":
		return RequisiteTypeWallet, nil
	case "SBP":
		return RequisiteTypeSBP, nil
	default:
		return "", ErrorUnknownRequisiteType
	}
}
