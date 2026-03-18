package repository

import (
	"banking-service/internal/dto"
	"banking-service/internal/model"
	"context"

	"gorm.io/gorm"
)

type paymentRepository struct {
	db *gorm.DB
}

func NewPaymentRepository(db *gorm.DB) PaymentRepository {
	return &paymentRepository{db: db}
}

func (r *paymentRepository) Create(ctx context.Context, payment *model.Payment) error {
	return r.db.WithContext(ctx).Create(payment).Error
}

func (r *paymentRepository) GetByID(ctx context.Context, id uint) (*model.Payment, error) {
	var payment model.Payment
	if err := r.db.WithContext(ctx).First(&payment, id).Error; err != nil {
		return nil, err
	}
	return &payment, nil
}

func (r *paymentRepository) Update(ctx context.Context, payment *model.Payment) error {
	return r.db.WithContext(ctx).Save(payment).Error
}

func (r *paymentRepository) FindByAccount(ctx context.Context, accountNumber string, filters *dto.PaymentFilters) ([]model.Payment, int64, error) {
	page := filters.Page
	if page < 1 {
		page = 1
	}
	pageSize := filters.PageSize
	if pageSize < 1 {
		pageSize = 10
	}

	query := r.db.WithContext(ctx).Model(&model.Payment{}).
		Where("payer_account = ? OR recipient_account = ?", accountNumber, accountNumber)

	if filters.Status != "" {
		query = query.Where("status = ?", filters.Status)
	}
	if !filters.StartDate.IsZero() {
		query = query.Where("created_at >= ?", filters.StartDate)
	}
	if !filters.EndDate.IsZero() {
		query = query.Where("created_at <= ?", filters.EndDate)
	}
	if filters.MinAmount > 0 {
		query = query.Where("amount >= ?", filters.MinAmount)
	}
	if filters.MaxAmount > 0 {
		query = query.Where("amount <= ?", filters.MaxAmount)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var payments []model.Payment
	offset := (page - 1) * pageSize
	err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&payments).Error
	return payments, total, err
}
