package routers

import (
	"github.com/gin-gonic/gin"
	"github.com/kshitij-404/dresstination-backend/controllers"
	"github.com/kshitij-404/dresstination-backend/middleware"
	"github.com/kshitij-404/dresstination-backend/modules"
)

func OutfitRoute(r *gin.Engine, firebaseClient *modules.FirebaseClient, fs *modules.FS) {
	r.POST("/outfits", middleware.TokenAuthMiddleware(), func(c *gin.Context) {
		controllers.CreateOutfit(c, firebaseClient, fs)
	})
	r.GET("/outfits", middleware.TokenAuthMiddleware(), func(c *gin.Context) {
		controllers.GetOutfit(c, firebaseClient)
	})
	r.GET("/outfits", middleware.TokenAuthMiddleware(), func(c *gin.Context) {
		controllers.GetAllOutfits(c, firebaseClient)
	})
}
