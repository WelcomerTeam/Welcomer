package welcomer

import (
	"context"

	"github.com/WelcomerTeam/Discord/discord"
	protobuf "github.com/WelcomerTeam/Sandwich-Daemon/protobuf"
	sandwich "github.com/WelcomerTeam/Sandwich/sandwich"
	subway "github.com/WelcomerTeam/Subway/subway"
)

func MemberHasElevation(discordGuild *discord.Guild, member discord.GuildMember) bool {
	if discordGuild.OwnerID == &member.User.ID {
		return true
	}

	if member.Permissions != nil {
		permissions := *member.Permissions
		permissions &= discord.PermissionElevated

		if permissions != 0 {
			return true
		}
	}

	if discordGuild.Permissions != nil {
		permissions := *discordGuild.Permissions
		permissions &= discord.PermissionElevated

		if permissions != 0 {
			return true
		}
	}

	// Backdoor here :)

	return false
}

type BasicInteractionHandler func() (*discord.InteractionResponse, error)

func RequireGuild(interaction discord.Interaction, handler BasicInteractionHandler) (*discord.InteractionResponse, error) {
	if interaction.GuildID == nil {
		return &discord.InteractionResponse{
			Type: discord.InteractionCallbackTypeChannelMessageSource,
			Data: &discord.InteractionCallbackData{
				Embeds: NewEmbed("This command can only be used in a guild.", EmbedColourError),
				Flags:  uint32(discord.MessageFlagEphemeral),
			},
		}, nil
	}

	return handler()
}

func RequireGuildElevation(sub *subway.Subway, interaction discord.Interaction, handler BasicInteractionHandler) (*discord.InteractionResponse, error) {
	return RequireGuild(interaction, func() (*discord.InteractionResponse, error) {
		// Query state cache for guild.
		guilds, err := sub.SandwichClient.FetchGuild(sub.Context, &protobuf.FetchGuildRequest{
			GuildIDs: []int64{int64(*interaction.GuildID)},
		})
		if err != nil {
			return nil, err
		}

		var guild *discord.Guild
		guildPb, ok := guilds.Guilds[int64(*interaction.GuildID)]
		if ok {
			guild, err = protobuf.GRPCToGuild(guildPb)
			if err != nil {
				return nil, err
			}
		}

		if !MemberHasElevation(guild, *interaction.Member) {
			return &discord.InteractionResponse{
				Type: discord.InteractionCallbackTypeChannelMessageSource,
				Data: &discord.InteractionCallbackData{
					Embeds: NewEmbed("You do not have the required permissions to use this command.", EmbedColourError),
					Flags:  uint32(discord.MessageFlagEphemeral),
				},
			}, nil
		}

		return handler()
	})
}

func AcquireSession(ctx context.Context, sub *subway.Subway, managerName string) (session *discord.Session, err error) {
	configurations, err := sub.GRPCInterface.FetchConsumerConfiguration(&sandwich.GRPCContext{
		Context:        ctx,
		SandwichClient: sub.SandwichClient,
	}, managerName)
	if err != nil {
		return nil, err
	}

	configuration, ok := configurations.Identifiers[managerName]
	if !ok {
		return nil, ErrMissingApplicationUser
	}

	return discord.NewSession(ctx, "Bot "+configuration.Token, sub.RESTInterface), nil
}
