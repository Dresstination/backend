package main

import (
	"github.com/gin-gonic/gin"
	"github.com/kshitij-404/dresstination-backend/routers"
)

func setupRouter() *gin.Engine {

	r := gin.Default()

	// Import routes
	routers.PingRoute(r)
	routers.UserRoute(r)
	routers.AuthRoutes(r)
	routers.OutfitRoute(r)

	return r
}

func main() {
	r := setupRouter()
	r.Run(":8080")
}
