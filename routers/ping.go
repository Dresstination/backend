package routers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func PingRoute(r *gin.Engine) {
	r.GET("/ping", func(c *gin.Context) {
		c.String(http.StatusOK, "pong")
	})
}
