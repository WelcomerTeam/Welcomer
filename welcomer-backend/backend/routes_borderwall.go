package backend

import (
	"database/sql"
	"encoding/json"
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
	"github.com/ua-parser/uap-go/uaparser"
)

var userAgentParser, _ = uaparser.NewFromBytes(uaparser.DefinitionYaml)

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
	GuildName string `json:"guild_name,omitempty"`
	Valid     bool   `json:"valid"`
}

// Route GET /api/borderwall/:key.
func getBorderwall(ctx *gin.Context) {
	requireOAuthAuthorization(ctx, func(ctx *gin.Context) {
		key := ctx.Param("key")

		if key == "" {
			ctx.JSON(http.StatusBadRequest, NewBaseResponse(ErrBorderwallInvalidKey, nil))

			return
		}

		borderwallRequest, err := welcomer.Queries.GetBorderwallRequest(ctx, uuid.FromStringOrNil(key))
		if err != nil {
			welcomer.Logger.Warn().Err(err).Msg("Failed to get borderwall request")
		}

		borderwallResponse := BorderwallResponse{
			Valid: !borderwallRequest.RequestUuid.IsNil() && !borderwallRequest.IsVerified,
		}

		if !borderwallRequest.RequestUuid.IsNil() {
			session := sessions.Default(ctx)
			user, _ := GetUserSession(session)

			if borderwallRequest.UserID != int64(user.ID) {
				welcomer.Logger.Error().
					Int64("userID", int64(user.ID)).
					Int64("borderwallRequestUserID", borderwallRequest.UserID).
					Msg("User ID does not match")

				ctx.JSON(http.StatusForbidden, NewBaseResponse(ErrBorderwallUserInvalid, borderwallResponse))

				return
			}

			guild, err := welcomer.GRPCInterface.FetchGuildByID(backend.GetBasicEventContext(ctx).ToGRPCContext(), discord.Snowflake(borderwallRequest.GuildID))
			if err != nil {
				welcomer.Logger.Warn().Err(err).Int64("guildID", borderwallRequest.GuildID).Msg("Failed to fetch guild")
			} else if !guild.ID.IsNil() {
				borderwallResponse.GuildName = guild.Name
			}
		}

		ctx.JSON(http.StatusOK, NewBaseResponse(nil, borderwallResponse))
	})
}

