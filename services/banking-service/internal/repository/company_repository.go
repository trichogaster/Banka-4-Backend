package repository

import (
	"banking-service/internal/model"
	"context"
)

type CompanyRepository interface {
	Create(ctx context.Context, company *model.Company) error
}
