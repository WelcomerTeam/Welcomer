package plugins

import (
	"bytes"
	"context"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/WelcomerTeam/Discord/discord"
	sandwich "github.com/WelcomerTeam/Sandwich-Daemon/proto"
	subway "github.com/WelcomerTeam/Subway/subway"
	"github.com/WelcomerTeam/Welcomer/welcomer-core"
	"github.com/WelcomerTeam/Welcomer/welcomer-core/database"
	"github.com/gofrs/uuid"
	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
)

const (
	giveawaySetupMenuTitleKey        = "title"
	giveawaySetupMenuDescriptionKey  = "description"
	giveawaySetupMenuAccentColourKey = "accent_colour"
	giveawaySetupMenuThumbnailURLKey = "thumbnail_url"

	giveawaySetupMenuPrizesKey = "prizes"

	giveawaySetupMenuDurationKey        = "duration"
	giveawaySetupMenuAnnounceWinnersKey = "announce_winners"

	giveawaySetupMenuRolesAllowedKey         = "roles_allowed"
	giveawaySetupMenuRolesAllowedIncludedKey = "roles_allowed_included"
	giveawaySetupMenuRolesAllowedExcludedKey = "roles_allowed_excluded"

	giveawaySetupMenuMinimumJoinDateKey = "minimum_join_date"
	giveawaySetupMenuStartKey           = "start"

	giveawaySetupMenuDisplayKey            = "display"
	giveawaySetupMenuDisplayShowPrizesKey  = "display_show_prizes"
	giveawaySetupMenuDisplayShowEntriesKey = "display_show_entries"

	giveawaySetupMenuPreviewOnKey  = "preview_on"
	giveawaySetupMenuPreviewOffKey = "preview_off"

	giveawaySetupMenuPingKey                    = "ping"
	giveawaySetupMenuPingEveryoneKey            = "ping_everyone"
	giveawaySetupMenuPingHereKey                = "ping_here"
	giveawaySetupMenuPingRolesAllowedToEnterKey = "ping_roles_allowed_to_enter"
	giveawaySetupMenuPingAdditionalRolesKey     = "additional_roles_to_ping"

	giveawayManageMenuToggleAllowEntriesKey = "toggle_allow_entries"
	giveawayManageMenuExtendDurationKey     = "extend_duration"
	giveawayManageMenuEndGiveawayKey        = "end_giveaway"
	giveawayManageMenuExportEntriesKey      = "export_entries"
	giveawayManageMenuExportWinnersKey      = "export_winners"
)

func NewGiveawaysCog() *GiveawaysCog {
	return &GiveawaysCog{
		InteractionCommands: subway.SetupInteractionCommandable(&subway.InteractionCommandable{}),
	}
}

type GiveawaysCog struct {
	InteractionCommands *subway.InteractionCommandable
}

// Assert types.

var (
	_ subway.Cog                        = (*GiveawaysCog)(nil)
	_ subway.CogWithInteractionCommands = (*GiveawaysCog)(nil)
)

func (cog *GiveawaysCog) CogInfo() *subway.CogInfo {
	return &subway.CogInfo{
		Name:        "Giveaways",
		Description: "Provides the functionality for the 'Giveaways' feature",
	}
}

func (cog *GiveawaysCog) GetInteractionCommandable() *subway.InteractionCommandable {
	return cog.InteractionCommands
}

func (cog *GiveawaysCog) RegisterCog(sub *subway.Subway) error {
	giveawaysGroup := subway.NewSubcommandGroup(
		"giveaways",
		"Giveaways commands",
	)

	giveawaysGroup.MustAddInteractionCommand(&subway.InteractionCommandable{
		Name:        "new",
		Description: "Makes a new giveaway",

		Type: subway.InteractionCommandableTypeSubcommand,

		DMPermission:            new(false),
		DefaultMemberPermission: new(discord.Int64(welcomer.PermissionElevated)),

		Handler: func(ctx context.Context, sub *subway.Subway, interaction discord.Interaction) (*discord.InteractionResponse, error) {
			return welcomer.RequireGuildElevation(sub, interaction, func() (*discord.InteractionResponse, error) {
				var giveaway *database.GuildGiveaways
				var err error

				err = welcomer.RetryWithFallback(
					func() error {
						giveaway, err = welcomer.Queries.CreateGiveaway(ctx, database.CreateGiveawayParams{
							GuildID:   int64(*interaction.GuildID),
							CreatedBy: int64(interaction.GetUser().ID),
							EndTime:   time.Time{},
						})

						return err
					},
					func() error {
						return welcomer.EnsureGuild(ctx, *interaction.GuildID)
					},
					nil,
				)
				if err != nil {
					welcomer.Logger.Error().Err(err).
						Int64("guild_id", int64(*interaction.GuildID)).
						Msg("Failed to create giveaway settings")

					return nil, err
				}

				welcomer.PusherGuildScience.Push(
					ctx,
					*interaction.GuildID,
					interaction.GetUser().ID,
					database.ScienceGuildEventTypeGiveawayCreated,
					&welcomer.GuildScienceGiveawayEvents{
						GiveawayUUID: giveaway.GiveawayUuid,
					},
				)

				return &discord.InteractionResponse{
					Type: discord.InteractionCallbackTypeModal,
					Data: &discord.InteractionCallbackData{
						Title:    "Create Giveaway",
						CustomID: "giveaway_edit:" + giveaway.GiveawayUuid.String(),
						Components: []discord.InteractionComponent{
							{
								Type:  discord.InteractionComponentTypeLabel,
								Label: "Title",
								Component: &discord.InteractionComponent{
									CustomID: giveawaySetupMenuTitleKey,
									Type:     discord.InteractionComponentTypeTextInput,
									Value:    giveaway.Title,
									Style:    discord.InteractionComponentStyleShort,
									Required: new(false),
								},
							},
							{
								Type:        discord.InteractionComponentTypeLabel,
								Label:       "Prizes",
								Description: "One prize per line, with optional count, e.g. 2x Discord Nitro",
								Component: &discord.InteractionComponent{
									CustomID:    giveawaySetupMenuPrizesKey,
									Type:        discord.InteractionComponentTypeTextInput,
									Style:       discord.InteractionComponentStyleParagraph,
									Placeholder: "Welcomer Pro\n2x Discord Nitro",
								},
							},
							{
								Type:        discord.InteractionComponentTypeLabel,
								Label:       "Duration",
								Description: "e.g. 1h, 30m, 2d. Only years, days, hours and minutes are supported.",
								Component: &discord.InteractionComponent{
									CustomID:    giveawaySetupMenuDurationKey,
									Type:        discord.InteractionComponentTypeTextInput,
									Placeholder: "7d 3h 60m",
									Style:       discord.InteractionComponentStyleShort,
									Required:    new(false),
								},
							},
						},
					},
				}, nil
			})
		},
	})

	giveawaysGroup.MustAddInteractionCommand(&subway.InteractionCommandable{
		Name:        "manage",
		Description: "Manage an existing giveaway",

		Type:        subway.InteractionCommandableTypeSubcommand,
		CommandType: new(discord.ApplicationCommandTypeMessage),

		DefaultMemberPermission: new(discord.Int64(welcomer.PermissionElevated)),
		DMPermission:            new(false),

		Handler: func(ctx context.Context, sub *subway.Subway, interaction discord.Interaction) (*discord.InteractionResponse, error) {
			return welcomer.RequireGuildElevation(sub, interaction, func() (*discord.InteractionResponse, error) {
				return &discord.InteractionResponse{
					Type: discord.InteractionCallbackTypeChannelMessageSource,
					Data: &discord.InteractionCallbackData{
						Components: []discord.InteractionComponent{
							{
								Type: discord.InteractionComponentTypeContainer,
								Components: []discord.InteractionComponent{
									{
										Type:    discord.InteractionComponentTypeTextDisplay,
										Content: "You can manage your giveaways settings such as disabling entries, extending the duration or ending the giveaway early by right clicking the giveaway message and selecting \"Manage Giveaway\".",
									},
									{
										Type: discord.InteractionComponentTypeMediaGallery,
										Items: []discord.InteractionComponentMediaGalleryItem{
											{
												Media: discord.MediaItem{
													URL: "https://welcomer.gg/assets/manage_giveaway.png",
												},
											},
										},
									},
								},
							},
						},
						Flags: uint32(discord.MessageFlagEphemeral + discord.MessageFlagIsComponentsV2),
					},
				}, nil
			})
		},
	})

	cog.InteractionCommands.MustAddInteractionCommand(&subway.InteractionCommandable{
		Name: "Manage Giveaway",

		Type:        subway.InteractionCommandableTypeCommand,
		CommandType: new(discord.ApplicationCommandTypeMessage),

		DefaultMemberPermission: new(discord.Int64(welcomer.PermissionElevated)),
		DMPermission:            new(false),

		Handler: func(ctx context.Context, sub *subway.Subway, interaction discord.Interaction) (*discord.InteractionResponse, error) {
			return welcomer.RequireGuildElevation(sub, interaction, func() (*discord.InteractionResponse, error) {
				if interaction.Data.TargetID == nil {
					return nil, nil
				}

				message, ok := interaction.Data.Resolved.Messages[*interaction.Data.TargetID]
				if !ok {
					welcomer.Logger.Error().
						Int64("guild_id", int64(*interaction.GuildID)).
						Int64("message_id", int64(*interaction.Data.TargetID)).
						Msg("Failed to find message for giveaway manage command")

					return nil, errors.New("failed to find message for giveaway manage command")
				}

				giveaway, err := welcomer.Queries.GetGiveawayFromMessageID(ctx, database.GetGiveawayFromMessageIDParams{
					GuildID:   int64(*interaction.GuildID),
					ChannelID: int64(message.ChannelID),
					MessageID: int64(message.ID),
				})
				if err != nil && !errors.Is(err, pgx.ErrNoRows) {
					welcomer.Logger.Error().Err(err).
						Int64("guild_id", int64(*interaction.GuildID)).
						Int64("channel_id", int64(message.ChannelID)).
						Int64("message_id", int64(message.ID)).
						Msg("Failed to get giveaway settings from message ID")

					return nil, err
				} else if errors.Is(err, pgx.ErrNoRows) {
					welcomer.Logger.Warn().
						Int64("guild_id", int64(*interaction.GuildID)).
						Int64("channel_id", int64(message.ChannelID)).
						Int64("message_id", int64(message.ID)).
						Msg("Giveaway not found for giveaway settings message")

					return &discord.InteractionResponse{
						Type: discord.InteractionCallbackTypeChannelMessageSource,
						Data: &discord.InteractionCallbackData{
							Embeds: welcomer.NewEmbed("This message is not associated with a giveaway. Please make sure you are using this command on the giveaway message.", welcomer.EmbedColourError),
							Flags:  uint32(discord.MessageFlagEphemeral),
						},
					}, nil
				}

				return &discord.InteractionResponse{
					Type: discord.InteractionCallbackTypeChannelMessageSource,
					Data: welcomer.WebhookMessageParamsToInteractionCallbackData(giveawayManageView(giveaway), uint32(discord.MessageFlagEphemeral+discord.MessageFlagIsComponentsV2)),
				}, nil
			})
		},
	})

	// TODO: reroll giveaway on message
	// TODO: resend giveaway message if accidentally deleted

	sub.RegisterComponentListener("giveaway_edit:*", handleGiveawayEditComponent)
	sub.RegisterComponentListener("giveaway_enter:*", handleGiveawayEnterComponent)
	sub.RegisterComponentListener("giveaway_manage:*", handleGiveawayManageComponent)

	cog.InteractionCommands.MustAddInteractionCommand(giveawaysGroup)

	return nil
}

