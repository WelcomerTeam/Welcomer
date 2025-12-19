package service

import "github.com/gin-gonic/gin"

// Route POST /generate
func (is *ImageService) generateHandler(context *gin.Context) {
}

func (is *ImageService) registerRoutes(g *gin.Engine) {
	g.POST("/generate", is.generateHandler)
}
