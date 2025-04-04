package welcomer

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/WelcomerTeam/Discord/discord"
	protobuf "github.com/WelcomerTeam/Sandwich-Daemon/protobuf"
	sandwich "github.com/WelcomerTeam/Sandwich/sandwich"
	subway "github.com/WelcomerTeam/Subway/subway"
	"github.com/WelcomerTeam/Welcomer/welcomer-core/database"
	utils "github.com/WelcomerTeam/Welcomer/welcomer-utils"
)

var elevatedUsers []discord.Snowflake

func init() {
	elevatedUsersStr := os.Getenv("ELEVATED_USERS")

	if elevatedUsersStr != "" {
		err := json.Unmarshal([]byte(elevatedUsersStr), &elevatedUsers)
		if err != nil {
			panic(fmt.Errorf("failed to parse ELEVATED_USERS: %w", err))
		}
	}
}

func CheckChannelGuild(ctx context.Context, c protobuf.SandwichClient, guildID, channelID discord.Snowflake) (bool, error) {
	channels, err := c.FetchGuildChannels(ctx, &protobuf.FetchGuildChannelsRequest{
		GuildID: int64(guildID),
	})
	if err != nil {
		return false, err
	}

	for channel := range channels.GuildChannels {
		if channel == int64(channelID) {
			return true, nil
		}
	}

	return false, nil
}

func CheckGuildMemberships(memberships []*database.GetUserMembershipsByGuildIDRow) (hasWelcomerPro, hasCustomBackgrounds bool) {
	for _, membership := range memberships {
		if IsCustomBackgroundsMembership(database.MembershipType(membership.MembershipType)) {
			hasCustomBackgrounds = true
		} else if IsWelcomerProMembership(database.MembershipType(membership.MembershipType)) {
			hasWelcomerPro = true
		}
	}

	return
}

func IsCustomBackgroundsMembership(membershipType database.MembershipType) bool {
	return membershipType == database.MembershipTypeLegacyCustomBackgrounds || membershipType == database.MembershipTypeCustomBackgrounds
}

func IsWelcomerProMembership(membershipType database.MembershipType) bool {
	return membershipType == database.MembershipTypeLegacyWelcomerPro || membershipType == database.MembershipTypeWelcomerPro
}

func MemberHasElevation(discordGuild discord.Guild, member discord.GuildMember) bool {
	if member.User != nil {
		for _, elevatedUser := range elevatedUsers {
			if member.User.ID == elevatedUser {
				return true
			}
		}
	}

	if discordGuild.Owner {
		return true
	}

	if discordGuild.OwnerID != nil && *discordGuild.OwnerID == member.User.ID {
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

	return false
}

type BasicInteractionHandler func() (*discord.InteractionResponse, error)

func RequireGuild(interaction discord.Interaction, handler BasicInteractionHandler) (*discord.InteractionResponse, error) {
	if interaction.GuildID == nil {
		return &discord.InteractionResponse{
			Type: discord.InteractionCallbackTypeChannelMessageSource,
			Data: &discord.InteractionCallbackData{
				Embeds: utils.NewEmbed("This command can only be used in a guild.", utils.EmbedColourError),
				Flags:  uint32(discord.MessageFlagEphemeral),
			},
		}, nil
	}

	return handler()
}

func RequireGuildElevation(sub *subway.Subway, interaction discord.Interaction, handler BasicInteractionHandler) (*discord.InteractionResponse, error) {
	return RequireGuild(interaction, func() (*discord.InteractionResponse, error) {
		// Query state cache for guild.
		guildsResponse, err := sub.SandwichClient.FetchGuild(sub.Context, &protobuf.FetchGuildRequest{
			GuildIDs: []int64{int64(*interaction.GuildID)},
		})
		if err != nil {
			return nil, err
		}

		var guild discord.Guild
		guildPb, ok := guildsResponse.Guilds[int64(*interaction.GuildID)]
		if ok {
			guild, err = protobuf.GRPCToGuild(guildPb)
			if err != nil {
				return nil, err
			}
		}

		if guild.ID.IsNil() {
			return nil, utils.ErrMissingGuild
		}

		if interaction.Member == nil || !MemberHasElevation(guild, *interaction.Member) {
			return &discord.InteractionResponse{
				Type: discord.InteractionCallbackTypeChannelMessageSource,
				Data: &discord.InteractionCallbackData{
					Embeds: utils.NewEmbed("You do not have the required permissions to use this command.", utils.EmbedColourError),
					Flags:  uint32(discord.MessageFlagEphemeral),
				},
			}, nil
		}

		return handler()
	})
}

// EnsureGuild will create or update a guild entry. This requires RequireMutualGuild to be called.
func EnsureGuild(ctx context.Context, queries *database.Queries, guildID discord.Snowflake) error {
	guild, _ := queries.GetGuild(ctx, int64(guildID))
	if guild == nil || guild.GuildID != 0 {
		return nil
	}

	_, err := queries.CreateGuild(ctx, database.CreateGuildParams{
		GuildID:          int64(guildID),
		EmbedColour:      database.DefaultGuild.EmbedColour,
		SiteSplashUrl:    database.DefaultGuild.SiteSplashUrl,
		SiteStaffVisible: database.DefaultGuild.SiteStaffVisible,
		SiteGuildVisible: database.DefaultGuild.SiteGuildVisible,
		SiteAllowInvites: database.DefaultGuild.SiteAllowInvites,
	})
	if err != nil {
		return fmt.Errorf("failed to ensure guild: %w", err)
	}

	return nil
}

func AcquireSession(ctx context.Context, grpcInterface sandwich.GRPC, restInterface discord.RESTInterface, client protobuf.SandwichClient, managerName string) (session *discord.Session, err error) {
	configurations, err := grpcInterface.FetchConsumerConfiguration(&sandwich.GRPCContext{
		Context:        ctx,
		SandwichClient: client,
	}, managerName)
	if err != nil {
		return nil, err
	}

	configuration, ok := configurations.Identifiers[managerName]
	if !ok {
		return nil, utils.ErrMissingApplicationUser
	}

	return discord.NewSession("Bot "+configuration.Token, restInterface), nil
}
