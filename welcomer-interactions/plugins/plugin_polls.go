package plugins

import (
	"context"
	"fmt"
	"strconv"
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
	pollsSetupMenuTitleKey        = "title"
	pollsSetupMenuDescriptionKey  = "description"
	pollsSetupMenuAccentColorKey  = "accent_color"
	pollsSetupMenuThumbnailURLKey = "thumbnail_url"

	pollsSetupMenuAnswersKey               = "answers"
	pollsSetupMenuOptionsKey               = "options"
	pollsSetupMenuDurationKey              = "duration"
	pollsSetupMenuToggleAnonymousVotingKey = "toggle_anonymous_voting"
	pollsSetupMenuMaximumAnswersKey        = "maximum_answers"

	// stub option during initial modal to set it to the maximum answers.
	pollsSetupMenuAllowMultipleAnswersKey = "allow_multiple_answers"

	pollsSetupMenuRolesAllowedKey         = "roles_allowed"
	pollsSetupMenuRolesAllowedIncludedKey = "roles_allowed_included"
	pollsSetupMenuRolesAllowedExcludedKey = "roles_allowed_excluded"

	pollsSetupMenuManageResubmissionsKey = "manage_resubmissions"
	pollsSetupMenuNoResubmissionsKey     = "no_resubmissions"
	pollsSetupMenuAllowAdditionsOnlyKey  = "allow_additions_only"
	pollsSetupMenuAllowResubmissionsKey  = "allow_resubmissions"

	pollsSetupMenuShowResultsKey                   = "show_results"
	pollsSetupMenuResultsAlwaysVisibleKey          = "results_always_visible"
	pollsSetupMenuResultsVisibleAfterVotingKey     = "results_visible_after_voting"
	pollsSetupMenuResultsVisibleAfterVotingEndsKey = "results_visible_after_voting_ends"

	pollsSetupMenuMinimumJoinDateKey = "minimum_join_date"
	pollsSetupMenuStartKey           = "start"

	pollsSetupMenuPreviewOnKey  = "preview_on"
	pollsSetupMenuPreviewOffKey = "preview_off"

	pollsSetupMenuPingKey                    = "ping"
	pollsSetupMenuPingEveryoneKey            = "ping_everyone"
	pollsSetupMenuPingHereKey                = "ping_here"
	pollsSetupMenuPingRolesAllowedToEnterKey = "ping_roles_allowed_to_enter"
	pollsSetupMenuPingAdditionalRolesKey     = "additional_roles_to_ping"
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
	pollsGroup := subway.NewSubcommandGroup(
		"polls",
		"Poll commands",
	)

	pollsGroup.MustAddInteractionCommand(&subway.InteractionCommandable{
		Name:        "new",
		Description: "Makes a new poll",

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
									CustomID: pollsSetupMenuTitleKey,
									Type:     discord.InteractionComponentTypeTextInput,
									Value:    poll.Title,
									Style:    discord.InteractionComponentStyleShort,
									Required: new(false),
								},
							},
							{
								Type:        discord.InteractionComponentTypeLabel,
								Label:       "Answers",
								Description: "One answer per line. Max of 10 answers is allowed.",
								Component: &discord.InteractionComponent{
									CustomID:    pollsSetupMenuAnswersKey,
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
									CustomID:    pollsSetupMenuDurationKey,
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
									CustomID: pollsSetupMenuOptionsKey,
									Type:     discord.InteractionComponentTypeCheckboxGroup,
									Options: []discord.ApplicationSelectOption{
										{
											Label:       "Allow Multiple Selections",
											Description: "You can limit the maximum number of answers a user can select in the next step.",
											Value:       pollsSetupMenuAllowMultipleAnswersKey,
										},
										{
											Label:       "Anonymous Poll",
											Description: "Resubmissions are not allowed and results will only be available when the poll ends.",
											Value:       pollsSetupMenuToggleAnonymousVotingKey,
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

	sub.RegisterComponentListener("poll_edit:*", handlePollEditComponent)

	cog.InteractionCommands.MustAddInteractionCommand(pollsGroup)

	return nil
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
			if titleArgument, err := subway.GetArgument(ctx, pollsSetupMenuTitleKey); err == nil {
				poll.Title = titleArgument.MustString()
			}

			if answersArgument, err := subway.GetArgument(ctx, pollsSetupMenuAnswersKey); err == nil {
				answers := answersArgument.MustString()
				answersList := strings.Split(answers, "\n")

				poll.PollOptions = pgtype.JSONB{
					Bytes:  welcomer.MarshalAnswersListJSON(answersList[:min(len(answersList), 10)]),
					Status: pgtype.Present,
				}
			}

			if durationArgument, err := subway.GetArgument(ctx, pollsSetupMenuDurationKey); err == nil {
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

			if optionsArgument, err := subway.GetArgument(ctx, pollsSetupMenuOptionsKey); err == nil {
				options := optionsArgument.MustStrings()

				for _, option := range options {
					switch option {
					case pollsSetupMenuAllowMultipleAnswersKey:
						poll.MaximumSelections = 0
					case pollsSetupMenuToggleAnonymousVotingKey:
						poll.IsAnonymous = true
					}
				}
			}
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
		Resubmissions:     poll.Resubmissions,
		ResultsVisibility: poll.ResultsVisibility,
		RolesAllowed:      poll.RolesAllowed,
		RolesExcluded:     poll.RolesExcluded,
		MinimumJoinDate:   poll.MinimumJoinDate,
		MessageID:         poll.MessageID,
		ChannelID:         poll.ChannelID,
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
				CustomID: customIDPrefix + pollsSetupMenuTitleKey,
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
				CustomID: customIDPrefix + pollsSetupMenuAnswersKey,
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
						"**Allow Multiple Selections:** " + welcomer.If(poll.MaximumSelections == 1, "No\n", "Yes"+welcomer.If(poll.MaximumSelections == 0, "", " ("+strconv.Itoa(int(poll.MaximumSelections))+")")+"\n") +
						"**Anonymous Poll:** " + welcomer.If(poll.IsAnonymous, "Yes\n", "No\n") +
						"**Resubmissions:** " +
						welcomer.If(poll.IsAnonymous, "Not Allowed (anonymous poll)\n",
							welcomer.If(poll.Resubmissions == string(welcomer.PollResubmissionOptionAlways), "Allowed\n",
								welcomer.If(poll.Resubmissions == string(welcomer.PollResubmissionOptionOnlyAdditions), "Allowed, but only for adding new answers\n",
									welcomer.If(poll.Resubmissions == string(welcomer.PollResubmissionOptionNever), "Not Allowed\n", ""),
								),
							)) +
						"**Results Visibility:** " +
						welcomer.If(poll.IsAnonymous, "Hidden until poll ends (anonymous poll)\n",
							welcomer.If(poll.ResultsVisibility == string(welcomer.PollResultVisibilityOptionAlways), "Always Visible\n",
								welcomer.If(poll.ResultsVisibility == string(welcomer.PollResultVisibilityOptionAfterVoting), "Only visible after voting\n",
									welcomer.If(poll.ResultsVisibility == string(welcomer.PollResultVisibilityOptionAfterEnd), "Hidden until poll ends\n", ""),
								),
							),
						),
				},
			},
			Accessory: &discord.InteractionComponent{
				Type:     discord.InteractionComponentTypeButton,
				Style:    discord.InteractionComponentStyleSecondary,
				Label:    "Edit",
				CustomID: customIDPrefix + pollsSetupMenuOptionsKey,
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
						welcomer.If(poll.EndTime.IsZero(), "\n-# Poll will run until ended manually with `/poll end`.", ""),
				},
			},
			Accessory: &discord.InteractionComponent{
				Type:     discord.InteractionComponentTypeButton,
				Style:    discord.InteractionComponentStyleSecondary,
				Label:    "Edit",
				CustomID: customIDPrefix + pollsSetupMenuDurationKey,
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
				CustomID: customIDPrefix + pollsSetupMenuRolesAllowedKey,
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
						welcomer.If(!poll.MinimumJoinDate.IsZero(), "\n-# Users who have joined the server within "+welcomer.HumanizeDuration(int(poll.MinimumJoinDate.Unix()), true)+" of the poll starting cannot enter the poll.", ""),
				},
			},
			Accessory: &discord.InteractionComponent{
				Type:     discord.InteractionComponentTypeButton,
				Style:    discord.InteractionComponentStyleSecondary,
				Label:    "Edit",
				CustomID: customIDPrefix + pollsSetupMenuMinimumJoinDateKey,
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
						CustomID: customIDPrefix + pollsSetupMenuStartKey,
						Disabled: len(pollAnswers) == 0,
					},
					{
						Type:     discord.InteractionComponentTypeButton,
						Style:    discord.InteractionComponentStyleSecondary,
						Label:    "Preview",
						CustomID: customIDPrefix + pollsSetupMenuPreviewOnKey,
					},
				},
			},
		},
	}
}

func getPollAnswersAsString(pollAnswers []string) string {
	if len(pollAnswers) == 0 {
		return "No Answers Configured"
	}

	result := ""

	for i, answer := range pollAnswers {
		result += fmt.Sprintf("**%d.** %s\n", i+1, answer)
	}

	return result
}
