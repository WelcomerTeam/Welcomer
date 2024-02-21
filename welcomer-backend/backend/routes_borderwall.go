package backend

import (
	"database/sql"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	discord "github.com/WelcomerTeam/Discord/discord"
	sandwich "github.com/WelcomerTeam/Sandwich-Daemon/protobuf"
	"github.com/WelcomerTeam/Welcomer/welcomer-core"
	"github.com/WelcomerTeam/Welcomer/welcomer-core/database"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/gofrs/uuid"
	"github.com/jackc/pgtype"
	jsoniter "github.com/json-iterator/go"
)

const (
	// Recaptcha must return a value higher than this threshold to be considered valid.
	// Anything above the value of 0.5 is considered as "low risk".
	RecaptchaThreshold = 0.5

	// IPIntel must return a value below this threshold to be considered valid.
	// Anything below the value of 0.90 is considered as "low risk".
	IPIntelThreshold = 0.9
)

type BorderwallRequest struct {
	Response        string `json:"response"`
	PlatformVersion string `json:"platform_version"`
}

type BorderwallResponse struct {
	Valid     bool   `json:"valid"`
	GuildName string `json:"guild_name"`
}

// Route GET /api/borderwall/:key
func getBorderwall(ctx *gin.Context) {
	requireOAuthAuthorization(ctx, func(ctx *gin.Context) {
		key := ctx.Param("key")

		if key == "" {
			ctx.JSON(http.StatusBadRequest, BaseResponse{
				Ok: false,
			})

			return
		}

		borderwallRequest, err := backend.Database.GetBorderwallRequest(ctx, uuid.FromStringOrNil(key))
		if err != nil {
			backend.Logger.Warn().Err(err).Msg("Failed to get borderwall request")
		}

		borderwallResponse := BorderwallResponse{
			Valid: borderwallRequest != nil && !borderwallRequest.RequestUuid.IsNil() && !borderwallRequest.IsVerified,
		}

		if borderwallRequest != nil && !borderwallRequest.RequestUuid.IsNil() {
			guild, err := backend.GRPCInterface.FetchGuildByID(backend.GetBasicEventContext().ToGRPCContext(), discord.Snowflake(borderwallRequest.GuildID))
			if err != nil {
				backend.Logger.Warn().Err(err).Int64("guildID", int64(borderwallRequest.GuildID)).Msg("Failed to fetch guild")
			} else if guild != nil {
				borderwallResponse.GuildName = guild.Name
			}
		}

		ctx.JSON(http.StatusOK, BaseResponse{
			Ok:   true,
			Data: borderwallResponse,
		})
	})
}

