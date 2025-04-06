package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/WelcomerTeam/Discord/discord"
	"github.com/WelcomerTeam/Welcomer/welcomer-core"
	core "github.com/WelcomerTeam/Welcomer/welcomer-core"
	"github.com/WelcomerTeam/Welcomer/welcomer-core/database"
	_ "github.com/joho/godotenv/autoload"
	"golang.org/x/oauth2"
)

func main() {
	var err error

	loggingLevel := flag.String("level", os.Getenv("LOGGING_LEVEL"), "Logging level")
	postgresURL := flag.String("postgresURL", os.Getenv("POSTGRES_URL"), "Postgres connection URL")

	patreonClientID := flag.String("patreonClientID", os.Getenv("PATREON_CLIENT_ID"), "Patreon client ID")
	patreonClientSecret := flag.String("patreonClientSecret", os.Getenv("PATREON_CLIENT_SECRET"), "Patreon client secret")

	patreonAccessToken := flag.String("patreonAccessToken", os.Getenv("PATREON_ACCESS_TOKEN"), "Patreon access token")
	patreonRefreshToken := flag.String("patreonRefreshToken", os.Getenv("PATREON_REFRESH_TOKEN"), "Patreon refresh token")

	patreonWebhookUrl := flag.String("patreonWebhookUrl", os.Getenv("PATREON_WEBHOOK_URL"), "Webhook URL for logging")

	flag.Parse()

	ctx, cancel := context.WithCancel(context.Background())

	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Recovered from panic:", r)

			err = welcomer.SendWebhookMessage(ctx, *patreonWebhookUrl, discord.WebhookMessageParams{
				Content: "<@143090142360371200>",
				Embeds: []discord.Embed{
					{
						Title:       "Patreon Service",
						Description: fmt.Sprintf("Recovered from panic: %v", r),
						Color:       int32(16760839),
						Timestamp:   welcomer.ToPointer(time.Now()),
					},
				},
			})
			if err != nil {
				welcomer.Logger.Warn().Err(err).Msg("Failed to send webhook message")
			}
		}
	}()

	welcomer.SetupLogger(*loggingLevel)
	welcomer.SetupDatabase(ctx, *postgresURL)

	entrypoint(ctx, *patreonClientID, *patreonClientSecret, *patreonAccessToken, *patreonRefreshToken, *patreonWebhookUrl)

	cancel()
}

