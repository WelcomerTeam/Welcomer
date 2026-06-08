package plugins

import (
	"bytes"
	"context"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"math/rand/v2"
	"slices"
	"strconv"
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
	pollSetupMenuTitleKey        = "title"
	pollSetupMenuDescriptionKey  = "description"
	pollSetupMenuAccentColourKey = "accent_colour"
	pollSetupMenuThumbnailURLKey = "thumbnail_url"

	pollSetupMenuAnswersKey               = "answers"
	pollSetupMenuOptionsKey               = "options"
	pollSetupMenuDurationKey              = "duration"
	pollSetupMenuToggleAnonymousVotingKey = "toggle_anonymous_voting"
	pollSetupMenuMaximumAnswersKey        = "maximum_answers"

	// stub option during initial modal to set it to the maximum answers.
	pollSetupMenuAllowMultipleAnswersKey = "allow_multiple_answers"

	pollSetupMenuRolesAllowedKey         = "roles_allowed"
	pollSetupMenuRolesAllowedIncludedKey = "roles_allowed_included"
	pollSetupMenuRolesAllowedExcludedKey = "roles_allowed_excluded"

	pollSetupMenuManageResubmissionsKey = "manage_resubmissions"
	pollSetupMenuShowResultsKey         = "show_results"

	pollSetupMenuMinimumJoinDateKey = "minimum_join_date"
	pollSetupMenuStartKey           = "start"

	pollSetupMenuPreviewOnKey  = "preview_on"
	pollSetupMenuPreviewOffKey = "preview_off"

	pollSetupMenuPingKey                    = "ping"
	pollSetupMenuPingEveryoneKey            = "ping_everyone"
	pollSetupMenuPingHereKey                = "ping_here"
	pollSetupMenuPingRolesAllowedToEnterKey = "ping_roles_allowed_to_enter"
	pollSetupMenuPingAdditionalRolesKey     = "additional_roles_to_ping"

	pollVoteSelectionKey = "poll_vote_selection"

	pollManageMenuToggleAllowEntriesKey = "toggle_allow_entries"
	pollManageMenuExtendDurationKey     = "extend_duration"
	pollManageMenuEndPollKey            = "end_poll"
	pollManageMenuExportEntriesKey      = "export_entries"

	pollMessageUpdateRate = 1 * time.Second
)

func NewPollsCog() *PollsCog {
	return &PollsCog{
		InteractionCommands: subway.SetupInteractionCommandable(&subway.InteractionCommandable{}),
	}
}

type PollsCog struct {
	InteractionCommands *subway.InteractionCommandable
}

// Assert types.

var (
	_ subway.Cog                        = (*PollsCog)(nil)
	_ subway.CogWithInteractionCommands = (*PollsCog)(nil)
)

func (cog *PollsCog) CogInfo() *subway.CogInfo {
	return &subway.CogInfo{
		Name:        "Polls",
		Description: "Provides the functionality for the 'Polls' feature",
	}
}

func (cog *PollsCog) GetInteractionCommandable() *subway.InteractionCommandable {
	return cog.InteractionCommands
}

func (cog *PollsCog) RegisterCog(sub *subway.Subway) error {
	if welcomer.GetEnvironmentType() == welcomer.EnvironmentTypeDevelopment {
		sectionEmojiIDs = [][]string{
			{
				"1499902373110354072",
				"1499902374259462145",
				"1499902371986280488",
				"1499902375455096932",
			},
			{
				"1499902377593929789",
				"1499902376364998697",
			},
			{
				"1499902365438967979",
				"1499902367028744372",
				"1499902364398653511",
				"1499902370358755500",
			},
		}
	} else {
		sectionEmojiIDs = [][]string{
			{
				"1499921790414356654",
				"1499921791483904041",
				"1499921784995188836",
				"1499921792805113886",
			},
			{
				"1499921795954901032",
				"1499921794994278601",
			},
			{
				"1499921786035376268",
				"1499921787130216519",
				"1499921789067989062",
				"1499921788048638014",
			},
		}
	}

	pollsGroup := subway.NewSubcommandGroup(
		"polls",
		"Poll commands",
	)

	pollsGroup.MustAddInteractionCommand(&subway.InteractionCommandable{
		Name:        "create",
		Description: "Creates a new poll",

		Type: subway.InteractionCommandableTypeSubcommand,

		DMPermission:            new(false),
		DefaultMemberPermission: new(discord.Int64(welcomer.PermissionElevated)),

		Handler: func(ctx context.Context, sub *subway.Subway, interaction discord.Interaction) (*discord.InteractionResponse, error) {
			return welcomer.RequireGuildElevation(sub, interaction, func() (*discord.InteractionResponse, error) {
				var poll *database.GuildPolls
				var err error

				err = welcomer.RetryWithFallback(
					func() error {
						poll, err = welcomer.Queries.CreatePoll(ctx, database.CreatePollParams{
							GuildID:           int64(*interaction.GuildID),
							CreatedBy:         int64(interaction.GetUser().ID),
							EndTime:           time.Time{},
							Resubmissions:     welcomer.PollResubmissionOptionAlways.String(),
							ResultsVisibility: welcomer.PollResultVisibilityOptionAlways.String(),
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
						Msg("Failed to create poll")

					return nil, err
				}

				welcomer.PusherGuildScience.Push(
					ctx,
					*interaction.GuildID,
					interaction.GetUser().ID,
					database.ScienceGuildEventTypePollCreated,
					&welcomer.GuildSciencePollEvents{
						PollUUID: poll.PollUuid,
					},
				)

				return &discord.InteractionResponse{
					Type: discord.InteractionCallbackTypeModal,
					Data: &discord.InteractionCallbackData{
						Title:    "Create Poll",
						CustomID: "poll_edit:" + poll.PollUuid.String(),
						Components: []discord.InteractionComponent{
							{
								Type:  discord.InteractionComponentTypeLabel,
								Label: "Question",
								Component: &discord.InteractionComponent{
									CustomID: pollSetupMenuTitleKey,
									Type:     discord.InteractionComponentTypeTextInput,
									Value:    welcomer.StringToJsonLiteral(poll.Title),
									Style:    discord.InteractionComponentStyleShort,
									Required: new(false),
								},
							},
							{
								Type:        discord.InteractionComponentTypeLabel,
								Label:       "Answers",
								Description: "One answer per line. Max of 10 answers is allowed.",
								Component: &discord.InteractionComponent{
									CustomID:    pollSetupMenuAnswersKey,
									Type:        discord.InteractionComponentTypeTextInput,
									Style:       discord.InteractionComponentStyleParagraph,
									Placeholder: "Answer 1\nAnswer 2\nAnswer 3",
								},
							},
							{
								Type:        discord.InteractionComponentTypeLabel,
								Label:       "Duration",
								Description: "e.g. 1h, 30m, 2d. Only years, days, hours and minutes are supported.",
								Component: &discord.InteractionComponent{
									CustomID:    pollSetupMenuDurationKey,
									Type:        discord.InteractionComponentTypeTextInput,
									Placeholder: "7d 3h 60m",
									Style:       discord.InteractionComponentStyleShort,
									Required:    new(false),
								},
							},
							{
								Type:  discord.InteractionComponentTypeLabel,
								Label: "Poll Options",
								Component: &discord.InteractionComponent{
									CustomID: pollSetupMenuOptionsKey,
									Type:     discord.InteractionComponentTypeCheckboxGroup,
									Options: []discord.ApplicationSelectOption{
										{
											Label: "Allow Multiple Answers",
											Value: pollSetupMenuAllowMultipleAnswersKey,
										},
										{
											Label:       "Anonymous Poll",
											Description: "Resubmissions are not allowed and results will only be available when the poll ends.",
											Value:       pollSetupMenuToggleAnonymousVotingKey,
										},
									},
									Required: new(false),
								},
							},
						},
					},
				}, nil
			})
		},
	})

	cog.InteractionCommands.MustAddInteractionCommand(&subway.InteractionCommandable{
		Name: "Manage Poll",

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
						Msg("Failed to find message for poll manage command")

					return nil, errors.New("failed to find message for poll manage command")
				}

				poll, err := welcomer.Queries.GetPollFromMessageID(ctx, database.GetPollFromMessageIDParams{
					GuildID:   int64(*interaction.GuildID),
					ChannelID: int64(message.ChannelID),
					MessageID: int64(message.ID),
				})
				if err != nil && !errors.Is(err, pgx.ErrNoRows) {
					welcomer.Logger.Error().Err(err).
						Int64("guild_id", int64(*interaction.GuildID)).
						Int64("channel_id", int64(message.ChannelID)).
						Int64("message_id", int64(message.ID)).
						Msg("Failed to get poll settings from message ID")

					return nil, err
				} else if errors.Is(err, pgx.ErrNoRows) {
					welcomer.Logger.Warn().
						Int64("guild_id", int64(*interaction.GuildID)).
						Int64("channel_id", int64(message.ChannelID)).
						Int64("message_id", int64(message.ID)).
						Msg("Poll not found for poll settings message")

					return &discord.InteractionResponse{
						Type: discord.InteractionCallbackTypeChannelMessageSource,
						Data: &discord.InteractionCallbackData{
							Embeds: welcomer.NewEmbed("This message is not associated with a poll. Please make sure you are using this command on the poll message.", welcomer.EmbedColourError),
							Flags:  uint32(discord.MessageFlagEphemeral),
						},
					}, nil
				}

				return &discord.InteractionResponse{
					Type: discord.InteractionCallbackTypeChannelMessageSource,
					Data: welcomer.WebhookMessageParamsToInteractionCallbackData(pollManageView(poll), uint32(discord.MessageFlagEphemeral+discord.MessageFlagIsComponentsV2)),
				}, nil
			})
		},
	})

	sub.RegisterComponentListener("poll_edit:*", handlePollEditComponent)
	sub.RegisterComponentListener("poll_enter:*", handlePollVoteComponent)
	sub.RegisterComponentListener("poll_manage:*", handlePollManageComponent)

	cog.InteractionCommands.MustAddInteractionCommand(pollsGroup)

	return nil
}

