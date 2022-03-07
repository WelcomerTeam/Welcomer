package welcomer

import (
	"github.com/WelcomerTeam/Discord/discord"
	sandwich "github.com/WelcomerTeam/Sandwich/sandwich"
)

func NewPogCog() *PogCog {
	return &PogCog{
		Commands:            sandwich.SetupCommandable(&sandwich.Commandable{}),
		InteractionCommands: sandwich.SetupInteractionCommandable(&sandwich.InteractionCommandable{}),
	}
}

type PogCog struct {
	Commands            *sandwich.Commandable
	InteractionCommands *sandwich.InteractionCommandable
}

// Assert types.

var (
	_ sandwich.Cog                        = (*PogCog)(nil)
	_ sandwich.CogWithCommands            = (*PogCog)(nil)
	_ sandwich.CogWithInteractionCommands = (*PogCog)(nil)
)

// CogInfo returns information about a cog, including name and description.
func (p *PogCog) CogInfo() *sandwich.CogInfo {
	return &sandwich.CogInfo{
		Name:        "PogCog",
		Description: "This cog is poggers",
	}
}

// GetCommandable returns all commands in a cog. You can optionally register commands here, however
// there is no guarantee this wont be called multiple times.
func (p *PogCog) GetCommandable() *sandwich.Commandable {
	return p.Commands
}

// GetInteractionCommandable returns all interaction commands in a cog. You can optionally register commands here, however
// there is no guarantee this wont be called multiple times.
func (p *PogCog) GetInteractionCommandable() *sandwich.InteractionCommandable {
	return p.InteractionCommands
}

// RegisterCog is called in bot.RegisterCog, the plugin should set itself up, including commands.
func (p *PogCog) RegisterCog(b *sandwich.Bot) (err error) {
	// Using MustAddCommand instead of AddCommand ensures that we have set the bot up properly.
	// Any errors that occur adding a command, such as name colliosion, will result in a panic
	// when using MustX functions.
	p.Commands.MustAddCommand(&sandwich.Commandable{
		Name:        "pog",
		Description: "This command is very poggers",
		Handler: func(ctx *sandwich.CommandContext) (err error) {
			_, err = ctx.Reply(ctx.EventContext.Session, discord.MessageParams{
				Content: "<:rock:732274836038221855>📣 pog",
			})

			return
		},
	})

	p.InteractionCommands.MustAddInteractionCommand(&sandwich.InteractionCommandable{
		Name:        "pog",
		Description: "This command is very poggers",
		Handler: func(ctx *sandwich.InteractionContext) (resp *sandwich.InteractionResponse, err error) {
			return &sandwich.InteractionResponse{
				Type: discord.InteractionCallbackTypeChannelMessageSource,
				Data: discord.InteractionCallbackData{
					WebhookMessageParams: discord.WebhookMessageParams{
						Content: "<:rock:732274836038221855>📣 pog",
					},
				},
			}, nil
		},
	})

	return nil
}