func entrypoint(ctx context.Context, patreonClientID, patreonClientSecret, patreonAccessToken, patreonRefreshToken, patreonWebhookUrl string) {
	config := oauth2.Config{
		ClientID:     patreonClientID,
		ClientSecret: patreonClientSecret,
		Scopes:       []string{"campaigns.members"},
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://www.patreon.com/oauth2/authorize",
			TokenURL: "https://www.patreon.com/oauth2/token",
		},
	}

	tc := config.Client(ctx, &oauth2.Token{
		AccessToken:  patreonAccessToken,
		RefreshToken: patreonRefreshToken,
		ExpiresIn:    3600,
	})
	tc.Transport = welcomer.NewUserAgentSetterTransport(tc.Transport, welcomer.UserAgent)

	membersList, err := core.GetAllPatreonMembers(ctx, tc)
	if err != nil {
		panic(fmt.Sprintf("GetAllPatreonMembers(): %v", err))
	}

	membersMap := make(map[int64]core.PatreonMember, len(membersList))
	for _, member := range membersList {
		membersMap[int64(member.PatreonUserID)] = member
	}

	patreonUsersList, err := welcomer.Queries.GetPatreonUsers(ctx)
	if err != nil {
		panic(fmt.Sprint("GetPatreonUsers(): %w", err))
	}

	patreonUsersMap := make(map[int64]database.PatreonUsers, len(patreonUsersList))
	for _, patreonUser := range patreonUsersList {
		patreonUsersMap[patreonUser.PatreonUserID] = *patreonUser
	}

	processPatreonUsersNewlyLinked := []discord.Snowflake{}
	processPatreonUsersTiersChanged := []discord.Snowflake{}
	processPatreonUsersNoLongerPledging := []discord.Snowflake{}
	processPatreonUsersActive := []discord.Snowflake{}
	processPatreonUsersDeclined := []discord.Snowflake{}
	processPatreonUsersMissing := []discord.Snowflake{}
	processHasWarning := false

	// Try auto-link patreon users if they have discord linked and are not in the database.
	for _, patreonMember := range membersList {
		_, ok := patreonUsersMap[int64(patreonMember.PatreonUserID)]
		if !ok && !patreonMember.Attributes.SocialConnections.Discord.UserID.IsNil() {
			patreonUser := database.PatreonUsers{
				PatreonUserID:    int64(patreonMember.PatreonUserID),
				CreatedAt:        time.Now(),
				UpdatedAt:        time.Now(),
				UserID:           int64(patreonMember.Attributes.SocialConnections.Discord.UserID),
				FullName:         patreonMember.Attributes.FullName,
				Email:            patreonMember.Attributes.Email,
				ThumbUrl:         patreonMember.Attributes.ThumbUrl,
				PledgeCreatedAt:  time.Time{},
				PledgeEndedAt:    time.Time{},
				TierID:           0,
				LastChargeStatus: "",
				PatronStatus:     "",
			}

			_, err = welcomer.Queries.CreatePatreonUser(ctx, database.CreatePatreonUserParams{
				PatreonUserID:    patreonUser.PatreonUserID,
				UserID:           patreonUser.UserID,
				FullName:         patreonUser.FullName,
				Email:            patreonUser.Email,
				ThumbUrl:         patreonUser.ThumbUrl,
				PledgeCreatedAt:  patreonUser.PledgeCreatedAt,
				PledgeEndedAt:    patreonUser.PledgeEndedAt,
				TierID:           patreonUser.TierID,
				LastChargeStatus: patreonUser.LastChargeStatus,
				PatronStatus:     patreonUser.PatronStatus,
			})
			if err != nil {
				welcomer.Logger.Warn().Err(err).Msg("Failed to create patreon user")

				processHasWarning = true
			} else {
				processPatreonUsersNewlyLinked = append(processPatreonUsersNewlyLinked, discord.Snowflake(patreonUser.PatreonUserID))

				err = core.OnPatreonLinked(ctx, welcomer.PatreonUser{
					ID:       discord.Snowflake(patreonUser.PatreonUserID),
					Email:    patreonUser.Email,
					FullName: patreonUser.FullName,
					SocialConnections: core.PatreonUser_SocialConnections{
						Discord: core.PatreonUser_Discord{
							UserID: discord.Snowflake(patreonUser.UserID),
						},
					},
					ThumbURL: patreonUser.ThumbUrl,
				}, true)
				if err != nil {
					welcomer.Logger.Warn().Err(err).Msg("Failed to trigger patreon linked")

					processHasWarning = true
				}
			}

			patreonUsersList = append(patreonUsersList, &patreonUser)
		}
	}

	// Update patreon users if they are in the database.
	for _, databasePatreonUser := range patreonUsersList {
		m, ok := membersMap[databasePatreonUser.PatreonUserID]

		var tierID core.PatreonTier
		for _, tier := range m.EntitledTiers {
			tierID = tier
		}

		if ok {
			// Patreon user exists

			switch tierID {
			case core.PatreonTierFree,
				core.PatreonTierUnpublishedWelcomerDonator,
				core.PatreonTierUnpublishedWelcomerPro1,
				core.PatreonTierUnpublishedWelcomerPro3,
				core.PatreonTierUnpublishedWelcomerPro5,
				core.PatreonTierWelcomerPro,
				0:
			default:
				welcomer.Logger.Warn().
					Int64("patreon_user_id", databasePatreonUser.PatreonUserID).
					Int64("tier_id", int64(tierID)).
					Str("charge_status", string(m.Attributes.LastChargeStatus)).
					Msgf("Unhandled tier")

				processHasWarning = true

				continue
			}

			if core.PatreonTier(databasePatreonUser.TierID) != tierID {
				// Tier ID has changed

				if tierID != core.PatreonTier(databasePatreonUser.TierID) {
					welcomer.Logger.Info().
						Int64("patreon_user_id", databasePatreonUser.PatreonUserID).
						Int64("old_tier_id", int64(databasePatreonUser.TierID)).
						Int64("new_tier_id", int64(tierID)).
						Msgf("Patreon user's tier has changed")

					newPatreonUser := database.CreateOrUpdatePatreonUserParams{
						PatreonUserID:    databasePatreonUser.PatreonUserID,
						UserID:           databasePatreonUser.UserID,
						FullName:         databasePatreonUser.FullName,
						Email:            databasePatreonUser.Email,
						ThumbUrl:         databasePatreonUser.ThumbUrl,
						PledgeCreatedAt:  databasePatreonUser.PledgeCreatedAt,
						PledgeEndedAt:    databasePatreonUser.PledgeEndedAt,
						TierID:           int64(tierID),
						LastChargeStatus: string(m.Attributes.LastChargeStatus),
						PatronStatus:     string(m.Attributes.PatronStatus),
					}

					processPatreonUsersTiersChanged = append(processPatreonUsersTiersChanged, discord.Snowflake(databasePatreonUser.PatreonUserID))

					err = core.OnPatreonTierChanged(ctx, databasePatreonUser, newPatreonUser)
					if err != nil {
						welcomer.Logger.Warn().Err(err).Msg("Failed to trigger patreon tier changed")

						processHasWarning = true

						err = core.OnPatreonTierChanged_Fallback(ctx, databasePatreonUser, newPatreonUser, err)
						if err != nil {
							welcomer.Logger.Warn().Err(err).Msg("Failed to trigger patreon tier changed fallback")
						}
					}
				}
			} else {
				// Tier ID has not changed, check if pledge status has changed.

				if core.PatronStatus(databasePatreonUser.PatronStatus) != m.Attributes.PatronStatus {
					welcomer.Logger.Info().
						Int64("patreon_user_id", databasePatreonUser.PatreonUserID).
						Str("old_patron_status", string(databasePatreonUser.PatronStatus)).
						Str("new_patron_status", string(m.Attributes.PatronStatus)).
						Msgf("Patreon user's patron status has changed")

					switch m.Attributes.PatronStatus {
					case core.PatreonStatusActive:

						processPatreonUsersActive = append(processPatreonUsersActive, discord.Snowflake(databasePatreonUser.PatreonUserID))

						err = core.OnPatreonActive(ctx, *databasePatreonUser, m)
						if err != nil {
							welcomer.Logger.Warn().Err(err).Msg("Failed to trigger patreon active")

							processHasWarning = true
						}
					case core.PatreonStatusFormer:

						processPatreonUsersNoLongerPledging = append(processPatreonUsersNoLongerPledging, discord.Snowflake(databasePatreonUser.PatreonUserID))

						err = core.OnPatreonNoLongerPledging(ctx, *databasePatreonUser, m)
						if err != nil {
							welcomer.Logger.Warn().Err(err).Msg("Failed to trigger patreon no longer pledging")

							processHasWarning = true
						}
					case core.PatreonStatusDeclined:
						// Update database user.

						processPatreonUsersDeclined = append(processPatreonUsersDeclined, discord.Snowflake(databasePatreonUser.PatreonUserID))

						_, err = welcomer.Queries.CreateOrUpdatePatreonUser(ctx, database.CreateOrUpdatePatreonUserParams{
							PatreonUserID:    databasePatreonUser.PatreonUserID,
							UserID:           databasePatreonUser.UserID,
							PledgeCreatedAt:  databasePatreonUser.PledgeCreatedAt,
							PledgeEndedAt:    databasePatreonUser.PledgeEndedAt,
							TierID:           databasePatreonUser.TierID,
							FullName:         welcomer.Coalesce(m.Attributes.FullName, databasePatreonUser.FullName),
							Email:            welcomer.Coalesce(m.Attributes.Email, databasePatreonUser.Email),
							ThumbUrl:         welcomer.Coalesce(m.Attributes.ThumbUrl, databasePatreonUser.ThumbUrl),
							LastChargeStatus: string(m.Attributes.LastChargeStatus),
							PatronStatus:     string(m.Attributes.PatronStatus),
						})
						if err != nil {
							welcomer.Logger.Error().Err(err).
								Int64("user_id", int64(databasePatreonUser.UserID)).
								Msg("Failed to create or update patreon user")
						}

						welcomer.Logger.Info().
							Int64("patreon_user_id", databasePatreonUser.PatreonUserID).
							Int64("tier_id", int64(tierID)).
							Str("name", databasePatreonUser.FullName).
							Str("email", databasePatreonUser.Email).
							Str("patron_status", string(m.Attributes.PatronStatus)).
							Str("last_charge_status", string(m.Attributes.LastChargeStatus)).
							Msgf("Patreon user's pledge has been declined")
					default:
						welcomer.Logger.Warn().
							Int64("patreon_user_id", int64(m.PatreonUserID)).
							Int64("tier_id", int64(tierID)).
							Str("patron_status", string(m.Attributes.PatronStatus)).
							Msgf("Unhandled patron status")

						processHasWarning = true
					}
				}
			}

		} else {
			// Patreon user no longer exists

			welcomer.Logger.Info().
				Int64("patreon_user_id", databasePatreonUser.PatreonUserID).
				Msgf("Patreon user no longer exists")

			processPatreonUsersMissing = append(processPatreonUsersMissing, discord.Snowflake(databasePatreonUser.PatreonUserID))

			err = core.OnPatreonNoLongerPledging(ctx, *databasePatreonUser, m)
			if err != nil {
				welcomer.Logger.Warn().Err(err).Msg("Failed to trigger patreon no longer pledging")

				processHasWarning = true
			}

		}
	}

	embedFields := make([]discord.EmbedField, 0)

	if len(processPatreonUsersNewlyLinked) > 0 {
		embedFields = append(embedFields, discord.EmbedField{
			Name:  "Newly Linked",
			Value: fmt.Sprintf("%d - %d", len(processPatreonUsersNewlyLinked), processPatreonUsersNewlyLinked),
		})
	}

	if len(processPatreonUsersTiersChanged) > 0 {
		embedFields = append(embedFields, discord.EmbedField{
			Name:  "Tiers Changed",
			Value: fmt.Sprintf("%d - %d", len(processPatreonUsersTiersChanged), processPatreonUsersTiersChanged),
		})
	}

	if len(processPatreonUsersNoLongerPledging) > 0 {
		embedFields = append(embedFields, discord.EmbedField{
			Name:  "No Longer Pledging",
			Value: fmt.Sprintf("%d - %d", len(processPatreonUsersNoLongerPledging), processPatreonUsersNoLongerPledging),
		})
	}

	if len(processPatreonUsersActive) > 0 {
		embedFields = append(embedFields, discord.EmbedField{
			Name:  "Active",
			Value: fmt.Sprintf("%d - %d", len(processPatreonUsersActive), processPatreonUsersActive),
		})
	}

	if len(processPatreonUsersDeclined) > 0 {
		embedFields = append(embedFields, discord.EmbedField{
			Name:  "Declined",
			Value: fmt.Sprintf("%d - %d", len(processPatreonUsersDeclined), processPatreonUsersDeclined),
		})
	}

	// Only include missing when there is other fields.
	if len(processPatreonUsersMissing) > 0 && len(embedFields) > 0 {
		embedFields = append(embedFields, discord.EmbedField{
			Name:  "Missing",
			Value: fmt.Sprintf("%d - %d", len(processPatreonUsersMissing), processPatreonUsersMissing),
		})
	}

	if len(embedFields) == 0 && processHasWarning {
		embedFields = append(embedFields, discord.EmbedField{
			Name:  "Warning",
			Value: "There was a warning but no users were processed",
		})
	}

	if len(embedFields) > 0 {
		err = welcomer.SendWebhookMessage(ctx, patreonWebhookUrl, discord.WebhookMessageParams{
			Content: welcomer.If(processHasWarning, "<@143090142360371200>", ""),
			Embeds: []discord.Embed{
				{
					Title:     "Patreon Service",
					Fields:    embedFields,
					Color:     welcomer.If(processHasWarning, int32(16760839), int32(5415248)),
					Timestamp: welcomer.ToPointer(time.Now()),
				},
			},
		})
		if err != nil {
			welcomer.Logger.Warn().Err(err).Msg("Failed to send webhook message")
		}
	}
}
