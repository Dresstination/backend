package routers

import (
    "github.com/gin-gonic/gin"
    "github.com/kshitij-404/dresstination-backend/controllers"
    "github.com/kshitij-404/dresstination-backend/middleware"
    "github.com/kshitij-404/dresstination-backend/modules"
)

func AuthRoutes(r *gin.Engine, firebaseClient *modules.FirebaseClient, fs *modules.FS) {
    authorized := r.Group("/admin")
    authorized.Use(middleware.TokenAuthMiddleware())

    authorized.GET("/secrets", func(c *gin.Context) {
        controllers.GetSecrets(c, firebaseClient, fs)
    })
}