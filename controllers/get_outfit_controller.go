package controllers

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/kshitij-404/dresstination-backend/modules"
)

func GetOutfit(c *gin.Context, firebaseClient *modules.FirebaseClient) {
	collection := "outfits" // Firestore collection name

	// Retrieve documents from Firestore
	params := c.Request.URL.Query()
	if len(params) > 0 {
		documentID := params.Get("id")
		if documentID != "" {
			document, err := firebaseClient.GetDocument(collection, documentID)
			if err != nil {
				log.Println("Retrieval error", err)
				c.JSON(http.StatusInternalServerError, gin.H{"Retrieval error": err.Error()})
				return

			}
			c.JSON(http.StatusOK, gin.H{"outfit": document})
			return
		}
	}

}
