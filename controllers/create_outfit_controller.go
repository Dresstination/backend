package controllers

import (
    "net/http"

    "github.com/gin-gonic/gin"
)

type Product struct {
    ID       string  `json:"id"`
    Image    string  `json:"image"`
    Title    string  `json:"title"`
    Price    float64 `json:"price"`
    Currency string  `json:"currency"`
    Link     string  `json:"link"`
}

type OutfitElement struct {
    ID          string    `json:"id"`
    Title       string    `json:"title"`
    Description string    `json:"description"`
    ImageLink   string    `json:"image_link"`
    SearchQuery string    `json:"search_query"`
    Products    []Product `json:"products"`
}

type Outfit struct {
    ID            string          `json:"id"`
    Title         string          `json:"title"`
    OutfitElements []OutfitElement `json:"outfit_elements"`
}

func CreateOutfits(c *gin.Context) {
    var newOutfit Outfit

    // Bind the JSON body to the newOutfit struct
    if err := c.ShouldBindJSON(&newOutfit); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    // Here you would typically save the newOutfit to a database
    // For this example, we'll just return the newOutfit as a response
    c.JSON(http.StatusCreated, newOutfit)
}