func handleGiveawayManageComponent(ctx context.Context, sub *subway.Subway, interaction discord.Interaction) (*discord.InteractionResponse, error) {
	if interaction.GuildID == nil {
		return nil, nil
	}

	if interaction.Data.CustomID == "" {
		return nil, nil
	}

	customIDSplit := strings.Split(interaction.Data.CustomID, ":")
	if len(customIDSplit) < 3 {
		return nil, nil
	}

	giveawayUUID, err := uuid.FromString(customIDSplit[1])
	if err != nil {
		return nil, err
	}

	giveaway, err := welcomer.Queries.GetGiveaway(ctx, database.GetGiveawayParams{
		GuildID:      int64(*interaction.GuildID),
		GiveawayUuid: giveawayUUID,
	})
	if err != nil {
		welcomer.Logger.Error().Err(err).
			Int64("guild_id", int64(*interaction.GuildID)).
			Str("giveaway_uuid", giveawayUUID.String()).
			Msg("Failed to get giveaway settings")

		return nil, err
	}

	switch interaction.Type {
	case discord.InteractionTypeMessageComponent:
		switch customIDSplit[2] {
		case giveawayManageMenuToggleAllowEntriesKey:
			giveaway.AllowEntries = !giveaway.AllowEntries
		case giveawayManageMenuExtendDurationKey:
			return &discord.InteractionResponse{
				Type: discord.InteractionCallbackTypeModal,
				Data: &discord.InteractionCallbackData{
					Title:    "Extend Giveaway Duration",
					CustomID: interaction.Data.CustomID,
					Components: []discord.InteractionComponent{
						{
							Type: discord.InteractionComponentTypeTextDisplay,
							Content: "Enter the new giveaway duration from the current time. Leave empty if you want the giveaway to run indefinitely.\n\nIf you would like to remove time from the current duration, put a '-' before the duration." +
								welcomer.If(giveaway.EndTime.IsZero(), "\n\nThis giveaway is currently set to run indefinitely so this will be the duration from the current time.", ""),
						},
						{
							Type:        discord.InteractionComponentTypeLabel,
							Label:       "Duration",
							Description: "e.g. 1h, 30m, 2d, -5m. Only years, days, hours and minutes are supported.",
							Component: &discord.InteractionComponent{
								CustomID:    giveawayManageMenuExtendDurationKey,
								Type:        discord.InteractionComponentTypeTextInput,
								Placeholder: "7d 3h 60m -2d",
								Style:       discord.InteractionComponentStyleShort,
								Required:    new(false),
							},
						},
					},
				},
			}, nil
		case giveawayManageMenuEndGiveawayKey:
			return &discord.InteractionResponse{
				Type: discord.InteractionCallbackTypeModal,
				Data: &discord.InteractionCallbackData{
					Title:    "End Giveaway",
					CustomID: interaction.Data.CustomID,
					Components: []discord.InteractionComponent{
						{
							Type:    discord.InteractionComponentTypeTextDisplay,
							Content: "Are you sure you want to end the giveaway early? This cannot be undone.",
						},
					},
				},
			}, nil
		case giveawayManageMenuExportEntriesKey:
			return exportGiveawayEntries(ctx, sub, interaction, giveaway)
		case giveawayManageMenuExportWinnersKey:
			return exportGiveawayWinners(ctx, sub, interaction, giveaway)
		default:
			welcomer.Logger.Warn().
				Int64("guild_id", int64(*interaction.GuildID)).
				Str("giveaway_uuid", giveawayUUID.String()).
				Str("custom_id", interaction.Data.CustomID).
				Msg("Unknown giveaway manage component interaction")
		}
	case discord.InteractionTypeModalSubmit:
		switch customIDSplit[2] {
		case giveawayManageMenuExtendDurationKey:
			durationArgument, err := subway.GetArgument(ctx, giveawayManageMenuExtendDurationKey)

			if err == nil {
				durationString := durationArgument.MustString()
				durationString, hasMinus := strings.CutPrefix(durationString, "-")

				seconds, err := welcomer.ParseDurationAsSeconds(durationString)
				if err != nil || seconds < 0 {
					welcomer.Logger.Error().Err(err).
						Int64("guild_id", int64(*interaction.GuildID)).
						Str("duration", durationString).
						Msg("Failed to parse duration")

					return nil, nil
				}

				// If the duration is indefinite, reset to current time.
				if giveaway.EndTime.IsZero() {
					giveaway.EndTime = time.Now()
				}

				if hasMinus {
					giveaway.EndTime = giveaway.EndTime.Add(-time.Duration(seconds) * time.Second)
				} else {
					giveaway.EndTime = giveaway.EndTime.Add(time.Duration(seconds) * time.Second)
				}

				if giveaway.EndTime.Before(time.Now()) {
					return &discord.InteractionResponse{
						Type: discord.InteractionCallbackTypeChannelMessageSource,
						Data: &discord.InteractionCallbackData{
							Embeds: welcomer.NewEmbed("The new end time cannot be in the past. Please end the giveaway if you want to do this.", welcomer.EmbedColourError),
							Flags:  uint32(discord.MessageFlagEphemeral),
						},
					}, nil
				}
			} else {
				// If no duration is passed, make the duration indefinite.
				giveaway.EndTime = time.Time{}
			}
		case giveawayManageMenuEndGiveawayKey:
			giveaway.EndTime = time.Now()

			data, _ := json.Marshal(welcomer.CustomEventInvokeEndGiveawayStructure{
				GiveawayUUID: giveaway.GiveawayUuid,
				GuildID:      *interaction.GuildID,
			})

			_, err = sub.SandwichClient.RelayMessage(ctx, &sandwich.RelayMessageRequest{
				Identifier: welcomer.GetManagerNameFromContext(ctx),
				Type:       welcomer.CustomEventInvokeEndGiveaway,
				Data:       data,
			})
			if err != nil {
				return nil, err
			}
		default:
			welcomer.Logger.Warn().
				Int64("guild_id", int64(*interaction.GuildID)).
				Str("giveaway_uuid", giveawayUUID.String()).
				Str("custom_id", interaction.Data.CustomID).
				Msg("Unknown giveaway manage modal submit interaction")
		}
	}

	_, err = welcomer.UpdateGiveawayGuildSettingsWithAudit(ctx, database.UpdateGiveawayParams{
		GiveawayUuid:    giveaway.GiveawayUuid,
		IsSetup:         giveaway.IsSetup,
		Title:           giveaway.Title,
		StartTime:       giveaway.StartTime,
		EndTime:         giveaway.EndTime,
		AnnounceWinners: giveaway.AnnounceWinners,
		GiveawayPrizes:  giveaway.GiveawayPrizes,
		RolesAllowed:    giveaway.RolesAllowed,
		RolesExcluded:   giveaway.RolesExcluded,
		MinimumJoinDate: giveaway.MinimumJoinDate,
		Description:     giveaway.Description,
		AccentColour:    giveaway.AccentColour,
		ImageUrl:        giveaway.ImageUrl,
		ShowPrizes:      giveaway.ShowPrizes,
		ShowEntries:     giveaway.ShowEntries,
		AllowEntries:    giveaway.AllowEntries,
		HasEnded:        giveaway.HasEnded,
	}, interaction.GetUser().ID, *interaction.GuildID)
	if err != nil {
		welcomer.Logger.Error().Err(err).
			Int64("guild_id", int64(*interaction.GuildID)).
			Str("giveaway_uuid", giveaway.GiveawayUuid.String()).
			Msg("Failed to update giveaway settings")

		return nil, err
	}

	if customIDSplit[2] == giveawayManageMenuEndGiveawayKey {
		giveaway.HasEnded = true
	}

	err = discord.CreateInteractionResponse(ctx, sub.EmptySession, interaction.ID, interaction.Token, discord.InteractionResponse{
		Type: discord.InteractionCallbackTypeUpdateMessage,
		Data: welcomer.WebhookMessageParamsToInteractionCallbackData(giveawayManageView(giveaway), uint32(discord.MessageFlagEphemeral+discord.MessageFlagIsComponentsV2)),
	})
	if err != nil {
		welcomer.Logger.Error().Err(err).
			Int64("guild_id", int64(*interaction.GuildID)).
			Str("giveaway_uuid", giveawayUUID.String()).
			Str("custom_id", interaction.Data.CustomID).
			Msg("Failed to create interaction response for giveaway manage component")
	}

	return nil, nil
}