func handlePollManageComponent(ctx context.Context, sub *subway.Subway, interaction discord.Interaction) (*discord.InteractionResponse, error) {
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

	pollUUID, err := uuid.FromString(customIDSplit[1])
	if err != nil {
		return nil, err
	}

	poll, err := welcomer.Queries.GetPoll(ctx, database.GetPollParams{
		GuildID:  int64(*interaction.GuildID),
		PollUuid: pollUUID,
	})
	if err != nil {
		welcomer.Logger.Error().Err(err).
			Int64("guild_id", int64(*interaction.GuildID)).
			Str("poll_uuid", pollUUID.String()).
			Msg("Failed to get poll settings")

		return nil, err
	}

	switch interaction.Type {
	case discord.InteractionTypeMessageComponent:
		switch customIDSplit[2] {
		case pollManageMenuToggleAllowEntriesKey:
			poll.AllowEntries = !poll.AllowEntries
		case pollManageMenuExtendDurationKey:
			return &discord.InteractionResponse{
				Type: discord.InteractionCallbackTypeModal,
				Data: &discord.InteractionCallbackData{
					Title:    "Extend Poll Duration",
					CustomID: interaction.Data.CustomID,
					Components: []discord.InteractionComponent{
						{
							Type: discord.InteractionComponentTypeTextDisplay,
							Content: "Enter the new poll duration from the current time. Leave empty if you want the poll to run indefinitely.\n\nIf you would like to remove time from the current duration, put a '-' before the duration." +
								welcomer.If(poll.EndTime.IsZero(), "\n\nThis poll is currently set to run indefinitely so this will be the duration from the current time.", ""),
						},
						{
							Type:        discord.InteractionComponentTypeLabel,
							Label:       "Duration",
							Description: "e.g. 1h, 30m, 2d, -5m. Only years, days, hours and minutes are supported.",
							Component: &discord.InteractionComponent{
								CustomID:    pollManageMenuExtendDurationKey,
								Type:        discord.InteractionComponentTypeTextInput,
								Placeholder: "7d 3h 60m -2d",
								Style:       discord.InteractionComponentStyleShort,
								Required:    new(false),
							},
						},
					},
				},
			}, nil
		case pollManageMenuEndPollKey:
			return &discord.InteractionResponse{
				Type: discord.InteractionCallbackTypeModal,
				Data: &discord.InteractionCallbackData{
					Title:    "End Poll",
					CustomID: interaction.Data.CustomID,
					Components: []discord.InteractionComponent{
						{
							Type:    discord.InteractionComponentTypeTextDisplay,
							Content: "Are you sure you want to end the poll early? This cannot be undone.",
						},
					},
				},
			}, nil
		case pollManageMenuExportEntriesKey:
			return exportPollEntries(ctx, sub, interaction, poll)
		default:
			welcomer.Logger.Warn().
				Int64("guild_id", int64(*interaction.GuildID)).
				Str("poll_uuid", pollUUID.String()).
				Str("custom_id", interaction.Data.CustomID).
				Msg("Unknown poll manage component interaction")
		}
	case discord.InteractionTypeModalSubmit:
		switch customIDSplit[2] {
		case pollManageMenuExtendDurationKey:
			durationArgument, err := subway.GetArgument(ctx, pollManageMenuExtendDurationKey)

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
				if poll.EndTime.IsZero() {
					poll.EndTime = time.Now()
				}

				if hasMinus {
					poll.EndTime = poll.EndTime.Add(-time.Duration(seconds) * time.Second)
				} else {
					poll.EndTime = poll.EndTime.Add(time.Duration(seconds) * time.Second)
				}

				if poll.EndTime.Before(time.Now()) {
					return &discord.InteractionResponse{
						Type: discord.InteractionCallbackTypeChannelMessageSource,
						Data: &discord.InteractionCallbackData{
							Embeds: welcomer.NewEmbed("The new end time cannot be in the past. Please end the poll if you want to do this.", welcomer.EmbedColourError),
							Flags:  uint32(discord.MessageFlagEphemeral),
						},
					}, nil
				}
			} else {
				// If no duration is passed, make the duration indefinite.
				poll.EndTime = time.Time{}
			}
		case pollManageMenuEndPollKey:
			poll.EndTime = time.Now()

			data, _ := json.Marshal(welcomer.CustomEventInvokeEndPollStructure{
				PollUUID: poll.PollUuid,
				GuildID:  *interaction.GuildID,
			})

			_, err = sub.SandwichClient.RelayMessage(ctx, &sandwich.RelayMessageRequest{
				Identifier: welcomer.GetManagerNameFromContext(ctx),
				Type:       welcomer.CustomEventInvokeEndPoll,
				Data:       data,
			})
			if err != nil {
				return nil, err
			}
		default:
			welcomer.Logger.Warn().
				Int64("guild_id", int64(*interaction.GuildID)).
				Str("poll_uuid", pollUUID.String()).
				Str("custom_id", interaction.Data.CustomID).
				Msg("Unknown poll manage modal submit interaction")
		}
	}

	_, err = welcomer.UpdatePollGuildSettingsWithAudit(ctx, database.UpdatePollParams{
		PollUuid:          poll.PollUuid,
		IsSetup:           poll.IsSetup,
		HasEnded:          poll.HasEnded,
		Title:             poll.Title,
		Description:       poll.Description,
		AccentColour:      poll.AccentColour,
		ImageUrl:          poll.ImageUrl,
		StartTime:         poll.StartTime,
		EndTime:           poll.EndTime,
		PollOptions:       poll.PollOptions,
		IsAnonymous:       poll.IsAnonymous,
		MaximumSelections: poll.MaximumSelections,
		AllowEntries:      poll.AllowEntries,
		Resubmissions:     poll.Resubmissions,
		ResultsVisibility: poll.ResultsVisibility,
		RolesAllowed:      poll.RolesAllowed,
		RolesExcluded:     poll.RolesExcluded,
		MinimumJoinDate:   poll.MinimumJoinDate,
	}, interaction.GetUser().ID, *interaction.GuildID)
	if err != nil {
		welcomer.Logger.Error().Err(err).
			Int64("guild_id", int64(*interaction.GuildID)).
			Str("poll_uuid", poll.PollUuid.String()).
			Msg("Failed to update poll settings")

		return nil, err
	}

	if customIDSplit[2] == pollManageMenuEndPollKey {
		poll.HasEnded = true
	}

	err = discord.CreateInteractionResponse(ctx, sub.EmptySession, interaction.ID, interaction.Token, discord.InteractionResponse{
		Type: discord.InteractionCallbackTypeUpdateMessage,
		Data: welcomer.WebhookMessageParamsToInteractionCallbackData(pollManageView(poll), uint32(discord.MessageFlagEphemeral+discord.MessageFlagIsComponentsV2)),
	})
	if err != nil {
		welcomer.Logger.Error().Err(err).
			Int64("guild_id", int64(*interaction.GuildID)).
			Str("poll_uuid", pollUUID.String()).
			Str("custom_id", interaction.Data.CustomID).
			Msg("Failed to create interaction response for poll manage component")
	}

	return nil, nil
}

func exportPollEntries(ctx context.Context, sub *subway.Subway, interaction discord.Interaction, poll *database.GuildPolls) (*discord.InteractionResponse, error) {
	pollAnswers := welcomer.UnmarshalAnswersListJSON(poll.PollOptions.Bytes)

	entries, err := welcomer.Queries.GetPollEntries(ctx, poll.PollUuid)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		welcomer.Logger.Error().Err(err).
			Str("poll_uuid", poll.PollUuid.String()).
			Msg("Failed to get poll entry users")

		return nil, err
	}

	var file bytes.Buffer

	writer := csv.NewWriter(&file)

	totalEntries := map[int32]int{}

	for i := range len(pollAnswers) {
		totalEntries[int32(i)] = 0
	}

	for _, entry := range entries {
		totalEntries[entry.OptionIndex]++
	}

	if poll.IsAnonymous {
		writer.Write([]string{"option", "count"})

		for entry_index, entry_count := range totalEntries {
			writer.Write([]string{
				pollAnswers[entry_index],
				welcomer.Itoa(int64(entry_count)),
			})
		}
	} else {
		writer.Write([]string{"user_id", "option", "entered_at"})

		for _, entry := range entries {
			writer.Write([]string{
				welcomer.Itoa(entry.UserID),
				pollAnswers[entry.OptionIndex],
				entry.CreatedAt.Format("2006-01-02 15:04:05"),
			})
		}
	}

	writer.Flush()

	if err := writer.Error(); err != nil {
		welcomer.Logger.Error().Err(err).
			Str("poll_uuid", poll.PollUuid.String()).
			Msg("Failed to write poll entries to csv")

		return nil, err
	}

	err = interaction.SendResponse(ctx, sub.EmptySession, discord.InteractionCallbackTypeChannelMessageSource, &discord.InteractionCallbackData{
		Content: "Here are the entries for this poll:",
		Files: []discord.File{
			{
				Reader:      &file,
				Name:        fmt.Sprintf("poll_entries_%s.csv", poll.PollUuid.String()),
				ContentType: "text/csv",
			},
		},
		Flags: uint32(discord.MessageFlagEphemeral),
	})
	if err != nil {
		welcomer.Logger.Error().Err(err).
			Str("poll_uuid", poll.PollUuid.String()).
			Msg("Failed to send poll entries response")

		return nil, err
	}

	return nil, nil
}

