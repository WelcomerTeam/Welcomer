package welcomer

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/WelcomerTeam/Discord/discord"
	sandwich_protobuf "github.com/WelcomerTeam/Sandwich-Daemon/proto"
	subway "github.com/WelcomerTeam/Subway/subway"
	"github.com/WelcomerTeam/Welcomer/welcomer-core/database"
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

func CheckChannelGuild(ctx context.Context, c sandwich_protobuf.SandwichClient, guildID, channelID discord.Snowflake) (bool, error) {
	channels, err := c.FetchGuildChannel(ctx, &sandwich_protobuf.FetchGuildChannelRequest{
		GuildId: int64(guildID),
	})
	if err != nil {
		return false, err
	}

	for channel := range channels.GetChannels() {
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

func MemberHasElevation(discordGuild *discord.Guild, member *discord.GuildMember) bool {
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
				Embeds: NewEmbed("This command can only be used in a guild.", EmbedColourError),
				Flags:  uint32(discord.MessageFlagEphemeral),
			},
		}, nil
	}

	return handler()
}

func IsInterationAuthorElevated(sub *subway.Subway, interaction discord.Interaction) bool {
	// Query state cache for guild.
	guild, err := FetchGuild(sub.Context, *interaction.GuildID)
	if err != nil {
		Logger.Error().Err(err).
			Int64("guild_id", int64(*interaction.GuildID)).
			Msg("Failed to fetch guild from state cache")

		return false
	}

	if guild.ID.IsNil() {
		return false
	}

	if interaction.Member == nil || !MemberHasElevation(guild, interaction.Member) {
		return false
	}

	return true
}

func RequireGuildElevation(sub *subway.Subway, interaction discord.Interaction, handler BasicInteractionHandler) (*discord.InteractionResponse, error) {
	return RequireGuild(interaction, func() (*discord.InteractionResponse, error) {
		if !IsInterationAuthorElevated(sub, interaction) {
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

// EnsureGuild will create or update a guild entry. This requires RequireMutualGuild to be called.
func EnsureGuild(ctx context.Context, guildID discord.Snowflake) error {
	guild, _ := Queries.GetGuild(ctx, int64(guildID))
	if guild == nil || guild.GuildID != 0 {
		return nil
	}

	_, err := CreateGuildWithAudit(ctx, database.CreateGuildParams{
		GuildID:          int64(guildID),
		EmbedColour:      DefaultGuild.EmbedColour,
		SiteSplashUrl:    DefaultGuild.SiteSplashUrl,
		SiteStaffVisible: DefaultGuild.SiteStaffVisible,
		SiteGuildVisible: DefaultGuild.SiteGuildVisible,
		SiteAllowInvites: DefaultGuild.SiteAllowInvites,
	}, 0)
	if err != nil {
		return fmt.Errorf("failed to ensure guild: %w", err)
	}

	return nil
}

func AcquireSession(ctx context.Context, applicationIdentifier string) (*discord.Session, error) {
	// If we have a session in the context, use it.
	session, sessionInContext := GetSessionFromContext(ctx)
	if sessionInContext {
		return session, nil
	}

	configurations, err := SandwichClient.FetchApplication(ctx, &sandwich_protobuf.ApplicationIdentifier{
		ApplicationIdentifier: applicationIdentifier,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to fetch application: %v", err)
	}

	configuration, sessionInContext := configurations.GetApplications()[applicationIdentifier]
	if !sessionInContext {
		return nil, ErrMissingApplicationUser
	}

	return discord.NewSession("Bot "+configuration.BotToken, RESTInterface), nil
}
