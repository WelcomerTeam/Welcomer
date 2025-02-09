package backend

import (
	"net/http"
	"time"

	"github.com/gin-contrib/cache"
	"github.com/gin-contrib/cache/persistence"
	"github.com/gin-gonic/gin"
)

func registerExampleRoutes(g *gin.Engine) {
	store := persistence.NewInMemoryStore(time.Second)

	g.GET("/cache", cache.CachePage(store, time.Minute, func(c *gin.Context) {
		c.String(http.StatusOK, "Hello")
	}))
}
