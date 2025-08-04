package welcomer

import (
	"context"

	"github.com/WelcomerTeam/Discord/discord"
	sandwich_protobuf "github.com/WelcomerTeam/Sandwich-Daemon/proto"
)

func FetchGuild(ctx context.Context, guildID discord.Snowflake) (*discord.Guild, error) {
	// Query state cache for guild.
	guilds, err := SandwichClient.FetchGuild(ctx, &sandwich_protobuf.FetchGuildRequest{
		GuildIds: []int64{int64(guildID)},
	})
	if err != nil {
		Logger.Error().Err(err).
			Int64("guild_id", int64(guildID)).
			Msg("Failed to fetch guild from state cache")
	}

	guildPb, ok := guilds.GetGuilds()[int64(guildID)]
	if ok {
		return sandwich_protobuf.PBToGuild(guildPb), nil
	}

	return nil, ErrMissingGuild
}

func FetchGuildChannels(ctx context.Context, guildID discord.Snowflake) ([]*discord.Channel, error) {
	channels, err := SandwichClient.FetchGuildChannel(ctx, &sandwich_protobuf.FetchGuildChannelRequest{
		GuildId: int64(guildID),
	})
	if err != nil {
		Logger.Error().Err(err).
			Int64("guild_id", int64(guildID)).
			Msg("Failed to fetch channels from state cache")
	}

	discordChannels := make([]*discord.Channel, 0, len(channels.GetChannels()))

	for _, channelPb := range channels.GetChannels() {
		discordChannels = append(discordChannels, sandwich_protobuf.PBToChannel(channelPb))
	}

	return discordChannels, nil
}

func FetchUser(ctx context.Context, userID discord.Snowflake) (*discord.User, error) {
	users, err := SandwichClient.FetchUser(ctx, &sandwich_protobuf.FetchUserRequest{
		UserIds: []int64{int64(userID)},
	})
	if err != nil {
		return nil, err
	}

	userPb, ok := users.GetUsers()[int64(userID)]
	if ok {
		return sandwich_protobuf.PBToUser(userPb), err
	}

	return nil, ErrMissingUser
}
