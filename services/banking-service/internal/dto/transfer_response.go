package dto

import (
	"banking-service/internal/model"
	"time"
)

type TransferResponse struct {
	TransactionID          uint                    `json:"transaction_id"`
	PayerAccountNumber     string                  `json:"payer_account_number"`
	RecipientAccountNumber string                  `json:"recipient_account_number"`
	StartAmount            float64                 `json:"start_amount"`
	StartCurrencyCode      model.CurrencyCode      `json:"start_currency_code"`
	EndAmount              float64                 `json:"end_amount"`
	EndCurrencyCode        model.CurrencyCode      `json:"end_currency_code"`
	Status                 model.TransactionStatus `json:"status"`
	CreatedAt              time.Time               `json:"created_at"`
}

func ToTransferResponse(t *model.Transaction) *TransferResponse {
	return &TransferResponse{
		TransactionID:          t.TransactionID,
		PayerAccountNumber:     t.PayerAccountNumber,
		RecipientAccountNumber: t.RecipientAccountNumber,
		StartAmount:            t.StartAmount,
		StartCurrencyCode:      t.StartCurrencyCode,
		EndAmount:              t.EndAmount,
		EndCurrencyCode:        t.EndCurrencyCode,
		Status:                 t.Status,
		CreatedAt:              t.CreatedAt,
	}
}
