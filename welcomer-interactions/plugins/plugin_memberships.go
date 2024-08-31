package plugins

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/WelcomerTeam/Discord/discord"
	pb "github.com/WelcomerTeam/Sandwich-Daemon/protobuf"
	sandwich "github.com/WelcomerTeam/Sandwich/sandwich"
	subway "github.com/WelcomerTeam/Subway/subway"
	welcomer "github.com/WelcomerTeam/Welcomer/welcomer-core"
	"github.com/WelcomerTeam/Welcomer/welcomer-core/database"
	utils "github.com/WelcomerTeam/Welcomer/welcomer-utils"
	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v4"
)

func NewMembershipCog() *MembershipCog {
	return &MembershipCog{
		InteractionCommands: subway.SetupInteractionCommandable(&subway.InteractionCommandable{}),
	}
}

type MembershipCog struct {
	InteractionCommands *subway.InteractionCommandable
}

// Assert types.

var (
	_ subway.Cog                        = (*MembershipCog)(nil)
	_ subway.CogWithInteractionCommands = (*MembershipCog)(nil)
)

const (
	NoMembershipsAvailable = "no_memberships"
)

func (p *MembershipCog) CogInfo() *subway.CogInfo {
	return &subway.CogInfo{
		Name:        "Memberships",
		Description: "Manage memberships.",
	}
}

func (p *MembershipCog) GetInteractionCommandable() *subway.InteractionCommandable {
	return p.InteractionCommands
}

type UserMembership struct {
	MembershipUUID   uuid.UUID
	GuildID          int64
	GuildName        string
	ExpiresAt        time.Time
	MembershipStatus database.MembershipStatus
	MembershipType   database.MembershipType
}

func getUserMembershipsByUserID(ctx context.Context, sub *subway.Subway, userID discord.Snowflake) ([]UserMembership, error) {
	queries := welcomer.GetQueriesFromContext(ctx)

	memberships, err := queries.GetUserMembershipsByUserID(ctx, int64(userID))
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		sub.Logger.Error().Err(err).
			Int64("user_id", int64(userID)).
			Msg("Failed to get user memberships")

		return nil, err
	}

	if len(memberships) == 0 {
		return []UserMembership{}, nil
	}

	var guildIDs []int64

	for _, membership := range memberships {
		if membership.GuildID != 0 {
			guildIDs = append(guildIDs, membership.GuildID)
		}
	}

	var guilds map[int64]*pb.Guild

	// Fetch all guilds in one request.
	if len(guildIDs) > 0 {
		guildResponse, err := sub.SandwichClient.FetchGuild(ctx, &pb.FetchGuildRequest{
			GuildIDs: guildIDs,
		})
		if err != nil {
			sub.Logger.Error().Err(err).
				Msg("Failed to fetch guilds via GRPC")

			guilds = map[int64]*pb.Guild{}
		} else {
			guilds = guildResponse.Guilds
		}
	} else {
		guilds = map[int64]*pb.Guild{}
	}

	userMemberships := make([]UserMembership, 0, len(memberships))

	for _, membership := range memberships {
		var guildName string

		if guild, ok := guilds[membership.GuildID]; ok {
			guildName = guild.Name
		} else {
			guildName = fmt.Sprintf("Unknown Guild %d", membership.GuildID)
		}

		userMemberships = append(userMemberships, UserMembership{
			MembershipUUID:   membership.MembershipUuid,
			GuildID:          membership.GuildID,
			GuildName:        guildName,
			ExpiresAt:        membership.ExpiresAt,
			MembershipStatus: database.MembershipStatus(membership.Status),
			MembershipType:   database.MembershipType(membership.MembershipType),
		})
	}

	return userMemberships, nil
}

