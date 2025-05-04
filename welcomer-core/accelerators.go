package welcomer

import (
	"context"

	"github.com/WelcomerTeam/Discord/discord"
	pb "github.com/WelcomerTeam/Sandwich-Daemon/protobuf"
)

func FetchGuild(ctx context.Context, guildID discord.Snowflake) (discord.Guild, error) {
	// Query state cache for guild.
	guilds, err := SandwichClient.FetchGuild(ctx, &pb.FetchGuildRequest{
		GuildIDs: []int64{int64(guildID)},
	})
	if err != nil {
		Logger.Error().Err(err).
			Int64("guild_id", int64(guildID)).
			Msg("Failed to fetch guild from state cache")
	}

	var guild discord.Guild

	guildPb, ok := guilds.Guilds[int64(guildID)]
	if ok {
		guild, err = pb.GRPCToGuild(guildPb)
		if err != nil {
			Logger.Error().Err(err).
				Int64("guild_id", int64(guildID)).
				Msg("Failed to convert guild from protobuf")
		}
	} else {
		return guild, ErrMissingGuild
	}

	return guild, nil
}

func FetchGuildChannels(ctx context.Context, guildID discord.Snowflake) ([]discord.Channel, error) {
	channels, err := SandwichClient.FetchGuildChannels(ctx, &pb.FetchGuildChannelsRequest{
		GuildID: int64(guildID),
	})
	if err != nil {
		Logger.Error().Err(err).
			Int64("guild_id", int64(guildID)).
			Msg("Failed to fetch channels from state cache")
	}

	var discordChannels []discord.Channel

	for _, channelPb := range channels.GetGuildChannels() {
		channel, err := pb.GRPCToChannel(channelPb)
		if err != nil {
			Logger.Error().Err(err).
				Int64("guild_id", int64(guildID)).
				Int64("channel_id", int64(channelPb.ID)).
				Msg("Failed to convert channel from protobuf")
			continue
		}

		discordChannels = append(discordChannels, channel)
	}

	return discordChannels, nil
}

func FetchUser(ctx context.Context, userID discord.Snowflake, createDMChannel bool) (discord.User, error) {
	var users *pb.UsersResponse

	var user discord.User

	users, err := SandwichClient.FetchUsers(ctx, &pb.FetchUsersRequest{
		UserIDs:         []int64{int64(userID)},
		CreateDMChannel: createDMChannel,
	})
	if err != nil {
		return user, err
	}

	userPb, ok := users.Users[int64(userID)]
	if ok {
		var pUser discord.User

		pUser, err = pb.GRPCToUser(userPb)
		if err != nil {
			return user, err
		}

		user = pUser
	}

	return user, nil
}