// Route POST /api/borderwall/:key.
func setBorderwall(ctx *gin.Context) {
	requireOAuthAuthorization(ctx, func(ctx *gin.Context) {
		key := ctx.Param("key")

		if key == "" {
			ctx.JSON(http.StatusBadRequest, NewBaseResponse(ErrBorderwallInvalidKey, nil))

			return
		}

		userAgent := ctx.GetHeader("User-Agent")
		if userAgent == "" {
			ctx.JSON(http.StatusBadRequest, NewBaseResponse(NewMissingParameterError("user-agent"), nil))

			return
		}

		// Read "request" text from the post json body.
		var request BorderwallRequest

		if err := ctx.ShouldBindJSON(&request); err != nil {
			welcomer.Logger.Warn().Err(err).Msg("Failed to bind JSON")

			ctx.JSON(http.StatusBadRequest, NewBaseResponse(ErrInvalidJSON, nil))

			return
		}

		if request.Response == "" {
			welcomer.Logger.Warn().Msg("Missing response")

			ctx.JSON(http.StatusBadRequest, NewBaseResponse(NewMissingParameterError("response"), nil))

			return
		}

		borderwallRequest, err := welcomer.Queries.GetBorderwallRequest(ctx, uuid.FromStringOrNil(key))
		if err != nil {
			welcomer.Logger.Warn().Err(err).Msg("Failed to get borderwall request")
		}

		if borderwallRequest.RequestUuid.IsNil() {
			ctx.JSON(http.StatusBadRequest, NewBaseResponse(ErrBorderwallInvalidKey, nil))

			return
		}

		session := sessions.Default(ctx)

		user, ok := GetUserSession(session)
		if !ok {
			welcomer.Logger.Warn().Msg("Failed to get user session")

			ctx.JSON(http.StatusUnauthorized, NewBaseResponse(ErrMissingUser, nil))

			return
		}

		if borderwallRequest.UserID != int64(user.ID) {
			welcomer.Logger.Error().
				Int64("userID", int64(user.ID)).
				Int64("borderwallRequestUserID", borderwallRequest.UserID).
				Msg("User ID does not match")

			ctx.JSON(http.StatusForbidden, NewBaseResponse(ErrBorderwallUserInvalid, nil))

			return
		}

		if borderwallRequest.IsVerified {
			ctx.JSON(http.StatusBadRequest, NewBaseResponse(ErrBorderwallRequestAlreadyVerified, nil))

			return
		}

		// Validate reCAPTCHA
		recaptchaScore, err := welcomer.ValidateRecaptcha(request.Response, ctx.ClientIP())
		if err != nil {
			welcomer.Logger.Error().Err(err).Msg("Failed to validate recaptcha")

			ctx.JSON(http.StatusBadRequest, NewBaseResponse(ErrRecaptchaValidationFailed, nil))

			return
		}

		if recaptchaScore < RecaptchaThreshold {
			welcomer.Logger.Warn().Float64("score", recaptchaScore).Float64("threshold", RecaptchaThreshold).Msg("Recaptcha score is too low")

			ctx.JSON(http.StatusBadRequest, NewBaseResponse(ErrInsecureUser, nil))

			return
		}

		// Validate IPIntel
		ipIntelResponse, err := backend.IPChecker.CheckIP(ctx, ctx.ClientIP(), welcomer.IPIntelFlagDynamicBanListDynamicChecks, welcomer.IPIntelOFlagShowCountry)
		if err != nil {
			welcomer.Logger.Warn().Err(err).Msg("Failed to validate IPIntel")
		}

		if ipIntelResponse.Result > IPIntelThreshold {
			welcomer.Logger.Warn().Float64("score", ipIntelResponse.Result).Float64("threshold", IPIntelThreshold).Msg("IPIntel score is too high")

			ctx.JSON(http.StatusBadRequest, NewBaseResponse(ErrInsecureUser, nil))

			return
		}

		// Broadcast borderwall completion.
		managers, err := fetchManagersForGuild(ctx, discord.Snowflake(borderwallRequest.GuildID))
		if err != nil || len(managers) == 0 {
			welcomer.Logger.Error().Err(err).Int64("guildID", borderwallRequest.GuildID).Int("len", len(managers)).Msg("Failed to get managers for guild")

			ctx.JSON(http.StatusInternalServerError, NewBaseResponse(NewGenericErrorWithLineNumber(), nil))

			return
		}

		data, err := json.Marshal(welcomer.CustomEventInvokeWelcomerStructure{
			Member: discord.GuildMember{
				User: &discord.User{
					ID:            user.ID,
					Username:      user.Username,
					Discriminator: user.Discriminator,
					GlobalName:    user.GlobalName,
					Avatar:        user.Avatar,
				},
				GuildID: welcomer.ToPointer(discord.Snowflake(borderwallRequest.GuildID)),
			},
		})
		if err != nil {
			welcomer.Logger.Warn().Err(err).Msg("Failed to marshal borderwall completion data")

			ctx.JSON(http.StatusInternalServerError, NewBaseResponse(NewGenericErrorWithLineNumber(), nil))

			return
		}

		_, err = welcomer.SandwichClient.RelayMessage(ctx, &sandwich.RelayMessageRequest{
			Manager: managers[0],
			Type:    welcomer.CustomEventInvokeBorderwallCompletion,
			Data:    data,
		})
		if err != nil {
			welcomer.Logger.Warn().Err(err).Msg("Failed to relay borderwall completion")

			ctx.JSON(http.StatusInternalServerError, NewBaseResponse(NewGenericErrorWithLineNumber(), nil))

			return
		}

		clientIP := net.ParseIP(ctx.ClientIP())

		client := userAgentParser.Parse(userAgent)

		osName := client.Os.Family
		osVersion := client.Os.ToVersionString()

		// If platform version is 13 or higher, we assume it's Windows 11.
		// https://learn.microsoft.com/en-us/microsoft-edge/web-platform/how-to-detect-win11
		if strings.ToLower(osName) == "windows" && osVersion == "10" && getMajor(request.PlatformVersion) >= 13 {
			osVersion = "11"
		}

		welcomer.Logger.Info().
			Str("key", key).
			Float64("recaptchaScore", recaptchaScore).
			Float64("ipIntelResponse", ipIntelResponse.Result).
			Msg("Borderwall request verified")

		// Update the borderwall request with the response
		if _, err := welcomer.Queries.UpdateBorderwallRequest(ctx, database.UpdateBorderwallRequestParams{
			RequestUuid:     borderwallRequest.RequestUuid,
			IsVerified:      true,
			VerifiedAt:      sql.NullTime{Time: time.Now(), Valid: true},
			IpAddress:       pgtype.Inet{IPNet: &net.IPNet{IP: clientIP, Mask: clientIP.DefaultMask()}, Status: pgtype.Present},
			RecaptchaScore:  sql.NullFloat64{Float64: recaptchaScore, Valid: true},
			IpintelScore:    sql.NullFloat64{Float64: ipIntelResponse.Result, Valid: true},
			UaFamily:        sql.NullString{String: client.UserAgent.Family, Valid: true},
			UaFamilyVersion: sql.NullString{String: client.UserAgent.ToVersionString(), Valid: true},
			UaOs:            sql.NullString{String: osName, Valid: true},
			UaOsVersion:     sql.NullString{String: osVersion, Valid: true},
		}); err != nil {
			welcomer.Logger.Warn().Err(err).Msg("Failed to update borderwall request")

			ctx.JSON(http.StatusInternalServerError, NewBaseResponse(NewGenericErrorWithLineNumber(), nil))

			return
		}

		ctx.JSON(http.StatusOK, NewBaseResponse(nil, nil))
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