func (p *MembershipCog) RegisterCog(sub *subway.Subway) error {
	membershipGroup := subway.NewSubcommandGroup(
		"membership",
		"Manage your welcomer subscriptions.",
	)

	membershipGroup.MustAddInteractionCommand(&subway.InteractionCommandable{
		Name:        "list",
		Description: "Lists all membershipss you have available.",

		Handler: func(ctx context.Context, sub *subway.Subway, interaction discord.Interaction) (*discord.InteractionResponse, error) {
			var userID discord.Snowflake
			if interaction.Member != nil {
				userID = interaction.Member.User.ID
			} else {
				userID = interaction.User.ID
			}

			memberships, err := getUserMembershipsByUserID(ctx, sub, userID)
			if err != nil {
				sub.Logger.Error().Err(err).
					Int64("user_id", int64(userID)).
					Msg("Failed to get user memberships")

				return nil, err
			}

			if len(memberships) == 0 {
				return &discord.InteractionResponse{
					Type: discord.InteractionCallbackTypeChannelMessageSource,
					Data: &discord.InteractionCallbackData{
						Embeds: []discord.Embed{
							{
								Title:       "Memberships",
								Description: "You don't have any memberships.",
								Color:       utils.EmbedColourInfo,
							},
						},
						Components: []discord.InteractionComponent{
							{
								Type: discord.InteractionComponentTypeActionRow,
								Components: []discord.InteractionComponent{
									{
										Type:  discord.InteractionComponentTypeButton,
										Style: discord.InteractionComponentStyleLink,
										Label: "Get Welcomer Pro",
										URL:   welcomer.WebsiteURL + "/premium",
									},
								},
							},
						},
					},
				}, nil
			}

			embeds := []discord.Embed{}
			embed := discord.Embed{Title: "Your Memberships", Color: utils.EmbedColourInfo}

			for _, membership := range memberships {
				switch membership.MembershipStatus {
				case database.MembershipStatusRefunded,
					database.MembershipStatusRemoved,
					database.MembershipStatusUnknown:
					continue
				}

				embed.Fields = append(embed.Fields, discord.EmbedField{
					Name: fmt.Sprintf(
						"%s – %s%s",
						membership.MembershipType.Label(),
						membership.MembershipStatus.Label(),
						utils.If(
							membership.ExpiresAt.After(time.Now()) && !welcomer.IsCustomBackgroundsMembership(membership.MembershipType),
							fmt.Sprintf(" (Expires **<t:%d:R>**)", membership.ExpiresAt.Unix()),
							"",
						),
					),
					Value:  utils.If(membership.GuildID != 0, fmt.Sprintf("%s `%d`", membership.GuildName, membership.GuildID), "Unassigned"),
					Inline: false,
				})
			}

			embeds = append(embeds, embed)

			return &discord.InteractionResponse{
				Type: discord.InteractionCallbackTypeChannelMessageSource,
				Data: &discord.InteractionCallbackData{
					Embeds: embeds,
					Flags:  uint32(discord.MessageFlagEphemeral),
					Components: []discord.InteractionComponent{
						{
							Type: discord.InteractionComponentTypeActionRow,
							Components: []discord.InteractionComponent{
								{
									Type:  discord.InteractionComponentTypeButton,
									Style: discord.InteractionComponentStyleLink,
									Label: "Get Welcomer Pro",
									URL:   welcomer.WebsiteURL + "/premium",
								},
							},
						},
					},
				},
			}, nil
		},
	})

	membershipGroup.MustAddInteractionCommand(&subway.InteractionCommandable{
		Name:        "add",
		Description: "Add a membership to a server.",

		AutocompleteHandler: func(ctx context.Context, sub *subway.Subway, interaction discord.Interaction) ([]discord.ApplicationCommandOptionChoice, error) {
			var userID discord.Snowflake
			if interaction.Member != nil {
				userID = interaction.Member.User.ID
			} else {
				userID = interaction.User.ID
			}

			memberships, err := getUserMembershipsByUserID(ctx, sub, userID)
			if err != nil {
				sub.Logger.Error().Err(err).
					Int64("user_id", int64(interaction.User.ID)).
					Msg("Failed to get user memberships")
			}

			choices := make([]discord.ApplicationCommandOptionChoice, 0, len(memberships))

			for _, membership := range memberships {
				switch membership.MembershipStatus {
				case database.MembershipStatusUnknown,
					database.MembershipStatusRefunded,
					database.MembershipStatusExpired:
					continue
				}

				choices = append(choices, discord.ApplicationCommandOptionChoice{
					Name: fmt.Sprintf(
						"%s – %s%s",
						membership.MembershipType.Label(),
						membership.MembershipStatus.Label(),
						utils.If(
							membership.GuildID != 0,
							" (Assigned to "+membership.GuildName+")",
							"",
						),
					),
					Value: utils.StringToJsonLiteral(membership.MembershipUUID.String()),
				})
			}

			if len(choices) == 0 {
				choices = append(choices, discord.ApplicationCommandOptionChoice{
					Name:  "No memberships are available",
					Value: utils.StringToJsonLiteral(NoMembershipsAvailable),
				})
			}

			return choices, nil
		},

		ArgumentParameter: []subway.ArgumentParameter{
			{
				Required:     true,
				ArgumentType: subway.ArgumentTypeString,
				Name:         "membership",
				Description:  "The membership to add.",
				Autocomplete: &utils.True,
			},
			{
				Required:     false,
				ArgumentType: subway.ArgumentTypeGuild,
				Name:         "guild",
				Description:  "The guild to add the membership to.",
			},
		},

		Handler: func(ctx context.Context, sub *subway.Subway, interaction discord.Interaction) (*discord.InteractionResponse, error) {
			membershipUuidString := subway.MustGetArgument(ctx, "membership").MustString()

			if membershipUuidString == NoMembershipsAvailable {
				return &discord.InteractionResponse{
					Type: discord.InteractionCallbackTypeChannelMessageSource,
					Data: &discord.InteractionCallbackData{
						Embeds: utils.NewEmbed("You do not have any memberships available.", utils.EmbedColourError),
						Flags:  uint32(discord.MessageFlagEphemeral),
						Components: []discord.InteractionComponent{
							{
								Type: discord.InteractionComponentTypeActionRow,
								Components: []discord.InteractionComponent{
									{
										Type:  discord.InteractionComponentTypeButton,
										Style: discord.InteractionComponentStyleLink,
										Label: "Get Welcomer Pro",
										URL:   welcomer.WebsiteURL + "/premium",
									},
								},
							},
						},
					},
				}, nil
			}

			membershipUuid := uuid.UUID{}

			guild := subway.MustGetArgument(ctx, "guild").MustGuild()

			var err error

			if guild.ID.IsNil() {
				if interaction.GuildID == nil {
					return &discord.InteractionResponse{
						Type: discord.InteractionCallbackTypeChannelMessageSource,
						Data: &discord.InteractionCallbackData{
							Embeds: utils.NewEmbed("You must specify a guild ID or run the command in the guild you would like to add the membership to.", utils.EmbedColourError),
							Flags:  uint32(discord.MessageFlagEphemeral),
						},
					}, nil
				}

				guild.ID = *interaction.GuildID

				sGuild, err := sandwich.FetchGuild(sub.NewGRPCContext(ctx), &guild)
				if err != nil {
					sub.Logger.Warn().Err(err).
						Int64("guild_id", int64(guild.ID)).
						Msg("Failed to fetch guild")
				}

				if sGuild != nil {
					guild = *sGuild
				}
			}

			if guild.Name == "" {
				guild.Name = fmt.Sprintf("Unknown Guild `%d`", guild.ID)
			}

			err = membershipUuid.Parse(membershipUuidString)
			if err != nil {
				sub.Logger.Error().Err(err).
					Str("membership_uuid", membershipUuidString).
					Msg("Failed to parse membership UUID")

				return nil, err
			}

			var userID discord.Snowflake
			if interaction.Member != nil {
				userID = interaction.Member.User.ID
			} else {
				userID = interaction.User.ID
			}

			memberships, err := getUserMembershipsByUserID(ctx, sub, userID)
			if err != nil {
				sub.Logger.Error().Err(err).
					Int64("user_id", int64(interaction.User.ID)).
					Msg("Failed to get user memberships")

				return nil, err
			}

			for _, membership := range memberships {
				if membership.MembershipUUID.String() == membershipUuidString {
					if interaction.GuildID != nil && membership.GuildID == int64(*interaction.GuildID) {
						return &discord.InteractionResponse{
							Type: discord.InteractionCallbackTypeChannelMessageSource,
							Data: &discord.InteractionCallbackData{
								Embeds: utils.NewEmbed("This membership is already in use by this server.", utils.EmbedColourInfo),
								Flags:  uint32(discord.MessageFlagEphemeral),
							},
						}, nil
					}

					if membership.GuildID != 0 {
						return &discord.InteractionResponse{
							Type: discord.InteractionCallbackTypeChannelMessageSource,
							Data: &discord.InteractionCallbackData{
								Embeds: utils.NewEmbed("This membership is already in use by `"+membership.GuildName+"`. Please use `/membership remove` to remove the existing membership before re-assigning it.", utils.EmbedColourWarn),
								Flags:  uint32(discord.MessageFlagEphemeral),
							},
						}, nil
					}

					switch membership.MembershipStatus {
					case database.MembershipStatusUnknown,
						database.MembershipStatusRefunded:
						return &discord.InteractionResponse{
							Type: discord.InteractionCallbackTypeChannelMessageSource,
							Data: &discord.InteractionCallbackData{
								Embeds: utils.NewEmbed("This membership is no longer valid.", utils.EmbedColourError),
								Flags:  uint32(discord.MessageFlagEphemeral),
							},
						}, nil
					case database.MembershipStatusExpired:
						return &discord.InteractionResponse{
							Type: discord.InteractionCallbackTypeChannelMessageSource,
							Data: &discord.InteractionCallbackData{
								Embeds: utils.NewEmbed("This membership has expired.", utils.EmbedColourWarn),
								Flags:  uint32(discord.MessageFlagEphemeral),
							},
						}, nil
					case database.MembershipStatusActive,
						database.MembershipStatusIdle:

						queries := welcomer.GetQueriesFromContext(ctx)

						membership, err := queries.GetUserMembership(ctx, membershipUuid)
						if err != nil {
							sub.Logger.Error().Err(err).
								Str("membership_uuid", membershipUuid.String()).
								Msg("Failed to get user membership")

							return nil, err
						}

						// Check if the transaction is completed.
						if database.TransactionStatus(membership.TransactionStatus.Int32) != database.TransactionStatusCompleted {
							sub.Logger.Error().
								Str("membership_uuid", membershipUuid.String()).
								Int32("transaction_status", membership.TransactionStatus.Int32).
								Msg("Membership transaction is not completed")

							return nil, utils.ErrTransactionNotComplete
						}

						membership.UpdatedAt = time.Now()
						membership.Status = int32(database.MembershipStatusActive)
						membership.GuildID = int64(guild.ID)

						isNewMembership := membership.StartedAt.IsZero()

						if membership.StartedAt.IsZero() {
							// If StartedAt is zero, the membership has not started yet.
							// Compare ExpiresAt against StartedAt to find the length of the membership.
							duration := membership.ExpiresAt.Sub(membership.StartedAt)

							membership.StartedAt = time.Now()
							membership.ExpiresAt = membership.StartedAt.Add(duration)
						} else {
							// If StartedAt is not zero, the membership has already started.
							// Do not modify the ExpiresAt, but reset the StartedAt.
							membership.StartedAt = time.Now()
						}

						_, err = queries.UpdateUserMembership(ctx, database.UpdateUserMembershipParams{
							MembershipUuid:  membership.MembershipUuid,
							StartedAt:       membership.StartedAt,
							ExpiresAt:       membership.ExpiresAt,
							Status:          membership.Status,
							TransactionUuid: membership.TransactionUuid,
							UserID:          membership.UserID,
							GuildID:         membership.GuildID,
						})
						if err != nil {
							sub.Logger.Error().Err(err).
								Str("membership_uuid", membershipUuid.String()).
								Msg("Failed to update user membership")

							return nil, err
						}

						if isNewMembership {
							return &discord.InteractionResponse{
								Type: discord.InteractionCallbackTypeChannelMessageSource,
								Data: &discord.InteractionCallbackData{
									Embeds: utils.NewEmbed(
										fmt.Sprintf(
											"🎉 Your membership has now been applied to `%s`.%s",
											guild.Name,
											utils.If(
												!welcomer.IsCustomBackgroundsMembership(database.MembershipType(membership.MembershipType)),
												fmt.Sprintf(
													" Your membership expires **<t:%d:R>**.",
													membership.ExpiresAt.Unix(),
												),
												"",
											),
										),
										utils.EmbedColourSuccess,
									),
								},
							}, nil
						} else {
							return &discord.InteractionResponse{
								Type: discord.InteractionCallbackTypeChannelMessageSource,
								Data: &discord.InteractionCallbackData{
									Embeds: utils.NewEmbed(
										fmt.Sprintf(
											"🎉 Your membership has now been applied to `%s`.%s",
											guild.Name,
											utils.If(
												!welcomer.IsCustomBackgroundsMembership(database.MembershipType(membership.MembershipType)),
												fmt.Sprintf(
													" You have used this membership previously and expires **<t:%d:R>**.",
													membership.ExpiresAt.Unix(),
												),
												"",
											),
										),
										utils.EmbedColourSuccess,
									),
								},
							}, nil
						}
					}
				}
			}

			return &discord.InteractionResponse{
				Type: discord.InteractionCallbackTypeChannelMessageSource,
				Data: &discord.InteractionCallbackData{
					Embeds: utils.NewEmbed("Invalid membership.", utils.EmbedColourError),
					Flags:  uint32(discord.MessageFlagEphemeral),
				},
			}, nil
		},
	})

	membershipGroup.MustAddInteractionCommand(&subway.InteractionCommandable{
		Name:        "remove",
		Description: "Removes a membership from a server.",

		AutocompleteHandler: func(ctx context.Context, sub *subway.Subway, interaction discord.Interaction) ([]discord.ApplicationCommandOptionChoice, error) {
			var userID discord.Snowflake
			if interaction.Member != nil {
				userID = interaction.Member.User.ID
			} else {
				userID = interaction.User.ID
			}

			memberships, err := getUserMembershipsByUserID(ctx, sub, userID)
			if err != nil {
				sub.Logger.Error().Err(err).
					Int64("user_id", int64(interaction.User.ID)).
					Msg("Failed to get user memberships")
			}

			choices := make([]discord.ApplicationCommandOptionChoice, 0, len(memberships))

			for _, membership := range memberships {
				if membership.GuildID == 0 {
					continue
				}

				switch membership.MembershipStatus {
				case database.MembershipStatusUnknown,
					database.MembershipStatusRefunded,
					database.MembershipStatusRemoved,
					database.MembershipStatusExpired:
					continue
				}

				choices = append(choices, discord.ApplicationCommandOptionChoice{
					Name: fmt.Sprintf(
						"%s – %s (Assigned to %s)",
						membership.MembershipType.Label(),
						membership.MembershipStatus.Label(),
						membership.GuildName,
					),
					Value: utils.StringToJsonLiteral(membership.MembershipUUID.String()),
				})
			}

			if len(choices) == 0 {
				choices = append(choices, discord.ApplicationCommandOptionChoice{
					Name:  "No active memberships are available",
					Value: utils.StringToJsonLiteral(NoMembershipsAvailable),
				})
			}

			return choices, nil
		},

		ArgumentParameter: []subway.ArgumentParameter{
			{
				Required:     true,
				ArgumentType: subway.ArgumentTypeString,
				Name:         "membership",
				Description:  "The membership to remove.",
				Autocomplete: &utils.True,
			},
		},

		Handler: func(ctx context.Context, sub *subway.Subway, interaction discord.Interaction) (*discord.InteractionResponse, error) {
			membershipUuidString := subway.MustGetArgument(ctx, "membership").MustString()

			if membershipUuidString == NoMembershipsAvailable {
				return &discord.InteractionResponse{
					Type: discord.InteractionCallbackTypeChannelMessageSource,
					Data: &discord.InteractionCallbackData{
						Embeds: utils.NewEmbed("You do not have any active memberships available.", utils.EmbedColourError),
						Flags:  uint32(discord.MessageFlagEphemeral),
						Components: []discord.InteractionComponent{
							{
								Type: discord.InteractionComponentTypeActionRow,
								Components: []discord.InteractionComponent{
									{
										Type:  discord.InteractionComponentTypeButton,
										Style: discord.InteractionComponentStyleLink,
										Label: "Get Welcomer Pro",
										URL:   welcomer.WebsiteURL + "/premium",
									},
								},
							},
						},
					},
				}, nil
			}

			membershipUuid := uuid.UUID{}

			err := membershipUuid.Parse(membershipUuidString)
			if err != nil {
				sub.Logger.Error().Err(err).
					Str("membership_uuid", membershipUuidString).
					Msg("Failed to parse membership UUID")

				return nil, err
			}

			var userID discord.Snowflake
			if interaction.Member != nil {
				userID = interaction.Member.User.ID
			} else {
				userID = interaction.User.ID
			}

			memberships, err := getUserMembershipsByUserID(ctx, sub, userID)
			if err != nil {
				sub.Logger.Error().Err(err).
					Int64("user_id", int64(interaction.User.ID)).
					Msg("Failed to get user memberships")

				return nil, err
			}

			for _, membership := range memberships {
				if membership.MembershipUUID.String() == membershipUuidString {
					if membership.GuildID == 0 {
						return &discord.InteractionResponse{
							Type: discord.InteractionCallbackTypeChannelMessageSource,
							Data: &discord.InteractionCallbackData{
								Embeds: utils.NewEmbed("This membership is not currently in use.", utils.EmbedColourInfo),
								Flags:  uint32(discord.MessageFlagEphemeral),
							},
						}, nil
					}

					switch membership.MembershipStatus {
					case database.MembershipStatusUnknown,
						database.MembershipStatusIdle,
						database.MembershipStatusActive,
						database.MembershipStatusExpired,
						database.MembershipStatusRefunded,
						database.MembershipStatusRemoved:
					}
					queries := welcomer.GetQueriesFromContext(ctx)

					membership, err := queries.GetUserMembership(ctx, membershipUuid)
					if err != nil {
						sub.Logger.Error().Err(err).
							Str("membership_uuid", membershipUuid.String()).
							Msg("Failed to get user membership")

						return nil, err
					}

					membership.UpdatedAt = time.Now()

					// Only set the status to Idle if the membership is currently Active.
					membership.Status = utils.If(membership.Status == int32(database.MembershipStatusActive), int32(database.MembershipStatusIdle), membership.Status)
					membership.GuildID = 0

					_, err = queries.UpdateUserMembership(ctx, database.UpdateUserMembershipParams{
						MembershipUuid:  membership.MembershipUuid,
						StartedAt:       membership.StartedAt,
						ExpiresAt:       membership.ExpiresAt,
						Status:          membership.Status,
						TransactionUuid: membership.TransactionUuid,
						UserID:          membership.UserID,
						GuildID:         membership.GuildID,
					})
					if err != nil {
						sub.Logger.Error().Err(err).
							Str("membership_uuid", membershipUuid.String()).
							Msg("Failed to update user membership")

						return nil, err
					}

					return &discord.InteractionResponse{
						Type: discord.InteractionCallbackTypeChannelMessageSource,
						Data: &discord.InteractionCallbackData{
							Embeds: utils.NewEmbed(
								fmt.Sprintf(
									"Your membership has been removed.%s",
									utils.If(
										!welcomer.IsCustomBackgroundsMembership(database.MembershipType(membership.MembershipType)),
										fmt.Sprintf(
											" This membership expires **<t:%d:R>**.",
											membership.ExpiresAt.Unix(),
										),
										"",
									),
								),
								utils.EmbedColourSuccess,
							),
						},
					}, nil
				}
			}

			return &discord.InteractionResponse{
				Type: discord.InteractionCallbackTypeChannelMessageSource,
				Data: &discord.InteractionCallbackData{
					Embeds: utils.NewEmbed("Invalid membership.", utils.EmbedColourError),
					Flags:  uint32(discord.MessageFlagEphemeral),
				},
			}, nil
		},
	})

	p.InteractionCommands.MustAddInteractionCommand(membershipGroup)

	return nil
}
