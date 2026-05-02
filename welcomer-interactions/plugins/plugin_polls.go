package plugins

import (
	"context"
	"fmt"
	"math"
	"math/rand"
	"slices"
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
	pollSetupMenuTitleKey        = "title"
	pollSetupMenuDescriptionKey  = "description"
	pollSetupMenuAccentColourKey = "accent_colour"
	pollSetupMenuThumbnailURLKey = "thumbnail_url"

	pollSetupMenuAnswersKey               = "answers"
	pollSetupMenuOptionsKey               = "options"
	pollSetupMenuDurationKey              = "duration"
	pollSetupMenuToggleAnonymousVotingKey = "toggle_anonymous_voting"
	pollSetupMenuMaximumAnswersKey        = "maximum_answers"
	pollSetupMenuMaximumAnswersValueKey   = "maximum_answers_value"

	// stub option during initial modal to set it to the maximum answers.
	pollSetupMenuAllowMultipleAnswersKey = "allow_multiple_answers"

	pollSetupMenuRolesAllowedKey         = "roles_allowed"
	pollSetupMenuRolesAllowedIncludedKey = "roles_allowed_included"
	pollSetupMenuRolesAllowedExcludedKey = "roles_allowed_excluded"

	pollSetupMenuManageResubmissionsKey = "manage_resubmissions"
	pollSetupMenuNoResubmissionsKey     = "no_resubmissions"
	pollSetupMenuAllowAdditionsOnlyKey  = "allow_additions_only"
	pollSetupMenuAllowResubmissionsKey  = "allow_resubmissions"

	pollSetupMenuShowResultsKey                   = "show_results"
	pollSetupMenuResultsAlwaysVisibleKey          = "results_always_visible"
	pollSetupMenuResultsVisibleAfterVotingKey     = "results_visible_after_voting"
	pollSetupMenuResultsVisibleAfterVotingEndsKey = "results_visible_after_voting_ends"

	pollSetupMenuMinimumJoinDateKey = "minimum_join_date"
	pollSetupMenuStartKey           = "start"

	pollSetupMenuPreviewOnKey  = "preview_on"
	pollSetupMenuPreviewOffKey = "preview_off"

	pollSetupMenuPingKey                    = "ping"
	pollSetupMenuPingEveryoneKey            = "ping_everyone"
	pollSetupMenuPingHereKey                = "ping_here"
	pollSetupMenuPingRolesAllowedToEnterKey = "ping_roles_allowed_to_enter"
	pollSetupMenuPingAdditionalRolesKey     = "additional_roles_to_ping"
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
							Content: "Once started, the poll message will be sent and entries will be allowed. You can end or extend the poll at any time, but you cannot edit the poll settings.\n\nBelow you can configure who should be pinged when the poll starts.",
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
				results[i] = rand.Intn(100)
			}

			message := pollView(poll, results)

			// Hack to disable poll button and add back button
			message.Components[len(message.Components)-1].Components[0].Disabled = true
			message.Components[len(message.Components)-1].Components = append(message.Components[len(message.Components)-1].Components, discord.InteractionComponent{
				CustomID: "poll_edit:" + poll.PollUuid.String() + ":preview_off",
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

			message, err := interaction.Channel.Send(ctx, session, welcomer.WebhookMessageParamsToMessageParams(pollView(poll, results)))
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

func pollView(poll *database.GuildPolls, results []int) discord.WebhookMessageParams {
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

	maxValue := 0
	entries := 0

	for _, result := range results {
		entries += result
		if result > maxValue {
			maxValue = result
		}
	}

	maxValue = int(math.Ceil(float64(maxValue)/4) * 4)

	answersString := "**Answers:**\n"
	answers := welcomer.UnmarshalAnswersListJSON(poll.PollOptions.Bytes)

	switch {
	case poll.ResultsVisibility == string(welcomer.PollResultVisibilityOptionAlways),
		poll.ResultsVisibility == string(welcomer.PollResultVisibilityOptionAfterEnd) && time.Now().Before(poll.EndTime):
		// Show answers and percentages
		for i, answer := range answers {
			truePercentage := float64(results[i]) / float64(entries) * 100
			resultPercentage := int(float64(results[i]) / float64(maxValue) * 100)

			answersString += fmt.Sprintf("\n%s (**%d - %.1f**%%)\n%s\n", answer, results[i], truePercentage, getEmojiCombination(resultPercentage, 10))
		}
	case poll.ResultsVisibility == string(welcomer.PollResultVisibilityOptionAfterEnd) && time.Now().After(poll.EndTime),
		poll.ResultsVisibility == string(welcomer.PollResultVisibilityOptionAfterVoting):
		// Show answers
		for _, answer := range answers {
			answersString += fmt.Sprintf("- %s\n", answer)
		}
	}

	if !poll.IsAnonymous {
		answersString += "\n**Total Entries:** " + strconv.Itoa(entries) + "\n"
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
						Label:    "Answer Poll",
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

	for i, answer := range pollAnswers {
		result += fmt.Sprintf("- %s\n", i+1, answer)
	}

	return result
}

var sectionEmojiIDs = [][]string{}

const maxSegmentsPerGroup = 4

func getEmojiCombination(value_of_100 int, length int) string {
	if value_of_100 <= 0 || length <= 0 {
		return ""
	}

	segments := int(math.Ceil(float64(value_of_100) * float64(length) / 25))

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
