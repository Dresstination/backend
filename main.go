package main

import (
	"context"
	"log"
	"github.com/gin-gonic/gin"
	"github.com/kshitij-404/dresstination-backend/routers"
	"github.com/joho/godotenv"
	"github.com/kshitij-404/dresstination-backend/modules"
	"os"
	"time"
)

var fs *modules.FS

func init() {
    // Load the .env file
    err := godotenv.Load()
    if err != nil {
        log.Fatal("Error loading .env file")
    }

	// var err error
    fs, err = modules.NewFS(os.Getenv("FIREBASE_SECRETS_JSON"), 10*time.Second)
    if err != nil {
        log.Fatalf("Failed to initialize Firebase Storage: %v", err)
    }
}

func setupRouter(firebaseClient *modules.FirebaseClient) *gin.Engine {

	r := gin.Default()

	// Import routes
	routers.PingRoute(r)
	routers.UserRoute(r, firebaseClient, fs)
	routers.AuthRoutes(r, firebaseClient, fs)
	routers.OutfitRoute(r, firebaseClient, fs)

	return r
}

func main() {
	// setup firebase client
	firebaseClient, err := modules.NewFirebaseClient(context.Background(), os.Getenv("FIREBASE_PROJECT_ID"), []byte(os.Getenv("FIREBASE_SECRETS_JSON")))

	if err != nil {
        log.Fatalf("Error initializing Firebase client: %v", err)
    }

	r := setupRouter(firebaseClient)

	r.Run(":8080")
}
