package service

import (
	"net/http"
	"os"
	"time"

	"github.com/WelcomerTeam/Discord/discord"
	"github.com/WelcomerTeam/Welcomer/welcomer-core"
	"github.com/gin-gonic/gin"
)

// Route POST /generate
func (is *ImageService) generateHandler(context *gin.Context) {
	ctx := context.Request.Context()

	onRequest()

	var requestBody welcomer.GenerateImageOptionsRaw
	if err := context.ShouldBindJSON(&requestBody); err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})

		return
	}

	start := time.Now()

	file, format, timing, err := is.GenerateImage(ctx, generateImageRequestToOptions(requestBody))
	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})

		return
	}

	onGenerationComplete(start, requestBody.GuildID, requestBody.Background, format)

	if is.Options.Debug {
		_ = os.WriteFile("output.png", file, 0o644)
	}

	context.Header("Server-Timing", timing.String())
	context.Data(http.StatusOK, format.String(), file)
}

func (is *ImageService) registerRoutes(g *gin.Engine) {
	g.POST("/generate", is.generateHandler)
}

func generateImageRequestToOptions(req welcomer.GenerateImageOptionsRaw) GenerateImageOptions {
	return GenerateImageOptions{
		ShowAvatar:         req.ShowAvatar,
		GuildID:            discord.Snowflake(req.GuildID),
		UserID:             discord.Snowflake(req.UserID),
		AllowAnimated:      req.AllowAnimated,
		AvatarURL:          req.AvatarURL,
		Theme:              welcomer.ImageTheme(req.Theme),
		Background:         req.Background,
		Text:               req.Text,
		TextFont:           req.TextFont,
		TextStroke:         welcomer.FormatTextStroke(req.TextStroke),
		TextAlign:          welcomer.ImageAlignment(req.TextAlign),
		TextColor:          welcomer.ConvertToRGBA(req.TextColor),
		TextStrokeColor:    welcomer.ConvertToRGBA(req.TextStrokeColor),
		ImageBorderColor:   welcomer.ConvertToRGBA(req.ImageBorderColor),
		ImageBorderWidth:   int(req.ImageBorderWidth),
		ProfileFloat:       welcomer.ImageAlignment(req.ProfileFloat),
		ProfileBorderColor: welcomer.ConvertToRGBA(req.ProfileBorderColor),
		ProfileBorderWidth: int(req.ProfileBorderWidth),
		ProfileBorderCurve: welcomer.ImageProfileBorderType(req.ProfileBorderCurve),
	}
}
