package offers

import (
	"errors"
	"fmt"

	consterrors "github.com/florin-rada/grm-back/src/const_errors"
	"github.com/florin-rada/grm-back/src/database"

	"gorm.io/gorm"
)

// The offer model will handle all the database related interactions
// Each user will have 1 of the packages active, by default, they will have the "Free" offer
// This will mean 3 max runners, and max sync rate of 60 seconds

// Offer represents one of the packages the tool offers
type Offer struct {
	gorm.Model
	Name           string `json:"name" gorm:"name"`
	MaxRunners     int    `json:"max_runners" gorm:"max_runners"`
	BasePrice      int    `json:"base_price" gorm:"base_price"` // price is in cents to avoid rounding errors (99 -> 99 cents 0.99 Dollars)
	Currency       string `json:"currency" gorm:"currency"`
	Description    string `json:"description" gorm:"description"`
	MaxSyncRate    uint   `json:"max_sync_rate" gorm:"max_sync_rate"` // how often can the system check the status
	Published      bool   `json:"published" gorm:"published"`
	IsDefaultOffer bool   `json:"is_default_offer" gorm:"is_default_offer,unique"`
}

type OfferRepository struct {
	db *gorm.DB
}

func NewOfferRepository(db *gorm.DB) *OfferRepository {
	return &OfferRepository{db}
}
func (o *OfferRepository) AddOffer(offer Offer) error {
	// Validate the offer fields
	if offer.Name == "" {
		return consterrors.ErrEmptyOfferName
	}
	if offer.MaxRunners < 0 {
		return consterrors.ErrNegativeMaxRunners
	}
	if offer.BasePrice < 0 {
		return consterrors.ErrNegativeBasePrice
	}
	if offer.Currency == "" {
		return consterrors.ErrEmptyCurrency
	}
	if offer.MaxSyncRate == 0 {
		return consterrors.ErrZerodMaxSyncRate
	}

	// Add the offer to the database
	result := o.db.Create(&offer)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

func (o *OfferRepository) EditOffer(id uint, updatedOffer Offer) error {
	// Find the offer to update
	var offer Offer
	result := o.db.First(&offer, id)
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
	result = o.db.Save(&offer)
	if result.Error != nil {
		return result.Error
	}

	return nil
}

func (o *OfferRepository) DeleteOffer(id uint) error {
	// Find the offer to delete
	var offer Offer
	result := o.db.First(&offer, id)
	if result.Error != nil {
		return result.Error
	}

	// Delete the offer from the database
	result = o.db.Delete(&offer)
	if result.Error != nil {
		return result.Error
	}

	return nil
}

func (o *OfferRepository) GetOffers() ([]Offer, error) {
	var offers []Offer
	result := o.db.Find(&offers)
	if result.Error != nil {
		return nil, result.Error
	}
	return offers, nil
}

func (o *OfferRepository) PublishOffer(id int) error {
	offer, err := o.GetOffer(id)
	if err != nil {
		return err
	}
	offer.Published = true
	result := o.db.Save(&offer)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

func (o *OfferRepository) UnpublishOffer(id int) error {
	offer, err := o.GetOffer(id)
	if err != nil {
		return err
	}
	offer.Published = true
	result := o.db.Save(&offer)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

func (o *OfferRepository) GetOffer(id int) (Offer, error) {
	var offer Offer
	result := o.db.First(&offer)
	if result.Error != nil {
		return Offer{}, result.Error
	}
	return offer, nil
}

func (o *OfferRepository) GetDefaultOffer() (Offer, error) {
	var offer Offer
	result := o.db.Model(&offer).Where("is_default_offer=?", 1).First(&offer)
	if result.Error != nil && errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return Offer{
			Name:           "Free Tier",
			MaxRunners:     5,
			BasePrice:      0.0,
			Currency:       "USD",
			MaxSyncRate:    60, // 60 seconds
			IsDefaultOffer: true,
			Published:      true,
		}, nil
	} else if result.Error != nil {
		sql := o.db.ToSQL(func(tx *gorm.DB) *gorm.DB {
			return tx.Model(&offer).Where("is_default_offer=?", 1).First(&offer)
		})
		fmt.Printf("The sql statement that failed: %s", sql)
		return Offer{}, result.Error
	}
	return offer, nil
}

/* // AddOffer adds a new offer to the database
func (r *OfferRepository) AddOffer(offer Offer) error {
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
	result := r.db.Create(&offer)
	if result.Error != nil {
		return result.Error
	}

	return nil
} */

/* func AddOffer(offer Offer) error {
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
} */

/*
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
*/
func init() {
	err := database.PublicDB.AutoMigrate(&Offer{})
	if err != nil {
		panic(fmt.Sprintf("Error migrating `offers` table: %s", err.Error()))
	}
}
