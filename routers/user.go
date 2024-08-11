package routers

import (
    "github.com/gin-gonic/gin"
    "github.com/kshitij-404/dresstination-backend/controllers"
    "github.com/kshitij-404/dresstination-backend/middleware"
)

func UserRoute(r *gin.Engine) {
    r.GET("/user", middleware.TokenAuthMiddleware(), controllers.GetUser)
}