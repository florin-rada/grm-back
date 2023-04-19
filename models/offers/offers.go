package offers

import (
	"back/database"
	"errors"
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
	MaxSyncRate time.Time `json:"max_sync_rate" gorm:"max_sync_rate"` // how often can the system check the status
	Published   bool      `json:"published" gorm:"published"`
}

func AddOffer(offer Offer) error {
	// Validate the offer fields
	if offer.Name == "" {
		return errors.New("Offer name cannot be empty")
	}
	if offer.MaxRunners < 0 {
		return errors.New("MaxRunners cannot be negative")
	}
	if offer.BasePrice < 0 {
		return errors.New("BasePrice cannot be negative")
	}
	if offer.Currency == "" {
		return errors.New("Currency cannot be empty")
	}
	if offer.MaxSyncRate.IsZero() {
		return errors.New("MaxSyncRate cannot be zero")
	}

	// Add the offer to the database
	result := database.PublicDB.Create(&offer)
	if result.Error != nil {
		return result.Error
	}

	return nil
}

func EditOffer(id uint, updatedOffer Offer) error {
	// Find the offer to update
	var offer Offer
	result := database.PublicDB.First(&offer, id)
	if result.Error != nil {
		return result.Error
	}

	// Update the offer fields
	offer.Name = updatedOffer.Name
	offer.MaxRunners = updatedOffer.MaxRunners
	offer.BasePrice = updatedOffer.BasePrice
	offer.Currency = updatedOffer.Currency
	offer.Description = updatedOffer.Description
	offer.MaxSyncRate = updatedOffer.MaxSyncRate
	offer.Published = updatedOffer.Published

	// Save the changes to the database
	result = database.PublicDB.Save(&offer)
	if result.Error != nil {
		return result.Error
	}

	return nil
}

func DeleteOffer(id uint) error {
	// Find the offer to delete
	var offer Offer
	result := database.PublicDB.First(&offer, id)
	if result.Error != nil {
		return result.Error
	}

	// Delete the offer from the database
	result = database.PublicDB.Delete(&offer)
	if result.Error != nil {
		return result.Error
	}

	return nil
}

func init() {
	err := database.PublicDB.AutoMigrate(&Offer{})
	if err != nil {
		panic(fmt.Sprintf("Error migrating `offers` table: %s", err.Error()))
	}
}
