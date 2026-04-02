package model

import "time"

type ForexPair struct {
	ForexPairID uint `gorm:"primaryKey"`

	ListingID uint `gorm:"not null;uniqueIndex"`
	Listing   Listing

	Base  string `gorm:"size:3;not null;uniqueIndex:idx_pair"`
	Quote string `gorm:"size:3;not null;uniqueIndex:idx_pair"`

	Rate float64 `gorm:"not null"`

	ProviderUpdatedAt    time.Time
	ProviderNextUpdateAt time.Time

	CreatedAt time.Time
	UpdatedAt time.Time
}
type ForexPairDailyPriceInfo struct {
	ID          uint `gorm:"primaryKey;autoIncrement"`
	ForexPairID uint `gorm:"not null;index"`
	ForexPair   ForexPair
	Date        time.Time `gorm:"not null;index"`
	Rate        float64   `gorm:"not null;default:0"` // close rate
	Ask         float64   `gorm:"not null;default:0"`
	Bid         float64   `gorm:"not null;default:0"`
	Change      float64   `gorm:"not null;default:0"`
	Volume      uint      `gorm:"not null;default:0"`
}
