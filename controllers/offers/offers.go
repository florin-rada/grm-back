package offers

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/florin-rada/grm-back/models/offers"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type OfferController struct {
	db              *gorm.DB
	offerRepository offers.OfferRepository
}

func NewOfferController(db *gorm.DB) *OfferController {
	return &OfferController{
		db:              db,
		offerRepository: *offers.NewOfferRepository(db),
	}
}

func (oc *OfferController) CreateOffer(c *gin.Context) {
	// Parse the request body to get the new offer data
	var offer offers.Offer
	if err := c.ShouldBindJSON(&offer); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Add the offer to the database
	if err := oc.offerRepository.AddOffer(offer); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Return a success response
	c.JSON(http.StatusCreated, gin.H{"data": offer})
}

func (oc *OfferController) GetOffers(c *gin.Context) {
	// Get all offers from the database
	var offers []offers.Offer
	result := oc.db.Find(&offers)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}

	// Return the offers as a JSON response
	c.JSON(http.StatusOK, gin.H{"data": offers})
}

func (oc *OfferController) GetOffer(c *gin.Context) {
	// Get the offer ID from the URL parameter
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid offer ID"})
		return
	}

	// Find the offer with the given ID in the database
	var offer offers.Offer
	result := oc.db.First(&offer, id)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Offer not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		}
		return
	}

	// Return the offer as a JSON response
	c.JSON(http.StatusOK, gin.H{"data": offer})
}

func (oc *OfferController) UpdateOffer(c *gin.Context) {
	// Get the offer ID from the URL parameter
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid offer ID"})
		return
	}

	// Parse the request body to get the updated offer data
	var updatedOffer offers.Offer
	if err := c.ShouldBindJSON(&updatedOffer); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Update the offer in the database
	result := oc.db.Model(&offers.Offer{Model: gorm.Model{ID: uint(id)}}).Updates(updatedOffer)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}

	// Return a success response
	c.JSON(http.StatusOK, gin.H{"data": "Offer updated successfully"})
}

func (oc *OfferController) DeleteOffer(c *gin.Context) {
	// Get the offer ID from the URL parameter
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid offer ID"})
		return
	}

	// Delete the offer from the database
	result := oc.db.Delete(&offers.Offer{}, id)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}

	// Return a success response
	c.JSON(http.StatusOK, gin.H{"data": "Offer deleted successfully"})
}
