package plugins

import (
	"context"
	"fmt"
	"sort"
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
	// TODO: dashboard

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

				for position, leaderboardUser := range leaderboard[:20] {
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

	// TODO: newcreation
	// TODO: newusers
	// TODO: oldusers

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