func exportGiveawayEntries(ctx context.Context, sub *subway.Subway, interaction discord.Interaction, giveaway *database.GuildGiveaways) (*discord.InteractionResponse, error) {
	entries, err := welcomer.Queries.GetGiveawayEntries(ctx, giveaway.GiveawayUuid)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		welcomer.Logger.Error().Err(err).
			Str("giveaway_uuid", giveaway.GiveawayUuid.String()).
			Msg("Failed to get giveaway entry users")

		return nil, err
	}

	var file bytes.Buffer

	writer := csv.NewWriter(&file)

	_ = writer.Write([]string{"user_id", "entered_at"})

	for _, entry := range entries {
		_ = writer.Write([]string{
			welcomer.Itoa(entry.UserID),
			entry.CreatedAt.Format("2006-01-02 15:04:05"),
		})
	}

	writer.Flush()

	if err := writer.Error(); err != nil {
		welcomer.Logger.Error().Err(err).
			Str("giveaway_uuid", giveaway.GiveawayUuid.String()).
			Msg("Failed to write giveaway entries to csv")

		return nil, err
	}

	err = interaction.SendResponse(ctx, sub.EmptySession, discord.InteractionCallbackTypeChannelMessageSource, &discord.InteractionCallbackData{
		Content: "Here are the entries for this giveaway:",
		Files: []discord.File{
			{
				Reader:      &file,
				Name:        fmt.Sprintf("giveaway_entries_%s.csv", giveaway.GiveawayUuid.String()),
				ContentType: "text/csv",
			},
		},
		Flags: uint32(discord.MessageFlagEphemeral),
	})
	if err != nil {
		welcomer.Logger.Error().Err(err).
			Str("giveaway_uuid", giveaway.GiveawayUuid.String()).
			Msg("Failed to send giveaway entries response")

		return nil, err
	}

	return nil, nil
}

func exportGiveawayWinners(ctx context.Context, sub *subway.Subway, interaction discord.Interaction, giveaway *database.GuildGiveaways) (*discord.InteractionResponse, error) {
	winners, err := welcomer.Queries.GetGiveawayWinners(ctx, giveaway.GiveawayUuid)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		welcomer.Logger.Error().Err(err).
			Str("giveaway_uuid", giveaway.GiveawayUuid.String()).
			Msg("Failed to get giveaway winners")

		return nil, err
	}

	var file bytes.Buffer

	writer := csv.NewWriter(&file)

	_ = writer.Write([]string{"user_id", "prize", "message_id"})

	for _, winner := range winners {
		_ = writer.Write([]string{
			welcomer.Itoa(winner.UserID),
			winner.Prize,
			welcomer.Itoa(winner.MessageID),
		})
	}

	writer.Flush()

	if err := writer.Error(); err != nil {
		welcomer.Logger.Error().Err(err).
			Str("giveaway_uuid", giveaway.GiveawayUuid.String()).
			Msg("Failed to write giveaway winners to csv")

		return nil, err
	}

	err = interaction.SendResponse(ctx, sub.EmptySession, discord.InteractionCallbackTypeChannelMessageSource, &discord.InteractionCallbackData{
		Content: "Here are the winners for this giveaway:\n" +
			"-# If the giveaway has ended and the winners is empty, it may mean it has not finished processing the giveaway yet.",
		Files: []discord.File{
			{
				Reader:      &file,
				Name:        fmt.Sprintf("giveaway_winners_%s.csv", giveaway.GiveawayUuid.String()),
				ContentType: "text/csv",
			},
		},
		Flags: uint32(discord.MessageFlagEphemeral),
	})
	if err != nil {
		welcomer.Logger.Error().Err(err).
			Str("giveaway_uuid", giveaway.GiveawayUuid.String()).
			Msg("Failed to send giveaway entries response")

		return nil, err
	}

	return nil, nil
}

func handleGiveawayEnterComponent(ctx context.Context, sub *subway.Subway, interaction discord.Interaction) (*discord.InteractionResponse, error) {
	if interaction.GuildID == nil {
		return nil, nil
	}

	if interaction.Data.CustomID == "" {
		return nil, nil
	}

	customIDSplit := strings.Split(interaction.Data.CustomID, ":")
	if len(customIDSplit) < 2 {
		return nil, nil
	}

	giveawayUUID, err := uuid.FromString(customIDSplit[1])
	if err != nil {
		return nil, err
	}

	giveaway, err := welcomer.Queries.GetGiveaway(ctx, database.GetGiveawayParams{
		GuildID:      int64(*interaction.GuildID),
		GiveawayUuid: giveawayUUID,
	})
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		welcomer.Logger.Error().Err(err).
			Int64("guild_id", int64(*interaction.GuildID)).
			Str("giveaway_uuid", giveawayUUID.String()).
			Msg("Failed to get giveaway settings")

		return nil, err
	} else if errors.Is(err, pgx.ErrNoRows) {
		welcomer.Logger.Warn().
			Int64("guild_id", int64(*interaction.GuildID)).
			Str("giveaway_uuid", giveawayUUID.String()).
			Msg("Giveaway not found for giveaway entry")

		return &discord.InteractionResponse{
			Type: discord.InteractionCallbackTypeChannelMessageSource,
			Data: &discord.InteractionCallbackData{
				Embeds: welcomer.NewEmbed("This giveaway no longer exists. It may have been deleted or ended.", welcomer.EmbedColourError),
				Flags:  uint32(discord.MessageFlagEphemeral),
			},
		}, nil
	}

	if giveaway.HasEnded {
		return &discord.InteractionResponse{
			Type: discord.InteractionCallbackTypeChannelMessageSource,
			Data: &discord.InteractionCallbackData{
				Embeds: welcomer.NewEmbed("This giveaway has already ended. Better luck next time!", welcomer.EmbedColourError),
				Flags:  uint32(discord.MessageFlagEphemeral),
			},
		}, nil
	}

	if !giveaway.AllowEntries {
		return &discord.InteractionResponse{
			Type: discord.InteractionCallbackTypeChannelMessageSource,
			Data: &discord.InteractionCallbackData{
				Embeds: welcomer.NewEmbed("This giveaway does not have entries enabled. Please try again later.", welcomer.EmbedColourError),
				Flags:  uint32(discord.MessageFlagEphemeral),
			},
		}, nil
	}

	rolesAllowed := welcomer.UnmarshalRolesListJSON(giveaway.RolesAllowed.Bytes)
	rolesExcluded := welcomer.UnmarshalRolesListJSON(giveaway.RolesExcluded.Bytes)

	if len(rolesAllowed) > 0 && !hasAnyRoles(rolesAllowed, interaction.Member.Roles) {
		return &discord.InteractionResponse{
			Type: discord.InteractionCallbackTypeChannelMessageSource,
			Data: &discord.InteractionCallbackData{
				Embeds: welcomer.NewEmbed("Sorry, you are missing a required role to enter this giveaway.", welcomer.EmbedColourError),
				Flags:  uint32(discord.MessageFlagEphemeral),
			},
		}, nil
	}

	if len(rolesExcluded) > 0 && hasAnyRoles(rolesExcluded, interaction.Member.Roles) {
		return &discord.InteractionResponse{
			Type: discord.InteractionCallbackTypeChannelMessageSource,
			Data: &discord.InteractionCallbackData{
				Embeds: welcomer.NewEmbed("Sorry, you have a role that disqualifies you from entering this giveaway.", welcomer.EmbedColourError),
				Flags:  uint32(discord.MessageFlagEphemeral),
			},
		}, nil
	}

	if !giveaway.MinimumJoinDate.IsZero() {
		joinBefore := giveaway.StartTime.Add(-(time.Duration(giveaway.MinimumJoinDate.Unix()) * time.Second))
		if interaction.Member.JoinedAt.After(joinBefore) {
			return &discord.InteractionResponse{
				Type: discord.InteractionCallbackTypeChannelMessageSource,
				Data: &discord.InteractionCallbackData{
					Embeds: welcomer.NewEmbed(fmt.Sprintf("Sorry, you must have joined the server before <t:%d:f> to enter this giveaway.", joinBefore.Unix()), welcomer.EmbedColourError),
					Flags:  uint32(discord.MessageFlagEphemeral),
				},
			}, nil
		}
	}

	_, err = welcomer.Queries.AddGiveawayEntry(ctx, database.AddGiveawayEntryParams{
		GiveawayUuid: giveawayUUID,
		UserID:       int64(interaction.Member.User.ID),
	})
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		welcomer.Logger.Error().Err(err).
			Int64("guild_id", int64(*interaction.GuildID)).
			Str("giveaway_uuid", giveawayUUID.String()).
			Msg("Failed to add giveaway entry")
	}

	if err != nil && errors.Is(err, pgx.ErrNoRows) {
		return &discord.InteractionResponse{
			Type: discord.InteractionCallbackTypeChannelMessageSource,
			Data: &discord.InteractionCallbackData{
				Embeds: welcomer.NewEmbed("You have already entered this giveaway! Good luck!", welcomer.EmbedColourInfo),
				Flags:  uint32(discord.MessageFlagEphemeral),
			},
		}, nil
	}

	entries, err := welcomer.Queries.CountGiveawayEntries(ctx, giveawayUUID)
	if err != nil {
		welcomer.Logger.Error().Err(err).
			Int64("guild_id", int64(*interaction.GuildID)).
			Str("giveaway_uuid", giveawayUUID.String()).
			Msg("Failed to count giveaway entries")
	}

	go func() {
		time.Sleep(5 * time.Second)

		newEntries, err := welcomer.Queries.CountGiveawayEntries(ctx, giveawayUUID)
		if err != nil {
			welcomer.Logger.Error().Err(err).
				Int64("guild_id", int64(*interaction.GuildID)).
				Str("giveaway_uuid", giveawayUUID.String()).
				Msg("Failed to count giveaway entries")
		}

		if entries == newEntries {
			message := discord.Message{
				ID:        discord.Snowflake(giveaway.MessageID),
				ChannelID: discord.Snowflake(giveaway.ChannelID),
			}

			session, err := welcomer.AcquireSession(ctx, welcomer.GetManagerNameFromContext(ctx))
			if err != nil {
				welcomer.Logger.Error().Err(err).
					Int64("guild_id", int64(*interaction.GuildID)).
					Str("giveaway_uuid", giveawayUUID.String()).
					Msg("Failed to acquire session to edit giveaway message after entry")

				return
			}

			_, err = message.Edit(ctx, session, welcomer.WebhookMessageParamsToMessageParams(giveawayView(giveaway, newEntries)))
			if err != nil {
				welcomer.Logger.Error().Err(err).
					Int64("guild_id", int64(*interaction.GuildID)).
					Str("giveaway_uuid", giveawayUUID.String()).
					Msg("Failed to edit giveaway message after entry")
			}

			welcomer.Logger.Info().
				Int64("guild_id", int64(*interaction.GuildID)).
				Str("giveaway_uuid", giveawayUUID.String()).
				Int32("entries", newEntries).
				Msg("Updated giveaway message after new entry")
		}
	}()

	return &discord.InteractionResponse{
		Type: discord.InteractionCallbackTypeChannelMessageSource,
		Data: &discord.InteractionCallbackData{
			Embeds: welcomer.NewEmbed("You have successfully entered the giveaway! Good luck!", welcomer.EmbedColourSuccess),
			Flags:  uint32(discord.MessageFlagEphemeral),
		},
	}, nil
}