func handlePollVoteComponent(ctx context.Context, sub *subway.Subway, interaction discord.Interaction) (*discord.InteractionResponse, error) {
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

	pollUUID, err := uuid.FromString(customIDSplit[1])
	if err != nil {
		return nil, err
	}

	poll, err := welcomer.Queries.GetPoll(ctx, database.GetPollParams{
		GuildID:  int64(*interaction.GuildID),
		PollUuid: pollUUID,
	})
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		welcomer.Logger.Error().Err(err).
			Int64("guild_id", int64(*interaction.GuildID)).
			Str("poll_uuid", pollUUID.String()).
			Msg("Failed to get poll settings")

		return nil, err
	} else if errors.Is(err, pgx.ErrNoRows) {
		welcomer.Logger.Warn().
			Int64("guild_id", int64(*interaction.GuildID)).
			Str("poll_uuid", pollUUID.String()).
			Msg("Poll not found for poll entry")

		return &discord.InteractionResponse{
			Type: discord.InteractionCallbackTypeChannelMessageSource,
			Data: &discord.InteractionCallbackData{
				Embeds: welcomer.NewEmbed("This poll no longer exists. It may have been deleted or ended.", welcomer.EmbedColourError),
				Flags:  uint32(discord.MessageFlagEphemeral),
			},
		}, nil
	}

	if poll.HasEnded {
		return &discord.InteractionResponse{
			Type: discord.InteractionCallbackTypeChannelMessageSource,
			Data: &discord.InteractionCallbackData{
				Embeds: welcomer.NewEmbed("This poll has already ended.", welcomer.EmbedColourError),
				Flags:  uint32(discord.MessageFlagEphemeral),
			},
		}, nil
	}

	if !poll.AllowEntries {
		return &discord.InteractionResponse{
			Type: discord.InteractionCallbackTypeChannelMessageSource,
			Data: &discord.InteractionCallbackData{
				Embeds: welcomer.NewEmbed("This poll does not have entries enabled. Please try again later.", welcomer.EmbedColourError),
				Flags:  uint32(discord.MessageFlagEphemeral),
			},
		}, nil
	}

	rolesAllowed := welcomer.UnmarshalRolesListJSON(poll.RolesAllowed.Bytes)
	rolesExcluded := welcomer.UnmarshalRolesListJSON(poll.RolesExcluded.Bytes)

	if len(rolesAllowed) > 0 && !hasAnyRoles(rolesAllowed, interaction.Member.Roles) {
		return &discord.InteractionResponse{
			Type: discord.InteractionCallbackTypeChannelMessageSource,
			Data: &discord.InteractionCallbackData{
				Embeds: welcomer.NewEmbed("Sorry, you are missing a required role to vote on this poll.", welcomer.EmbedColourError),
				Flags:  uint32(discord.MessageFlagEphemeral),
			},
		}, nil
	}

	if len(rolesExcluded) > 0 && hasAnyRoles(rolesExcluded, interaction.Member.Roles) {
		return &discord.InteractionResponse{
			Type: discord.InteractionCallbackTypeChannelMessageSource,
			Data: &discord.InteractionCallbackData{
				Embeds: welcomer.NewEmbed("Sorry, you have a role that disqualifies you from voting on this poll.", welcomer.EmbedColourError),
				Flags:  uint32(discord.MessageFlagEphemeral),
			},
		}, nil
	}

	if poll.MinimumJoinDate.Unix() > 0 {
		joinBefore := poll.StartTime.Add(-(time.Duration(poll.MinimumJoinDate.Unix()) * time.Second))
		if interaction.Member.JoinedAt.After(joinBefore) {
			return &discord.InteractionResponse{
				Type: discord.InteractionCallbackTypeChannelMessageSource,
				Data: &discord.InteractionCallbackData{
					Embeds: welcomer.NewEmbed(fmt.Sprintf("Sorry, you must have joined the server before <t:%d:f> to vote on this poll.", joinBefore.Unix()), welcomer.EmbedColourError),
					Flags:  uint32(discord.MessageFlagEphemeral),
				},
			}, nil
		}
	}

	pollEntries, err := welcomer.Queries.GetPollEntriesForUser(ctx, database.GetPollEntriesForUserParams{
		PollUuid: pollUUID,
		UserID:   int64(interaction.GetUser().ID),
	})
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		welcomer.Logger.Error().Err(err).
			Int64("guild_id", int64(*interaction.GuildID)).
			Str("poll_uuid", pollUUID.String()).
			Int64("user_id", int64(interaction.GetUser().ID)).
			Msg("Failed to get poll entries for user")

		return nil, err
	}

	if len(pollEntries) > 0 && poll.Resubmissions == string(welcomer.PollResubmissionOptionNever) {
		return &discord.InteractionResponse{
			Type: discord.InteractionCallbackTypeChannelMessageSource,
			Data: &discord.InteractionCallbackData{
				Embeds: welcomer.NewEmbed("Sorry, you have already submitted an entry for this poll and resubmissions are not allowed.", welcomer.EmbedColourError),
				Flags:  uint32(discord.MessageFlagEphemeral),
			},
		}, nil
	}

	options := welcomer.UnmarshalAnswersListJSON(poll.PollOptions.Bytes)

	componentOptions := make([]discord.ApplicationSelectOption, len(options))

	if interaction.Type == discord.InteractionTypeMessageComponent {
		if len(options) == 1 {
			return &discord.InteractionResponse{
				Type: discord.InteractionCallbackTypeModal,
				Data: &discord.InteractionCallbackData{
					Title:    "Submit Vote",
					CustomID: interaction.Data.CustomID,
					Components: []discord.InteractionComponent{
						{
							Type:  discord.InteractionComponentTypeLabel,
							Label: options[0],
							Component: &discord.InteractionComponent{
								CustomID: pollVoteSelectionKey,
								Type:     discord.InteractionComponentTypeCheckbox,
								Default:  new(len(pollEntries) > 0),
							},
						},
					},
				},
			}, nil
		}

		for i, option := range options {
			componentOptions[i] = discord.ApplicationSelectOption{
				Label: option,
				Value: strconv.Itoa(i),
				Default: slices.ContainsFunc(pollEntries, func(e *database.GuildPollsEntries) bool {
					return int(e.OptionIndex) == i
				}),
			}
		}

		switch poll.MaximumSelections {
		case 1:
			return &discord.InteractionResponse{
				Type: discord.InteractionCallbackTypeModal,
				Data: &discord.InteractionCallbackData{
					Title:    "Submit Vote",
					CustomID: interaction.Data.CustomID,
					Components: []discord.InteractionComponent{
						{
							Type:  discord.InteractionComponentTypeLabel,
							Label: "Select Your Vote",
							Component: &discord.InteractionComponent{
								CustomID: pollVoteSelectionKey,
								Type:     discord.InteractionComponentTypeRadioGroup,
								Options:  componentOptions,
								Required: new(len(pollEntries) == 0 || poll.Resubmissions != string(welcomer.PollResubmissionOptionAlways)),
							},
						},
					},
				},
			}, nil
		default:
			return &discord.InteractionResponse{
				Type: discord.InteractionCallbackTypeModal,
				Data: &discord.InteractionCallbackData{
					Title:    "Submit Votes",
					CustomID: interaction.Data.CustomID,
					Components: []discord.InteractionComponent{
						{
							Type:  discord.InteractionComponentTypeLabel,
							Label: "Select Your Votes",
							Component: &discord.InteractionComponent{
								CustomID:  pollVoteSelectionKey,
								Type:      discord.InteractionComponentTypeCheckboxGroup,
								Options:   componentOptions,
								MaxValues: new(welcomer.If(poll.MaximumSelections == 0, int32(len(componentOptions)), poll.MaximumSelections)),
								Required:  new(len(pollEntries) == 0 || poll.Resubmissions != string(welcomer.PollResubmissionOptionAlways)),
							},
						},
					},
				},
			}, nil
		}
	}

	pollOptionsArgument, err := subway.GetArgument(ctx, pollVoteSelectionKey)
	if err != nil {
		welcomer.Logger.Warn().
			Int64("guild_id", int64(*interaction.GuildID)).
			Str("poll_uuid", pollUUID.String()).
			Msg("Poll options argument not found in poll entry interaction")

		return nil, nil
	}

	var pollOptions []int32

	switch pollOptionsArgument.ArgumentType {
	case subway.ArgumentTypeBool:
		pollOptionsBool := pollOptionsArgument.MustBool()
		if pollOptionsBool {
			pollOptions = []int32{0}
		}
	case subway.ArgumentTypeStrings:
		pollOptionsStrings := pollOptionsArgument.MustStrings()
		for _, optionStr := range pollOptionsStrings {
			optionIndex, err := strconv.ParseInt(optionStr, 10, 32)
			if err != nil {
				welcomer.Logger.Warn().
					Int64("guild_id", int64(*interaction.GuildID)).
					Str("poll_uuid", pollUUID.String()).
					Str("option_str", optionStr).
					Msg("Invalid option index submitted for poll entry")

				continue
			}

			pollOptions = append(pollOptions, int32(optionIndex))
		}
	case subway.ArgumentTypeString:
		optionIndex, err := strconv.ParseInt(pollOptionsArgument.MustString(), 10, 32)
		if err != nil {
			welcomer.Logger.Warn().
				Int64("guild_id", int64(*interaction.GuildID)).
				Str("poll_uuid", pollUUID.String()).
				Str("option_str", pollOptionsArgument.MustString()).
				Msg("Invalid option index submitted for poll entry")
		}

		pollOptions = []int32{int32(optionIndex)}
	default:
		welcomer.Logger.Warn().
			Int64("guild_id", int64(*interaction.GuildID)).
			Str("poll_uuid", pollUUID.String()).
			Int("argument_type", int(pollOptionsArgument.ArgumentType)).
			Msg("Invalid argument type for poll options in poll entry interaction")

		return nil, nil
	}

	err = welcomer.Queries.RemovePollEntriesNotMatching(ctx, database.RemovePollEntriesNotMatchingParams{
		PollUuid: pollUUID,
		UserID:   int64(interaction.GetUser().ID),
		Options:  pollOptions,
	})
	if err != nil {
		welcomer.Logger.Error().Err(err).
			Int64("guild_id", int64(*interaction.GuildID)).
			Str("poll_uuid", pollUUID.String()).
			Int64("user_id", int64(interaction.GetUser().ID)).
			Msg("Failed to remove poll entries not matching options")

		return nil, err
	}

	for _, pollOption := range pollOptions {
		_, err = welcomer.Queries.AddPollEntry(ctx, database.AddPollEntryParams{
			PollUuid:    pollUUID,
			UserID:      int64(interaction.GetUser().ID),
			OptionIndex: pollOption,
		})
		if err != nil && !strings.Contains(err.Error(), "unique constraint") {
			welcomer.Logger.Warn().Err(err).
				Int64("guild_id", int64(*interaction.GuildID)).
				Str("poll_uuid", pollUUID.String()).
				Int64("user_id", int64(interaction.GetUser().ID)).
				Int32("option_index", pollOption).
				Msg("Failed to add poll entry")

			continue
		}
	}

	var showResultsToUser bool

	var updateMainMessage bool

	switch welcomer.PollResultVisibilityOption(poll.ResultsVisibility) {
	case welcomer.PollResultVisibilityOptionAlways:
		showResultsToUser = true
		updateMainMessage = true
	case welcomer.PollResultVisibilityOptionAfterVoting:
		showResultsToUser = true
	}

	entriesCounts, err := welcomer.Queries.GetPollEntriesCounts(ctx, poll.PollUuid)
	if err != nil {
		welcomer.Logger.Error().Err(err).
			Int64("guild_id", int64(*interaction.GuildID)).
			Str("poll_uuid", pollUUID.String()).
			Msg("Failed to get poll entries counts")

		return nil, err
	}

	results := make([]int, len(options))

	for _, entryCount := range entriesCounts {
		results[entryCount.OptionIndex] = int(entryCount.EntryCount)
	}

	if updateMainMessage {
		go func() {
			time.Sleep(pollMessageUpdateRate)

			newEntries, err := welcomer.Queries.GetPollEntriesCounts(ctx, poll.PollUuid)

			totalOldEntries := 0
			totalNewEntries := 0

			for _, entryCount := range entriesCounts {
				totalOldEntries += int(entryCount.EntryCount)
			}

			for _, entryCount := range newEntries {
				totalNewEntries += int(entryCount.EntryCount)
			}

			if totalOldEntries != totalNewEntries {
				return
			}

			message := discord.Message{
				ID:        discord.Snowflake(poll.MessageID),
				ChannelID: discord.Snowflake(poll.ChannelID),
			}

			session, err := welcomer.AcquireSession(ctx, welcomer.GetManagerNameFromContext(ctx))
			if err != nil {
				welcomer.Logger.Error().Err(err).
					Int64("guild_id", int64(*interaction.GuildID)).
					Str("poll_uuid", pollUUID.String()).
					Msg("Failed to acquire session to edit poll message after entry")

				return
			}

			_, err = message.Edit(ctx, session, welcomer.WebhookMessageParamsToMessageParams(pollView(poll, results, false, false)))
			if err != nil {
				welcomer.Logger.Error().Err(err).
					Int64("guild_id", int64(*interaction.GuildID)).
					Str("poll_uuid", pollUUID.String()).
					Msg("Failed to edit poll message after entry")
			}

			welcomer.Logger.Info().
				Int64("guild_id", int64(*interaction.GuildID)).
				Str("poll_uuid", pollUUID.String()).
				Int("entries", totalNewEntries).
				Msg("Updated poll message after new entry")
		}()
	}

	if showResultsToUser {
		view := pollView(poll, results, true, false)
		view.Components[0].Content = "Your vote has been submitted! Here are the current results:"

		return &discord.InteractionResponse{
			Type: discord.InteractionCallbackTypeChannelMessageSource,
			Data: welcomer.WebhookMessageParamsToInteractionCallbackData(view, uint32(discord.MessageFlagEphemeral+discord.MessageFlagIsComponentsV2)),
		}, nil
	}

	return &discord.InteractionResponse{
		Type: discord.InteractionCallbackTypeChannelMessageSource,
		Data: &discord.InteractionCallbackData{
			Embeds: welcomer.NewEmbed("Your vote has been submitted!", welcomer.EmbedColourSuccess),
			Flags:  uint32(discord.MessageFlagEphemeral),
		},
	}, nil
}

