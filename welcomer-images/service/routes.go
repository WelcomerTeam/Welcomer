package service

import (
	"net/http"
	"os"
	"time"

	"github.com/WelcomerTeam/Discord/discord"
	utils "github.com/WelcomerTeam/Welcomer/welcomer-utils"
	"github.com/gin-gonic/gin"
)

// Route POST /generate
func (is *ImageService) generateHandler(c *gin.Context) {
	onRequest()

	var requestBody utils.GenerateImageOptionsRaw
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

func generateImageRequestToOptions(req utils.GenerateImageOptionsRaw) GenerateImageOptions {
	return GenerateImageOptions{
		GuildID:            discord.Snowflake(req.GuildID),
		UserID:             discord.Snowflake(req.UserID),
		AllowAnimated:      req.AllowAnimated,
		AvatarURL:          req.AvatarURL,
		Theme:              utils.ImageTheme(req.Theme),
		Background:         req.Background,
		Text:               req.Text,
		TextFont:           req.TextFont,
		TextStroke:         utils.FormatTextStroke(req.TextStroke),
		TextAlign:          utils.ImageAlignment(req.TextAlign),
		TextColor:          utils.ConvertToRGBA(req.TextColor),
		TextStrokeColor:    utils.ConvertToRGBA(req.TextStrokeColor),
		ImageBorderColor:   utils.ConvertToRGBA(req.ImageBorderColor),
		ImageBorderWidth:   int(req.ImageBorderWidth),
		ProfileFloat:       utils.ImageAlignment(req.ProfileFloat),
		ProfileBorderColor: utils.ConvertToRGBA(req.ProfileBorderColor),
		ProfileBorderWidth: int(req.ProfileBorderWidth),
		ProfileBorderCurve: utils.ImageProfileBorderType(req.ProfileBorderCurve),
	}
}
