package routers

import (
    "github.com/gin-gonic/gin"
    "github.com/kshitij-404/dresstination-backend/controllers"
    "github.com/kshitij-404/dresstination-backend/middleware"
)

func AuthRoutes(r *gin.Engine) {
    authorized := r.Group("/admin")
    authorized.Use(middleware.TokenAuthMiddleware())

    authorized.GET("/secrets", controllers.GetSecrets)
}