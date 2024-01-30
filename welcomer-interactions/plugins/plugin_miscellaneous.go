package plugins

import (
	"context"
	"fmt"
	"sort"
	"strconv"
	"time"

	"github.com/WelcomerTeam/Discord/discord"
	sandwich "github.com/WelcomerTeam/Sandwich-Daemon/protobuf"
	subway "github.com/WelcomerTeam/Subway/subway"
	"github.com/WelcomerTeam/Welcomer/welcomer-core"
)

func NewMiscellaneousCog() *MiscellaneousCog {
	return &MiscellaneousCog{
		InteractionCommands: subway.SetupInteractionCommandable(&subway.InteractionCommandable{}),
	}
}

type MiscellaneousCog struct {
	InteractionCommands *subway.InteractionCommandable
}

// Assert types.

var (
	_ subway.Cog                        = (*MiscellaneousCog)(nil)
	_ subway.CogWithInteractionCommands = (*MiscellaneousCog)(nil)
)

func (m *MiscellaneousCog) CogInfo() *subway.CogInfo {
	return &subway.CogInfo{
		Name:        "Miscellaneous",
		Description: "Miscellaneous commands.",
	}
}

func (m *MiscellaneousCog) GetInteractionCommandable() *subway.InteractionCommandable {
	return m.InteractionCommands
}

