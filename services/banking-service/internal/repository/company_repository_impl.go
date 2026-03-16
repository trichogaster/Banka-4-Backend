package repository

import (
	"banking-service/internal/model"
	"context"

	"gorm.io/gorm"
)

type companyRepository struct {
	db *gorm.DB
}

func NewCompanyRepository(db *gorm.DB) CompanyRepository {
	return &companyRepository{db: db}
}

func (r *companyRepository) Create(ctx context.Context, company *model.Company) error {
	return r.db.WithContext(ctx).Create(company).Error
}

