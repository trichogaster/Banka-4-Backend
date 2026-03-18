package handler

import (
	"banking-service/internal/dto"
	"banking-service/internal/service"
	"common/pkg/auth"
	"common/pkg/errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

type AccountHandler struct {
	service *service.AccountService
}

func NewAccountHandler(service *service.AccountService) *AccountHandler {
	return &AccountHandler{service: service}
}

func (h *AccountHandler) Create(c *gin.Context) {
	var req dto.CreateAccountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(errors.BadRequestErr(err.Error()))
		return
	}

	account, err := h.service.Create(c.Request.Context(), req)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusCreated, dto.ToAccountResponse(account))
}

func (h *AccountHandler) GetClientAccounts(c *gin.Context) {
	authCtx := auth.GetAuth(c)
	if authCtx == nil || authCtx.ClientID == nil {
		c.Error(errors.BadRequestErr("client authentication required"))
		return
	}

	accounts, err := h.service.GetClientAccounts(c.Request.Context(), *authCtx.ClientID)
	if err != nil {
		c.Error(err)
		return
	}

	resp := make([]dto.AccountSummaryResponse, len(accounts))
	for i := range accounts {
		resp[i] = dto.ToAccountSummaryResponse(&accounts[i])
	}
	c.JSON(http.StatusOK, resp)
}

func (h *AccountHandler) GetAccountDetails(c *gin.Context) {
	authCtx := auth.GetAuth(c)
	if authCtx == nil || authCtx.ClientID == nil {
		c.Error(errors.BadRequestErr("client authentication required"))
		return
	}

	accountNumber := c.Param("accountNumber")
	account, err := h.service.GetAccountDetails(c.Request.Context(), accountNumber, *authCtx.ClientID)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, dto.ToAccountResponse(account))
}

func (h *AccountHandler) UpdateAccountName(c *gin.Context) {
	authCtx := auth.GetAuth(c)
	if authCtx == nil || authCtx.ClientID == nil {
		c.Error(errors.BadRequestErr("client authentication required"))
		return
	}

	accountNumber := c.Param("accountNumber")

	var req dto.UpdateAccountNameRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(errors.BadRequestErr(err.Error()))
		return
	}

	if err := h.service.UpdateAccountName(c.Request.Context(), accountNumber, *authCtx.ClientID, req.Name); err != nil {
		c.Error(err)
		return
	}

	c.Status(http.StatusOK)
}

func (h *AccountHandler) RequestLimitsChange(c *gin.Context) {
	authCtx := auth.GetAuth(c)
	if authCtx == nil || authCtx.ClientID == nil {
		c.Error(errors.BadRequestErr("client authentication required"))
		return
	}

	accountNumber := c.Param("accountNumber")

	var req dto.RequestLimitsChangeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(errors.BadRequestErr(err.Error()))
		return
	}

	code, err := h.service.RequestLimitsChange(c.Request.Context(), accountNumber, *authCtx.ClientID, req.DailyLimit, req.MonthlyLimit)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": code}) //only for testing purposes, in a real app it would be sent to mobile app
}

func (h *AccountHandler) ConfirmLimitsChange(c *gin.Context) {
	authCtx := auth.GetAuth(c)
	if authCtx == nil || authCtx.ClientID == nil {
		c.Error(errors.BadRequestErr("client authentication required"))
		return
	}

	accountNumber := c.Param("accountNumber")

	var req dto.ConfirmLimitsChangeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(errors.BadRequestErr(err.Error()))
		return
	}

	if err := h.service.ConfirmLimitsChange(c.Request.Context(), accountNumber, *authCtx.ClientID, req.Code); err != nil {
		c.Error(err)
		return
	}

	c.Status(http.StatusOK)
}