func handlePollEditComponent(ctx context.Context, sub *subway.Subway, interaction discord.Interaction) (*discord.InteractionResponse, error) {
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

	pollUUID, err := uuid.FromString(customIDSplit[1])
	if err != nil {
		return nil, err
	}

	poll, err := welcomer.Queries.GetPoll(ctx, database.GetPollParams{
		GuildID:  int64(*interaction.GuildID),
		PollUuid: pollUUID,
	})
	if err != nil {
		welcomer.Logger.Error().Err(err).
			Int64("guild_id", int64(*interaction.GuildID)).
			Str("poll_uuid", pollUUID.String()).
			Msg("Failed to get poll settings")

		return nil, err
	}

	switch interaction.Type {
	case discord.InteractionTypeMessageComponent:
		switch customIDSplit[2] {
		case pollSetupMenuTitleKey:
			return &discord.InteractionResponse{
				Data: &discord.InteractionCallbackData{
					Title:    "Customise Poll Message",
					CustomID: interaction.Data.CustomID,
					Components: []discord.InteractionComponent{
						{
							Type:  discord.InteractionComponentTypeLabel,
							Label: "Title",
							Component: &discord.InteractionComponent{
								CustomID: pollSetupMenuTitleKey,
								Type:     discord.InteractionComponentTypeTextInput,
								Value:    welcomer.StringToJsonLiteral(poll.Title),
								Style:    discord.InteractionComponentStyleShort,
								Required: new(false),
							},
						},
						{
							Type:  discord.InteractionComponentTypeLabel,
							Label: "Description",
							Component: &discord.InteractionComponent{
								CustomID: pollSetupMenuDescriptionKey,
								Type:     discord.InteractionComponentTypeTextInput,
								Value:    welcomer.StringToJsonLiteral(poll.Description),
								Style:    discord.InteractionComponentStyleParagraph,
								Required: new(false),
							},
						},
						{
							Type:        discord.InteractionComponentTypeLabel,
							Label:       "Accent Colour",
							Description: "If specified, the left side of the poll message will be this colour. Accepts #HEX format.",
							Component: &discord.InteractionComponent{
								CustomID:    pollSetupMenuAccentColourKey,
								Type:        discord.InteractionComponentTypeTextInput,
								Placeholder: "#4CD787",
								Value:       welcomer.StringToJsonLiteral(welcomer.If(poll.AccentColour < 0, "", fmt.Sprintf("#%06X", poll.AccentColour))),
								Style:       discord.InteractionComponentStyleShort,
								Required:    new(false),
							},
						},
						{
							Type:        discord.InteractionComponentTypeLabel,
							Label:       "Image URL",
							Description: "If specified, this image will show below your title and description.",
							Component: &discord.InteractionComponent{
								CustomID:    pollSetupMenuThumbnailURLKey,
								Type:        discord.InteractionComponentTypeTextInput,
								Placeholder: "https://example.com/image.png",
								Value:       welcomer.StringToJsonLiteral(poll.ImageUrl),
								Style:       discord.InteractionComponentStyleShort,
								Required:    new(false),
							},
						},
					},
				},
				Type: discord.InteractionCallbackTypeModal,
			}, nil
		case pollSetupMenuAnswersKey:
			return &discord.InteractionResponse{
				Data: &discord.InteractionCallbackData{
					Title:    "Edit Poll Answers",
					CustomID: interaction.Data.CustomID,
					Components: []discord.InteractionComponent{
						{
							Type:        discord.InteractionComponentTypeLabel,
							Label:       "Answers",
							Description: "One answer per line. Max of 10 answers is allowed.",
							Component: &discord.InteractionComponent{
								CustomID:    pollSetupMenuAnswersKey,
								Type:        discord.InteractionComponentTypeTextInput,
								Placeholder: "Answer 1\nAnswer 2\nAnswer 3",
								Value:       welcomer.StringToJsonLiteral(welcomer.Coalesce(strings.Join(welcomer.UnmarshalAnswersListJSON(poll.PollOptions.Bytes), "\n"), "")),
								Style:       discord.InteractionComponentStyleParagraph,
							},
						},
					},
				},
				Type: discord.InteractionCallbackTypeModal,
			}, nil
		case pollSetupMenuOptionsKey:
			optionComponents := []discord.InteractionComponent{
				{
					Type:  discord.InteractionComponentTypeLabel,
					Label: "Answers Options",
					Component: &discord.InteractionComponent{
						CustomID: pollSetupMenuMaximumAnswersKey,
						Type:     discord.InteractionComponentTypeRadioGroup,
						Options: []discord.ApplicationSelectOption{
							{
								Label:   "Single Answer",
								Value:   "1",
								Default: poll.MaximumSelections == 1,
							},
							{
								Label:   "Allow Multiple Answers",
								Value:   "0",
								Default: poll.MaximumSelections == 0,
							},
						},
					},
				},
				{
					Type:        discord.InteractionComponentTypeLabel,
					Label:       "Anonymous Poll",
					Description: "Resubmissions are not allowed and results will only be available when the poll ends.",
					Component: &discord.InteractionComponent{
						Type:     discord.InteractionComponentTypeCheckbox,
						CustomID: pollSetupMenuToggleAnonymousVotingKey,
						Required: new(false),
						Default:  &poll.IsAnonymous,
					},
				},
			}

			if !poll.IsAnonymous {
				optionComponents = append(optionComponents, []discord.InteractionComponent{
					{
						Type:        discord.InteractionComponentTypeLabel,
						Label:       "Resubmissions",
						Description: "Manage if users can change answers or can only add additional answers.",
						Component: &discord.InteractionComponent{
							CustomID: pollSetupMenuManageResubmissionsKey,
							Type:     discord.InteractionComponentTypeRadioGroup,
							Options: []discord.ApplicationSelectOption{
								{
									Label:   "Not Allowed",
									Value:   string(welcomer.PollResubmissionOptionNever),
									Default: poll.Resubmissions == string(welcomer.PollResubmissionOptionNever) || poll.IsAnonymous,
								},
								{
									Label:   "Allowed",
									Value:   string(welcomer.PollResubmissionOptionAlways),
									Default: poll.Resubmissions == string(welcomer.PollResubmissionOptionAlways) && !poll.IsAnonymous,
								},
								{
									Label:       "Allow Additions Only",
									Description: "Only allows additional answers to be selected and existing options cannot be removed.",
									Value:       string(welcomer.PollResubmissionOptionOnlyAdditions),
									Default:     poll.Resubmissions == string(welcomer.PollResubmissionOptionOnlyAdditions) && !poll.IsAnonymous,
								},
							},
						},
						Disabled: poll.IsAnonymous,
					},
					{
						Type:        discord.InteractionComponentTypeLabel,
						Label:       "Results Visibility",
						Description: "Manage when poll results are visible to voters in the poll message.",
						Component: &discord.InteractionComponent{
							CustomID: pollSetupMenuShowResultsKey,
							Type:     discord.InteractionComponentTypeRadioGroup,
							Options: []discord.ApplicationSelectOption{
								{
									Label:       "Always Visible",
									Description: "Shows live poll results on the main poll message when submitted.",
									Value:       string(welcomer.PollResultVisibilityOptionAlways),
									Default:     poll.ResultsVisibility == string(welcomer.PollResultVisibilityOptionAlways) && !poll.IsAnonymous,
								},
								{
									Label:       "Visible After Voting",
									Description: "Poll results will only be visible to a user after submitting.",
									Value:       string(welcomer.PollResultVisibilityOptionAfterVoting),
									Default:     poll.ResultsVisibility == string(welcomer.PollResultVisibilityOptionAfterVoting) && !poll.IsAnonymous,
								},
								{
									Label:       "Hidden Until Poll Ends",
									Description: "Only updates the main poll message with results when ended.",
									Value:       string(welcomer.PollResultVisibilityOptionAfterEnd),
									Default:     poll.ResultsVisibility == string(welcomer.PollResultVisibilityOptionAfterEnd) || poll.IsAnonymous,
								},
							},
						},
						Disabled: poll.IsAnonymous,
					},
				}...)
			}

			return &discord.InteractionResponse{
				Data: &discord.InteractionCallbackData{
					Title:      "Edit Poll Options",
					CustomID:   interaction.Data.CustomID,
					Components: optionComponents,
				},
				Type: discord.InteractionCallbackTypeModal,
			}, nil
		case pollSetupMenuDurationKey:
			return &discord.InteractionResponse{
				Data: &discord.InteractionCallbackData{
					Title:    "Edit Poll Duration",
					CustomID: interaction.Data.CustomID,
					Components: []discord.InteractionComponent{
						{
							Type:        discord.InteractionComponentTypeLabel,
							Label:       "Duration",
							Description: "e.g. 1h, 30m, 2d. Only years, days, hours and minutes are supported.",
							Component: &discord.InteractionComponent{
								CustomID:    pollSetupMenuDurationKey,
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
		case pollSetupMenuRolesAllowedKey:
			return &discord.InteractionResponse{
				Data: &discord.InteractionCallbackData{
					Title:    "Edit Poll Entry Rules",
					CustomID: interaction.Data.CustomID,
					Components: []discord.InteractionComponent{
						{
							Type:        discord.InteractionComponentTypeLabel,
							Label:       "Roles Allowed to Enter",
							Description: "Users must have at least one of these roles to enter. Ignored if empty.",
							Component: &discord.InteractionComponent{
								CustomID:  pollSetupMenuRolesAllowedIncludedKey,
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
								CustomID:  pollSetupMenuRolesAllowedExcludedKey,
								Type:      discord.InteractionComponentTypeRoleSelect,
								Required:  new(false),
								MaxValues: new(int32(25)),
							},
						},
					},
				},
				Type: discord.InteractionCallbackTypeModal,
			}, nil
		case pollSetupMenuMinimumJoinDateKey:
			return &discord.InteractionResponse{
				Data: &discord.InteractionCallbackData{
					Title:    "Edit Poll Minimum Join Date",
					CustomID: interaction.Data.CustomID,
					Components: []discord.InteractionComponent{
						{
							Type:        discord.InteractionComponentTypeLabel,
							Label:       "Minimum Join Date",
							Description: "Users joined within the duration specified cannot enter. Ignored if empty. e.g. 1h, 30m, 2d. ",
							Component: &discord.InteractionComponent{
								CustomID:    pollSetupMenuMinimumJoinDateKey,
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
		case pollSetupMenuStartKey:
			roles := welcomer.UnmarshalRolesListJSON(poll.RolesAllowed.Bytes)

			var options []discord.ApplicationSelectOption

			if len(roles) == 0 {
				options = []discord.ApplicationSelectOption{
					{
						Label: "Ping @everyone",
						Value: pollSetupMenuPingEveryoneKey,
					},
					{
						Label: "Ping @here",
						Value: pollSetupMenuPingHereKey,
					},
				}
			} else {
				options = []discord.ApplicationSelectOption{
					{
						Label:       "Ping Roles Allowed to Enter",
						Value:       pollSetupMenuPingRolesAllowedToEnterKey,
						Description: "Pings roles you have configured in \"roles allowed to enter\"",
					},
				}
			}

			return &discord.InteractionResponse{
				Data: &discord.InteractionCallbackData{
					Title:    "Start Poll",
					CustomID: interaction.Data.CustomID,
					Components: []discord.InteractionComponent{
						{
							Type:    discord.InteractionComponentTypeTextDisplay,
							Content: "Once started, the poll message will be sent and votes will be allowed. You can end or extend the poll at any time, but you cannot edit the poll settings.\n\nBelow you can configure who should be pinged when the poll starts.",
						},
						{
							Type:  discord.InteractionComponentTypeLabel,
							Label: "Delivery Option",
							Component: &discord.InteractionComponent{
								Type:      discord.InteractionComponentTypeCheckboxGroup,
								CustomID:  pollSetupMenuPingKey,
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
								CustomID:  pollSetupMenuPingAdditionalRolesKey,
								Required:  new(false),
								MaxValues: new(int32(25)),
							},
						},
					},
				},
				Type: discord.InteractionCallbackTypeModal,
			}, nil
		case pollSetupMenuPreviewOnKey:
			poll.StartTime = time.Now()

			if poll.EndTime.Unix() > 0 {
				poll.EndTime = time.Now().Add(time.Duration(poll.EndTime.Unix()) * time.Second)
			}

			answers := welcomer.UnmarshalAnswersListJSON(poll.PollOptions.Bytes)
			results := make([]int, len(answers))

			for i := range results {
				results[i] = rand.IntN(20)
			}

			message := pollView(poll, results, false, false)

			// Hack to disable poll button and add back button
			message.Components[len(message.Components)-1].Components[0].Disabled = true
			message.Components[len(message.Components)-1].Components = append(message.Components[len(message.Components)-1].Components, discord.InteractionComponent{
				CustomID: "poll_edit:" + poll.PollUuid.String() + ":" + pollSetupMenuPreviewOffKey,
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
					Str("poll_uuid", poll.PollUuid.String()).
					Msg("Failed to edit poll message")

				return nil, err
			}

			return nil, nil
		case pollSetupMenuPreviewOffKey:
		default:
			welcomer.Logger.Warn().
				Int64("guild_id", int64(*interaction.GuildID)).
				Str("poll_uuid", poll.PollUuid.String()).
				Str("custom_id", interaction.Data.CustomID).
				Msg("Unknown poll manage component interaction")
		}
	case discord.InteractionTypeModalSubmit:
		switch customIDSplit[2] {
		case "":
			if titleArgument, err := subway.GetArgument(ctx, pollSetupMenuTitleKey); err == nil {
				poll.Title = titleArgument.MustString()
			}

			if answersArgument, err := subway.GetArgument(ctx, pollSetupMenuAnswersKey); err == nil {
				answers := answersArgument.MustString()
				answersList := strings.Split(answers, "\n")

				poll.PollOptions = pgtype.JSONB{
					Bytes:  welcomer.MarshalAnswersListJSON(answersList[:min(len(answersList), 10)]),
					Status: pgtype.Present,
				}
			}

			if durationArgument, err := subway.GetArgument(ctx, pollSetupMenuDurationKey); err == nil {
				seconds, err := welcomer.ParseDurationAsSeconds(durationArgument.MustString())
				if err != nil || seconds < 0 {
					welcomer.Logger.Error().Err(err).
						Int64("guild_id", int64(*interaction.GuildID)).
						Str("duration", durationArgument.MustString()).
						Msg("Failed to parse duration")

					return nil, nil
				}

				poll.EndTime = time.Unix(int64(seconds), 0)
			}

			if optionsArgument, err := subway.GetArgument(ctx, pollSetupMenuOptionsKey); err == nil {
				options := optionsArgument.MustStrings()

				for _, option := range options {
					switch option {
					case pollSetupMenuAllowMultipleAnswersKey:
						poll.MaximumSelections = 0
					case pollSetupMenuToggleAnonymousVotingKey:
						poll.IsAnonymous = true
					}
				}
			}
		case pollSetupMenuTitleKey:
			if titleArgument, err := subway.GetArgument(ctx, pollSetupMenuTitleKey); err == nil {
				poll.Title = titleArgument.MustString()
			} else {
				poll.Title = ""
			}

			if descriptionArgument, err := subway.GetArgument(ctx, pollSetupMenuDescriptionKey); err == nil {
				poll.Description = descriptionArgument.MustString()
			} else {
				poll.Description = ""
			}

			if accentColourArgument, err := subway.GetArgument(ctx, pollSetupMenuAccentColourKey); err == nil {
				rgba, err := welcomer.ParseColour(accentColourArgument.MustString(), "#000000")
				if err != nil {
					welcomer.Logger.Error().Err(err).
						Int64("guild_id", int64(*interaction.GuildID)).
						Str("accent_colour", accentColourArgument.MustString()).
						Msg("Failed to parse accent colour")

					poll.AccentColour = -1
				} else {
					poll.AccentColour = int64(int32(rgba.R)<<16 + int32(rgba.G)<<8 + int32(rgba.B))
				}
			} else {
				poll.AccentColour = -1
			}

			if thumbnailURLArgument, err := subway.GetArgument(ctx, pollSetupMenuThumbnailURLKey); err == nil {
				poll.ImageUrl = thumbnailURLArgument.MustString()

				if _, ok := welcomer.IsValidURL(thumbnailURLArgument.MustString()); !ok {
					welcomer.Logger.Error().Err(err).
						Int64("guild_id", int64(*interaction.GuildID)).
						Str("thumbnail_url", thumbnailURLArgument.MustString()).
						Msg("Failed to parse thumbnail URL")

					poll.ImageUrl = ""
				}
			} else {
				poll.ImageUrl = ""
			}
		case pollSetupMenuAnswersKey:
			if answersArgument, err := subway.GetArgument(ctx, pollSetupMenuAnswersKey); err == nil {
				answers := strings.Split(answersArgument.MustString(), "\n")
				poll.PollOptions = pgtype.JSONB{
					Bytes:  welcomer.MarshalAnswersListJSON(answers[:min(len(answers), 10)]),
					Status: pgtype.Present,
				}
			} else {
				poll.PollOptions = pgtype.JSONB{
					Bytes:  nil,
					Status: pgtype.Null,
				}
			}
		case pollSetupMenuOptionsKey:
			if maximumAnswersKey, err := subway.GetArgument(ctx, pollSetupMenuMaximumAnswersKey); err == nil {
				maximumAnswersValue := maximumAnswersKey.MustString()

				if maximumAnswersValue == "1" {
					poll.MaximumSelections = 1
				} else {
					poll.MaximumSelections = 0
				}
			}

			if manageResubmissionsKey, err := subway.GetArgument(ctx, pollSetupMenuManageResubmissionsKey); err == nil {
				manageResubmissionsValue := manageResubmissionsKey.MustString()
				poll.Resubmissions = manageResubmissionsValue
			}

			if showResultsKey, err := subway.GetArgument(ctx, pollSetupMenuShowResultsKey); err == nil {
				showResultsValue := showResultsKey.MustString()
				poll.ResultsVisibility = showResultsValue
			}

			// Move toggle anonymous to the bottom so it can override show results and manage resubmissions options.
			if toggleAnonymousVotingKey, err := subway.GetArgument(ctx, pollSetupMenuToggleAnonymousVotingKey); err == nil {
				if toggleAnonymousVotingKey.MustBool() {
					poll.IsAnonymous = true
					poll.Resubmissions = welcomer.PollResubmissionOptionNever.String()
					poll.ResultsVisibility = welcomer.PollResultVisibilityOptionAfterEnd.String()
				} else {
					poll.IsAnonymous = false
				}
			}
		case pollSetupMenuDurationKey:
			if durationArgument, err := subway.GetArgument(ctx, pollSetupMenuDurationKey); err == nil {
				seconds, err := welcomer.ParseDurationAsSeconds(durationArgument.MustString())
				if err != nil || seconds < 0 {
					welcomer.Logger.Error().Err(err).
						Int64("guild_id", int64(*interaction.GuildID)).
						Str("duration", durationArgument.MustString()).
						Msg("Failed to parse duration")

					return nil, nil
				}

				poll.EndTime = time.Unix(int64(seconds), 0)
			} else {
				poll.EndTime = time.Time{}
			}
		case pollSetupMenuRolesAllowedKey:
			if allowedRoles, err := subway.GetArgument(ctx, pollSetupMenuRolesAllowedIncludedKey); err == nil {
				allowedRolesList := make([]discord.Snowflake, 0, len(allowedRoles.MustStrings()))

				for _, roleString := range allowedRoles.MustStrings() {
					roleSnowflake, err := welcomer.Atoi(roleString)
					if err == nil {
						allowedRolesList = append(allowedRolesList, discord.Snowflake(roleSnowflake))
					}
				}

				poll.RolesAllowed = pgtype.JSONB{
					Bytes:  welcomer.MarshalRolesListJSON(allowedRolesList),
					Status: pgtype.Present,
				}
			} else {
				poll.RolesAllowed = pgtype.JSONB{
					Bytes:  []byte{123, 125}, // []
					Status: pgtype.Present,
				}
			}

			if excludedRoles, err := subway.GetArgument(ctx, pollSetupMenuRolesAllowedExcludedKey); err == nil {
				excludedRolesList := make([]discord.Snowflake, 0, len(excludedRoles.MustStrings()))

				for _, roleString := range excludedRoles.MustStrings() {
					roleSnowflake, err := welcomer.Atoi(roleString)
					if err == nil {
						excludedRolesList = append(excludedRolesList, discord.Snowflake(roleSnowflake))
					}
				}

				poll.RolesExcluded = pgtype.JSONB{
					Bytes:  welcomer.MarshalRolesListJSON(excludedRolesList),
					Status: pgtype.Present,
				}
			} else {
				poll.RolesExcluded = pgtype.JSONB{
					Bytes:  []byte{123, 125}, // []
					Status: pgtype.Present,
				}
			}
		case pollSetupMenuMinimumJoinDateKey:
			if durationArgument, err := subway.GetArgument(ctx, pollSetupMenuMinimumJoinDateKey); err == nil {
				seconds, err := welcomer.ParseDurationAsSeconds(durationArgument.MustString())
				if err != nil || seconds < 0 {
					welcomer.Logger.Error().Err(err).
						Int64("guild_id", int64(*interaction.GuildID)).
						Str("duration", durationArgument.MustString()).
						Msg("Failed to parse duration")

					return nil, nil
				}

				poll.MinimumJoinDate = time.Unix(int64(seconds), 0)
			} else {
				poll.MinimumJoinDate = time.Time{}
			}
		case pollSetupMenuStartKey:
			var pingOptions []string

			if pingOptionsArgument, err := subway.GetArgument(ctx, pollSetupMenuPingKey); err == nil {
				pingOptions = pingOptionsArgument.MustStrings()
			}

			poll.StartTime = time.Now()

			if poll.EndTime.Unix() > 0 {
				poll.EndTime = time.Now().Add(time.Duration(poll.EndTime.Unix()) * time.Second)
			}

			poll.IsSetup = false

			session, err := welcomer.AcquireSession(ctx, welcomer.GetManagerNameFromContext(ctx))
			if err != nil {
				return nil, err
			}

			answers := welcomer.UnmarshalAnswersListJSON(poll.PollOptions.Bytes)
			results := make([]int, len(answers))

			message, err := interaction.Channel.Send(ctx, session, welcomer.WebhookMessageParamsToMessageParams(pollView(poll, results, false, false)))
			if err != nil {
				welcomer.Logger.Error().Err(err).
					Int64("guild_id", int64(*interaction.GuildID)).
					Msg("Failed to send poll message")

				return nil, err
			}

			pingMessage := ""

			if len(pingOptions) > 0 {
				switch {
				case slices.Contains(pingOptions, pollSetupMenuPingEveryoneKey):
					pingMessage += "@everyone"
				case slices.Contains(pingOptions, pollSetupMenuPingHereKey):
					pingMessage += "@here"
				case slices.Contains(pingOptions, pollSetupMenuPingRolesAllowedToEnterKey):
					rolesAllowedToEnter := welcomer.UnmarshalRolesListJSON(poll.RolesAllowed.Bytes)

					if len(rolesAllowedToEnter) > 0 {
						for _, role := range rolesAllowedToEnter {
							pingMessage += fmt.Sprintf(" <@&%d>", role)
						}
					}
				}
			}

			if additionalRolesArgument, err := subway.GetArgument(ctx, pollSetupMenuPingAdditionalRolesKey); err == nil {
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
						Msg("Failed to send poll ping message")
				}
			}

			poll.MessageID = int64(message.ID)
			poll.ChannelID = int64(message.ChannelID)

			_, err = welcomer.Queries.UpdatePollMessage(ctx, database.UpdatePollMessageParams{
				PollUuid:  pollUUID,
				MessageID: int64(message.ID),
				ChannelID: int64(message.ChannelID),
			})
			if err != nil {
				welcomer.Logger.Error().Err(err).
					Int64("guild_id", int64(*interaction.GuildID)).
					Str("poll_uuid", poll.PollUuid.String()).
					Msg("Failed to update poll message and channel")

				return nil, err
			}

			welcomer.PusherGuildScience.Push(
				ctx,
				*interaction.GuildID,
				interaction.GetUser().ID,
				database.ScienceGuildEventTypePollStarted,
				&welcomer.GuildSciencePollEvents{
					PollUUID: poll.PollUuid,
				},
			)
		default:
			welcomer.Logger.Warn().
				Int64("guild_id", int64(*interaction.GuildID)).
				Str("poll_uuid", poll.PollUuid.String()).
				Str("custom_id", interaction.Data.CustomID).
				Msg("Unknown poll modal submit interaction")
		}

	default:
		welcomer.Logger.Warn().
			Int64("guild_id", int64(*interaction.GuildID)).
			Int("interaction_type", int(interaction.Type)).
			Msg("Unknown interaction type for poll edit menu")

		return nil, nil
	}

	_, err = welcomer.UpdatePollGuildSettingsWithAudit(ctx, database.UpdatePollParams{
		PollUuid:          poll.PollUuid,
		IsSetup:           poll.IsSetup,
		HasEnded:          poll.HasEnded,
		Title:             poll.Title,
		Description:       poll.Description,
		AccentColour:      poll.AccentColour,
		ImageUrl:          poll.ImageUrl,
		StartTime:         poll.StartTime,
		EndTime:           poll.EndTime,
		PollOptions:       poll.PollOptions,
		IsAnonymous:       poll.IsAnonymous,
		MaximumSelections: poll.MaximumSelections,
		AllowEntries:      poll.AllowEntries,
		Resubmissions:     poll.Resubmissions,
		ResultsVisibility: poll.ResultsVisibility,
		RolesAllowed:      poll.RolesAllowed,
		RolesExcluded:     poll.RolesExcluded,
		MinimumJoinDate:   poll.MinimumJoinDate,
	}, interaction.GetUser().ID, *interaction.GuildID)
	if err != nil {
		welcomer.Logger.Error().Err(err).
			Int64("guild_id", int64(*interaction.GuildID)).
			Str("poll_uuid", poll.PollUuid.String()).
			Msg("Failed to update poll settings")

		return nil, err
	}

	if poll.IsSetup {
		err = discord.CreateInteractionResponse(ctx, sub.EmptySession, interaction.ID, interaction.Token, discord.InteractionResponse{
			Type: welcomer.If(customIDSplit[2] == "", discord.InteractionCallbackTypeChannelMessageSource, discord.InteractionCallbackTypeUpdateMessage),
			Data: welcomer.WebhookMessageParamsToInteractionCallbackData(pollSetupView(poll), uint32(discord.MessageFlagEphemeral+discord.MessageFlagIsComponentsV2)),
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
								Content: "Your poll has now started!\n\nYou can manage your poll settings such as disabling entries, extending the duration or ending the poll early by right clicking the poll message and selecting \"Manage Poll\".\n\n-# How was your experience? Let us know in our feedback channel: https://discord.gg/t2Ye8jBfPh",
							},
							{
								Type: discord.InteractionComponentTypeMediaGallery,
								Items: []discord.InteractionComponentMediaGalleryItem{
									{
										Media: discord.MediaItem{
											URL: "https://welcomer.gg/assets/manage_poll.png",
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
			Str("poll_uuid", poll.PollUuid.String()).
			Msg("Failed to edit poll message")

		return nil, err
	}

	return &discord.InteractionResponse{
		Type: discord.InteractionCallbackTypeDeferredUpdateMessage,
	}, nil
}

func pollManageView(poll *database.GuildPolls) discord.WebhookMessageParams {
	customIDPrefix := "poll_manage:" + poll.PollUuid.String() + ":"

	return discord.WebhookMessageParams{
		Components: []discord.InteractionComponent{
			{
				Type: discord.InteractionComponentTypeContainer,
				Components: []discord.InteractionComponent{
					{
						Type:    discord.InteractionComponentTypeTextDisplay,
						Content: fmt.Sprintf("### Manage poll **%s**", welcomer.Coalesce(poll.Title, "New Poll")),
					},
					{
						Type: discord.InteractionComponentTypeSeparator,
					},
					{
						Type: discord.InteractionComponentTypeSection,
						Components: []discord.InteractionComponent{
							{
								Type: discord.InteractionComponentTypeTextDisplay,
								Content: "**Allow Poll Entries:**\n" +
									welcomer.If(poll.AllowEntries, "True", "False") +
									welcomer.If(!poll.AllowEntries, "\n-# When disabled, users cannot enter the poll. This is useful to temporarily pause entries without ending the poll.", ""),
							},
						},
						Accessory: &discord.InteractionComponent{
							Type:     discord.InteractionComponentTypeButton,
							Style:    discord.InteractionComponentStyleSecondary,
							Label:    welcomer.If(poll.AllowEntries, "Disable", "Enable"),
							CustomID: customIDPrefix + pollManageMenuToggleAllowEntriesKey,
							Disabled: poll.HasEnded,
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
								Content: "**Poll " + welcomer.If(poll.HasEnded, "Ended", "Ends") + ":**\n" +
									welcomer.If(poll.EndTime.Unix() > 0, "<t:"+welcomer.Itoa(poll.EndTime.Unix())+":R> (<t:"+welcomer.Itoa(poll.EndTime.Unix())+":f>)", "No end time (runs indefinitely)") + "\n" +
									welcomer.If(
										poll.HasEnded,
										"-# This poll has already ended, so the duration cannot be extended.",
										"-# Extends the poll end time.",
									),
							},
						},
						Accessory: &discord.InteractionComponent{
							Type:     discord.InteractionComponentTypeButton,
							Style:    discord.InteractionComponentStyleSecondary,
							Label:    "Extend",
							CustomID: customIDPrefix + pollManageMenuExtendDurationKey,
							Disabled: poll.HasEnded,
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
								Content: "**End Poll**",
							},
						},
						Accessory: &discord.InteractionComponent{
							Type:     discord.InteractionComponentTypeButton,
							Style:    discord.InteractionComponentStyleDanger,
							Label:    "End Poll",
							CustomID: customIDPrefix + pollManageMenuEndPollKey,
							Disabled: poll.HasEnded,
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
								Content: "**Export Poll Entries**\n" +
									"-# Exports a CSV file of all poll entries.",
							},
						},
						Accessory: &discord.InteractionComponent{
							Type:     discord.InteractionComponentTypeButton,
							Style:    discord.InteractionComponentStylePrimary,
							Label:    "Export Entries",
							CustomID: customIDPrefix + pollManageMenuExportEntriesKey,
						},
					},
					{
						Type: discord.InteractionComponentTypeSeparator,
					},
				},
			},
		},
	}
}

func pollSetupView(poll *database.GuildPolls) discord.WebhookMessageParams {
	pollAnswers := welcomer.UnmarshalAnswersListJSON(poll.PollOptions.Bytes)
	rolesAllowed := welcomer.UnmarshalRolesListJSON(poll.RolesAllowed.Bytes)
	rolesExcluded := welcomer.UnmarshalRolesListJSON(poll.RolesExcluded.Bytes)

	customIDPrefix := "poll_edit:" + poll.PollUuid.String() + ":"

	containerComponents := []discord.InteractionComponent{
		{
			Type:    discord.InteractionComponentTypeTextDisplay,
			Content: "### Create Poll",
		},
		{
			Type: discord.InteractionComponentTypeSeparator,
		},
		{
			Type: discord.InteractionComponentTypeSection,
			Components: []discord.InteractionComponent{
				{
					Type:    discord.InteractionComponentTypeTextDisplay,
					Content: "**" + welcomer.Coalesce(poll.Title, "New Poll") + "**\n" + poll.Description,
				},
			},
			Accessory: &discord.InteractionComponent{
				Type:     discord.InteractionComponentTypeButton,
				Style:    discord.InteractionComponentStylePrimary,
				Label:    "Customise Message",
				CustomID: customIDPrefix + pollSetupMenuTitleKey,
			},
		},
	}

	if poll.ImageUrl != "" {
		containerComponents = append(containerComponents, discord.InteractionComponent{
			Type: discord.InteractionComponentTypeMediaGallery,
			Items: []discord.InteractionComponentMediaGalleryItem{
				{
					Media: discord.MediaItem{
						URL: poll.ImageUrl,
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
					Content: "**Answers:**\n" + getPollAnswersAsString(pollAnswers),
				},
			},
			Accessory: &discord.InteractionComponent{
				Type:     discord.InteractionComponentTypeButton,
				Style:    discord.InteractionComponentStyleSecondary,
				Label:    "Edit",
				CustomID: customIDPrefix + pollSetupMenuAnswersKey,
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
					Content: "**Poll Options:**\n\n" +
						"**Allow Multiple Answers:** " + welcomer.If(poll.MaximumSelections == 1, "No\n\n", "Yes"+welcomer.If(poll.MaximumSelections == 0, "", " ("+strconv.Itoa(int(poll.MaximumSelections))+")")+"\n\n") +
						"**Anonymous Poll:** " + welcomer.If(poll.IsAnonymous, "Yes\n", "No\n") +
						"**Resubmissions:** " +
						welcomer.If(poll.IsAnonymous, "Not Allowed (anonymous poll)\n",
							welcomer.If(poll.Resubmissions == string(welcomer.PollResubmissionOptionAlways), "Allowed\n",
								welcomer.If(poll.Resubmissions == string(welcomer.PollResubmissionOptionOnlyAdditions), "Allow Additions Only\n",
									welcomer.If(poll.Resubmissions == string(welcomer.PollResubmissionOptionNever), "Not Allowed\n", ""),
								),
							)) +
						"**Results Visibility:** " +
						welcomer.If(poll.IsAnonymous, "Hidden Until Poll Ends (anonymous poll)\n",
							welcomer.If(poll.ResultsVisibility == string(welcomer.PollResultVisibilityOptionAlways), "Always Visible\n",
								welcomer.If(poll.ResultsVisibility == string(welcomer.PollResultVisibilityOptionAfterVoting), "Only visible after voting\n",
									welcomer.If(poll.ResultsVisibility == string(welcomer.PollResultVisibilityOptionAfterEnd), "Hidden Until Poll Ends\n", ""),
								),
							),
						),
				},
			},
			Accessory: &discord.InteractionComponent{
				Type:     discord.InteractionComponentTypeButton,
				Style:    discord.InteractionComponentStyleSecondary,
				Label:    "Edit",
				CustomID: customIDPrefix + pollSetupMenuOptionsKey,
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
					Content: "**Duration:**\n" + welcomer.If(poll.EndTime.Unix() > 0, welcomer.HumanizeDuration(int(poll.EndTime.Unix()), true), "No end time (runs indefinitely)") +
						welcomer.If(poll.EndTime.IsZero(), "\n-# Poll will run until ended manually.", ""),
				},
			},
			Accessory: &discord.InteractionComponent{
				Type:     discord.InteractionComponentTypeButton,
				Style:    discord.InteractionComponentStyleSecondary,
				Label:    "Edit",
				CustomID: customIDPrefix + pollSetupMenuDurationKey,
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
					Content: "**Roles Allowed to Enter:**\n" + welcomer.Coalesce(joinRolesList(rolesAllowed), "All") + "\n\n**Roles Excluded from Entering:**\n" + welcomer.Coalesce(joinRolesList(rolesExcluded), "None"),
				},
			},
			Accessory: &discord.InteractionComponent{
				Type:     discord.InteractionComponentTypeButton,
				Style:    discord.InteractionComponentStyleSecondary,
				Label:    "Edit",
				CustomID: customIDPrefix + pollSetupMenuRolesAllowedKey,
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
					Content: "**Minimum Join Date:**\n" + welcomer.Coalesce(welcomer.HumanizeDuration(int(poll.MinimumJoinDate.Unix()), true), "None") +
						welcomer.If(poll.MinimumJoinDate.Unix() > 0, "\n-# Users who have joined the server within "+welcomer.HumanizeDuration(int(poll.MinimumJoinDate.Unix()), true)+" of the poll starting cannot enter the poll.", ""),
				},
			},
			Accessory: &discord.InteractionComponent{
				Type:     discord.InteractionComponentTypeButton,
				Style:    discord.InteractionComponentStyleSecondary,
				Label:    "Edit",
				CustomID: customIDPrefix + pollSetupMenuMinimumJoinDateKey,
			},
		},
	}...)

	return discord.WebhookMessageParams{
		Flags: discord.MessageFlagEphemeral + discord.MessageFlagIsComponentsV2,
		Components: []discord.InteractionComponent{
			{
				Type:        discord.InteractionComponentTypeContainer,
				AccentColor: new(uint32(welcomer.If(poll.AccentColour >= 0, poll.AccentColour, welcomer.EmbedColourInfo))),
				Components:  containerComponents,
			},
			{
				Type: discord.InteractionComponentTypeActionRow,
				Components: []discord.InteractionComponent{
					{
						Type:     discord.InteractionComponentTypeButton,
						Style:    discord.InteractionComponentStyleSuccess,
						Label:    "Start Poll",
						CustomID: customIDPrefix + pollSetupMenuStartKey,
						Disabled: len(pollAnswers) == 0,
					},
					{
						Type:     discord.InteractionComponentTypeButton,
						Style:    discord.InteractionComponentStyleSecondary,
						Label:    "Preview",
						CustomID: customIDPrefix + pollSetupMenuPreviewOnKey,
					},
				},
			},
		},
	}
}

func getPollResultString(poll *database.GuildPolls, answers []string, results []int, hasFinished bool) string {
	maxValue := 0
	entries := 0

	for _, result := range results {
		entries += result

		if result > maxValue {
			maxValue = result
		}
	}

	answersString := "**Votes:**\n"

	var truePercentage float64

	var resultPercentage int

	for answerIndex, answer := range answers {
		if results[answerIndex] > 0 {
			truePercentage = float64(results[answerIndex]) / float64(entries) * 100
			resultPercentage = int(float64(results[answerIndex]) / float64(maxValue) * 100)
		}

		answersString += fmt.Sprintf("\n%s (**%d vote%s - %.1f**%%)%s\n%s\n", answer, results[answerIndex], welcomer.If(results[answerIndex] == 1, "", "s"), truePercentage, welcomer.If(results[answerIndex] == maxValue && hasFinished, " ⭐", ""), getEmojiCombination(resultPercentage, 10))
	}

	return answersString
}

func getPollResultsMinimal(poll *database.GuildPolls, answers []string, results []int) string {
	answersString := "**Votes:**\n"
	for _, answer := range answers {
		answersString += fmt.Sprintf("- %s\n", answer)
	}

	return answersString
}

func pollView(poll *database.GuildPolls, results []int, isUser, hasFinished bool) discord.WebhookMessageParams {
	containerComponents := []discord.InteractionComponent{
		{
			Type:    discord.InteractionComponentTypeTextDisplay,
			Content: "**" + welcomer.Coalesce(poll.Title, "New Poll") + "**\n" + poll.Description,
		},
	}

	if poll.ImageUrl != "" {
		containerComponents = append(containerComponents, discord.InteractionComponent{
			Type: discord.InteractionComponentTypeMediaGallery,
			Items: []discord.InteractionComponentMediaGalleryItem{
				{
					Media: discord.MediaItem{
						URL: poll.ImageUrl,
					},
				},
			},
		})
	}

	answers := welcomer.UnmarshalAnswersListJSON(poll.PollOptions.Bytes)

	var votes int
	for _, result := range results {
		votes += result
	}

	var answersString string

	if (poll.ResultsVisibility == string(welcomer.PollResultVisibilityOptionAlways) && !poll.IsAnonymous) ||
		poll.ResultsVisibility == string(welcomer.PollResultVisibilityOptionAfterEnd) && time.Now().After(poll.EndTime) ||
		(poll.ResultsVisibility == string(welcomer.PollResultVisibilityOptionAfterVoting) && isUser && !poll.IsAnonymous) {
		// Show answers and percentages
		answersString = getPollResultString(poll, answers, results, hasFinished)
	} else {
		// Show answers
		answersString = getPollResultsMinimal(poll, answers, results)
	}

	if !poll.IsAnonymous {
		answersString += "\n**Total Votes:** " + strconv.Itoa(votes) + "\n"
	}

	containerComponents = append(containerComponents, []discord.InteractionComponent{
		{
			Type: discord.InteractionComponentTypeSeparator,
		},
		{
			Type:    discord.InteractionComponentTypeTextDisplay,
			Content: answersString,
		},
	}...)

	containerComponents = append(containerComponents, []discord.InteractionComponent{
		{
			Type: discord.InteractionComponentTypeSeparator,
		},
		{
			Type:    discord.InteractionComponentTypeTextDisplay,
			Content: "**Poll Ends:** " + welcomer.If(poll.EndTime.Unix() > 0, "<t:"+welcomer.Itoa(poll.EndTime.Unix())+":R> (<t:"+welcomer.Itoa(poll.EndTime.Unix())+":f>)", "No end time (runs indefinitely)"),
		},
	}...)

	message := discord.WebhookMessageParams{
		Components: []discord.InteractionComponent{
			{
				Type:    discord.InteractionComponentTypeTextDisplay,
				Content: fmt.Sprintf("-# <@%d> has started a new poll!", poll.CreatedBy),
			},
			{
				Type:        discord.InteractionComponentTypeContainer,
				AccentColor: new(uint32(welcomer.If(poll.AccentColour >= 0, poll.AccentColour, welcomer.EmbedColourInfo))),
				Components:  containerComponents,
			},
			{
				Type: discord.InteractionComponentTypeActionRow,
				Components: []discord.InteractionComponent{
					{
						Type:     discord.InteractionComponentTypeButton,
						Style:    discord.InteractionComponentStyleSuccess,
						CustomID: "poll_enter:" + poll.PollUuid.String(),
						Label:    "Vote",
						Disabled: !poll.AllowEntries && !poll.IsSetup,
					},
				},
			},
		},
		Flags: discord.MessageFlagIsComponentsV2,
	}

	return message
}

func getPollAnswersAsString(pollAnswers []string) string {
	if len(pollAnswers) == 0 {
		return "No Answers Configured"
	}

	result := ""

	for _, answer := range pollAnswers {
		result += fmt.Sprintf("- %s\n", answer)
	}

	return result
}

var sectionEmojiIDs = [][]string{}

const maxSegmentsPerGroup = 4

func getEmojiCombination(value int, length int) string {
	if value <= 0 || length <= 0 {
		return ""
	}

	segments := int(math.Ceil(float64(value) * float64(length) / 25))

	if segments <= maxSegmentsPerGroup {
		return "<:_:" + sectionEmojiIDs[0][segments-1] + ">"
	}

	var out string

	for {
		if segments > maxSegmentsPerGroup {
			if out == "" {
				out += "<:_:" + sectionEmojiIDs[1][1] + ">"
			} else {
				out += "<:_:" + sectionEmojiIDs[1][0] + ">"
			}

			segments -= maxSegmentsPerGroup
		} else {
			out += "<:_:" + sectionEmojiIDs[2][segments-1] + ">"

			return out
		}
	}
}
