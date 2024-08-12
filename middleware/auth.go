package middleware

import (
	"net/http"
	"strings"

	"log"

	"github.com/gin-gonic/gin"
)

var tokens = map[string]string{
	"token_for_foo":    "foo",
	"token_for_austin": "austin",
	"token_for_lena":   "lena",
	"token_for_manu":   "manu",
}

// TokenAuthMiddleware validates Bearer Tokens
func TokenAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			log.Println("Authorization header missing")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			c.Abort()
			return
		}

		idToken := strings.TrimPrefix(authHeader, "Bearer ")

		client, err := app.Auth(ctx.Background())
		if err != nil {
			log.Fatalf("error getting Auth client: %v\n", err)
		}

		token, err := client.VerifyIDToken(ctx, idToken)
		if err != nil {
			log.Printf("error verifying ID token: %v\n", err)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}

		log.Printf("Verified ID token: %v\n", token)

		// Set the user in the context
		c.Set("user", token)
		c.Next()
	}
}
