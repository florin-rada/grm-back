package offers

import (
	"back/database"
	"fmt"
	"time"

	"gorm.io/gorm"
)

// The offer model will handle all the database related interactions
// Each user will have 1 of the packages active, by default, they will have the "Free" offer
// This will mean 3 max runners, and max sync rate of 60 seconds

// Offer represents one of the packages the tool offers
type Offer struct {
	gorm.Model
	Name        string    `json:"name" gorm:"name"`
	MaxRunners  int       `json:"max_runners" gorm:"max_runners"`
	BasePrice   int       `json:"base_price" gorm:"base_price"` // price is in cents to avoid rounding errors (99 -> 99 cents 0.99 Dollars)
	Currency    string    `json:"currency" gorm:"currency"`
	Description string    `json:"description" gorm:"description"`
	MaxSyncRate time.Time `json:"max_sync_rate" gorm:"max_sync_rate"`
	Published   bool      `json:"published" gorm:"published"`
}

func AddOffer(offer Offer) (Offer, error) {
	err := database.PublicDB.Save(&offer)
	if err.Error != nil {
		return Offer{}, err.Error
	}

	return offer, nil
}

func EditOffer(offer Offer) (Offer, error) {

	return offer, nil
}

func init() {
	err := database.PublicDB.AutoMigrate(&Offer{})
	if err != nil {
		panic(fmt.Sprintf("Error migrating `offers` table: %s", err.Error()))
	}
}
