package plugins

import (
	"archive/zip"
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"sort"
	"time"

	"github.com/WelcomerTeam/Discord/discord"
	sandwich_protobuf "github.com/WelcomerTeam/Sandwich-Daemon/proto"
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
		Name:        "cleanup",
		Description: "Remove messages from the bot.",

		Handler: func(ctx context.Context, sub *subway.Subway, interaction discord.Interaction) (*discord.InteractionResponse, error) {
			session, err := welcomer.AcquireSession(ctx, welcomer.GetManagerNameFromContext(ctx))
			if err != nil {
				return nil, err
			}

			channel := discord.Channel{ID: *interaction.ChannelID, GuildID: interaction.GuildID}

			messageHistory, err := channel.History(ctx, session, nil, nil, nil, nil)
			if err != nil {
				return nil, err
			}

			messagesToDelete := make([]discord.Snowflake, 0, len(messageHistory))
			for _, message := range messageHistory {
				// Skip message if it is over 14 days old.
				if message.ID.Time().Before(time.Now().Add(-time.Hour * 24 * 14)) {
					continue
				}

				if message.Author.ID == interaction.ApplicationID {
					messagesToDelete = append(messagesToDelete, message.ID)
				}
			}

			go func() {
				time.Sleep(time.Second * 15)

				err = interaction.DeleteOriginalResponse(ctx, sub.EmptySession)
				if err != nil {
					welcomer.Logger.Error().Err(err).Msg("Failed to delete original response")
				}
			}()

			if len(messagesToDelete) == 0 {
				return &discord.InteractionResponse{
					Type: discord.InteractionCallbackTypeChannelMessageSource,
					Data: &discord.InteractionCallbackData{
						Embeds: welcomer.NewEmbed("No messages to delete", welcomer.EmbedColourInfo),
						Flags:  uint32(discord.MessageFlagEphemeral),
					},
				}, nil
			}

			if len(messagesToDelete) == 1 {
				message := discord.Message{ID: messagesToDelete[0], ChannelID: *interaction.ChannelID, GuildID: interaction.GuildID}

				err = message.Delete(ctx, session, nil)
				if err != nil {
					welcomer.Logger.Error().Err(err).Msg("Failed to delete message")

					return nil, err
				}
			} else if len(messagesToDelete) > 1 {
				if interaction.GuildID == nil {
					for _, messageID := range messagesToDelete {
						message := discord.Message{ID: messageID, ChannelID: *interaction.ChannelID}

						err = message.Delete(ctx, session, nil)
						if err != nil {
							welcomer.Logger.Error().Err(err).Msg("Failed to delete message")

							return nil, err
						}
					}
				} else {
					err = channel.DeleteMessages(ctx, session, messagesToDelete, nil)
					if err != nil {
						welcomer.Logger.Error().Err(err).Msg("Failed to delete messages")

						return nil, err
					}
				}
			}

			return &discord.InteractionResponse{
				Type: discord.InteractionCallbackTypeChannelMessageSource,
				Data: &discord.InteractionCallbackData{
					Embeds: welcomer.NewEmbed(
						fmt.Sprintf(
							"%d message%s been deleted",
							len(messagesToDelete),
							welcomer.If(len(messagesToDelete) == 1, " has", "s have"),
						),
						welcomer.EmbedColourInfo,
					),
				},
			}, nil
		},
	})

	m.InteractionCommands.MustAddInteractionCommand(&subway.InteractionCommandable{
		Name:        "dashboard",
		Description: "Get a link to the Welcomer dashboard",

		Handler: func(ctx context.Context, sub *subway.Subway, interaction discord.Interaction) (*discord.InteractionResponse, error) {
			if interaction.GuildID == nil {
				return &discord.InteractionResponse{
					Type: discord.InteractionCallbackTypeChannelMessageSource,
					Data: &discord.InteractionCallbackData{
						Embeds: []discord.Embed{
							{
								Description: "### **Configure your guild with the website dashboard**\n\nManage your guild settings and memberships at " + welcomer.WebsiteURL + "/dashboard",
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
							Embeds: []discord.Embed{
								{
									Description: fmt.Sprintf("### **Configure your guild with the website dashboard**\n\nManage this guild's settings and memberships ib our dashboard [**here**](%s).", welcomer.WebsiteURL+"/dashboard/"+interaction.GuildID.String()),
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
					Embeds: []discord.Embed{
						{
							Description: "### **Everything you need to boost your guild's engagement**\n\nGet Welcomer Pro and support Welcomer development.",
							Color:       welcomer.EmbedColourInfo,
						},
					},
					Components: []discord.InteractionComponent{
						{
							Type: discord.InteractionComponentTypeActionRow,
							Components: []discord.InteractionComponent{
								{
									Type:  discord.InteractionComponentTypeButton,
									Style: discord.InteractionComponentStylePremium,
									SKUID: discord.Snowflake(welcomer.TryParseInt(os.Getenv("WELCOMER_PRO_DISCORD_SKU_ID"))),
								},
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

	m.InteractionCommands.MustAddInteractionCommand(&subway.InteractionCommandable{
		Name:        "emojis",
		Description: "Get a list of all the emojis in the guild",

		DMPermission: &welcomer.False,

		Handler: func(ctx context.Context, sub *subway.Subway, interaction discord.Interaction) (*discord.InteractionResponse, error) {
			return welcomer.RequireGuild(interaction, func() (*discord.InteractionResponse, error) {
				embeds := []discord.Embed{}
				embed := discord.Embed{Title: "Emojis", Color: welcomer.EmbedColourInfo}

				guildEmojis, err := sub.SandwichClient.FetchGuildEmoji(sub.Context, &sandwich_protobuf.FetchGuildEmojiRequest{
					GuildId: int64(*interaction.GuildID),
				})
				if err != nil {
					return nil, err
				}

				// Flatten map into slice

				i := 0
				emojis := make([]*sandwich_protobuf.Emoji, len(guildEmojis.GetEmojis()))
				for _, emoji := range guildEmojis.GetEmojis() {
					emojis[i] = emoji
					i++
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
						embed = discord.Embed{Color: welcomer.EmbedColourInfo}
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

		DMPermission: &welcomer.False,

		Handler: func(ctx context.Context, sub *subway.Subway, interaction discord.Interaction) (*discord.InteractionResponse, error) {
			return welcomer.RequireGuild(interaction, func() (*discord.InteractionResponse, error) {
				session, err := welcomer.AcquireSession(ctx, welcomer.GetManagerNameFromContext(ctx))
				if err != nil {
					return nil, err
				}

				guild := discord.Guild{ID: *interaction.GuildID}
				invites, err := guild.Invites(ctx, session)
				if err != nil {
					welcomer.Logger.Error().Err(err).Msg("Failed to get invites")

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

				embeds := []discord.Embed{}
				embed := discord.Embed{Title: "Invite Leaderboard", Color: welcomer.EmbedColourInfo}

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
						embed = discord.Embed{Color: welcomer.EmbedColourInfo}
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

		DMPermission: &welcomer.False,

		Handler: func(ctx context.Context, sub *subway.Subway, interaction discord.Interaction) (*discord.InteractionResponse, error) {
			return welcomer.RequireGuild(interaction, func() (*discord.InteractionResponse, error) {
				go func(ctx context.Context, sub *subway.Subway, interaction discord.Interaction) {
					embeds := []discord.Embed{}
					embed := discord.Embed{Title: "Newly Created Users", Color: welcomer.EmbedColourInfo}

					// Chunk users
					_, err := sub.SandwichClient.RequestGuildChunk(sub.Context, &sandwich_protobuf.RequestGuildChunkRequest{
						GuildId: int64(*interaction.GuildID),
					})
					if err != nil {
						welcomer.Logger.Error().Err(err).
							Int64("guild_id", int64(*interaction.GuildID)).
							Msg("Failed to chunk guild")
					}

					guildMembersResp, err := sub.SandwichClient.FetchGuildMember(ctx, &sandwich_protobuf.FetchGuildMemberRequest{
						GuildId: int64(*interaction.GuildID),
					})
					if err != nil {
						welcomer.Logger.Error().Err(err).
							Int64("guild_id", int64(*interaction.GuildID)).
							Msg("Failed to fetch guild members")
					}

					lastMonth := time.Now().Add(-time.Hour * 24 * 30)

					guildMembers := make([]*sandwich_protobuf.GuildMember, 0, len(guildMembersResp.GuildMembers))
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
							"<@"+welcomer.Itoa(guildMember.User.ID)+">",
							discord.Snowflake(guildMember.User.ID).Time().Unix(),
						)

						// If the embed content will go over 4000 characters then create a new embed and continue from that one.
						if len(embed.Description)+len(newCreationWithNumber) > 4000 {
							embeds = append(embeds, embed)
							embed = discord.Embed{Color: welcomer.EmbedColourInfo}
						}

						embed.Description += newCreationWithNumber
					}

					embeds = append(embeds, embed)

					_, err = interaction.EditOriginalResponse(ctx, sub.EmptySession, discord.WebhookMessageParams{
						Embeds: embeds,
					})
					if err != nil {
						welcomer.Logger.Error().Err(err).Msg("Failed to edit original response")
					}
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

		DMPermission: &welcomer.False,

		Handler: func(ctx context.Context, sub *subway.Subway, interaction discord.Interaction) (*discord.InteractionResponse, error) {
			return welcomer.RequireGuild(interaction, func() (*discord.InteractionResponse, error) {
				go func(ctx context.Context, sub *subway.Subway, interaction discord.Interaction) {
					embeds := []discord.Embed{}
					embed := discord.Embed{Title: "Newly Joined Members", Color: welcomer.EmbedColourInfo}

					// Chunk users
					_, err := sub.SandwichClient.RequestGuildChunk(sub.Context, &sandwich_protobuf.RequestGuildChunkRequest{
						GuildId: int64(*interaction.GuildID),
					})
					if err != nil {
						welcomer.Logger.Error().Err(err).
							Int64("guild_id", int64(*interaction.GuildID)).
							Msg("Failed to chunk guild")
					}

					guildMembersResp, err := sub.SandwichClient.FetchGuildMember(ctx, &sandwich_protobuf.FetchGuildMemberRequest{
						GuildId: int64(*interaction.GuildID),
					})
					if err != nil {
						welcomer.Logger.Error().Err(err).
							Int64("guild_id", int64(*interaction.GuildID)).
							Msg("Failed to fetch guild members")
					}

					i := 0
					guildMembers := make([]*sandwich_protobuf.GuildMember, len(guildMembersResp.GetGuildMembers()))
					for _, guildMember := range guildMembersResp.GuildMembers {
						guildMembers[i] = guildMember
						i++
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
							"<@"+welcomer.Itoa(guildMember.User.ID)+">",
							joinedAt.Unix(),
						)

						// If the embed content will go over 4000 characters then create a new embed and continue from that one.
						if len(embed.Description)+len(newMemberWithNumber) > 4000 {
							embeds = append(embeds, embed)
							embed = discord.Embed{Color: welcomer.EmbedColourInfo}
						}

						embed.Description += newMemberWithNumber
					}

					embeds = append(embeds, embed)

					_, err = interaction.EditOriginalResponse(ctx, sub.EmptySession, discord.WebhookMessageParams{
						Embeds: embeds,
					})
					if err != nil {
						welcomer.Logger.Error().Err(err).Msg("Failed to edit original response")
					}
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

		DMPermission: &welcomer.False,

		Handler: func(ctx context.Context, sub *subway.Subway, interaction discord.Interaction) (*discord.InteractionResponse, error) {
			return welcomer.RequireGuild(interaction, func() (*discord.InteractionResponse, error) {
				go func(ctx context.Context, sub *subway.Subway, interaction discord.Interaction) {
					embeds := []discord.Embed{}
					embed := discord.Embed{Title: "Oldest Members", Color: welcomer.EmbedColourInfo}

					// Chunk users
					_, err := sub.SandwichClient.RequestGuildChunk(sub.Context, &sandwich_protobuf.RequestGuildChunkRequest{
						GuildId: int64(*interaction.GuildID),
					})
					if err != nil {
						welcomer.Logger.Error().Err(err).
							Int64("guild_id", int64(*interaction.GuildID)).
							Msg("Failed to chunk guild")
					}

					guildMembersResp, err := sub.SandwichClient.FetchGuildMember(ctx, &sandwich_protobuf.FetchGuildMemberRequest{
						GuildId: int64(*interaction.GuildID),
					})
					if err != nil {
						welcomer.Logger.Error().Err(err).
							Int64("guild_id", int64(*interaction.GuildID)).
							Msg("Failed to fetch guild members")
					}

					i := 0
					guildMembers := make([]*sandwich_protobuf.GuildMember, len(guildMembersResp.GuildMembers))
					for _, guildMember := range guildMembersResp.GuildMembers {
						guildMembers[i] = guildMember
						i++
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
							"<@"+welcomer.Itoa(guildMember.User.ID)+">",
							joinedAt.Unix(),
						)

						// If the embed content will go over 4000 characters then create a new embed and continue from that one.
						if len(embed.Description)+len(newMemberWithNumber) > 4000 {
							embeds = append(embeds, embed)
							embed = discord.Embed{Color: welcomer.EmbedColourInfo}
						}

						embed.Description += newMemberWithNumber
					}

					embeds = append(embeds, embed)

					_, err = interaction.EditOriginalResponse(ctx, sub.EmptySession, discord.WebhookMessageParams{
						Embeds: embeds,
					})
					if err != nil {
						welcomer.Logger.Error().Err(err).Msg("Failed to edit original response")
					}
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
				ctx, sub.EmptySession,
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
				ctx, sub.EmptySession,
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

	m.InteractionCommands.MustAddInteractionCommand(&subway.InteractionCommandable{
		Name:        "purge",
		Description: "Remove messages from a channel, based on criteria.",

		ArgumentParameter: []subway.ArgumentParameter{
			{
				Name:         "user",
				ArgumentType: subway.ArgumentTypeUser,
				Description:  "Remove messages from this user",
			},
			{
				Name:         "limit",
				ArgumentType: subway.ArgumentTypeInt,
				Description:  "Limit of messages to search for",
				MaxValue:     welcomer.ToPointer(int32(100)),
				MinValue:     welcomer.ToPointer(int32(1)),
			},
			{
				Name:         "bot",
				ArgumentType: subway.ArgumentTypeBool,
				Description:  "Remove messages from bots",
			},
			{
				Name:         "webhooks",
				ArgumentType: subway.ArgumentTypeBool,
				Description:  "Removes messages from webhooks",
			},
			{
				Name:         "newusers",
				ArgumentType: subway.ArgumentTypeBool,
				Description:  "Removes messages from new users who have joined in the last week",
			},
			{
				Name:         "timeout",
				ArgumentType: subway.ArgumentTypeInt,
				Description:  "When supplied, the user will also be timed out for this number of hours",
				MinValue:     welcomer.ToPointer(int32(1)),
			},
			{
				Name:         "reason",
				ArgumentType: subway.ArgumentTypeString,
				Description:  "Reason for the purge. Included in the audit log",
			},
		},

		Handler: func(ctx context.Context, sub *subway.Subway, interaction discord.Interaction) (*discord.InteractionResponse, error) {
			return welcomer.RequireGuildElevation(sub, interaction, func() (*discord.InteractionResponse, error) {
				argumentUser := subway.MustGetArgument(ctx, "user").MustUser()
				argumentLimit := subway.MustGetArgument(ctx, "limit").MustInt()
				argumentBot := subway.MustGetArgument(ctx, "bot").MustBool()
				argumentWebhooks := subway.MustGetArgument(ctx, "webhooks").MustBool()
				argumentNewUsers := subway.MustGetArgument(ctx, "newusers").MustBool()
				argumentTimeout := subway.MustGetArgument(ctx, "timeout").MustInt()
				argumentReason := subway.MustGetArgument(ctx, "reason").MustString()

				session, err := welcomer.AcquireSession(ctx, welcomer.GetManagerNameFromContext(ctx))
				if err != nil {
					return nil, err
				}

				channel := discord.Channel{ID: *interaction.ChannelID, GuildID: interaction.GuildID}

				var limit *int32
				if argumentLimit > 0 {
					limit = welcomer.ToPointer(int32(argumentLimit))
				}

				messageHistory, err := channel.History(ctx, session, nil, nil, nil, limit)
				if err != nil {
					return nil, err
				}

				messagesToDelete := make([]discord.Snowflake, 0, len(messageHistory))
				usersToTimeout := make(map[discord.Snowflake]bool)

				for _, message := range messageHistory {
					// Skip message if it is over 14 days old.
					if message.ID.Time().Before(time.Now().Add(-time.Hour * 24 * 14)) {
						continue
					}

					if (argumentNewUsers && message.Author.ID.Time().Before(time.Now().Add(-time.Hour*24*7))) ||
						(!argumentUser.ID.IsNil() && message.Author.ID == argumentUser.ID) ||
						(argumentBot && message.Author.Bot) ||
						(argumentWebhooks && message.WebhookID != nil) ||
						(!argumentNewUsers && argumentUser.ID.IsNil() && !argumentBot && !argumentWebhooks) {
						messagesToDelete = append(messagesToDelete, message.ID)

						if argumentTimeout > 0 {
							usersToTimeout[message.Author.ID] = true
						}
					}
				}

				if len(messagesToDelete) == 0 {
					return &discord.InteractionResponse{
						Type: discord.InteractionCallbackTypeChannelMessageSource,
						Data: &discord.InteractionCallbackData{
							Embeds: welcomer.NewEmbed("No messages found to delete", welcomer.EmbedColourInfo),
							Flags:  uint32(discord.MessageFlagEphemeral),
						},
					}, nil
				}

				if len(messagesToDelete) == 1 {
					message := discord.Message{ID: messagesToDelete[0], ChannelID: *interaction.ChannelID, GuildID: interaction.GuildID}

					err = message.Delete(
						ctx, session,
						welcomer.ToPointer(fmt.Sprintf(
							"Purge by %s (%d). Reason: %s",
							welcomer.GetUserDisplayName(interaction.Member.User),
							interaction.Member.User.ID,
							welcomer.If(argumentReason == "", "No reason provided", argumentReason),
						)),
					)
					if err != nil {
						welcomer.Logger.Error().Err(err).Msg("Failed to delete message")

						return nil, err
					}
				} else if len(messageHistory) > 1 {
					err = channel.DeleteMessages(
						ctx, session,
						messagesToDelete,
						welcomer.ToPointer(fmt.Sprintf(
							"Purge by %s (%d). Reason: %s",
							welcomer.GetUserDisplayName(interaction.Member.User),
							interaction.Member.User.ID,
							welcomer.If(argumentReason == "", "No reason provided", argumentReason),
						)),
					)
					if err != nil {
						welcomer.Logger.Error().Err(err).Msg("Failed to delete messages")

						return nil, err
					}
				}

				communicationDisabledUntil := welcomer.ToPointer(time.Now().Add(time.Hour * time.Duration(argumentTimeout)))

				for userID := range usersToTimeout {
					guildMember := discord.GuildMember{GuildID: interaction.GuildID, User: &discord.User{ID: userID}}
					err = guildMember.Edit(ctx, session,
						discord.GuildMemberParams{
							CommunicationDisabledUntil: communicationDisabledUntil,
						},
						welcomer.ToPointer(fmt.Sprintf(
							"Timeout from purge by %s (%d). Reason: %s",
							welcomer.GetUserDisplayName(interaction.Member.User),
							interaction.Member.User.ID,
							welcomer.If(argumentReason == "", "No reason provided", argumentReason),
						)),
					)
					if err != nil {
						welcomer.Logger.Error().Err(err).
							Int64("guild_id", int64(*interaction.GuildID)).
							Int64("user_id", int64(userID)).
							Msg("Failed to timeout user")
					}
				}

				return &discord.InteractionResponse{
					Type: discord.InteractionCallbackTypeChannelMessageSource,
					Data: &discord.InteractionCallbackData{
						Embeds: welcomer.NewEmbed(
							fmt.Sprintf(
								"%d message%s been deleted",
								len(messagesToDelete),
								welcomer.If(len(messagesToDelete) == 1, " has", "s have"),
							),
							welcomer.EmbedColourInfo,
						),
					},
				}, nil
			})
		},
	})

	m.InteractionCommands.MustAddInteractionCommand(&subway.InteractionCommandable{
		Name:        "pog",
		Description: "???",

		Handler: func(ctx context.Context, sub *subway.Subway, interaction discord.Interaction) (*discord.InteractionResponse, error) {
			return &discord.InteractionResponse{
				Type: discord.InteractionCallbackTypeChannelMessageSource,
				Data: &discord.InteractionCallbackData{
					Content: welcomer.EmojiRock + "📣 pog",
				},
			}, nil
		},
	})

	m.InteractionCommands.MustAddInteractionCommand(&subway.InteractionCommandable{
		Name:        "support",
		Description: "Need help with the bot?",

		DMPermission:            &welcomer.False,
		DefaultMemberPermission: welcomer.ToPointer(discord.Int64(discord.PermissionElevated)),

		Handler: func(ctx context.Context, sub *subway.Subway, interaction discord.Interaction) (*discord.InteractionResponse, error) {
			return &discord.InteractionResponse{
				Type: discord.InteractionCallbackTypeChannelMessageSource,
				Data: &discord.InteractionCallbackData{
					Embeds: []discord.Embed{
						{
							Description: fmt.Sprintf("### **Welcomer Support Guild**\n\nGet support with using Welcomer on our support server [**here**](%s).", welcomer.WebsiteURL+"/support"),
							Color:       welcomer.EmbedColourInfo,
						},
					},
				},
			}, nil
		},
	})

	m.InteractionCommands.MustAddInteractionCommand(&subway.InteractionCommandable{
		Name:        "zipemojis",
		Description: "Get all the emojis in the guild as a zip file",

		DMPermission: &welcomer.False,

		Handler: func(ctx context.Context, sub *subway.Subway, interaction discord.Interaction) (*discord.InteractionResponse, error) {
			return welcomer.RequireGuild(interaction, func() (*discord.InteractionResponse, error) {
				go func() {
					guildEmojis, err := sub.SandwichClient.FetchGuildEmoji(sub.Context, &sandwich_protobuf.FetchGuildEmojiRequest{
						GuildId: int64(*interaction.GuildID),
					})
					if err != nil {
						return
					}

					var buf bytes.Buffer
					zipWriter := zip.NewWriter(&buf)

					client := http.Client{Timeout: time.Second * 5}

					for _, emoji := range guildEmojis.GetEmojis() {
						url := discord.EndpointCDN + welcomer.If(
							emoji.Animated,
							discord.EndpointEmojiAnimated(welcomer.Itoa(emoji.ID)),
							discord.EndpointEmoji(welcomer.Itoa(emoji.ID)),
						)
						resp, err := client.Get(url)
						if err != nil {
							welcomer.Logger.Warn().Err(err).
								Int64("emoji_id", emoji.ID).
								Msg("Failed to get emoji")

							continue
						}

						if resp.StatusCode >= 200 && resp.StatusCode < 300 {
							writer, err := zipWriter.Create(
								fmt.Sprintf(
									"%s_%d.%s",
									emoji.Name,
									emoji.ID,
									welcomer.If(emoji.Animated, "gif", "png"),
								),
							)
							if err != nil {
								welcomer.Logger.Error().Err(err).Msg("Failed to create file in zip")

								continue
							}

							_, err = io.Copy(writer, resp.Body)
							if err != nil {
								welcomer.Logger.Error().Err(err).Msg("Failed to copy file to zip")
							}
						}
					}

					err = zipWriter.Close()
					if err != nil {
						welcomer.Logger.Error().Err(err).Msg("Failed to close zip")
					}

					responseMessage := discord.WebhookMessageParams{
						Embeds: welcomer.NewEmbed("Here is a list of the guild emojis!", welcomer.EmbedColourInfo),
					}
					responseMessage.Files = append(responseMessage.Files, discord.File{
						Name:        "emojis.zip",
						ContentType: "application/zip",
						Reader:      &buf,
					})

					_, err = interaction.EditOriginalResponse(ctx, sub.EmptySession, responseMessage)
					if err != nil {
						welcomer.Logger.Error().Err(err).Msg("Failed to edit original response")
					}
				}()

				return &discord.InteractionResponse{
					Type: discord.InteractionCallbackTypeDeferredChannelMessageSource,
				}, nil
			})
		},
	})

	return nil
}

type InviteLeaderboardEntry struct {
	InviterID discord.Snowflake
	Uses      int
}
