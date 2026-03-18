package handler

import (
	"banking-service/internal/dto"
	"banking-service/internal/service"
	"common/pkg/auth"
	"common/pkg/errors"
	"math"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type PaymentHandler struct {
	service        *service.PaymentService
	accountService *service.AccountService
}

func NewPaymentHandler(paymentService *service.PaymentService, accountService *service.AccountService) *PaymentHandler {
	return &PaymentHandler{service: paymentService, accountService: accountService}
}

func (h *PaymentHandler) CreatePayment(c *gin.Context) {
	var req dto.CreatePaymentRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(errors.BadRequestErr(err.Error()))
		return
	}

	payment, err := h.service.CreatePayment(c.Request.Context(), req)
	if err != nil {
		c.Error(errors.InternalErr(err))
		return
	}

	c.JSON(http.StatusOK, dto.CreatePaymentResponse{
		PaymentID: payment.PaymentID,
	})
}

func (h *PaymentHandler) VerifyPayment(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("payment_id"), 10, 64)
	if err != nil {
		c.Error(errors.BadRequestErr("invalid payment id"))
		return
	}

	var req dto.VerifyPaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(errors.BadRequestErr(err.Error()))
		return
	}

	payment, err := h.service.VerifyPayment(c.Request.Context(), uint(id), req.Code)
	if err != nil {
		c.Error(errors.InternalErr(err))
		return
	}

	c.JSON(http.StatusOK, dto.VerifyPaymentResponse{
		PaymentID: payment.PaymentID,
	})
}

func (h *PaymentHandler) GetAccountPayments(c *gin.Context) {
	authCtx := auth.GetAuth(c)
	if authCtx == nil || authCtx.ClientID == nil {
		c.Error(errors.BadRequestErr("client authentication required"))
		return
	}

	accountNumber := c.Param("accountNumber")

	if _, err := h.accountService.GetAccountDetails(c.Request.Context(), accountNumber, *authCtx.ClientID); err != nil {
		c.Error(err)
		return
	}

	var filters dto.PaymentFilters
	if err := c.ShouldBindQuery(&filters); err != nil {
		c.Error(errors.BadRequestErr(err.Error()))
		return
	}
	if filters.Page < 1 {
		filters.Page = 1
	}
	if filters.PageSize < 1 {
		filters.PageSize = 10
	}

	payments, total, err := h.service.GetAccountPayments(c.Request.Context(), accountNumber, &filters)
	if err != nil {
		c.Error(err)
		return
	}

	data := make([]dto.PaymentSummaryResponse, len(payments))
	for i, p := range payments {
		data[i] = dto.PaymentSummaryResponse{
			ID:               p.PaymentID,
			RecipientName:    p.RecipientName,
			RecipientAccount: p.Transaction.RecipientAccountNumber,
			PayerAccount:     p.Transaction.PayerAccountNumber,
			Amount:           p.Transaction.StartAmount,
			Currency:         string(p.Transaction.StartCurrencyCode),
			Status:           string(p.Transaction.Status),
			Purpose:          p.Purpose,
			PaymentCode:      p.PaymentCode,
			CreatedAt:        p.Transaction.CreatedAt,
		}
	}

	totalPages := int(math.Ceil(float64(total) / float64(filters.PageSize)))
	c.JSON(http.StatusOK, dto.ListPaymentsResponse{
		Data:       data,
		Total:      total,
		Page:       filters.Page,
		PageSize:   filters.PageSize,
		TotalPages: totalPages,
	})
}
