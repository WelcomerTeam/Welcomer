package plugins

import (
	"context"

	"github.com/WelcomerTeam/Discord/discord"
	sandwich "github.com/WelcomerTeam/Sandwich-Daemon/protobuf"
	subway "github.com/WelcomerTeam/Subway/subway"
	"github.com/WelcomerTeam/Welcomer/welcomer-core"
	jsoniter "github.com/json-iterator/go"
)

func NewWelcomerCog() *WelcomerCog {
	return &WelcomerCog{
		InteractionCommands: subway.SetupInteractionCommandable(&subway.InteractionCommandable{}),
	}
}

type WelcomerCog struct {
	InteractionCommands *subway.InteractionCommandable
}

// Assert types.

var (
	_ subway.Cog                        = (*GeneralCog)(nil)
	_ subway.CogWithInteractionCommands = (*GeneralCog)(nil)
)

func (p *WelcomerCog) CogInfo() *subway.CogInfo {
	return &subway.CogInfo{
		Name:        "Welcomer",
		Description: "Provides the functionality for the 'Welcomer' feature",
	}
}

func (p *WelcomerCog) GetInteractionCommandable() *subway.InteractionCommandable {
	return p.InteractionCommands
}

func (p *WelcomerCog) RegisterCog(sub *subway.Subway) error {
	welcomerGroup := subway.NewSubcommandGroup(
		"welcomer",
		"Welcome new users to your server with fancy images, text or send them a direct message.",
	)

	welcomerGroup.MustAddInteractionCommand(&subway.InteractionCommandable{
		Name:        "test",
		Description: "Tests the Welcomer functionality.",

		Type: subway.InteractionCommandableTypeSubcommand,

		ArgumentParameter: []subway.ArgumentParameter{
			{
				Required:     false,
				ArgumentType: subway.ArgumentTypeMember,
				Name:         "user",
				Description:  "The user you would like to send the welcome message for.",
			},
		},

		Handler: func(ctx context.Context, sub *subway.Subway, interaction discord.Interaction) (*discord.InteractionResponse, error) {
			return welcomer.RequireGuildElevation(sub, interaction, func() (*discord.InteractionResponse, error) {
				member := subway.MustGetArgument(ctx, "user").MustMember()
				if member == nil {
					member = interaction.Member
				}

				// GuildID may be missing, fill it in.
				member.GuildID = interaction.GuildID

				data, err := jsoniter.Marshal(welcomer.CustomEventInvokeWelcomerStructure{
					Interaction: &interaction,
					Member:      member,
				})
				if err != nil {
					return nil, err
				}

				_, err = sub.SandwichClient.RelayMessage(ctx, &sandwich.RelayMessageRequest{
					Manager: welcomer.GetManagerNameFromContext(ctx),
					Type:    welcomer.CustomEventInvokeWelcomer,
					Data:    data,
				})
				if err != nil {
					return nil, err
				}

				return &discord.InteractionResponse{
					Type: discord.InteractionCallbackTypeDeferredChannelMessageSource,
				}, nil
			})
		},
	})

	welcomerGroup.MustAddInteractionCommand(&subway.InteractionCommandable{
		Name:        "enable",
		Description: "Enables the Welcomer functionality.",
	})

	p.InteractionCommands.MustAddInteractionCommand(welcomerGroup)

	return nil
}
