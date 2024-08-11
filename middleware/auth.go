package middleware

import (
    "net/http"
    "strings"

    "github.com/gin-gonic/gin"
    "log"
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

        token := strings.TrimPrefix(authHeader, "Bearer ")
        user, exists := tokens[token]
        if !exists {
            log.Printf("Invalid token: %s\n", token)
            c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
            c.Abort()
            return
        }

        // Set the user in the context
        c.Set("user", user)
        c.Next()
    }
}