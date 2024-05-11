package plugins

import (
	"context"
	"errors"
	"fmt"

	"github.com/WelcomerTeam/Discord/discord"
	subway "github.com/WelcomerTeam/Subway/subway"
	welcomer "github.com/WelcomerTeam/Welcomer/welcomer-core"
	"github.com/WelcomerTeam/Welcomer/welcomer-core/database"
	"github.com/jackc/pgx/v4"
)

func NewMembershipCog() *MembershipCog {
	return &MembershipCog{
		InteractionCommands: subway.SetupInteractionCommandable(&subway.InteractionCommandable{}),
	}
}

type MembershipCog struct {
	InteractionCommands *subway.InteractionCommandable
}

// Assert types.

var (
	_ subway.Cog                        = (*MembershipCog)(nil)
	_ subway.CogWithInteractionCommands = (*MembershipCog)(nil)
)

const (
	MembershipModuleCustomBackgrounds = "custom_backgrounds"
	MembershipModulePro               = "pro"
)

func (p *MembershipCog) CogInfo() *subway.CogInfo {
	return &subway.CogInfo{
		Name:        "Memberships",
		Description: "Manage memberships.",
	}
}

func (p *MembershipCog) GetInteractionCommandable() *subway.InteractionCommandable {
	return p.InteractionCommands
}

func (p *MembershipCog) RegisterCog(sub *subway.Subway) error {
	membershipGroup := subway.NewSubcommandGroup(
		"membership",
		"Manage your welcomer subscriptions.",
	)

	membershipGroup.MustAddInteractionCommand(&subway.InteractionCommandable{
		Name:        "list",
		Description: "Lists all membershipss you have available.",

		Handler: func(ctx context.Context, sub *subway.Subway, interaction discord.Interaction) (*discord.InteractionResponse, error) {
			queries := welcomer.GetQueriesFromContext(ctx)

			memberships, err := queries.GetUserMembershipsByUserID(ctx, int64(interaction.User.ID))
			if err != nil && !errors.Is(err, pgx.ErrNoRows) {
				sub.Logger.Error().Err(err).
					Int64("user_id", int64(interaction.User.ID)).
					Msg("Failed to get user memberships.")
			}

			if len(memberships) == 0 {
				return &discord.InteractionResponse{
					Type: discord.InteractionCallbackTypeChannelMessageSource,
					Data: &discord.InteractionCallbackData{
						Embeds: []*discord.Embed{
							{
								Title:       "Memberships",
								Description: "You don't have any memberships.",
								Color:       welcomer.EmbedColourInfo,
							},
						},
						Components: []*discord.InteractionComponent{
							{
								Type: discord.InteractionComponentTypeActionRow,
								Components: []*discord.InteractionComponent{
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
			}

			embeds := []*discord.Embed{}
			embed := &discord.Embed{Title: "Memberships", Color: welcomer.EmbedColourInfo}

			for _, membership := range memberships {
				membershipStatus := database.MembershipStatus(membership.Status)
				membershipType := database.MembershipType(membership.MembershipType)

				embed.Fields = append(embed.Fields, &discord.EmbedField{
					Name:   fmt.Sprintf("%s – %s", membershipType.String(), membershipStatus.String()),
					Value:  welcomer.If(membership.GuildID != 0, fmt.Sprintf("Guild: %s `%d`", "TODO", membership.GuildID), "Unassigned"),
					Inline: false,
				})
			}

			embeds = append(embeds, embed)

			return &discord.InteractionResponse{
				Type: discord.InteractionCallbackTypeChannelMessageSource,
				Data: &discord.InteractionCallbackData{
					Embeds: embeds,
					Flags:  uint32(discord.MessageFlagEphemeral),
				},
			}, nil
		},
	})

	p.InteractionCommands.MustAddInteractionCommand(membershipGroup)

	return nil
}
