# use go install go.uber.org/mock/mockgen@latest
mockgen:
	mockgen -destination ./internal/mock/invoice/invoice_mock.go --source ./internal/service/invoice/invoice.go Store
	mockgen -destination ./internal/mock/merchant/merchant_mock.go --source ./internal/service/merchant/merchant.go Store
	mockgen -destination ./internal/mock/requisite/requisite_mock.go --source ./internal/service/requisite/requisite.go Store