func hasAnyRoles(roleList, userRoles []discord.Snowflake) bool {
	for _, role := range roleList {
		if slices.Contains(userRoles, role) {
			return true
		}
	}

	return false
}

func handleGiveawayEditComponent(ctx context.Context, sub *subway.Subway, interaction discord.Interaction) (*discord.InteractionResponse, error) {
	if interaction.GuildID == nil {
		return nil, nil
	}

	if interaction.Data.CustomID == "" {
		return nil, nil
	}

	customIDSplit := strings.Split(interaction.Data.CustomID, ":")
	if len(customIDSplit) < 2 {
		return nil, nil
	}

	if len(customIDSplit) < 3 {
		customIDSplit = append(customIDSplit, "")
	}

	giveawayUUID, err := uuid.FromString(customIDSplit[1])
	if err != nil {
		return nil, err
	}

	giveaway, err := welcomer.Queries.GetGiveaway(ctx, database.GetGiveawayParams{
		GuildID:      int64(*interaction.GuildID),
		GiveawayUuid: giveawayUUID,
	})
	if err != nil {
		welcomer.Logger.Error().Err(err).
			Int64("guild_id", int64(*interaction.GuildID)).
			Str("giveaway_uuid", giveawayUUID.String()).
			Msg("Failed to get giveaway settings")

		return nil, err
	}

	switch interaction.Type {
	case discord.InteractionTypeMessageComponent:
		switch customIDSplit[2] {
		case giveawaySetupMenuTitleKey:
			return &discord.InteractionResponse{
				Data: &discord.InteractionCallbackData{
					Title:    "Customise Giveaway Message",
					CustomID: interaction.Data.CustomID,
					Components: []discord.InteractionComponent{
						{
							Type:  discord.InteractionComponentTypeLabel,
							Label: "Title",
							Component: &discord.InteractionComponent{
								CustomID: giveawaySetupMenuTitleKey,
								Type:     discord.InteractionComponentTypeTextInput,
								Value:    giveaway.Title,
								Style:    discord.InteractionComponentStyleShort,
								Required: new(false),
							},
						},
						{
							Type:  discord.InteractionComponentTypeLabel,
							Label: "Description",
							Component: &discord.InteractionComponent{
								CustomID: giveawaySetupMenuDescriptionKey,
								Type:     discord.InteractionComponentTypeTextInput,
								Value:    giveaway.Description,
								Style:    discord.InteractionComponentStyleParagraph,
								Required: new(false),
							},
						},
						{
							Type:        discord.InteractionComponentTypeLabel,
							Label:       "Accent Colour",
							Description: "If specified, the left side of the giveaway message will be this colour. Accepts #HEX format.",
							Component: &discord.InteractionComponent{
								CustomID:    giveawaySetupMenuAccentColourKey,
								Type:        discord.InteractionComponentTypeTextInput,
								Placeholder: "#4CD787",
								Value:       welcomer.If(giveaway.AccentColour < 0, "", fmt.Sprintf("#%06X", giveaway.AccentColour)),
								Style:       discord.InteractionComponentStyleShort,
								Required:    new(false),
							},
						},
						{
							Type:        discord.InteractionComponentTypeLabel,
							Label:       "Image URL",
							Description: "If specified, this image will show below your title and description.",
							Component: &discord.InteractionComponent{
								CustomID:    giveawaySetupMenuThumbnailURLKey,
								Type:        discord.InteractionComponentTypeTextInput,
								Placeholder: "https://example.com/image.png",
								Value:       giveaway.ImageUrl,
								Style:       discord.InteractionComponentStyleShort,
								Required:    new(false),
							},
						},
						{
							Type:  discord.InteractionComponentTypeLabel,
							Label: "Giveaway Message Display Options",
							Component: &discord.InteractionComponent{
								Type:     discord.InteractionComponentTypeCheckboxGroup,
								CustomID: giveawaySetupMenuDisplayKey,
								Required: new(false),
								Options: []discord.ApplicationSelectOption{
									{
										Label:       "Show Prizes",
										Value:       giveawaySetupMenuDisplayShowPrizesKey,
										Description: "When enabled, the list of prizes will show in the giveaway message shown to users.",
										Default:     giveaway.ShowPrizes,
									},
									{
										Label:       "Show Entries",
										Value:       giveawaySetupMenuDisplayShowEntriesKey,
										Description: "When enabled, the number of entries will show in the giveaway message shown to users.",
										Default:     giveaway.ShowEntries,
									},
								},
							},
						},
					},
				},
				Type: discord.InteractionCallbackTypeModal,
			}, nil
		case giveawaySetupMenuPrizesKey:
			return &discord.InteractionResponse{
				Data: &discord.InteractionCallbackData{
					Title:    "Edit Giveaway Prizes",
					CustomID: interaction.Data.CustomID,
					Components: []discord.InteractionComponent{
						{
							Type:        discord.InteractionComponentTypeLabel,
							Label:       "Prizes",
							Description: "One prize per line, with optional count, e.g. 2x Discord Nitro",
							Component: &discord.InteractionComponent{
								CustomID:    giveawaySetupMenuPrizesKey,
								Type:        discord.InteractionComponentTypeTextInput,
								Placeholder: "Welcomer Pro\n2x Discord Nitro",
								Value:       formatGiveawayPrizesAsString(welcomer.UnmarshalGiveawayPrizeJSON(giveaway.GiveawayPrizes.Bytes)),
								Style:       discord.InteractionComponentStyleParagraph,
							},
						},
					},
				},
				Type: discord.InteractionCallbackTypeModal,
			}, nil
		case giveawaySetupMenuDurationKey:
			return &discord.InteractionResponse{
				Data: &discord.InteractionCallbackData{
					Title:    "Edit Giveaway Duration",
					CustomID: interaction.Data.CustomID,
					Components: []discord.InteractionComponent{
						{
							Type:        discord.InteractionComponentTypeLabel,
							Label:       "Duration",
							Description: "e.g. 1h, 30m, 2d. Only years, days, hours and minutes are supported.",
							Component: &discord.InteractionComponent{
								CustomID:    giveawaySetupMenuDurationKey,
								Type:        discord.InteractionComponentTypeTextInput,
								Placeholder: "7d 3h 60m",
								Style:       discord.InteractionComponentStyleShort,
								Required:    new(false),
							},
						},
					},
				},
				Type: discord.InteractionCallbackTypeModal,
			}, nil
		case giveawaySetupMenuAnnounceWinnersKey:
			giveaway.AnnounceWinners = !giveaway.AnnounceWinners
		case giveawaySetupMenuRolesAllowedKey:
			return &discord.InteractionResponse{
				Data: &discord.InteractionCallbackData{
					Title:    "Edit Giveaway Entry Rules",
					CustomID: interaction.Data.CustomID,
					Components: []discord.InteractionComponent{
						{
							Type:        discord.InteractionComponentTypeLabel,
							Label:       "Roles Allowed to Enter",
							Description: "Users must have at least one of these roles to enter. Ignored if empty.",
							Component: &discord.InteractionComponent{
								CustomID:  giveawaySetupMenuRolesAllowedIncludedKey,
								Type:      discord.InteractionComponentTypeRoleSelect,
								Required:  new(false),
								MaxValues: new(int32(25)),
							},
						},
						{
							Type:        discord.InteractionComponentTypeLabel,
							Label:       "Roles Excluded from Entering",
							Description: "Users with any of these roles cannot enter. Ignored if empty.",
							Component: &discord.InteractionComponent{
								CustomID:  giveawaySetupMenuRolesAllowedExcludedKey,
								Type:      discord.InteractionComponentTypeRoleSelect,
								Required:  new(false),
								MaxValues: new(int32(25)),
							},
						},
					},
				},
				Type: discord.InteractionCallbackTypeModal,
			}, nil
		case giveawaySetupMenuMinimumJoinDateKey:
			return &discord.InteractionResponse{
				Data: &discord.InteractionCallbackData{
					Title:    "Edit Giveaway Minimum Join Date",
					CustomID: interaction.Data.CustomID,
					Components: []discord.InteractionComponent{
						{
							Type:        discord.InteractionComponentTypeLabel,
							Label:       "Minimum Join Date",
							Description: "Users joined within the duration specified cannot enter. Ignored if empty. e.g. 1h, 30m, 2d. ",
							Component: &discord.InteractionComponent{
								CustomID:    giveawaySetupMenuMinimumJoinDateKey,
								Type:        discord.InteractionComponentTypeTextInput,
								Placeholder: "7d 3h 60m",
								Style:       discord.InteractionComponentStyleShort,
								Required:    new(false),
							},
						},
					},
				},
				Type: discord.InteractionCallbackTypeModal,
			}, nil
		case giveawaySetupMenuStartKey:
			roles := welcomer.UnmarshalRolesListJSON(giveaway.RolesAllowed.Bytes)

			var options []discord.ApplicationSelectOption

			if len(roles) == 0 {
				options = []discord.ApplicationSelectOption{
					{
						Label: "Ping @everyone",
						Value: giveawaySetupMenuPingEveryoneKey,
					},
					{
						Label: "Ping @here",
						Value: giveawaySetupMenuPingHereKey,
					},
				}
			} else {
				options = []discord.ApplicationSelectOption{
					{
						Label:       "Ping Roles Allowed to Enter",
						Value:       giveawaySetupMenuPingRolesAllowedToEnterKey,
						Description: "Pings roles you have configured in \"roles allowed to enter\"",
					},
				}
			}

			return &discord.InteractionResponse{
				Data: &discord.InteractionCallbackData{
					Title:    "Start Giveaway",
					CustomID: interaction.Data.CustomID,
					Components: []discord.InteractionComponent{
						{
							Type:    discord.InteractionComponentTypeTextDisplay,
							Content: "Once started, the giveaway message will be sent and entries will be allowed. You can end or extend the giveaway at any time, but you cannot edit the giveaway settings.\n\nBelow you can configure who should be pinged when the giveaway starts.",
						},
						{
							Type:  discord.InteractionComponentTypeLabel,
							Label: "Delivery Option",
							Component: &discord.InteractionComponent{
								Type:      discord.InteractionComponentTypeCheckboxGroup,
								CustomID:  giveawaySetupMenuPingKey,
								Required:  new(false),
								MaxValues: new(int32(1)),
								Options:   options,
							},
						},
						{
							Type:  discord.InteractionComponentTypeLabel,
							Label: "Additional Roles to Ping",
							Component: &discord.InteractionComponent{
								Type:      discord.InteractionComponentTypeRoleSelect,
								CustomID:  giveawaySetupMenuPingAdditionalRolesKey,
								Required:  new(false),
								MaxValues: new(int32(25)),
							},
						},
					},
				},
				Type: discord.InteractionCallbackTypeModal,
			}, nil
		case giveawaySetupMenuPreviewOnKey:
			giveaway.StartTime = time.Now()

			if giveaway.EndTime.Unix() > 0 {
				giveaway.EndTime = time.Now().Add(time.Duration(giveaway.EndTime.Unix()) * time.Second)
			}

			message := giveawayView(giveaway, 0)

			// Hack to disable giveaway button and add back button
			message.Components[len(message.Components)-1].Components[0].Disabled = true
			message.Components[len(message.Components)-1].Components = append(message.Components[len(message.Components)-1].Components, discord.InteractionComponent{
				CustomID: "giveaway_edit:" + giveaway.GiveawayUuid.String() + ":preview_off",
				Type:     discord.InteractionComponentTypeButton,
				Label:    "Back to Edit Menu",
				Style:    discord.InteractionComponentStyleSecondary,
			})

			err = discord.CreateInteractionResponse(ctx, sub.EmptySession, interaction.ID, interaction.Token, discord.InteractionResponse{
				Type: welcomer.If(customIDSplit[2] == "", discord.InteractionCallbackTypeChannelMessageSource, discord.InteractionCallbackTypeUpdateMessage),
				Data: welcomer.WebhookMessageParamsToInteractionCallbackData(message, uint32(discord.MessageFlagEphemeral+discord.MessageFlagIsComponentsV2)),
			})
			if err != nil {
				welcomer.Logger.Error().Err(err).
					Int64("guild_id", int64(*interaction.GuildID)).
					Str("giveaway_uuid", giveaway.GiveawayUuid.String()).
					Msg("Failed to edit giveaway message")

				return nil, err
			}

			return nil, nil
		case giveawaySetupMenuPreviewOffKey:
			err = discord.CreateInteractionResponse(ctx, sub.EmptySession, interaction.ID, interaction.Token, discord.InteractionResponse{
				Type: welcomer.If(customIDSplit[2] == "", discord.InteractionCallbackTypeChannelMessageSource, discord.InteractionCallbackTypeUpdateMessage),
				Data: welcomer.WebhookMessageParamsToInteractionCallbackData(giveawaySetupView(giveaway), uint32(discord.MessageFlagEphemeral+discord.MessageFlagIsComponentsV2)),
			})
			if err != nil {
				welcomer.Logger.Error().Err(err).
					Int64("guild_id", int64(*interaction.GuildID)).
					Str("giveaway_uuid", giveaway.GiveawayUuid.String()).
					Msg("Failed to edit giveaway message")

				return nil, err
			}

			return nil, nil
		default:
			welcomer.Logger.Warn().
				Int64("guild_id", int64(*interaction.GuildID)).
				Str("custom_id", interaction.Data.CustomID).
				Msg("Unknown giveaway edit menu option")

			return nil, nil
		}
	case discord.InteractionTypeModalSubmit:
		switch customIDSplit[2] {
		case "":
			if titleArgument, err := subway.GetArgument(ctx, giveawaySetupMenuTitleKey); err == nil {
				giveaway.Title = titleArgument.MustString()
			}

			if prizesArgument, err := subway.GetArgument(ctx, giveawaySetupMenuPrizesKey); err == nil {
				prizes := parsePrizesFromString(prizesArgument.MustString())
				giveaway.GiveawayPrizes = pgtype.JSONB{
					Bytes:  welcomer.MarshalGiveawayPrizeJSON(prizes),
					Status: pgtype.Present,
				}
			}

			if durationArgument, err := subway.GetArgument(ctx, giveawaySetupMenuDurationKey); err == nil {
				seconds, err := welcomer.ParseDurationAsSeconds(durationArgument.MustString())
				if err != nil || seconds < 0 {
					welcomer.Logger.Error().Err(err).
						Int64("guild_id", int64(*interaction.GuildID)).
						Str("duration", durationArgument.MustString()).
						Msg("Failed to parse duration")

					return nil, nil
				}

				giveaway.EndTime = time.Unix(int64(seconds), 0)
			}
		case giveawaySetupMenuTitleKey:
			if titleArgument, err := subway.GetArgument(ctx, giveawaySetupMenuTitleKey); err == nil {
				giveaway.Title = titleArgument.MustString()
			} else {
				giveaway.Title = ""
			}

			if descriptionArgument, err := subway.GetArgument(ctx, giveawaySetupMenuDescriptionKey); err == nil {
				giveaway.Description = descriptionArgument.MustString()
			} else {
				giveaway.Description = ""
			}

			if accentColourArgument, err := subway.GetArgument(ctx, giveawaySetupMenuAccentColourKey); err == nil {
				rgba, err := welcomer.ParseColour(accentColourArgument.MustString(), "#000000")
				if err != nil {
					welcomer.Logger.Error().Err(err).
						Int64("guild_id", int64(*interaction.GuildID)).
						Str("accent_colour", accentColourArgument.MustString()).
						Msg("Failed to parse accent colour")

					giveaway.AccentColour = -1
				} else {
					giveaway.AccentColour = int64(int32(rgba.R)<<16 + int32(rgba.G)<<8 + int32(rgba.B))
				}
			} else {
				giveaway.AccentColour = -1
			}

			if thumbnailURLArgument, err := subway.GetArgument(ctx, giveawaySetupMenuThumbnailURLKey); err == nil {
				giveaway.ImageUrl = thumbnailURLArgument.MustString()

				if _, ok := welcomer.IsValidURL(thumbnailURLArgument.MustString()); !ok {
					welcomer.Logger.Error().Err(err).
						Int64("guild_id", int64(*interaction.GuildID)).
						Str("thumbnail_url", thumbnailURLArgument.MustString()).
						Msg("Failed to parse thumbnail URL")

					giveaway.ImageUrl = ""
				}
			} else {
				giveaway.ImageUrl = ""
			}

			if displayOptionsArgument, err := subway.GetArgument(ctx, giveawaySetupMenuDisplayKey); err == nil {
				displayOptions := displayOptionsArgument.MustStrings()

				giveaway.ShowPrizes = slices.Contains(displayOptions, giveawaySetupMenuDisplayShowPrizesKey)
				giveaway.ShowEntries = slices.Contains(displayOptions, giveawaySetupMenuDisplayShowEntriesKey)
			} else {
				giveaway.ShowPrizes = false
				giveaway.ShowEntries = false
			}
		case giveawaySetupMenuPrizesKey:
			if prizesArgument, err := subway.GetArgument(ctx, giveawaySetupMenuPrizesKey); err == nil {
				prizes := parsePrizesFromString(prizesArgument.MustString())
				giveaway.GiveawayPrizes = pgtype.JSONB{
					Bytes:  welcomer.MarshalGiveawayPrizeJSON(prizes),
					Status: pgtype.Present,
				}
			} else {
				giveaway.GiveawayPrizes = pgtype.JSONB{
					Bytes:  nil,
					Status: pgtype.Null,
				}
			}
		case giveawaySetupMenuDurationKey:
			if durationArgument, err := subway.GetArgument(ctx, giveawaySetupMenuDurationKey); err == nil {
				seconds, err := welcomer.ParseDurationAsSeconds(durationArgument.MustString())
				if err != nil || seconds < 0 {
					welcomer.Logger.Error().Err(err).
						Int64("guild_id", int64(*interaction.GuildID)).
						Str("duration", durationArgument.MustString()).
						Msg("Failed to parse duration")

					return nil, nil
				}

				giveaway.EndTime = time.Unix(int64(seconds), 0)
			} else {
				giveaway.EndTime = time.Time{}
			}
		case giveawaySetupMenuRolesAllowedKey:
			if allowedRoles, err := subway.GetArgument(ctx, giveawaySetupMenuRolesAllowedIncludedKey); err == nil {
				allowedRolesList := make([]discord.Snowflake, 0, len(allowedRoles.MustStrings()))

				for _, roleString := range allowedRoles.MustStrings() {
					roleSnowflake, err := welcomer.Atoi(roleString)
					if err == nil {
						allowedRolesList = append(allowedRolesList, discord.Snowflake(roleSnowflake))
					}
				}

				giveaway.RolesAllowed = pgtype.JSONB{
					Bytes:  welcomer.MarshalRolesListJSON(allowedRolesList),
					Status: pgtype.Present,
				}
			} else {
				giveaway.RolesAllowed = pgtype.JSONB{
					Bytes:  []byte{123, 125}, // []
					Status: pgtype.Present,
				}
			}

			if excludedRoles, err := subway.GetArgument(ctx, giveawaySetupMenuRolesAllowedExcludedKey); err == nil {
				excludedRolesList := make([]discord.Snowflake, 0, len(excludedRoles.MustStrings()))

				for _, roleString := range excludedRoles.MustStrings() {
					roleSnowflake, err := welcomer.Atoi(roleString)
					if err == nil {
						excludedRolesList = append(excludedRolesList, discord.Snowflake(roleSnowflake))
					}
				}

				giveaway.RolesExcluded = pgtype.JSONB{
					Bytes:  welcomer.MarshalRolesListJSON(excludedRolesList),
					Status: pgtype.Present,
				}
			} else {
				giveaway.RolesExcluded = pgtype.JSONB{
					Bytes:  []byte{123, 125}, // []
					Status: pgtype.Present,
				}
			}
		case giveawaySetupMenuMinimumJoinDateKey:
			if durationArgument, err := subway.GetArgument(ctx, giveawaySetupMenuMinimumJoinDateKey); err == nil {
				seconds, err := welcomer.ParseDurationAsSeconds(durationArgument.MustString())
				if err != nil || seconds < 0 {
					welcomer.Logger.Error().Err(err).
						Int64("guild_id", int64(*interaction.GuildID)).
						Str("duration", durationArgument.MustString()).
						Msg("Failed to parse duration")

					return nil, nil
				}

				giveaway.MinimumJoinDate = time.Unix(int64(seconds), 0)
			} else {
				giveaway.MinimumJoinDate = time.Time{}
			}
		case giveawaySetupMenuStartKey:
			var pingOptions []string

			if pingOptionsArgument, err := subway.GetArgument(ctx, giveawaySetupMenuPingKey); err == nil {
				pingOptions = pingOptionsArgument.MustStrings()
			}

			giveaway.StartTime = time.Now()

			if giveaway.EndTime.Unix() > 0 {
				giveaway.EndTime = time.Now().Add(time.Duration(giveaway.EndTime.Unix()) * time.Second)
			}

			giveaway.IsSetup = false

			session, err := welcomer.AcquireSession(ctx, welcomer.GetManagerNameFromContext(ctx))
			if err != nil {
				return nil, err
			}

			message, err := interaction.Channel.Send(ctx, session, welcomer.WebhookMessageParamsToMessageParams(giveawayView(giveaway, 0)))
			if err != nil {
				welcomer.Logger.Error().Err(err).
					Int64("guild_id", int64(*interaction.GuildID)).
					Msg("Failed to send giveaway message")

				return nil, err
			}

			pingMessage := ""

			if len(pingOptions) > 0 {
				switch {
				case slices.Contains(pingOptions, giveawaySetupMenuPingEveryoneKey):
					pingMessage += "@everyone"
				case slices.Contains(pingOptions, giveawaySetupMenuPingHereKey):
					pingMessage += "@here"
				case slices.Contains(pingOptions, giveawaySetupMenuPingRolesAllowedToEnterKey):
					rolesAllowedToEnter := welcomer.UnmarshalRolesListJSON(giveaway.RolesAllowed.Bytes)

					if len(rolesAllowedToEnter) > 0 {
						for _, role := range rolesAllowedToEnter {
							pingMessage += fmt.Sprintf(" <@&%d>", role)
						}
					}
				}
			}

			if additionalRolesArgument, err := subway.GetArgument(ctx, giveawaySetupMenuPingAdditionalRolesKey); err == nil {
				additionalRoles := additionalRolesArgument.MustStrings()

				for _, roleString := range additionalRoles {
					roleSnowflake, err := welcomer.Atoi(roleString)
					if err == nil {
						pingMessage += fmt.Sprintf(" <@&%d>", roleSnowflake)
					}
				}
			}

			if pingMessage != "" {
				_, err = interaction.Channel.Send(ctx, session, discord.MessageParams{
					Content: pingMessage,
				})
				if err != nil {
					welcomer.Logger.Error().Err(err).
						Int64("guild_id", int64(*interaction.GuildID)).
						Msg("Failed to send giveaway ping message")
				}
			}

			_, err = welcomer.Queries.UpdateGiveawayMessage(ctx, database.UpdateGiveawayMessageParams{
				GiveawayUuid: giveawayUUID,
				MessageID:    int64(message.ID),
				ChannelID:    int64(message.ChannelID),
			})
			if err != nil {
				welcomer.Logger.Error().Err(err).
					Int64("guild_id", int64(*interaction.GuildID)).
					Str("giveaway_uuid", giveaway.GiveawayUuid.String()).
					Msg("Failed to update giveaway message and channel")

				return nil, err
			}

			welcomer.PusherGuildScience.Push(
				ctx,
				*interaction.GuildID,
				interaction.GetUser().ID,
				database.ScienceGuildEventTypeGiveawayStarted,
				&welcomer.GuildScienceGiveawayEvents{
					GiveawayUUID: giveaway.GiveawayUuid,
				},
			)
		default:
			welcomer.Logger.Warn().
				Int64("guild_id", int64(*interaction.GuildID)).
				Str("custom_id", interaction.Data.CustomID).
				Msg("Unknown giveaway edit menu option")

			return nil, nil
		}
	default:
		welcomer.Logger.Warn().
			Int64("guild_id", int64(*interaction.GuildID)).
			Int("interaction_type", int(interaction.Type)).
			Msg("Unknown interaction type for giveaway edit menu")

		return nil, nil
	}

	_, err = welcomer.UpdateGiveawayGuildSettingsWithAudit(ctx, database.UpdateGiveawayParams{
		GiveawayUuid:    giveaway.GiveawayUuid,
		IsSetup:         giveaway.IsSetup,
		Title:           giveaway.Title,
		StartTime:       giveaway.StartTime,
		EndTime:         giveaway.EndTime,
		AnnounceWinners: giveaway.AnnounceWinners,
		GiveawayPrizes:  giveaway.GiveawayPrizes,
		RolesAllowed:    giveaway.RolesAllowed,
		RolesExcluded:   giveaway.RolesExcluded,
		MinimumJoinDate: giveaway.MinimumJoinDate,
		Description:     giveaway.Description,
		AccentColour:    giveaway.AccentColour,
		ImageUrl:        giveaway.ImageUrl,
		ShowPrizes:      giveaway.ShowPrizes,
		ShowEntries:     giveaway.ShowEntries,
		AllowEntries:    giveaway.AllowEntries,
		HasEnded:        giveaway.HasEnded,
	}, interaction.GetUser().ID, *interaction.GuildID)
	if err != nil {
		welcomer.Logger.Error().Err(err).
			Int64("guild_id", int64(*interaction.GuildID)).
			Str("giveaway_uuid", giveaway.GiveawayUuid.String()).
			Msg("Failed to update giveaway settings")

		return nil, err
	}

	if giveaway.IsSetup {
		err = discord.CreateInteractionResponse(ctx, sub.EmptySession, interaction.ID, interaction.Token, discord.InteractionResponse{
			Type: welcomer.If(customIDSplit[2] == "", discord.InteractionCallbackTypeChannelMessageSource, discord.InteractionCallbackTypeUpdateMessage),
			Data: welcomer.WebhookMessageParamsToInteractionCallbackData(giveawaySetupView(giveaway), uint32(discord.MessageFlagEphemeral+discord.MessageFlagIsComponentsV2)),
		})
	} else {
		err = discord.CreateInteractionResponse(ctx, sub.EmptySession, interaction.ID, interaction.Token, discord.InteractionResponse{
			Type: discord.InteractionCallbackTypeUpdateMessage,
			Data: &discord.InteractionCallbackData{
				Components: []discord.InteractionComponent{
					{
						Type: discord.InteractionComponentTypeContainer,
						Components: []discord.InteractionComponent{
							{
								Type:    discord.InteractionComponentTypeTextDisplay,
								Content: "Your giveaway has now started!\n\nYou can manage your giveaways settings such as disabling entries, extending the duration or ending the giveaway early by right clicking the giveaway message and selecting \"Manage Giveaway\".\n\n-# How was your experience? Let us know in our feedback channel: https://discord.gg/t2Ye8jBfPh",
							},
							{
								Type: discord.InteractionComponentTypeMediaGallery,
								Items: []discord.InteractionComponentMediaGalleryItem{
									{
										Media: discord.MediaItem{
											URL: "https://welcomer.gg/assets/manage_giveaway.png",
										},
									},
								},
							},
						},
					},
				},
			},
		})
	}

	if err != nil {
		welcomer.Logger.Error().Err(err).
			Int64("guild_id", int64(*interaction.GuildID)).
			Str("giveaway_uuid", giveaway.GiveawayUuid.String()).
			Msg("Failed to edit giveaway message")

		return nil, err
	}

	return &discord.InteractionResponse{
		Type: discord.InteractionCallbackTypeDeferredUpdateMessage,
	}, nil
}

