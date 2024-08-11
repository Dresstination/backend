package routers

import (
	"github.com/gin-gonic/gin"
	"github.com/kshitij-404/dresstination-backend/controllers"
	"github.com/kshitij-404/dresstination-backend/middleware"
)

func OutfitRoute(r *gin.Engine) {
	r.POST("/outfits", middleware.TokenAuthMiddleware(), controllers.CreateOutfits)
	// r.GET("/outfits/:id", middleware.TokenAuthMiddleware(), controllers.GetOutfits)
	// r.GET("/outfits", middleware.TokenAuthMiddleware(), controllers.GetOutfitsHistory)
}
