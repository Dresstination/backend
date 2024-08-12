package main

import (
	"log"
	"github.com/gin-gonic/gin"
	"github.com/kshitij-404/dresstination-backend/routers"
	"github.com/joho/godotenv"
)

func init() {
    // Load the .env file
    err := godotenv.Load()
    if err != nil {
        log.Fatal("Error loading .env file")
    }
}

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