func joinRolesList(roles []discord.Snowflake) string {
	result := ""

	for i, role := range roles {
		result += fmt.Sprintf("<@&%d>", role)

		if i < len(roles)-1 {
			result += ", "
		}
	}

	return result
}

func getGiveawayPrizesAsString(giveawayPrizes []welcomer.GiveawayPrize) string {
	if len(giveawayPrizes) == 0 {
		return "No Prizes Configured"
	}

	result := ""

	for _, prize := range giveawayPrizes {
		result += fmt.Sprintf("**%d** x **%s**\n", prize.Count, prize.Title)
	}

	return result
}

func giveawayView(giveaway *database.GuildGiveaways, entries int32) discord.WebhookMessageParams {
	giveawayPrizes := welcomer.UnmarshalGiveawayPrizeJSON(giveaway.GiveawayPrizes.Bytes)

	containerComponents := []discord.InteractionComponent{
		{
			Type:    discord.InteractionComponentTypeTextDisplay,
			Content: "**" + welcomer.Coalesce(giveaway.Title, "New Giveaway") + "**\n" + giveaway.Description,
		},
	}

	if giveaway.ImageUrl != "" {
		containerComponents = append(containerComponents, discord.InteractionComponent{
			Type: discord.InteractionComponentTypeMediaGallery,
			Items: []discord.InteractionComponentMediaGalleryItem{
				{
					Media: discord.MediaItem{
						URL: giveaway.ImageUrl,
					},
				},
			},
		})
	}

	if giveaway.ShowPrizes {
		containerComponents = append(containerComponents, []discord.InteractionComponent{
			{
				Type: discord.InteractionComponentTypeSeparator,
			},
			{
				Type:    discord.InteractionComponentTypeTextDisplay,
				Content: "**Prizes:**\n" + getGiveawayPrizesAsString(giveawayPrizes),
			},
		}...)
	}

	containerComponents = append(containerComponents, []discord.InteractionComponent{
		{
			Type: discord.InteractionComponentTypeSeparator,
		},
		{
			Type: discord.InteractionComponentTypeTextDisplay,
			Content: "**Giveaway Ends:** " + welcomer.If(giveaway.EndTime.Unix() > 0, "<t:"+welcomer.Itoa(giveaway.EndTime.Unix())+":R> (<t:"+welcomer.Itoa(giveaway.EndTime.Unix())+":f>)", "No end time (runs indefinitely)") +
				"\n" + welcomer.If(giveaway.ShowEntries, fmt.Sprintf("**Entries:** %d", entries), ""),
		},
	}...)

	message := discord.WebhookMessageParams{
		Components: []discord.InteractionComponent{
			{
				Type:    discord.InteractionComponentTypeTextDisplay,
				Content: fmt.Sprintf("-# <@%d> has started a new giveaway!", giveaway.CreatedBy),
			},
			{
				Type:        discord.InteractionComponentTypeContainer,
				AccentColor: new(uint32(welcomer.If(giveaway.AccentColour >= 0, giveaway.AccentColour, welcomer.EmbedColourInfo))),
				Components:  containerComponents,
			},
			{
				Type: discord.InteractionComponentTypeActionRow,
				Components: []discord.InteractionComponent{
					{
						Type:     discord.InteractionComponentTypeButton,
						Style:    discord.InteractionComponentStyleSuccess,
						CustomID: "giveaway_enter:" + giveaway.GiveawayUuid.String(),
						Label:    "Enter Giveaway",
						Disabled: !giveaway.AllowEntries && !giveaway.IsSetup,
						Emoji: &discord.Emoji{
							Name: "🎉",
						},
					},
				},
			},
		},
		Flags: discord.MessageFlagIsComponentsV2,
	}

	return message
}

