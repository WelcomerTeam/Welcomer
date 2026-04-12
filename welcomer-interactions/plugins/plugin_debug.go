package plugins

import (
	"bytes"
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
		Name:        "json",
		Description: "Returns JSON payload for message",

		Type: subway.InteractionCommandableTypeSubcommand,

		DMPermission: new(true),

		ArgumentParameter: []subway.ArgumentParameter{
			{
				Name:         "message_id",
				Description:  "The ID of the message to fetch",
				ArgumentType: subway.ArgumentTypeString,
				Required:     true,
			},
		},

		Handler: func(ctx context.Context, sub *subway.Subway, interaction discord.Interaction) (*discord.InteractionResponse, error) {
			messageID := subway.MustGetArgument(ctx, "message_id").MustString()

			messageIDInt, err := welcomer.Atoi(messageID)
			if err != nil {
				return &discord.InteractionResponse{
					Type: discord.InteractionCallbackTypeChannelMessageSource,
					Data: &discord.InteractionCallbackData{
						Content: "Invalid message ID",
					},
				}, nil
			}

			session, err := welcomer.AcquireSession(ctx, welcomer.GetManagerNameFromContext(ctx))
			if err != nil {
				return nil, err
			}

			message, err := discord.GetChannelMessage(ctx, session, *interaction.ChannelID, discord.Snowflake(messageIDInt))
			if err != nil {
				return &discord.InteractionResponse{
					Type: discord.InteractionCallbackTypeChannelMessageSource,
					Data: &discord.InteractionCallbackData{
						Content: "Failed to fetch message",
					},
				}, nil
			}

			messageJSON, err := json.MarshalIndent(message, "", "  ")
			if err != nil {
				return &discord.InteractionResponse{
					Type: discord.InteractionCallbackTypeChannelMessageSource,
					Data: &discord.InteractionCallbackData{
						Content: "Failed to marshal message",
					},
				}, nil
			}

			err = interaction.SendResponse(ctx, session, discord.InteractionCallbackTypeChannelMessageSource, &discord.InteractionCallbackData{
				Files: []discord.File{
					{
						Name:        "message.json",
						ContentType: "application/json",
						Reader:      bytes.NewBuffer(messageJSON),
					},
				},
				Flags: uint32(discord.MessageFlagEphemeral),
			})
			if err != nil {
				return &discord.InteractionResponse{
					Type: discord.InteractionCallbackTypeChannelMessageSource,
					Data: &discord.InteractionCallbackData{
						Content: "Failed to send response",
					},
				}, nil
			}

			return nil, nil
		},
	})

	debugGroup.MustAddInteractionCommand(&subway.InteractionCommandable{
		Name:        "testinvite",
		Description: "Relays a BOT_ADD and JOIN event to consumers",

		Type: subway.InteractionCommandableTypeSubcommand,

		DMPermission:            new(false),
		DefaultMemberPermission: new(discord.Int64(welcomer.PermissionElevated)),

		Handler: func(ctx context.Context, sub *subway.Subway, interaction discord.Interaction) (*discord.InteractionResponse, error) {
			auditEvent := discord.AuditLogEntry{
				TargetID:   &interaction.ApplicationID,
				UserID:     &interaction.GetUser().ID,
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
		Description: "Relays a GUILD_MEMBER_ADD event to consumers",

		Type: subway.InteractionCommandableTypeSubcommand,

		ArgumentParameter: []subway.ArgumentParameter{
			{
				Required:     false,
				ArgumentType: subway.ArgumentTypeMember,
				Name:         "user",
				Description:  "The user to test",
			},
		},

		DMPermission:            new(false),
		DefaultMemberPermission: new(discord.Int64(welcomer.PermissionElevated)),

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
		Description: "Relays a GUILD_MEMBER_REMOVE event to consumers",

		Type: subway.InteractionCommandableTypeSubcommand,

		ArgumentParameter: []subway.ArgumentParameter{
			{
				Required:     false,
				ArgumentType: subway.ArgumentTypeMember,
				Name:         "user",
				Description:  "The user to test",
			},
		},

		DMPermission:            new(false),
		DefaultMemberPermission: new(discord.Int64(welcomer.PermissionElevated)),

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
