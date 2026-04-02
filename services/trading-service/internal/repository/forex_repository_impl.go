package repository

import (
	"context"
	"errors"
	"time"

	"github.com/RAF-SI-2025/Banka-4-Backend/services/trading-service/internal/model"
	"gorm.io/gorm"
)

type forexRepository struct {
	db *gorm.DB
}

func NewForexRepository(db *gorm.DB) ForexRepository {
	return &forexRepository{db: db}
}

func (r *forexRepository) Count(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&model.ForexPair{}).
		Count(&count).Error

	return count, err
}

func (r *forexRepository) Upsert(ctx context.Context, pair model.ForexPair) error {
	return r.db.WithContext(ctx).
		Where("base = ? AND quote = ?", pair.Base, pair.Quote).
		Assign(pair).
		FirstOrCreate(&pair).Error
}

func (r *forexRepository) FindAll(ctx context.Context, filter ListingFilter) ([]model.ForexPair, int64, error) {
	q := r.db.WithContext(ctx).Model(&model.ForexPair{})

	if filter.Search != "" {
		like := "%" + filter.Search + "%"
		q = q.Where("base LIKE ? OR quote LIKE ?", like, like)
	}
	if filter.PriceMin > 0 {
		q = q.Where("rate >= ?", filter.PriceMin)
	}
	if filter.PriceMax > 0 {
		q = q.Where("rate <= ?", filter.PriceMax)
	}

	var count int64
	if err := q.Count(&count).Error; err != nil {
		return nil, 0, err
	}

	dir := "ASC"
	if filter.SortDir == "desc" {
		dir = "DESC"
	}

	var pairs []model.ForexPair
	err := q.Order("rate " + dir).
		Limit(filter.PageSize).
		Offset((filter.Page - 1) * filter.PageSize).
		Find(&pairs).Error

	return pairs, count, err
}

func (r *forexRepository) FindByListingIDs(ctx context.Context, listingIDs []uint) ([]model.ForexPair, error) {
	var pairs []model.ForexPair
	if err := r.db.WithContext(ctx).Where("listing_id IN ?", listingIDs).Preload("Listing").Find(&pairs).Error; err != nil {
		return nil, err
	}
	return pairs, nil
}

func (r *forexRepository) CreateDailyPriceInfo(ctx context.Context, info *model.ForexPairDailyPriceInfo) error {
	return r.db.WithContext(ctx).Create(&info).Error
}

func (r *forexRepository) FindLastDailyPriceInfo(ctx context.Context, forexPairID uint, beforeDate time.Time) (*model.ForexPairDailyPriceInfo, error) {
	var info model.ForexPairDailyPriceInfo
	err := r.db.WithContext(ctx).
		Where("forex_pair_id = ? AND date < ?", forexPairID, beforeDate).
		Order("date DESC").
		First(&info).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &info, err
}

func (r *forexRepository) FindAllForexPairs(ctx context.Context) ([]model.ForexPair, error) {
	var pairs []model.ForexPair
	err := r.db.WithContext(ctx).Find(&pairs).Error
	return pairs, err
}
