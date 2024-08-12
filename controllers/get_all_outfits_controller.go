package controllers

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/kshitij-404/dresstination-backend/modules"
)

func GetAllOutfits(c *gin.Context, firebaseClient *modules.FirebaseClient) {
	collection := "outfits" // Firestore collection name

	// Retrieve documents from Firestore
	userId := c.GetString("userId")
	documents, err := firebaseClient.GetAllDocuments(collection, userId, "userId")
	if err != nil {
		log.Println("Retrieval error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"Retrieval error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"outfits": documents})
}
