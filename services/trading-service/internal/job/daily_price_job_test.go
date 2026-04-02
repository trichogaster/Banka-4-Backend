package job

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/RAF-SI-2025/Banka-4-Backend/services/trading-service/internal/model"
	"github.com/RAF-SI-2025/Banka-4-Backend/services/trading-service/internal/repository"
	"github.com/stretchr/testify/require"
)

type mockListingRepo struct {
	listings              []model.Listing
	forexListings         []model.Listing
	findAllErr            error
	findByTypeErr         error
	findLastDailyInfo     *model.ListingDailyPriceInfo
	findLastDailyInfoErr  error
	createDailyInfoCalled []model.ListingDailyPriceInfo
	createDailyInfoErr    error
}

func (m *mockListingRepo) FindStocks(ctx context.Context, filter repository.ListingFilter) ([]model.Listing, int64, error) {
	return nil, 0, nil
}

func (m *mockListingRepo) FindFutures(ctx context.Context, filter repository.ListingFilter) ([]model.Listing, int64, error) {
	return nil, 0, nil
}

func (m *mockListingRepo) FindOptions(ctx context.Context, filter repository.ListingFilter) ([]model.Listing, int64, error) {
	return nil, 0, nil
}

func (m *mockListingRepo) FindByID(ctx context.Context, id uint) (*model.Listing, error) {
	return nil, nil
}

func (m *mockListingRepo) Upsert(ctx context.Context, listing *model.Listing) error {
	return nil
}

func (m *mockListingRepo) UpdatePriceAndAsk(ctx context.Context, listing *model.Listing, price, ask float64) error {
	return nil
}

func (m *mockListingRepo) Count(ctx context.Context) (int64, error) {
	return 0, nil
}

func (m *mockListingRepo) FindAll(ctx context.Context) ([]model.Listing, error) {
	return m.listings, m.findAllErr
}

func (m *mockListingRepo) FindByType(ctx context.Context, typ model.ListingType) ([]model.Listing, error) {
	if typ == model.ListingTypeForexPair {
		return m.forexListings, m.findByTypeErr
	}
	return nil, nil
}

func (m *mockListingRepo) FindLastDailyPriceInfo(ctx context.Context, listingID uint, beforeDate time.Time) (*model.ListingDailyPriceInfo, error) {
	return m.findLastDailyInfo, m.findLastDailyInfoErr
}

func (m *mockListingRepo) CreateDailyPriceInfo(ctx context.Context, info *model.ListingDailyPriceInfo) error {
	if m.createDailyInfoErr != nil {
		return m.createDailyInfoErr
	}
	m.createDailyInfoCalled = append(m.createDailyInfoCalled, *info)
	return nil
}

type mockForexRepo struct {
	forexPairs            map[uint]model.ForexPair
	findByListingIDsErr   error
	findLastDailyInfo     *model.ForexPairDailyPriceInfo
	findLastDailyInfoErr  error
	createDailyInfoCalled []model.ForexPairDailyPriceInfo
	createDailyInfoErr    error
}

func (m *mockForexRepo) Count(ctx context.Context) (int64, error) {
	return 0, nil
}

func (m *mockForexRepo) Upsert(ctx context.Context, pair model.ForexPair) error {
	return nil
}

func (m *mockForexRepo) FindAll(ctx context.Context, filter repository.ListingFilter) ([]model.ForexPair, int64, error) {
	return nil, 0, nil
}

func (m *mockForexRepo) FindAllForexPairs(ctx context.Context) ([]model.ForexPair, error) {
	return nil, nil
}

func (m *mockForexRepo) FindByListingIDs(ctx context.Context, listingIDs []uint) ([]model.ForexPair, error) {
	if m.findByListingIDsErr != nil {
		return nil, m.findByListingIDsErr
	}
	var pairs []model.ForexPair
	for _, id := range listingIDs {
		if p, ok := m.forexPairs[id]; ok {
			pairs = append(pairs, p)
		}
	}
	return pairs, nil
}

