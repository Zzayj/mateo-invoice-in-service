package domain

import "errors"

var (
	ErrorUnknownRequisiteType = errors.New("unknown requisite type")

	ErrorMerchantNotFound    = errors.New("merchant not found")
	ErrorFailedFindMerchant  = errors.New("failed to find merchant")
	ErrorAmountLessThanLimit = errors.New("amount less than limit")
	ErrorFailedCreateInvoice = errors.New("failed to create invoice")

	ErrorFailedGetExchangeRate = errors.New("failed to get exchange rate")

	ErrorNoAvailableRequisites = errors.New("no available requisites")

	ErrorInvalidAmount      = errors.New("invalid amount")
	ErrorInvalidMerchantID  = errors.New("invalid merchant id")
	ErrorInvalidCallbackURL = errors.New("invalid callback url")
	ErrorInvalidUserID      = errors.New("invalid user id")
)
