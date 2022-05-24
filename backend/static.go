package backend

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func registerStaticRoutes(g *gin.Engine) {
	g.GET("/", func(ctx *gin.Context) {
		ctx.String(http.StatusOK, "Hello World")
	})
}