func (m *mockForexRepo) FindLastDailyPriceInfo(ctx context.Context, forexPairID uint, beforeDate time.Time) (*model.ForexPairDailyPriceInfo, error) {
	return m.findLastDailyInfo, m.findLastDailyInfoErr
}

func (m *mockForexRepo) CreateDailyPriceInfo(ctx context.Context, info *model.ForexPairDailyPriceInfo) error {
	if m.createDailyInfoErr != nil {
		return m.createDailyInfoErr
	}
	m.createDailyInfoCalled = append(m.createDailyInfoCalled, *info)
	return nil
}

func TestDailyPriceJob_Run(t *testing.T) {
	tests := []struct {
		name                 string
		listings             []model.Listing
		forexListings        []model.Listing
		forexPairs           map[uint]model.ForexPair
		prevListingInfo      *model.ListingDailyPriceInfo
		prevForexInfo        *model.ForexPairDailyPriceInfo
		expectedListingInfos int
		expectedForexInfos   int
		expectError          bool
	}{
		{
			name: "happy path - stocks and forex",
			listings: []model.Listing{
				{ListingID: 1, ListingType: model.ListingTypeStock, Ticker: "AAPL", Price: 150.0, Ask: 151.0},
				{ListingID: 2, ListingType: model.ListingTypeFuture, Ticker: "CLJ22", Price: 210.0, Ask: 211.0},
			},
			forexListings: []model.Listing{
				{ListingID: 3, ListingType: model.ListingTypeForexPair, Ticker: "EUR/USD", Price: 1.20, Ask: 1.21},
			},
			forexPairs: map[uint]model.ForexPair{
				3: {ForexPairID: 100, ListingID: 3, Base: "EUR", Quote: "USD", Rate: 1.20},
			},
			prevListingInfo:      nil, // no previous day -> change = 0
			prevForexInfo:        nil,
			expectedListingInfos: 2,
			expectedForexInfos:   1,
			expectError:          false,
		},
		{
			name: "with previous day data - change calculated",
			listings: []model.Listing{
				{ListingID: 1, ListingType: model.ListingTypeStock, Price: 150.0},
			},
			forexListings: []model.Listing{},
			prevListingInfo: &model.ListingDailyPriceInfo{
				Price: 100.0,
			},
			expectedListingInfos: 1,
			expectedForexInfos:   0,
			expectError:          false,
		},
		{
			name: "error on listing create - continues",
			listings: []model.Listing{
				{ListingID: 1, ListingType: model.ListingTypeStock, Price: 150.0},
			},
			forexListings:        []model.Listing{},
			expectedListingInfos: 0, // create will fail, but job should continue
			expectError:          false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockListing := &mockListingRepo{
				listings:              tt.listings,
				forexListings:         tt.forexListings,
				findLastDailyInfo:     tt.prevListingInfo,
				createDailyInfoCalled: []model.ListingDailyPriceInfo{},
			}
			if tt.name == "error on listing create - continues" {
				mockListing.createDailyInfoErr = errors.New("db error")
			}

			mockForex := &mockForexRepo{
				forexPairs:            tt.forexPairs,
				findLastDailyInfo:     tt.prevForexInfo,
				createDailyInfoCalled: []model.ForexPairDailyPriceInfo{},
			}

			job := NewDailyPriceJob(mockListing, mockForex)

			ctx := context.Background()
			err := job.Run(ctx)

			if tt.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			// Proveri broj kreiranih listing zapisa
			require.Len(t, mockListing.createDailyInfoCalled, tt.expectedListingInfos)

			// Proveri da li je change ispravno izračunat za listing (ako postoji prethodni dan)
			if tt.prevListingInfo != nil && len(mockListing.createDailyInfoCalled) > 0 {
				info := mockListing.createDailyInfoCalled[0]
				expectedChange := (150.0 - 100.0) / 100.0 * 100 // 50%
				require.InDelta(t, expectedChange, info.Change, 0.001)
			}

			// Proveri forex zapise
			require.Len(t, mockForex.createDailyInfoCalled, tt.expectedForexInfos)
		})
	}
}