func giveawayManageView(giveaway *database.GuildGiveaways) discord.WebhookMessageParams {
	customIDPrefix := "giveaway_manage:" + giveaway.GiveawayUuid.String() + ":"

	return discord.WebhookMessageParams{
		Components: []discord.InteractionComponent{
			{
				Type: discord.InteractionComponentTypeContainer,
				Components: []discord.InteractionComponent{
					{
						Type:    discord.InteractionComponentTypeTextDisplay,
						Content: fmt.Sprintf("### Manage entries for giveaway **%s**", welcomer.Coalesce(giveaway.Title, "New Giveaway")),
					},
					{
						Type: discord.InteractionComponentTypeSeparator,
					},
					{
						Type: discord.InteractionComponentTypeSection,
						Components: []discord.InteractionComponent{
							{
								Type: discord.InteractionComponentTypeTextDisplay,
								Content: "**Allow Giveaway Entries**:\n" +
									welcomer.If(giveaway.AllowEntries, "True", "False") +
									welcomer.If(!giveaway.AllowEntries, "\n-# When disabled, users cannot enter the giveaway. This is useful to temporarily pause entries without ending the giveaway.", ""),
							},
						},
						Accessory: &discord.InteractionComponent{
							Type:     discord.InteractionComponentTypeButton,
							Style:    discord.InteractionComponentStyleSecondary,
							Label:    welcomer.If(giveaway.AllowEntries, "Disable", "Enable"),
							CustomID: customIDPrefix + giveawayManageMenuToggleAllowEntriesKey,
							Disabled: giveaway.HasEnded,
						},
					},
					{
						Type: discord.InteractionComponentTypeSeparator,
					},
					{
						Type: discord.InteractionComponentTypeSection,
						Components: []discord.InteractionComponent{
							{
								Type: discord.InteractionComponentTypeTextDisplay,
								Content: "**Giveaway " + welcomer.If(giveaway.HasEnded, "Ended", "Ends") + ":**\n" +
									welcomer.If(giveaway.EndTime.Unix() > 0, "<t:"+welcomer.Itoa(giveaway.EndTime.Unix())+":R> (<t:"+welcomer.Itoa(giveaway.EndTime.Unix())+":f>)", "No end time (runs indefinitely)") + "\n" +
									welcomer.If(
										giveaway.HasEnded,
										"-# This giveaway has already ended, so the duration cannot be extended.",
										"-# Extends the giveaway end time.",
									),
							},
						},
						Accessory: &discord.InteractionComponent{
							Type:     discord.InteractionComponentTypeButton,
							Style:    discord.InteractionComponentStyleSecondary,
							Label:    "Extend",
							CustomID: customIDPrefix + giveawayManageMenuExtendDurationKey,
							Disabled: giveaway.HasEnded,
						},
					},
					{
						Type: discord.InteractionComponentTypeSeparator,
					},
					{
						Type: discord.InteractionComponentTypeSection,
						Components: []discord.InteractionComponent{
							{
								Type:    discord.InteractionComponentTypeTextDisplay,
								Content: "**End Giveaway**",
							},
						},
						Accessory: &discord.InteractionComponent{
							Type:     discord.InteractionComponentTypeButton,
							Style:    discord.InteractionComponentStyleDanger,
							Label:    "End Giveaway",
							CustomID: customIDPrefix + giveawayManageMenuEndGiveawayKey,
							Disabled: giveaway.HasEnded,
						},
					},
					{
						Type: discord.InteractionComponentTypeSeparator,
					},
					{
						Type: discord.InteractionComponentTypeSection,
						Components: []discord.InteractionComponent{
							{
								Type: discord.InteractionComponentTypeTextDisplay,
								Content: "**Export Giveaway Entries**\n" +
									"-# Exports a CSV file of all giveaway entries.",
							},
						},
						Accessory: &discord.InteractionComponent{
							Type:     discord.InteractionComponentTypeButton,
							Style:    discord.InteractionComponentStylePrimary,
							Label:    "Export Entries",
							CustomID: customIDPrefix + giveawayManageMenuExportEntriesKey,
						},
					},
					{
						Type: discord.InteractionComponentTypeSeparator,
					},
					{
						Type: discord.InteractionComponentTypeSection,
						Components: []discord.InteractionComponent{
							{
								Type: discord.InteractionComponentTypeTextDisplay,
								Content: "**Export Giveaway Winners**\n" +
									"-# Exports a CSV file of giveaway winners. Only available after the giveaway has ended.",
							},
						},
						Accessory: &discord.InteractionComponent{
							Type:     discord.InteractionComponentTypeButton,
							Style:    discord.InteractionComponentStylePrimary,
							Label:    "Export Winners",
							CustomID: customIDPrefix + giveawayManageMenuExportWinnersKey,
							Disabled: !giveaway.HasEnded,
						},
					},
					{
						Type: discord.InteractionComponentTypeSeparator,
					},
					// {
					// 	Type: discord.InteractionComponentTypeTextDisplay,
					// 	Content: "**Reroll Giveaway Winners**\n" +
					// 		"-# Want to reroll a giveaway winner? Right click the announced message and select \"Reroll Giveaway Winner\" to select a new winner.",
					// },
				},
			},
		},
	}
}

