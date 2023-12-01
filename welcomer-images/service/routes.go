package service

import (
	"net/http"
	"os"
	"time"

	"github.com/WelcomerTeam/Discord/discord"
	core "github.com/WelcomerTeam/Welcomer/welcomer-core"
	"github.com/gin-gonic/gin"
)

// Rouye POST /generate
func (is *ImageService) generateHandler(c *gin.Context) {
	onRequest()

	var requestBody GenerateImageOptionsRaw
	if err := c.ShouldBindJSON(&requestBody); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	start := time.Now()

	file, format, timing, err := is.GenerateImage(generateImageRequestToOptions(requestBody))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	onGenerationComplete(start, requestBody.GuildID, requestBody.Background, format)

	if is.Options.Debug {
		os.WriteFile("output.png", file, 0o644)
	}

	c.Header("Server-Timing", timing.String())
	c.Data(http.StatusOK, format.String(), file)
}

func (is *ImageService) registerRoutes(g *gin.Engine) {
	g.POST("/generate", is.generateHandler)
}

func generateImageRequestToOptions(req GenerateImageOptionsRaw) GenerateImageOptions {
	return GenerateImageOptions{
		GuildID:            discord.Snowflake(req.GuildID),
		UserID:             discord.Snowflake(req.UserID),
		AllowAnimated:      req.AllowAnimated,
		AvatarURL:          req.AvatarURL,
		Theme:              core.ImageTheme(req.Theme),
		Background:         req.Background,
		Text:               req.Text,
		TextFont:           req.TextFont,
		TextStroke:         core.FormatTextStroke(req.TextStroke),
		TextAlign:          core.ImageAlignment(req.TextAlign),
		TextColor:          core.ConvertToRGBA(req.TextColor),
		TextStrokeColor:    core.ConvertToRGBA(req.TextStrokeColor),
		ImageBorderColor:   core.ConvertToRGBA(req.ImageBorderColor),
		ImageBorderWidth:   int(req.ImageBorderWidth),
		ProfileFloat:       core.ImageAlignment(req.ProfileFloat),
		ProfileBorderColor: core.ConvertToRGBA(req.ProfileBorderColor),
		ProfileBorderWidth: int(req.ProfileBorderWidth),
		ProfileBorderCurve: core.ImageProfileBorderType(req.ProfileBorderCurve),
	}
}
