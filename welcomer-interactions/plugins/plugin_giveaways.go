package plugins

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/WelcomerTeam/Discord/discord"
	subway "github.com/WelcomerTeam/Subway/subway"
	"github.com/WelcomerTeam/Welcomer/welcomer-core"
	"github.com/WelcomerTeam/Welcomer/welcomer-core/database"
	"github.com/gofrs/uuid"
	"github.com/jackc/pgtype"
)

const (
	giveawaySetupMenuTitleKey        = "title"
	giveawaySetupMenuDescriptionKey  = "description"
	giveawaySetupMenuAccentColourKey = "accent_colour"
	giveawaySetupMenuThumbnailURLKey = "thumbnail_url"

	giveawaySetupMenuPrizesKey          = "prizes"
	giveawaySetupMenuDurationKey        = "duration"
	giveawaySetupMenuAnnounceWinnersKey = "announce_winners"
	giveawaySetupMenuRolesAllowedKey    = "roles_allowed"
	giveawaySetupMenuMinimumJoinDateKey = "minimum_join_date"
	giveawaySetupMenuStartKey           = "start"

	giveawaySetupMenuPreviewOnKey  = "preview_on"
	giveawaySetupMenuPreviewOffKey = "preview_off"
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
		Name:        "create",
		Description: "Creates a new giveaway",

		Type: subway.InteractionCommandableTypeSubcommand,

		ArgumentParameter: []subway.ArgumentParameter{
			{
				Name:         "title",
				Description:  "The title of the giveaway.",
				ArgumentType: subway.ArgumentTypeString,
				Required:     false,
			},
		},

		DMPermission:            new(false),
		DefaultMemberPermission: new(discord.Int64(welcomer.PermissionElevated)),

		Handler: func(ctx context.Context, sub *subway.Subway, interaction discord.Interaction) (*discord.InteractionResponse, error) {
			title := subway.MustGetArgument(ctx, "title").MustString()

			var giveaway *database.GuildGiveaways
			var err error

			err = welcomer.RetryWithFallback(
				func() error {
					giveaway, err = welcomer.Queries.CreateGiveaway(ctx, database.CreateGiveawayParams{
						GuildID:   int64(*interaction.GuildID),
						CreatedBy: int64(interaction.GetUser().ID),
						Title:     title,
						EndTime:   time.Time{},
					})

					return err
				},
				func() error {
					return welcomer.EnsureGuild(ctx, discord.Snowflake(*interaction.GuildID))
				},
				nil,
			)
			if err != nil {
				welcomer.Logger.Error().Err(err).
					Int64("guild_id", int64(*interaction.GuildID)).
					Str("title", title).
					Msg("Failed to create giveaway settings")

				return nil, err
			}

			return &discord.InteractionResponse{
				Type: discord.InteractionCallbackTypeModal,
				Data: &discord.InteractionCallbackData{
					Title:    "Create Giveaway",
					CustomID: "giveaway_edit:" + giveaway.GiveawayUuid.String(),
					Components: []discord.InteractionComponent{
						{
							Type:        discord.InteractionComponentTypeLabel,
							Label:       "Prizes",
							Description: "one prize per line, with optional count, e.g. 2x Discord Nitro",
							Component: &discord.InteractionComponent{
								CustomID: giveawaySetupMenuPrizesKey,
								Type:     discord.InteractionComponentTypeTextInput,
								Style:    discord.InteractionComponentStyleParagraph,
							},
						},
						{
							Type:        discord.InteractionComponentTypeLabel,
							Label:       "Duration",
							Description: "e.g., 1h, 30m, 2d. Only years, days, hours and minutes are supported.",
							Component: &discord.InteractionComponent{
								CustomID:    giveawaySetupMenuDurationKey,
								Type:        discord.InteractionComponentTypeTextInput,
								Placeholder: "7d 3h 60m",
								Style:       discord.InteractionComponentStyleShort,
								Required:    new(false),
							},
						},
						{
							Type:        discord.InteractionComponentTypeLabel,
							Label:       "Roles Allowed to Enter",
							Description: "Users must have at least one of these roles to enter. Ignored if empty.",
							Component: &discord.InteractionComponent{
								CustomID:  "allowed",
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
								CustomID:  "excluded",
								Type:      discord.InteractionComponentTypeRoleSelect,
								Required:  new(false),
								MaxValues: new(int32(25)),
							},
						},
						{
							Type:        discord.InteractionComponentTypeLabel,
							Label:       "Minimum Join Date",
							Description: "Users joined within the duration specified cannot enter. Ignored if empty. e.g., 1h, 30m, 2d. ",
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
			}, nil
		},
	})

	sub.RegisterComponentListener("giveaway_edit:*", handleGiveawayEditComponent)
	sub.RegisterComponentListener("giveaway_enter:*", handleGiveawayEnterComponent)

	// close - stops more entries being added
	// open  - allows more entries to be added

	// extend - extends the duration of the giveaway by a specified amount of time
	// end    - ends the giveaway immediately

	// export - exports the giveaway results

	// TODO: science events
	// giveawaycreate
	// giveawaystart
	// giveawayended

	cog.InteractionCommands.MustAddInteractionCommand(giveawaysGroup)

	return nil
}

func handleGiveawayEnterComponent(ctx context.Context, sub *subway.Subway, interaction discord.Interaction) (*discord.InteractionResponse, error) {
	return nil, nil
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

	if customIDSplit[0] != "giveaway_edit" {
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
		case giveawaySetupMenuTitleKey:
			return &discord.InteractionResponse{
				Data: &discord.InteractionCallbackData{
					Title:    "Customise Giveaway Display",
					CustomID: interaction.Data.CustomID,
					Components: []discord.InteractionComponent{
						{
							Type:  discord.InteractionComponentTypeLabel,
							Label: "Title",
							Component: &discord.InteractionComponent{
								CustomID:    giveawaySetupMenuTitleKey,
								Type:        discord.InteractionComponentTypeTextInput,
								Placeholder: giveaway.Title,
								Value:       giveaway.Title,
								Style:       discord.InteractionComponentStyleShort,
								Required:    new(false),
							},
						},
						{
							Type:  discord.InteractionComponentTypeLabel,
							Label: "Description",
							Component: &discord.InteractionComponent{
								CustomID:    giveawaySetupMenuDescriptionKey,
								Type:        discord.InteractionComponentTypeTextInput,
								Placeholder: giveaway.Description,
								Value:       giveaway.Description,
								Style:       discord.InteractionComponentStyleParagraph,
								Required:    new(false),
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
							Description: "If specified, this image will show at the bottom of your giveaway message.",
							Component: &discord.InteractionComponent{
								CustomID:    giveawaySetupMenuThumbnailURLKey,
								Type:        discord.InteractionComponentTypeTextInput,
								Placeholder: "https://example.com/image.png",
								Value:       giveaway.ImageUrl,
								Style:       discord.InteractionComponentStyleShort,
								Required:    new(false),
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
							Description: "one prize per line, with optional count, e.g. 2x Discord Nitro",
							Component: &discord.InteractionComponent{
								CustomID:    giveawaySetupMenuPrizesKey,
								Type:        discord.InteractionComponentTypeTextInput,
								Placeholder: formatGiveawayPrizesAsString(welcomer.UnmarshalGiveawayPrizeJSON(giveaway.GiveawayPrizes.Bytes)),
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
							Description: "e.g., 1h, 30m, 2d. Only years, days, hours and minutes are supported.",
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
								CustomID:  "allowed",
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
								CustomID:  "excluded",
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
							Description: "Users joined within the duration specified cannot enter. Ignored if empty. e.g., 1h, 30m, 2d. ",
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
			// TODO

			return &discord.InteractionResponse{
				Data: &discord.InteractionCallbackData{},
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
		if customIDSplit[2] == giveawaySetupMenuTitleKey || customIDSplit[2] == "" {
			if titleString, err := subway.GetArgument(ctx, giveawaySetupMenuTitleKey); err == nil {
				giveaway.Title = titleString.MustString()
			} else {
				giveaway.Title = ""
			}

			if descriptionString, err := subway.GetArgument(ctx, giveawaySetupMenuDescriptionKey); err == nil {
				giveaway.Description = descriptionString.MustString()
			} else {
				giveaway.Description = ""
			}

			if accentColourString, err := subway.GetArgument(ctx, giveawaySetupMenuAccentColourKey); err == nil {
				rgba, err := welcomer.ParseColour(accentColourString.MustString(), "#000000")
				if err != nil {
					welcomer.Logger.Error().Err(err).
						Int64("guild_id", int64(*interaction.GuildID)).
						Str("accent_colour", accentColourString.MustString()).
						Msg("Failed to parse accent colour")

					giveaway.AccentColour = -1
				} else {
					giveaway.AccentColour = int64(int32(rgba.R)<<16 + int32(rgba.G)<<8 + int32(rgba.B))
				}
			} else {
				giveaway.AccentColour = -1
			}

			if thumbnailURLString, err := subway.GetArgument(ctx, giveawaySetupMenuThumbnailURLKey); err == nil {
				giveaway.ImageUrl = thumbnailURLString.MustString()

				if _, ok := welcomer.IsValidURL(thumbnailURLString.MustString()); !ok {
					welcomer.Logger.Error().Err(err).
						Int64("guild_id", int64(*interaction.GuildID)).
						Str("thumbnail_url", thumbnailURLString.MustString()).
						Msg("Failed to parse thumbnail URL")

					giveaway.ImageUrl = ""
				}
			} else {
				giveaway.ImageUrl = ""
			}
		}

		if customIDSplit[2] == giveawaySetupMenuPrizesKey || customIDSplit[2] == "" {
			if prizesString, err := subway.GetArgument(ctx, giveawaySetupMenuPrizesKey); err == nil {
				prizes := parsePrizesFromString(prizesString.MustString())
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
		}

		if customIDSplit[2] == giveawaySetupMenuDurationKey || customIDSplit[2] == "" {
			if durationString, err := subway.GetArgument(ctx, giveawaySetupMenuDurationKey); err == nil {
				seconds, err := welcomer.ParseDurationAsSeconds(durationString.MustString())
				if err != nil || seconds < 0 {
					welcomer.Logger.Error().Err(err).
						Int64("guild_id", int64(*interaction.GuildID)).
						Str("duration", durationString.MustString()).
						Msg("Failed to parse duration")

					return nil, nil
				}

				giveaway.EndTime = time.Unix(int64(seconds), 0)
			} else {
				giveaway.EndTime = time.Time{}
			}
		}

		if customIDSplit[2] == giveawaySetupMenuRolesAllowedKey || customIDSplit[2] == "" {
			if allowedRoles, err := subway.GetArgument(ctx, "allowed"); err == nil {
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

			if excludedRoles, err := subway.GetArgument(ctx, "excluded"); err == nil {
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
		}

		if customIDSplit[2] == giveawaySetupMenuMinimumJoinDateKey || customIDSplit[2] == "" {
			if durationString, err := subway.GetArgument(ctx, giveawaySetupMenuMinimumJoinDateKey); err == nil {
				seconds, err := welcomer.ParseDurationAsSeconds(durationString.MustString())
				if err != nil || seconds < 0 {
					welcomer.Logger.Error().Err(err).
						Int64("guild_id", int64(*interaction.GuildID)).
						Str("duration", durationString.MustString()).
						Msg("Failed to parse duration")

					return nil, nil
				}

				giveaway.MinimumJoinDate = time.Unix(int64(seconds), 0)
			} else {
				giveaway.MinimumJoinDate = time.Time{}
			}
		}

		if customIDSplit[2] == giveawaySetupMenuStartKey {
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

			err = discord.DeleteOriginalInteractionResponse(ctx, session, interaction.ApplicationID, interaction.Token)
			if err != nil {
				welcomer.Logger.Error().Err(err).
					Int64("guild_id", int64(*interaction.GuildID)).
					Msg("Failed to delete original interaction response")
			}

			println(message.ID)
			// TODO
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
	}, interaction.GetUser().ID, *interaction.GuildID)
	if err != nil {
		welcomer.Logger.Error().Err(err).
			Int64("guild_id", int64(*interaction.GuildID)).
			Str("giveaway_uuid", giveaway.GiveawayUuid.String()).
			Msg("Failed to update giveaway settings")

		return nil, err
	}

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
	result := "**Prizes:**\n"

	if len(giveawayPrizes) == 0 {
		result += "No Prizes Configured"

		return result
	}

	for i, prize := range giveawayPrizes {
		result += fmt.Sprintf("**%d** x **%s**", prize.Count, prize.Title)

		if i < len(giveawayPrizes)-1 {
			result += "\n"
		}
	}

	return result
}

func giveawayView(giveaway *database.GuildGiveaways, entries int) discord.WebhookMessageParams {
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

	containerComponents = append(containerComponents, []discord.InteractionComponent{
		{
			Type: discord.InteractionComponentTypeSeparator,
		},
		{
			Type:    discord.InteractionComponentTypeTextDisplay,
			Content: getGiveawayPrizesAsString(giveawayPrizes),
		},
		{
			Type: discord.InteractionComponentTypeSeparator,
		},
		{
			Type: discord.InteractionComponentTypeTextDisplay,
			Content: "**Giveaway Ends:** " + welcomer.If(giveaway.EndTime.Unix() > 0, "<t:"+welcomer.Itoa(giveaway.EndTime.Unix())+":R> (<t:"+welcomer.Itoa(giveaway.EndTime.Unix())+":f>)", "Never") +
				"\n" + fmt.Sprintf("**Entries:** %d", entries),
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

func giveawaySetupView(giveaway *database.GuildGiveaways) discord.WebhookMessageParams {
	giveawayPrizes := welcomer.UnmarshalGiveawayPrizeJSON(giveaway.GiveawayPrizes.Bytes)
	rolesAllowed := welcomer.UnmarshalRolesListJSON(giveaway.RolesAllowed.Bytes)
	rolesExcluded := welcomer.UnmarshalRolesListJSON(giveaway.RolesExcluded.Bytes)

	customIDPrefix := "giveaway_edit:" + giveaway.GiveawayUuid.String() + ":"

	message := discord.WebhookMessageParams{
		Flags: discord.MessageFlagEphemeral + discord.MessageFlagIsComponentsV2,
		Components: []discord.InteractionComponent{
			{
				Type:        discord.InteractionComponentTypeContainer,
				AccentColor: new(uint32(welcomer.If(giveaway.AccentColour >= 0, giveaway.AccentColour, welcomer.EmbedColourInfo))),
				Components: []discord.InteractionComponent{
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
								Content: "**" + welcomer.Coalesce(giveaway.Title, "Unnamed Giveaway") + "**\n" + giveaway.Description,
							},
						},
						Accessory: &discord.InteractionComponent{
							Type:     discord.InteractionComponentTypeButton,
							Style:    discord.InteractionComponentStyleSecondary,
							Label:    "Customise",
							CustomID: customIDPrefix + giveawaySetupMenuTitleKey,
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
								Content: getGiveawayPrizesAsString(giveawayPrizes),
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
								Type:    discord.InteractionComponentTypeTextDisplay,
								Content: "**Duration**:\n" + welcomer.Coalesce(welcomer.HumanizeDuration(int(giveaway.EndTime.Unix()), true), "Forever"),
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
								Type:    discord.InteractionComponentTypeTextDisplay,
								Content: "**Announce Winners**:\n" + welcomer.If(giveaway.AnnounceWinners, "True", "False"),
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
								Content: "**Roles Allowed to Enter**:\n" + welcomer.Coalesce(joinRolesList(rolesAllowed), "None") + "\n\n**Roles Excluded from Entering**:\n" + welcomer.Coalesce(joinRolesList(rolesExcluded), "None"),
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
								Type:    discord.InteractionComponentTypeTextDisplay,
								Content: "**Minimum Join Date**:\n" + welcomer.Coalesce(welcomer.HumanizeDuration(int(giveaway.MinimumJoinDate.Unix()), true), "None"),
							},
						},
						Accessory: &discord.InteractionComponent{
							Type:     discord.InteractionComponentTypeButton,
							Style:    discord.InteractionComponentStyleSecondary,
							Label:    "Edit",
							CustomID: customIDPrefix + giveawaySetupMenuMinimumJoinDateKey,
						},
					},
				},
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

	if giveaway.ImageUrl != "" {
		message.Components[0].Components = append(message.Components[0].Components, discord.InteractionComponent{
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

	return message
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
