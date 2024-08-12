package routers

import (
	"github.com/gin-gonic/gin"
	"github.com/kshitij-404/dresstination-backend/controllers"
	"github.com/kshitij-404/dresstination-backend/middleware"
	"github.com/kshitij-404/dresstination-backend/modules"
)

func UserRoute(r *gin.Engine, firebaseClient *modules.FirebaseClient, fs *modules.FS) {
	r.GET("/user", middleware.TokenAuthMiddleware(), func(c *gin.Context) {
		controllers.GetUser(c, firebaseClient, fs)
	})
}
