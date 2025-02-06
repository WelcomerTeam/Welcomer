package backend

import (
	"github.com/gin-contrib/cache"
	"github.com/gin-contrib/cache/persistence"
	"github.com/gin-gonic/gin"
	"net/http"
	"time"
)

func registerExampleRoutes(g *gin.Engine) {
	store := persistence.NewInMemoryStore(time.Second)

	g.GET("/cache", cache.CachePage(store, time.Minute, func(c *gin.Context) {
		c.String(http.StatusOK, "Hello")
	}))
}
