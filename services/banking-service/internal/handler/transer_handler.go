package handler

import (
	"banking-service/internal/dto"
	"banking-service/internal/service"
	"common/pkg/errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

type TransferHandler struct {
	service *service.TransferService
}

func NewTransferHandler(service *service.TransferService) *TransferHandler {
	return &TransferHandler{service: service}
}

// ExecuteTransfer godoc
// @Summary Execute a transfer between accounts
// @Description Transfers funds between two accounts of the same client. Same currency transfers are direct, different currencies use Exchange Office with 0-1% commission.
// @Tags transfers
// @Accept json
// @Produce json
// @Param request body dto.CreateTransferRequest true "Transfer details"
// @Success 201 {object} dto.TransferResponse
// @Failure 400 {object} errors.AppError
// @Failure 422 {object} errors.AppError
// @Failure 500 {object} errors.AppError
// @Security BearerAuth
// @Router /api/transfers [post]
func (h *TransferHandler) ExecuteTransfer(c *gin.Context) {
	var req dto.CreateTransferRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(errors.BadRequestErr(err.Error()))
		return
	}

	transfer, err := h.service.ExecuteTransfer(c.Request.Context(), req)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusCreated, transfer)
}

// GetTransferHistory godoc
// @Summary Get transfer history
// @Description Retrieves client's transfer history with optional filtering by status and date range
// @Tags transfers
// @Produce json
// @Param account_num query string false "Filter by account number"
// @Param status query string false "Filter by status (Pending, Completed, Failed)"
// @Param start_date query string false "Filter from date (RFC3339 format)"
// @Param end_date query string false "Filter to date (RFC3339 format)"
// @Param page query int false "Page number" minimum(1)
// @Param page_size query int false "Page size" minimum(1) maximum(100)
// @Success 200 {object} dto.ListTransfersResponse
// @Failure 400 {object} errors.AppError
// @Failure 401 {object} errors.AppError
// @Security BearerAuth
// @Router /api/transfers [get]
func (h *TransferHandler) GetTransferHistory(c *gin.Context) {
	var query dto.ListTransfersQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		c.Error(errors.BadRequestErr(err.Error()))
		return
	}

	if query.Page == 0 {
		query.Page = 1
	}
	if query.PageSize == 0 {
		query.PageSize = 10
	}

	result, err := h.service.GetTransferHistory(
		c.Request.Context(),
		query.AccountNum,
		query.Status,
		query.StartDate,
		query.EndDate,
		query.Page,
		query.PageSize,
	)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, result)
}
