package plugins

import (
	"context"
	"encoding/json"

	"github.com/WelcomerTeam/Discord/discord"
	sandwich "github.com/WelcomerTeam/Sandwich-Daemon/proto"
	subway "github.com/WelcomerTeam/Subway/subway"
	"github.com/WelcomerTeam/Welcomer/welcomer-core"
)

func NewDebugCog() *DebugCog {
	return &DebugCog{
		InteractionCommands: subway.SetupInteractionCommandable(&subway.InteractionCommandable{}),
	}
}

type DebugCog struct {
	InteractionCommands *subway.InteractionCommandable
}

// Assert types.

var (
	_ subway.Cog                        = (*DebugCog)(nil)
	_ subway.CogWithInteractionCommands = (*DebugCog)(nil)
)

func (cog *DebugCog) CogInfo() *subway.CogInfo {
	return &subway.CogInfo{
		Name:        "Debug",
		Description: "Provides the functionality for the 'Debug' feature",
	}
}

func (cog *DebugCog) GetInteractionCommandable() *subway.InteractionCommandable {
	return cog.InteractionCommands
}

func (cog *DebugCog) RegisterCog(sub *subway.Subway) error {
	debugGroup := subway.NewSubcommandGroup(
		"debug",
		"Debugging commands",
	)

	debugGroup.MustAddInteractionCommand(&subway.InteractionCommandable{
		Name:        "testinvite",
		Description: "Relays a BOT_ADD and JOIN event to consumers",

		Type: subway.InteractionCommandableTypeSubcommand,

		DMPermission:            &welcomer.False,
		DefaultMemberPermission: welcomer.ToPointer(discord.Int64(discord.PermissionElevated)),

		Handler: func(ctx context.Context, sub *subway.Subway, interaction discord.Interaction) (*discord.InteractionResponse, error) {
			auditEvent := discord.AuditLogEntry{
				TargetID:   &interaction.ApplicationID,
				UserID:     &interaction.Member.User.ID,
				ActionType: discord.AuditLogActionBotAdd,
			}

			data, err := json.Marshal(auditEvent)
			if err != nil {
				return nil, err
			}

			_, err = sub.SandwichClient.RelayMessage(ctx, &sandwich.RelayMessageRequest{
				Identifier: welcomer.GetManagerNameFromContext(ctx),
				Type:       discord.DiscordEventGuildAuditLogEntryCreate,
				Data:       data,
			})
			if err != nil {
				return nil, err
			}

			data, err = json.Marshal(discord.Guild{ID: *interaction.GuildID})
			if err != nil {
				return nil, err
			}

			_, err = sub.SandwichClient.RelayMessage(ctx, &sandwich.RelayMessageRequest{
				Identifier: welcomer.GetManagerNameFromContext(ctx),
				Type:       discord.DiscordEventGuildJoin,
				Data:       data,
			})
			if err != nil {
				return nil, err
			}

			return &discord.InteractionResponse{
				Type: discord.InteractionCallbackTypeChannelMessageSource,
				Data: &discord.InteractionCallbackData{
					Embeds: welcomer.NewEmbed("Event relayed", welcomer.EmbedColourSuccess),
				},
			}, nil
		},
	})

	debugGroup.MustAddInteractionCommand(&subway.InteractionCommandable{
		Name:        "testjoin",
		Description: "Relays an GUILD_MEMBER_ADD event to consumers",

		Type: subway.InteractionCommandableTypeSubcommand,

		ArgumentParameter: []subway.ArgumentParameter{
			{
				Required:     false,
				ArgumentType: subway.ArgumentTypeMember,
				Name:         "user",
				Description:  "The user to test",
			},
		},

		DMPermission:            &welcomer.False,
		DefaultMemberPermission: welcomer.ToPointer(discord.Int64(discord.PermissionElevated)),

		Handler: func(ctx context.Context, sub *subway.Subway, interaction discord.Interaction) (*discord.InteractionResponse, error) {
			return welcomer.RequireGuildElevation(sub, interaction, func() (*discord.InteractionResponse, error) {
				member := subway.MustGetArgument(ctx, "user").MustMember()
				if member.User == nil || member.User.ID.IsNil() {
					member = *interaction.Member
				}

				// GuildID may be missing, fill it in.
				member.GuildID = interaction.GuildID

				data, err := json.Marshal(member)
				if err != nil {
					return nil, err
				}

				_, err = sub.SandwichClient.RelayMessage(ctx, &sandwich.RelayMessageRequest{
					Identifier: welcomer.GetManagerNameFromContext(ctx),
					Type:       discord.DiscordEventGuildMemberAdd,
					Data:       data,
				})
				if err != nil {
					return nil, err
				}

				return &discord.InteractionResponse{
					Type: discord.InteractionCallbackTypeChannelMessageSource,
					Data: &discord.InteractionCallbackData{
						Embeds: welcomer.NewEmbed("Event relayed", welcomer.EmbedColourSuccess),
					},
				}, nil
			})
		},
	})

	debugGroup.MustAddInteractionCommand(&subway.InteractionCommandable{
		Name:        "testleave",
		Description: "Relays an GUILD_MEMBER_REMOVE event to consumers",

		Type: subway.InteractionCommandableTypeSubcommand,

		ArgumentParameter: []subway.ArgumentParameter{
			{
				Required:     false,
				ArgumentType: subway.ArgumentTypeMember,
				Name:         "user",
				Description:  "The user to test",
			},
		},

		DMPermission:            &welcomer.False,
		DefaultMemberPermission: welcomer.ToPointer(discord.Int64(discord.PermissionElevated)),

		Handler: func(ctx context.Context, sub *subway.Subway, interaction discord.Interaction) (*discord.InteractionResponse, error) {
			return welcomer.RequireGuildElevation(sub, interaction, func() (*discord.InteractionResponse, error) {
				member := subway.MustGetArgument(ctx, "user").MustMember()
				if member.User == nil || member.User.ID.IsNil() {
					member = *interaction.Member
				}

				// GuildID may be missing, fill it in.
				member.GuildID = interaction.GuildID

				data, err := json.Marshal(discord.GuildMemberRemove{
					User:    *member.User,
					GuildID: *member.GuildID,
				})
				if err != nil {
					return nil, err
				}

				_, err = sub.SandwichClient.RelayMessage(ctx, &sandwich.RelayMessageRequest{
					Identifier: welcomer.GetManagerNameFromContext(ctx),
					Type:       discord.DiscordEventGuildMemberRemove,
					Data:       data,
				})
				if err != nil {
					return nil, err
				}

				return &discord.InteractionResponse{
					Type: discord.InteractionCallbackTypeChannelMessageSource,
					Data: &discord.InteractionCallbackData{
						Embeds: welcomer.NewEmbed("Event relayed", welcomer.EmbedColourSuccess),
					},
				}, nil
			})
		},
	})

	cog.InteractionCommands.MustAddInteractionCommand(debugGroup)

	return nil
}
