package http

import (
	"github.com/gofiber/fiber/v3"
	"github.com/pkg/errors"
	"github.com/shopspring/decimal"
	"mateo/internal/domain"
	"time"
)

const (
	currencyCode = "RUB"
)

var (
	ErrorInvalidAmount      = errors.New("invalid amount")
	ErrorEmptyMerchantID    = errors.New("empty merchantId field")
	ErrorEmptyCallbackURL   = errors.New("empty callbackUrl field")
	ErrorEmptyUserID        = errors.New("empty userId field")
	ErrorEmptyRequisiteType = errors.New("empty requisiteType field")
)

type CreateInvoiceRequest struct {
	Amount              int    `json:"amount"`
	InternalRequestID   string `json:"internalRequestID"`
	CallbackUrl         string `json:"callbackUrl"`
	CallbackKey         string `json:"callbackKey"`
	MerchantID          string `json:"merchantID"`
	Type                string `json:"type"`
	ActiveTime          int    `json:"activeTime"`
	BankID              string `json:"bankId"`
	FlexibleRange       int    `json:"flexibleRange"`
	AllowFlexibleAmount bool   `json:"allowFlexibleAmount"`
}

func (req *CreateInvoiceRequest) Validate() error {
	if req.Amount <= 0 {
		return ErrorInvalidAmount
	}
	if req.MerchantID == "" {
		return ErrorEmptyMerchantID
	}
	if req.CallbackUrl == "" {
		return ErrorEmptyCallbackURL
	}
	if req.Type == "" {
		return ErrorEmptyRequisiteType
	}
	return nil
}

type CreateInvoiceResponse struct {
	Status  string                     `json:"status"`
	Error   bool                       `json:"error"`
	Message string                     `json:"message"`
	Data    *CreateInvoiceResponseData `json:"data,omitempty"`
}

type CreateInvoiceResponseData struct {
	InvoiceId         string    `json:"invoiceId"`
	InvoiceStatus     string    `json:"invoiceStatus"`
	Amount            string    `json:"amount"`
	IsFlexibleAmount  bool      `json:"isFlexibleAmount"`
	CurrencyCode      string    `json:"currencyCode"`
	MerchantId        string    `json:"merchantId"`
	InternalRequestId string    `json:"internalRequestId"`
	CallbackUrl       string    `json:"callbackUrl"`
	CallbackKey       string    `json:"callbackKey"`
	PhoneNumber       string    `json:"phoneNumber"`
	WalletNumber      string    `json:"walletNumber"`
	CardNumber        string    `json:"cardNumber"`
	CardName          string    `json:"cardName"`
	Issuer            string    `json:"issuer"`
	TimeExperies      time.Time `json:"timeExperies"`
}

func (s *Server) CreateInvoice(fiberContext fiber.Ctx) error {
	ctx := fiberContext.Context()
	req := &CreateInvoiceRequest{}
	if err := fiberContext.Bind().Body(req); err != nil {
		return fiberContext.Status(fiber.StatusBadRequest).
			JSON(buildCreateInvoiceResponseWithError(errors.Wrap(err, "parse request body")))
	}

	if err := req.Validate(); err != nil {
		return fiberContext.Status(fiber.StatusBadRequest).JSON(buildCreateInvoiceResponseWithError(err))
	}

	requisiteType, err := domain.ParseRequisiteType(req.Type)
	if err != nil {
		return fiberContext.Status(fiber.StatusBadRequest).JSON(buildCreateInvoiceResponseWithError(err))
	}

	invoice, requisite, err := s.app.CreateInvoice(
		ctx,
		decimal.NewFromInt(int64(req.Amount)),
		req.MerchantID,
		requisiteType,
		req.InternalRequestID,
		req.CallbackUrl,
		req.CallbackKey,
		time.Duration(req.ActiveTime)*time.Minute,
		req.BankID,
		req.FlexibleRange,
		req.AllowFlexibleAmount,
	)
	if err != nil {
		return fiberContext.Status(fiber.StatusInternalServerError).JSON(buildCreateInvoiceResponseWithError(err))
	}

	return fiberContext.Status(fiber.StatusOK).JSON(buildCreateInvoiceResponseWithInvoice(invoice, requisite))
}

func buildCreateInvoiceResponseWithInvoice(invoice *domain.Invoice, requisite *domain.Requisite) *CreateInvoiceResponse {
	return &CreateInvoiceResponse{
		Status:  "ok",
		Error:   false,
		Message: "success",
		Data: &CreateInvoiceResponseData{
			InvoiceId:         invoice.ID,
			InvoiceStatus:     string(invoice.Status),
			Amount:            invoice.Amount.String(),
			IsFlexibleAmount:  invoice.IsFlexibleAmount,
			CurrencyCode:      currencyCode,
			MerchantId:        invoice.MerchantID,
			InternalRequestId: invoice.InternalRequestID,
			CallbackUrl:       invoice.CallbackURL,
			CallbackKey:       invoice.CallbackKey,
			PhoneNumber:       requisite.PhoneNumber,
			WalletNumber:      requisite.WalletNumber,
			CardNumber:        requisite.CardNumber,
			CardName:          requisite.RecipientName,
			Issuer:            requisite.BankName,
			TimeExperies:      invoice.TimeExpires,
		},
	}
}

func buildCreateInvoiceResponseWithError(err error) *CreateInvoiceResponse {
	return &CreateInvoiceResponse{
		Status:  "error",
		Error:   true,
		Message: err.Error(),
	}
}