// Route POST /api/borderwall/:key
func setBorderwall(ctx *gin.Context) {
	requireOAuthAuthorization(ctx, func(ctx *gin.Context) {
		key := ctx.Param("key")

		if key == "" {
			ctx.JSON(http.StatusBadRequest, BaseResponse{
				Ok: false,
			})

			return
		}

		logger := backend.Logger.With().Str("key", key).Logger()

		userAgent := ctx.GetHeader("User-Agent")
		if userAgent == "" {
			ctx.JSON(http.StatusBadRequest, BaseResponse{
				Ok: false,
			})

			return
		}

		// Read "response" text from the post json body
		var response BorderwallRequest

		if err := ctx.ShouldBindJSON(&response); err != nil {
			logger.Warn().Err(err).Msg("Failed to bind JSON")

			ctx.JSON(http.StatusBadRequest, BaseResponse{
				Ok: false,
			})

			return
		}

		if response.Response == "" {
			logger.Warn().Msg("Missing response")

			ctx.JSON(http.StatusBadRequest, BaseResponse{
				Ok: false,
			})

			return
		}

		borderwallRequest, err := backend.Database.GetBorderwallRequest(ctx, uuid.FromStringOrNil(key))
		if err != nil {
			logger.Warn().Err(err).Msg("Failed to get borderwall request")
		}

		if borderwallRequest == nil || borderwallRequest.RequestUuid.IsNil() {
			ctx.JSON(http.StatusBadRequest, BaseResponse{
				Ok:    false,
				Error: ErrBorderwallInvalidKey.Error(),
			})

			return
		}

		session := sessions.Default(ctx)

		user, ok := GetUserSession(session)
		if !ok {
			logger.Warn().Msg("Missing user in session")

			ctx.JSON(http.StatusUnauthorized, BaseResponse{
				Ok:    false,
				Error: ErrMissingUser.Error(),
			})

			return
		}

		if borderwallRequest.UserID != int64(user.ID) {
			logger.Warn().Int64("userID", int64(user.ID)).Int64("borderwallRequestUserID", borderwallRequest.UserID).Msg("User ID does not match")

			ctx.JSON(http.StatusForbidden, BaseResponse{
				Ok: false,
			})

			return
		}

		if borderwallRequest.IsVerified {
			ctx.JSON(http.StatusBadRequest, BaseResponse{
				Ok:    false,
				Error: ErrBorderwallRequestAlreadyVerified.Error(),
			})

			return
		}

		// Validate reCAPTCHA
		recaptchaScore, err := welcomer.ValidateRecaptcha(backend.Logger, response.Response, ctx.ClientIP())
		if err != nil {
			logger.Error().Err(err).Msg("Failed to validate recaptcha")

			ctx.JSON(http.StatusBadRequest, BaseResponse{
				Ok:    false,
				Error: ErrRecaptchaValidationFailed.Error(),
			})

			return
		}

		if recaptchaScore < RecaptchaThreshold {
			logger.Warn().Float64("score", recaptchaScore).Float64("threshold", RecaptchaThreshold).Msg("Recaptcha score is too low")

			ctx.JSON(http.StatusBadRequest, BaseResponse{
				Ok:    false,
				Error: ErrInsecureUser.Error(),
			})

			return
		}

		// Validate IPIntel
		ipIntelResponse, err := welcomer.CheckIPIntel(backend.Logger, ctx.ClientIP())
		if err != nil {
			logger.Warn().Err(err).Msg("Failed to validate IPIntel")
		}

		if ipIntelResponse > IPIntelThreshold {
			logger.Warn().Float64("score", ipIntelResponse).Float64("threshold", IPIntelThreshold).Msg("IPIntel score is too high")

			ctx.JSON(http.StatusBadRequest, BaseResponse{
				Ok:    false,
				Error: ErrInsecureUser.Error(),
			})

			return
		}

		// Broadcast borderwall completion.
		managers, err := fetchManagersForGuild(discord.Snowflake(borderwallRequest.GuildID))
		if err != nil || len(managers) == 0 {
			logger.Error().Err(err).Int64("guildID", int64(borderwallRequest.GuildID)).Int("len", len(managers)).Msg("Failed to get managers for guild")

			ctx.JSON(http.StatusInternalServerError, BaseResponse{
				Ok: false,
			})

			return
		}

		data, err := jsoniter.Marshal(welcomer.CustomEventInvokeBorderwallCompletionStructure{
			UserId: discord.Snowflake(borderwallRequest.UserID),
		})
		if err != nil {
			logger.Warn().Err(err).Msg("Failed to marshal borderwall completion data")

			ctx.JSON(http.StatusInternalServerError, BaseResponse{
				Ok: false,
			})

			return
		}

		_, err = backend.SandwichClient.RelayMessage(ctx, &sandwich.RelayMessageRequest{
			Manager: managers[0],
			Type:    welcomer.CustomEventInvokeBorderwallCompletion,
			Data:    data,
		})
		if err != nil {
			logger.Warn().Err(err).Msg("Failed to relay borderwall completion")

			ctx.JSON(http.StatusInternalServerError, BaseResponse{
				Ok: false,
			})

			return
		}

		ip := net.ParseIP(ctx.ClientIP())
		family, familyVersion, os, osVersion := welcomer.ParseUserAgent(userAgent)

		// If platform version is 13 or higher, we assume it's Windows 11.
		// https://learn.microsoft.com/en-us/microsoft-edge/web-platform/how-to-detect-win11
		if strings.ToLower(os) == "windows" && osVersion == "10" && getMajor(response.PlatformVersion) >= 13 {
			osVersion = "11"
		}

		logger.Info().
			Str("key", key).
			Float64("recaptchaScore", recaptchaScore).
			Float64("ipIntelResponse", ipIntelResponse).
			Msg("Borderwall request verified")

		// Update the borderwall request with the response
		if _, err := backend.Database.UpdateBorderwallRequest(ctx, &database.UpdateBorderwallRequestParams{
			RequestUuid:     borderwallRequest.RequestUuid,
			IsVerified:      true,
			VerifiedAt:      sql.NullTime{Time: time.Now(), Valid: true},
			IpAddress:       pgtype.Inet{IPNet: &net.IPNet{IP: ip, Mask: ip.DefaultMask()}, Status: pgtype.Present},
			RecaptchaScore:  sql.NullFloat64{Float64: recaptchaScore, Valid: true},
			IpintelScore:    sql.NullFloat64{Float64: ipIntelResponse, Valid: true},
			UaFamily:        sql.NullString{String: family, Valid: true},
			UaFamilyVersion: sql.NullString{String: familyVersion, Valid: true},
			UaOs:            sql.NullString{String: os, Valid: true},
			UaOsVersion:     sql.NullString{String: osVersion, Valid: true},
		}); err != nil {
			logger.Warn().Err(err).Msg("Failed to update borderwall request")

			ctx.JSON(http.StatusInternalServerError, BaseResponse{
				Ok: false,
			})

			return
		}

		ctx.JSON(http.StatusOK, BaseResponse{
			Ok: true,
		})
	})
}

func getMajor(input string) int {
	parts := strings.Split(input, ".")
	if len(parts) >= 1 {
		integerPart, _ := strconv.Atoi(parts[0])

		return integerPart
	}

	return 0
}

func registerBorderwallRoutes(g *gin.Engine) {
	g.GET("/api/borderwall/:key", getBorderwall)
	g.POST("/api/borderwall/:key", setBorderwall)
}