func (m *MiscellaneousCog) RegisterCog(sub *subway.Subway) error {
	m.InteractionCommands.MustAddInteractionCommand(&subway.InteractionCommandable{
		Name:        "dashboard",
		Description: "Get a link to the Welcomer dashboard",

		Handler: func(ctx context.Context, sub *subway.Subway, interaction discord.Interaction) (*discord.InteractionResponse, error) {
			if interaction.GuildID == nil {
				return &discord.InteractionResponse{
					Type: discord.InteractionCallbackTypeChannelMessageSource,
					Data: &discord.InteractionCallbackData{
						Embeds: []*discord.Embed{
							{
								Description: fmt.Sprintf("Manage your guild settings and memberships at %s", welcomer.WebsiteGuildURL("")),
								Color:       welcomer.EmbedColourInfo,
							},
						},
					},
				}, nil
			} else {
				return welcomer.RequireGuildElevation(sub, interaction, func() (*discord.InteractionResponse, error) {
					return &discord.InteractionResponse{
						Type: discord.InteractionCallbackTypeChannelMessageSource,
						Data: &discord.InteractionCallbackData{
							Embeds: []*discord.Embed{
								{
									Description: fmt.Sprintf("Manage this guild's settings and memberships [**here**](%s)", welcomer.WebsiteGuildURL(interaction.GuildID.String())),
									Color:       welcomer.EmbedColourInfo,
								},
							},
						},
					}, nil
				})
			}
		},
	})

	m.InteractionCommands.MustAddInteractionCommand(&subway.InteractionCommandable{
		Name:        "donate",
		Description: "Get Welcomer Pro and support Welcomer development",

		Handler: func(ctx context.Context, sub *subway.Subway, interaction discord.Interaction) (*discord.InteractionResponse, error) {
			return &discord.InteractionResponse{
				Type: discord.InteractionCallbackTypeChannelMessageSource,
				Data: &discord.InteractionCallbackData{
					Embeds: []*discord.Embed{
						{
							Description: "### **Everything you need to boost your guild's engagement**\n\nGet Welcomer Pro and support Welcomer development.\nFind out more [**here**](https://beta-dev.welcomer.gg/premium)",
							Color:       welcomer.EmbedColourInfo,
						},
					},
				},
			}, nil
		},
	})

	m.InteractionCommands.MustAddInteractionCommand(&subway.InteractionCommandable{
		Name:        "emojis",
		Description: "Get a list of all the emojis in the guild",

		Handler: func(ctx context.Context, sub *subway.Subway, interaction discord.Interaction) (*discord.InteractionResponse, error) {
			return welcomer.RequireGuild(interaction, func() (*discord.InteractionResponse, error) {
				embeds := []*discord.Embed{}
				embed := &discord.Embed{Title: "Emojis", Color: welcomer.EmbedColourInfo}

				guildEmojis, err := sub.SandwichClient.FetchGuildEmojis(sub.Context, &sandwich.FetchGuildEmojisRequest{
					GuildID: int64(*interaction.GuildID),
				})
				if err != nil {
					return nil, err
				}

				// Flatten map into slice
				emojis := make([]*sandwich.Emoji, 0, len(guildEmojis.GuildEmojis))
				for _, emoji := range guildEmojis.GuildEmojis {
					emojis = append(emojis, emoji)
				}

				// Sort emojis by animated and then by name
				sort.Slice(emojis, func(i, j int) bool {
					if emojis[i].Animated != emojis[j].Animated {
						return emojis[i].Animated
					}

					return emojis[i].Name < emojis[j].Name
				})

				for _, emoji := range emojis {
					var emojiLine string

					if emoji.Animated {
						emojiLine = fmt.Sprintf("- <a:%s:%d> **%s** `%d`\n", emoji.Name, emoji.ID, emoji.Name, emoji.ID)
					} else {
						emojiLine = fmt.Sprintf("- <:%s:%d> **%s** `%d`\n", emoji.Name, emoji.ID, emoji.Name, emoji.ID)
					}

					// If the embed content will go over 4000 characters then create a new embed and continue from that one.
					if len(embed.Description)+len(emojiLine) > 4000 {
						embeds = append(embeds, embed)
						embed = &discord.Embed{Color: welcomer.EmbedColourInfo}
					}

					embed.Description += emojiLine
				}

				embeds = append(embeds, embed)

				embeds[len(embeds)-1].Footer = &discord.EmbedFooter{
					Text: "Run /zipemojis to download all the emojis in a zip file.",
				}

				return &discord.InteractionResponse{
					Type: discord.InteractionCallbackTypeChannelMessageSource,
					Data: &discord.InteractionCallbackData{
						Embeds: embeds,
						Flags:  uint32(discord.MessageFlagEphemeral),
					},
				}, nil
			})
		},
	})

	m.InteractionCommands.MustAddInteractionCommand(&subway.InteractionCommandable{
		Name:        "invites",
		Description: "Get a leaderboard of the top inviters on this server",

		Handler: func(ctx context.Context, sub *subway.Subway, interaction discord.Interaction) (*discord.InteractionResponse, error) {
			return welcomer.RequireGuild(interaction, func() (*discord.InteractionResponse, error) {
				session, err := welcomer.AcquireSession(ctx, sub, welcomer.GetManagerNameFromContext(ctx))
				if err != nil {
					return nil, err
				}

				guild := discord.Guild{ID: *interaction.GuildID}
				invites, err := guild.Invites(session)
				if err != nil {
					sub.Logger.Error().Err(err).Msg("Failed to get invites")

					return nil, err
				}

				doesUserHaveInvites := false

				leaderboardMap := make(map[discord.Snowflake]int32)
				for _, invite := range invites {
					if invite.Inviter == nil || invite.Uses == 0 {
						continue
					}

					leaderboardMap[invite.Inviter.ID] += invite.Uses

					if invite.Inviter.ID == interaction.Member.User.ID {
						doesUserHaveInvites = true
					}
				}

				leaderboard := make([]InviteLeaderboardEntry, 0, len(leaderboardMap))
				for inviterID, uses := range leaderboardMap {
					leaderboard = append(leaderboard, InviteLeaderboardEntry{
						InviterID: inviterID,
						Uses:      int(uses),
					})
				}

				sort.Slice(leaderboard, func(i, j int) bool {
					return leaderboard[i].Uses > leaderboard[j].Uses || (leaderboard[i].Uses == leaderboard[j].Uses && leaderboard[i].InviterID > leaderboard[j].InviterID)
				})

				userPosition := -1
				userTotalInvites := 0

				if doesUserHaveInvites {
					for i, entry := range leaderboard {
						if entry.InviterID == interaction.Member.User.ID {
							userPosition = i + 1
							userTotalInvites = entry.Uses

							break
						}
					}
				}

				guildMemberIDs := []int64{}
				for inviterID := range leaderboardMap {
					guildMemberIDs = append(guildMemberIDs, int64(inviterID))
				}

				embeds := []*discord.Embed{}
				embed := &discord.Embed{Title: "Invite Leaderboard", Color: welcomer.EmbedColourInfo}

				embed.Description += fmt.Sprintf(
					"You have invited %d user%s to this server.\n",
					userTotalInvites,
					welcomer.If(userTotalInvites == 1, "", "s"),
				)

				if userPosition > 0 && userPosition <= 100 {
					embed.Description += fmt.Sprintf("You are currently **#%d** on the leaderboard.\n\n", userPosition)
				} else {
					embed.Description += "You are not on the leaderboard. Invite more users!\n\n"
				}

				if len(leaderboard) > 20 {
					leaderboard = leaderboard[:20]
				}

				for position, leaderboardUser := range leaderboard {
					leaderboardWithNumber := fmt.Sprintf(
						"%d. %s – **%d** invite%s\n",
						position+1,
						"<@"+leaderboardUser.InviterID.String()+">",
						leaderboardUser.Uses,
						welcomer.If(leaderboardUser.Uses == 1, "", "s"),
					)

					// If the embed content will go over 4000 characters then create a new embed and continue from that one.
					if len(embed.Description)+len(leaderboardWithNumber) > 4000 {
						embeds = append(embeds, embed)
						embed = &discord.Embed{Color: welcomer.EmbedColourInfo}
					}

					embed.Description += leaderboardWithNumber
				}

				embeds = append(embeds, embed)

				return &discord.InteractionResponse{
					Type: discord.InteractionCallbackTypeChannelMessageSource,
					Data: &discord.InteractionCallbackData{
						Embeds: embeds,
					},
				}, nil
			})
		},
	})

	m.InteractionCommands.MustAddInteractionCommand(&subway.InteractionCommandable{
		Name:        "newcreation",
		Description: "Returns a list of newly created users on discord",

		Handler: func(ctx context.Context, sub *subway.Subway, interaction discord.Interaction) (*discord.InteractionResponse, error) {
			return welcomer.RequireGuild(interaction, func() (*discord.InteractionResponse, error) {
				go func(ctx context.Context, sub *subway.Subway, interaction discord.Interaction) {
					embeds := []*discord.Embed{}
					embed := &discord.Embed{Title: "Newly Created Users", Color: welcomer.EmbedColourInfo}

					// Chunk users
					_, err := sub.SandwichClient.RequestGuildChunk(sub.Context, &sandwich.RequestGuildChunkRequest{
						GuildId: int64(*interaction.GuildID),
					})
					if err != nil {
						sub.Logger.Error().Err(err).
							Int64("guild_id", int64(*interaction.GuildID)).
							Msg("Failed to chunk guild")
					}

					guildMembersResp, err := sub.SandwichClient.FetchGuildMembers(ctx, &sandwich.FetchGuildMembersRequest{
						GuildID: int64(*interaction.GuildID),
					})
					if err != nil {
						sub.Logger.Error().Err(err).
							Int64("guild_id", int64(*interaction.GuildID)).
							Msg("Failed to fetch guild members")
					}

					lastMonth := time.Now().Add(-time.Hour * 24 * 30)

					guildMembers := make([]*sandwich.GuildMember, 0, len(guildMembersResp.GuildMembers))
					for _, guildMember := range guildMembersResp.GuildMembers {
						joinedAt, _ := time.Parse(time.RFC3339, guildMember.JoinedAt)
						if joinedAt.After(lastMonth) {
							guildMembers = append(guildMembers, guildMember)
						}
					}

					sort.Slice(guildMembers, func(i, j int) bool {
						return discord.Snowflake(guildMembers[i].User.ID).Time().After(discord.Snowflake(guildMembers[j].User.ID).Time())
					})

					if len(guildMembers) > 20 {
						guildMembers = guildMembers[:20]
					}

					if len(guildMembers) == 0 {
						embed.Description = "There are no newly created users."
					}

					for position, guildMember := range guildMembers {
						newCreationWithNumber := fmt.Sprintf(
							"%d. %s – **<t:%d:R>**\n",
							position+1,
							"<@"+strconv.FormatInt(guildMember.User.ID, 10)+">",
							discord.Snowflake(guildMember.User.ID).Time().Unix(),
						)

						// If the embed content will go over 4000 characters then create a new embed and continue from that one.
						if len(embed.Description)+len(newCreationWithNumber) > 4000 {
							embeds = append(embeds, embed)
							embed = &discord.Embed{Color: welcomer.EmbedColourInfo}
						}

						embed.Description += newCreationWithNumber
					}

					embeds = append(embeds, embed)

					_, err = interaction.EditOriginalResponse(sub.EmptySession, discord.WebhookMessageParams{
						Embeds: embeds,
					})
				}(ctx, sub, interaction)

				return &discord.InteractionResponse{
					Type: discord.InteractionCallbackTypeDeferredChannelMessageSource,
				}, nil
			})
		},
	})

	m.InteractionCommands.MustAddInteractionCommand(&subway.InteractionCommandable{
		Name:        "newmembers",
		Description: "Returns a list of new members on this guild",

		Handler: func(ctx context.Context, sub *subway.Subway, interaction discord.Interaction) (*discord.InteractionResponse, error) {
			return welcomer.RequireGuild(interaction, func() (*discord.InteractionResponse, error) {
				go func(ctx context.Context, sub *subway.Subway, interaction discord.Interaction) {
					embeds := []*discord.Embed{}
					embed := &discord.Embed{Title: "Newly Joined Members", Color: welcomer.EmbedColourInfo}

					// Chunk users
					_, err := sub.SandwichClient.RequestGuildChunk(sub.Context, &sandwich.RequestGuildChunkRequest{
						GuildId: int64(*interaction.GuildID),
					})
					if err != nil {
						sub.Logger.Error().Err(err).
							Int64("guild_id", int64(*interaction.GuildID)).
							Msg("Failed to chunk guild")
					}

					guildMembersResp, err := sub.SandwichClient.FetchGuildMembers(ctx, &sandwich.FetchGuildMembersRequest{
						GuildID: int64(*interaction.GuildID),
					})
					if err != nil {
						sub.Logger.Error().Err(err).
							Int64("guild_id", int64(*interaction.GuildID)).
							Msg("Failed to fetch guild members")
					}

					guildMembers := make([]*sandwich.GuildMember, 0, len(guildMembersResp.GuildMembers))
					for _, guildMember := range guildMembersResp.GuildMembers {
						guildMembers = append(guildMembers, guildMember)
					}

					sort.Slice(guildMembers, func(i, j int) bool {
						return guildMembers[i].JoinedAt > guildMembers[j].JoinedAt
					})

					if len(guildMembers) > 20 {
						guildMembers = guildMembers[:20]
					}

					for position, guildMember := range guildMembers {
						joinedAt, _ := time.Parse(time.RFC3339, guildMember.JoinedAt)
						newMemberWithNumber := fmt.Sprintf(
							"%d. %s – **<t:%d:R>**\n",
							position+1,
							"<@"+strconv.FormatInt(guildMember.User.ID, 10)+">",
							joinedAt.Unix(),
						)

						// If the embed content will go over 4000 characters then create a new embed and continue from that one.
						if len(embed.Description)+len(newMemberWithNumber) > 4000 {
							embeds = append(embeds, embed)
							embed = &discord.Embed{Color: welcomer.EmbedColourInfo}
						}

						embed.Description += newMemberWithNumber
					}

					embeds = append(embeds, embed)

					_, err = interaction.EditOriginalResponse(sub.EmptySession, discord.WebhookMessageParams{
						Embeds: embeds,
					})
				}(ctx, sub, interaction)

				return &discord.InteractionResponse{
					Type: discord.InteractionCallbackTypeDeferredChannelMessageSource,
				}, nil
			})
		},
	})

	m.InteractionCommands.MustAddInteractionCommand(&subway.InteractionCommandable{
		Name:        "oldmembers",
		Description: "Returns a list of the oldest members on this guild",

		Handler: func(ctx context.Context, sub *subway.Subway, interaction discord.Interaction) (*discord.InteractionResponse, error) {
			return welcomer.RequireGuild(interaction, func() (*discord.InteractionResponse, error) {
				go func(ctx context.Context, sub *subway.Subway, interaction discord.Interaction) {
					embeds := []*discord.Embed{}
					embed := &discord.Embed{Title: "Oldest Members", Color: welcomer.EmbedColourInfo}

					// Chunk users
					_, err := sub.SandwichClient.RequestGuildChunk(sub.Context, &sandwich.RequestGuildChunkRequest{
						GuildId: int64(*interaction.GuildID),
					})
					if err != nil {
						sub.Logger.Error().Err(err).
							Int64("guild_id", int64(*interaction.GuildID)).
							Msg("Failed to chunk guild")
					}

					guildMembersResp, err := sub.SandwichClient.FetchGuildMembers(ctx, &sandwich.FetchGuildMembersRequest{
						GuildID: int64(*interaction.GuildID),
					})
					if err != nil {
						sub.Logger.Error().Err(err).
							Int64("guild_id", int64(*interaction.GuildID)).
							Msg("Failed to fetch guild members")
					}

					guildMembers := make([]*sandwich.GuildMember, 0, len(guildMembersResp.GuildMembers))
					for _, guildMember := range guildMembersResp.GuildMembers {
						guildMembers = append(guildMembers, guildMember)
					}

					sort.Slice(guildMembers, func(i, j int) bool {
						return guildMembers[i].JoinedAt < guildMembers[j].JoinedAt
					})

					if len(guildMembers) > 20 {
						guildMembers = guildMembers[:20]
					}

					for position, guildMember := range guildMembers {
						joinedAt, _ := time.Parse(time.RFC3339, guildMember.JoinedAt)
						newMemberWithNumber := fmt.Sprintf(
							"%d. %s – **<t:%d:R>**\n",
							position+1,
							"<@"+strconv.FormatInt(guildMember.User.ID, 10)+">",
							joinedAt.Unix(),
						)

						// If the embed content will go over 4000 characters then create a new embed and continue from that one.
						if len(embed.Description)+len(newMemberWithNumber) > 4000 {
							embeds = append(embeds, embed)
							embed = &discord.Embed{Color: welcomer.EmbedColourInfo}
						}

						embed.Description += newMemberWithNumber
					}

					embeds = append(embeds, embed)

					_, err = interaction.EditOriginalResponse(sub.EmptySession, discord.WebhookMessageParams{
						Embeds: embeds,
					})
				}(ctx, sub, interaction)

				return &discord.InteractionResponse{
					Type: discord.InteractionCallbackTypeDeferredChannelMessageSource,
				}, nil
			})
		},
	})

	m.InteractionCommands.MustAddInteractionCommand(&subway.InteractionCommandable{
		Name:        "ping",
		Description: "Gets round trip API latency",

		Handler: func(ctx context.Context, sub *subway.Subway, interaction discord.Interaction) (*discord.InteractionResponse, error) {
			interactionDelay := time.Since(interaction.ID.Time()).Milliseconds()
			start := time.Now()

			err := interaction.SendResponse(
				sub.EmptySession,
				discord.InteractionCallbackTypeChannelMessageSource,
				&discord.InteractionCallbackData{
					Content: fmt.Sprintf(
						"Interaction Delay: %dms\nHTTP Latency: ...",
						interactionDelay,
					),
				})
			if err != nil {
				return nil, err
			}

			_, err = interaction.EditOriginalResponse(
				sub.EmptySession,
				discord.WebhookMessageParams{
					Content: fmt.Sprintf(
						"Interaction Delay: %dms\nHTTP Latency: %dms",
						interactionDelay,
						time.Since(start).Milliseconds(),
					),
				})
			if err != nil {
				return nil, err
			}

			return nil, nil
		},
	})

	// TODO: pog
	// TODO: purge
	// TODO: support
	// TODO: zipemojis

	return nil
}

type InviteLeaderboardEntry struct {
	InviterID discord.Snowflake
	Uses      int
}