func giveawaySetupView(giveaway *database.GuildGiveaways) discord.WebhookMessageParams {
	giveawayPrizes := welcomer.UnmarshalGiveawayPrizeJSON(giveaway.GiveawayPrizes.Bytes)
	rolesAllowed := welcomer.UnmarshalRolesListJSON(giveaway.RolesAllowed.Bytes)
	rolesExcluded := welcomer.UnmarshalRolesListJSON(giveaway.RolesExcluded.Bytes)

	customIDPrefix := "giveaway_edit:" + giveaway.GiveawayUuid.String() + ":"

	containerComponents := []discord.InteractionComponent{
		{
			Type:    discord.InteractionComponentTypeTextDisplay,
			Content: "### Create Giveaway",
		},
		{
			Type: discord.InteractionComponentTypeSeparator,
		},
		{
			Type: discord.InteractionComponentTypeSection,
			Components: []discord.InteractionComponent{
				{
					Type:    discord.InteractionComponentTypeTextDisplay,
					Content: "**" + welcomer.Coalesce(giveaway.Title, "New Giveaway") + "**\n" + giveaway.Description,
				},
			},
			Accessory: &discord.InteractionComponent{
				Type:     discord.InteractionComponentTypeButton,
				Style:    discord.InteractionComponentStylePrimary,
				Label:    "Customise Message",
				CustomID: customIDPrefix + giveawaySetupMenuTitleKey,
			},
		},
	}

	if giveaway.ImageUrl != "" {
		containerComponents = append(containerComponents, discord.InteractionComponent{
			Type: discord.InteractionComponentTypeMediaGallery,
			Items: []discord.InteractionComponentMediaGalleryItem{
				{
					Media: discord.MediaItem{
						URL: giveaway.ImageUrl,
					},
				},
			},
		})
	}

	containerComponents = append(containerComponents, []discord.InteractionComponent{
		{
			Type: discord.InteractionComponentTypeSeparator,
		},
		{
			Type: discord.InteractionComponentTypeSection,
			Components: []discord.InteractionComponent{
				{
					Type:    discord.InteractionComponentTypeTextDisplay,
					Content: "**Prizes:**\n" + getGiveawayPrizesAsString(giveawayPrizes),
				},
			},
			Accessory: &discord.InteractionComponent{
				Type:     discord.InteractionComponentTypeButton,
				Style:    discord.InteractionComponentStyleSecondary,
				Label:    "Edit",
				CustomID: customIDPrefix + giveawaySetupMenuPrizesKey,
			},
		},
		{
			Type: discord.InteractionComponentTypeSeparator,
		},
		{
			Type: discord.InteractionComponentTypeSection,
			Components: []discord.InteractionComponent{
				{
					Type: discord.InteractionComponentTypeTextDisplay,
					Content: "**Duration**:\n" + welcomer.If(giveaway.EndTime.Unix() > 0, welcomer.HumanizeDuration(int(giveaway.EndTime.Unix()), true), "No end time (runs indefinitely)") +
						welcomer.If(giveaway.EndTime.IsZero(), "\n-# Giveaway will run until ended manually.", ""),
				},
			},
			Accessory: &discord.InteractionComponent{
				Type:     discord.InteractionComponentTypeButton,
				Style:    discord.InteractionComponentStyleSecondary,
				Label:    "Edit",
				CustomID: customIDPrefix + giveawaySetupMenuDurationKey,
			},
		},
		{
			Type: discord.InteractionComponentTypeSeparator,
		},
		{
			Type: discord.InteractionComponentTypeSection,
			Components: []discord.InteractionComponent{
				{
					Type: discord.InteractionComponentTypeTextDisplay,
					Content: "**Announce Winners**:\n" +
						welcomer.If(giveaway.AnnounceWinners, "True", "False") +
						welcomer.If(!giveaway.AnnounceWinners, "\n-# When disabled, winners will not be announced.", ""),
				},
			},
			Accessory: &discord.InteractionComponent{
				Type:     discord.InteractionComponentTypeButton,
				Style:    discord.InteractionComponentStyleSecondary,
				Label:    welcomer.If(giveaway.AnnounceWinners, "Disable", "Enable"),
				CustomID: customIDPrefix + giveawaySetupMenuAnnounceWinnersKey,
			},
		},
		{
			Type: discord.InteractionComponentTypeSeparator,
		},
		{
			Type: discord.InteractionComponentTypeSection,
			Components: []discord.InteractionComponent{
				{
					Type:    discord.InteractionComponentTypeTextDisplay,
					Content: "**Roles Allowed to Enter**:\n" + welcomer.Coalesce(joinRolesList(rolesAllowed), "All") + "\n\n**Roles Excluded from Entering**:\n" + welcomer.Coalesce(joinRolesList(rolesExcluded), "None"),
				},
			},
			Accessory: &discord.InteractionComponent{
				Type:     discord.InteractionComponentTypeButton,
				Style:    discord.InteractionComponentStyleSecondary,
				Label:    "Edit",
				CustomID: customIDPrefix + giveawaySetupMenuRolesAllowedKey,
			},
		},
		{
			Type: discord.InteractionComponentTypeSeparator,
		},
		{
			Type: discord.InteractionComponentTypeSection,
			Components: []discord.InteractionComponent{
				{
					Type: discord.InteractionComponentTypeTextDisplay,
					Content: "**Minimum Join Date**:\n" + welcomer.Coalesce(welcomer.HumanizeDuration(int(giveaway.MinimumJoinDate.Unix()), true), "None") +
						welcomer.If(!giveaway.MinimumJoinDate.IsZero(), "\n-# Users who have joined the server within "+welcomer.HumanizeDuration(int(giveaway.MinimumJoinDate.Unix()), true)+" of the giveaway starting cannot enter the giveaway.", ""),
				},
			},
			Accessory: &discord.InteractionComponent{
				Type:     discord.InteractionComponentTypeButton,
				Style:    discord.InteractionComponentStyleSecondary,
				Label:    "Edit",
				CustomID: customIDPrefix + giveawaySetupMenuMinimumJoinDateKey,
			},
		},
	}...)

	return discord.WebhookMessageParams{
		Flags: discord.MessageFlagEphemeral + discord.MessageFlagIsComponentsV2,
		Components: []discord.InteractionComponent{
			{
				Type:        discord.InteractionComponentTypeContainer,
				AccentColor: new(uint32(welcomer.If(giveaway.AccentColour >= 0, giveaway.AccentColour, welcomer.EmbedColourInfo))),
				Components:  containerComponents,
			},
			{
				Type: discord.InteractionComponentTypeActionRow,
				Components: []discord.InteractionComponent{
					{
						Type:     discord.InteractionComponentTypeButton,
						Style:    discord.InteractionComponentStyleSuccess,
						Label:    "Start Giveaway",
						CustomID: customIDPrefix + giveawaySetupMenuStartKey,
						Disabled: len(giveawayPrizes) == 0,
					},
					{
						Type:     discord.InteractionComponentTypeButton,
						Style:    discord.InteractionComponentStyleSecondary,
						Label:    "Preview",
						CustomID: customIDPrefix + giveawaySetupMenuPreviewOnKey,
					},
				},
			},
		},
	}
}

func parsePrizesFromString(prizesString string) []welcomer.GiveawayPrize {
	prizes := make([]welcomer.GiveawayPrize, 0)

	lines := strings.Split(strings.TrimSpace(prizesString), "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		var prize string

		var count int

		// split by the first space
		parts := strings.SplitN(line, " ", 2)

		if len(parts) == 2 {
			// Allow for 2x ...
			parts[0] = strings.TrimSuffix(parts[0], "x")

			if newCount, err := welcomer.Atoi(parts[0]); err == nil && newCount > 0 {
				count = int(newCount)
				prize = strings.TrimSpace(parts[1])
			} else {
				prize = strings.TrimSpace(line)
				count = 1
			}
		} else {
			prize = strings.TrimSpace(line)
			count = 1
		}

		if prize == "" || count < 1 {
			continue
		}

		prizes = append(prizes, welcomer.GiveawayPrize{
			Count: count,
			Title: prize,
		})
	}

	return prizes
}

func formatGiveawayPrizesAsString(prizes []welcomer.GiveawayPrize) string {
	lines := make([]string, len(prizes))

	for i, prize := range prizes {
		lines[i] = fmt.Sprintf("%dx %s", prize.Count, prize.Title)
	}

	return strings.Join(lines, "\n")
}
