package dto

import (
	"banking-service/internal/model"
	_ "time"
)

type CreateTransferRequest struct {
	PayerAccountNumber     string             `json:"payer_account_number"     binding:"required"`
	RecipientAccountNumber string             `json:"recipient_account_number" binding:"required"`
	StartAmount            float64            `json:"start_amount"             binding:"required,min=0.01"`
	StartCurrencyCode      model.CurrencyCode `json:"start_currency_code"      binding:"required,currency_code"`
	EndCurrencyCode        model.CurrencyCode `json:"end_currency_code"        binding:"required,currency_code"`
}
