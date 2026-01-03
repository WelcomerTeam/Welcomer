package service

import (
	"net/http"

	"github.com/WelcomerTeam/Welcomer/welcomer-core"
	"github.com/gin-gonic/gin"
)

// Route POST /generate
func (is *ImageService) generateHandler(ctx *gin.Context) {
	var req welcomer.CustomWelcomerImageGenerateRequest

	if err := ctx.BindJSON(&req); err != nil {
		ctx.Data(http.StatusBadRequest, "text/plain", []byte("invalid request payload"))

		return
	}

	igctx := &ImageGenerationContext{
		Context:                            ctx.Request.Context(),
		CustomWelcomerImageGenerateRequest: req,
	}

	builder := is.GenerateCanvas(igctx)
	html := builder.String()

	resp, elapsed, err := is.ScreenshotFromHTML(ctx, html)
	if err != nil {
		ctx.Data(http.StatusInternalServerError, "text/plain", []byte(err.Error()))

		return
	}

	ctx.Header("X-Generation-Time-Ms", welcomer.Itoa(elapsed.Milliseconds()))
	ctx.Data(http.StatusOK, "image/png", resp)
}

func (is *ImageService) registerRoutes(g *gin.Engine) {
	g.POST("/generate", is.generateHandler)
}
