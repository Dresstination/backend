package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/kshitij-404/dresstination-backend/modules"
)

var secrets = gin.H{
	"foo":    gin.H{"email": "foo@bar.com", "phone": "123433"},
	"austin": gin.H{"email": "austin@example.com", "phone": "666"},
	"lena":   gin.H{"email": "lena@guapa.com", "phone": "523443"},
	"manu":   gin.H{"email": "manu@pizza.com", "phone": "888"},
}

// GetSecrets handles the GET /admin/secrets route
func GetSecrets(c *gin.Context, firebaseClient *modules.FirebaseClient, fs *modules.FS) {
	user := c.MustGet("user").(string)
	if secret, ok := secrets[user]; ok {
		c.JSON(http.StatusOK, gin.H{"user": user, "secret": secret})
	} else {
		c.JSON(http.StatusOK, gin.H{"user": user, "secret": "NO SECRET :("})
	}
}
