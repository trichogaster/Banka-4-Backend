package job

import (
	"context"
	"log"
	"time"

	"github.com/RAF-SI-2025/Banka-4-Backend/services/trading-service/internal/model"
	"github.com/RAF-SI-2025/Banka-4-Backend/services/trading-service/internal/repository"
)

type DailyPriceJob struct {
	listingRepo repository.ListingRepository
	forexRepo   repository.ForexRepository
}

func NewDailyPriceJob(
	listingRepo repository.ListingRepository,
	forexRepo repository.ForexRepository,
) *DailyPriceJob {
	return &DailyPriceJob{
		listingRepo: listingRepo,
		forexRepo:   forexRepo,
	}
}

// Run se poziva jednom dnevno (npr. u 00:00 UTC)
func (j *DailyPriceJob) Run(ctx context.Context) error {
	today := time.Now().Truncate(24 * time.Hour)

	// 1. ListingDailyPriceInfo za sve listinge
	if err := j.createListingDailyInfos(ctx, today); err != nil {
		log.Printf("Failed to create listing daily infos: %v", err)
		// Ne vraćamo grešku da bi forex deo ipak pokušao
	}

	// 2. ForexPairDailyPriceInfo za sve forex parove
	if err := j.createForexDailyInfos(ctx, today); err != nil {
		log.Printf("Failed to create forex daily infos: %v", err)
		return err
	}

	log.Println("Daily price info job completed successfully")
	return nil
}

func (j *DailyPriceJob) createListingDailyInfos(ctx context.Context, date time.Time) error {
	// Dohvati sve listinge (možeš koristiti FindAll ako postoji, ili FindStocks/FindFutures/FindOptions)
	listings, err := j.listingRepo.FindAll(ctx)
	if err != nil {
		return err
	}

	for _, listing := range listings {
		if listing.ListingType == model.ListingTypeForexPair {
			continue
		}
		// Izračunaj change u odnosu na prethodni dan (opciono)
		prevInfo, _ := j.listingRepo.FindLastDailyPriceInfo(ctx, listing.ListingID, date)
		change := 0.0
		if prevInfo != nil && prevInfo.Price != 0 {
			change = (listing.Price - prevInfo.Price) / prevInfo.Price * 100
		}

		dailyInfo := &model.ListingDailyPriceInfo{
			ListingID: listing.ListingID,
			Date:      date,
			Price:     listing.Price,
			Ask:       listing.Ask,
			Bid:       0.0, // Listing model nema Bid, pa ostavi 0 ili ga dodaj
			Change:    change,
			Volume:    0, // Ako nemaš podatak o volumenu
		}

		if err := j.listingRepo.CreateDailyPriceInfo(ctx, dailyInfo); err != nil {
			log.Printf("Failed to save daily price for listing %d: %v", listing.ListingID, err)
			// Nastavi sa ostalima
		}
	}
	return nil
}

func (j *DailyPriceJob) createForexDailyInfos(ctx context.Context, date time.Time) error {
	listings, err := j.listingRepo.FindByType(ctx, model.ListingTypeForexPair)
	if err != nil {
		return err
	}

	for _, listing := range listings {

		pairs, err := j.forexRepo.FindByListingIDs(ctx, []uint{listing.ListingID})
		if err != nil {
			log.Printf("Forex pair not found for listing %d: %v", listing.ListingID, err)
			continue
		}
		if len(pairs) == 0 {
			log.Printf("No forex pair for listing %d", listing.ListingID)
			continue
		}
		forexPair := pairs[0]

		// Izračunaj change u odnosu na prethodni dan
		prevInfo, _ := j.forexRepo.FindLastDailyPriceInfo(ctx, forexPair.ForexPairID, date)
		change := 0.0
		if prevInfo != nil && prevInfo.Rate != 0 {
			change = (forexPair.Rate - prevInfo.Rate) / prevInfo.Rate * 100
		}

		dailyInfo := &model.ForexPairDailyPriceInfo{
			ForexPairID: forexPair.ForexPairID,
			Date:        date,
			Rate:        forexPair.Rate,
			Ask:         listing.Ask,
			Bid:         0.0,
			Change:      change,
			Volume:      0,
		}

		if err := j.forexRepo.CreateDailyPriceInfo(ctx, dailyInfo); err != nil {
			log.Printf("Failed to save daily forex price for pair %d: %v", forexPair.ForexPairID, err)
		}
	}
	return nil
}
