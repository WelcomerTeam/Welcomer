package plugins

import (
	"context"
	"errors"
	"fmt"

	"github.com/WelcomerTeam/Discord/discord"
	sandwich "github.com/WelcomerTeam/Sandwich-Daemon/protobuf"
	subway "github.com/WelcomerTeam/Subway/subway"
	welcomer "github.com/WelcomerTeam/Welcomer/welcomer-core"
	"github.com/WelcomerTeam/Welcomer/welcomer-core/database"
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
	MembershipModuleCustomBackgrounds = "custom_backgrounds"
	MembershipModulePro               = "pro"
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
	MembershipStatus database.MembershipStatus
	MembershipType   database.MembershipType
}

func getUserMembershipsByUserID(ctx context.Context, sub *subway.Subway, userID discord.Snowflake) ([]UserMembership, error) {
	queries := welcomer.GetQueriesFromContext(ctx)

	memberships, err := queries.GetUserMembershipsByUserID(ctx, int64(userID))
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		sub.Logger.Error().Err(err).
			Int64("user_id", int64(userID)).
			Msg("Failed to get user memberships.")

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

	var guilds map[int64]*sandwich.Guild

	// Fetch all guilds in one request.
	if len(guildIDs) > 0 {
		guildResponse, err := sub.SandwichClient.FetchGuild(ctx, &sandwich.FetchGuildRequest{
			GuildIDs: guildIDs,
		})
		if err != nil {
			sub.Logger.Error().Err(err).
				Msg("Failed to fetch guilds via GRPC.")

			guilds = map[int64]*sandwich.Guild{}
		} else {
			guilds = guildResponse.Guilds
		}
	} else {
		guilds = map[int64]*sandwich.Guild{}
	}

	userMemberships := make([]UserMembership, 0, len(memberships))

	for _, membership := range memberships {
		var guildName string

		if guild, ok := guilds[membership.GuildID]; ok {
			guildName = guild.Name
		} else {
			guildName = fmt.Sprintf("Unknown Guild `%d`", membership.GuildID)
		}

		userMemberships = append(userMemberships, UserMembership{
			MembershipUUID:   membership.MembershipUuid,
			GuildID:          membership.GuildID,
			GuildName:        guildName,
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
					Msg("Failed to get user memberships.")

				return nil, err
			}

			if len(memberships) == 0 {
				return &discord.InteractionResponse{
					Type: discord.InteractionCallbackTypeChannelMessageSource,
					Data: &discord.InteractionCallbackData{
						Embeds: []*discord.Embed{
							{
								Title:       "Memberships",
								Description: "You don't have any memberships.",
								Color:       welcomer.EmbedColourInfo,
							},
						},
						Components: []*discord.InteractionComponent{
							{
								Type: discord.InteractionComponentTypeActionRow,
								Components: []*discord.InteractionComponent{
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

			embeds := []*discord.Embed{}
			embed := &discord.Embed{Title: "Your Memberships", Color: welcomer.EmbedColourInfo}

			for _, membership := range memberships {
				switch membership.MembershipStatus {
				case database.MembershipStatusRefunded,
					database.MembershipStatusRemoved,
					database.MembershipStatusUnknown:
					continue
				}

				embed.Fields = append(embed.Fields, &discord.EmbedField{
					Name:   fmt.Sprintf("%s – %s", membership.MembershipType.Label(), membership.MembershipStatus.Label()),
					Value:  welcomer.If(membership.GuildID != 0, fmt.Sprintf("Guild: %s `%d`", membership.GuildName, membership.GuildID), "Unassigned"),
					Inline: false,
				})
			}

			embeds = append(embeds, embed)

			return &discord.InteractionResponse{
				Type: discord.InteractionCallbackTypeChannelMessageSource,
				Data: &discord.InteractionCallbackData{
					Embeds: embeds,
					Flags:  uint32(discord.MessageFlagEphemeral),
					Components: []*discord.InteractionComponent{
						{
							Type: discord.InteractionComponentTypeActionRow,
							Components: []*discord.InteractionComponent{
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

		AutocompleteHandler: func(ctx context.Context, sub *subway.Subway, interaction discord.Interaction) ([]*discord.ApplicationCommandOptionChoice, error) {
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
					Msg("Failed to get user memberships.")
			}

			choices := make([]*discord.ApplicationCommandOptionChoice, 0, len(memberships))

			for _, membership := range memberships {
				switch membership.MembershipStatus {
				case database.MembershipStatusRefunded,
					database.MembershipStatusRemoved,
					database.MembershipStatusUnknown,
					database.MembershipStatusActive,
					database.MembershipStatusExpired:
					continue
				}

				choices = append(choices, &discord.ApplicationCommandOptionChoice{
					Name:  fmt.Sprintf("%s – %s", membership.MembershipType.Label(), membership.MembershipStatus.Label()),
					Value: welcomer.StringToJsonLiteral(membership.MembershipUUID.String()),
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
				Autocomplete: &welcomer.True,
			},
		},

		Handler: func(ctx context.Context, sub *subway.Subway, interaction discord.Interaction) (*discord.InteractionResponse, error) {
			membership := subway.MustGetArgument(ctx, "membership").MustString()

			println(membership)
			return nil, nil
		},
	})

	p.InteractionCommands.MustAddInteractionCommand(membershipGroup)

	return nil
}
