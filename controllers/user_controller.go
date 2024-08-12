package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/kshitij-404/dresstination-backend/modules"
)

var db = map[string]string{
	"foo":    "foo_value",
	"austin": "austin_value",
	"lena":   "lena_value",
	"manu":   "manu_value",
}

// GetUser handles the GET /user route
func GetUser(c *gin.Context, firebaseClient *modules.FirebaseClient, fs *modules.FS) {
	user := c.MustGet("user").(string)
	value, ok := db[user]
	if ok {
		c.JSON(http.StatusOK, gin.H{"user": user, "value": value})
	} else {
		c.JSON(http.StatusOK, gin.H{"user": user, "status": "no value"})
	}
}
