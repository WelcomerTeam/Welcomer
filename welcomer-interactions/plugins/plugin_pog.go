package plugins

import (
	"bytes"
	"context"
	"fmt"
	"time"

	"github.com/WelcomerTeam/Discord/discord"
	sandwich "github.com/WelcomerTeam/Sandwich-Daemon/protobuf"
	subway "github.com/WelcomerTeam/Subway/subway"
	"github.com/WelcomerTeam/Welcomer/welcomer-interactions/internal"
	jsoniter "github.com/json-iterator/go"
)

func NewGeneralCog() *GeneralCog {
	return &GeneralCog{
		InteractionCommands: subway.SetupInteractionCommandable(&subway.InteractionCommandable{}),
	}
}

type GeneralCog struct {
	InteractionCommands *subway.InteractionCommandable
}

// Assert types.

var (
	_ subway.Cog                        = (*GeneralCog)(nil)
	_ subway.CogWithInteractionCommands = (*GeneralCog)(nil)
)

func (p *GeneralCog) CogInfo() *subway.CogInfo {
	return &subway.CogInfo{
		Name:        "GeneralCog",
		Description: "General commands",
	}
}

func (p *GeneralCog) GetInteractionCommandable() *subway.InteractionCommandable {
	return p.InteractionCommands
}

func (p *GeneralCog) RegisterCog(sub *subway.Subway) error {
	p.InteractionCommands.MustAddInteractionCommand(&subway.InteractionCommandable{
		Name:        "ping",
		Description: "Gets round trip API latency",
		Handler: func(ctx context.Context, sub *subway.Subway, interaction discord.Interaction) (*discord.InteractionResponse, error) {
			ttfb := time.Since(interaction.ID.Time()).Milliseconds()
			start := time.Now()

			err := interaction.SendResponse(
				sub.EmptySession,
				discord.InteractionCallbackTypeChannelMessageSource,
				&discord.InteractionCallbackData{
					Content: fmt.Sprintf(
						"Interaction TTFB: %dms\nHTTP Latency: ...",
						ttfb,
					),
				})
			if err != nil {
				return nil, err
			}

			_, err = interaction.EditOriginalResponse(
				sub.EmptySession,
				discord.WebhookMessageParams{
					Content: fmt.Sprintf(
						"Interaction TTFB: %dms\nHTTP Latency: %dms",
						ttfb,
						time.Since(start).Milliseconds(),
					),
				})
			if err != nil {
				return nil, err
			}

			return nil, nil
		},
	})

	p.InteractionCommands.MustAddInteractionCommand(&subway.InteractionCommandable{
		Name:        "file_test",
		Description: "Test file uploads",
		Handler: func(ctx context.Context, sub *subway.Subway, interaction discord.Interaction) (*discord.InteractionResponse, error) {
			return &discord.InteractionResponse{
				Type: discord.InteractionCallbackTypeDeferredUpdateMessage,
				Data: &discord.InteractionCallbackData{
					Content: "Hello World",
					Files: []*discord.File{
						{
							Name:        "test.txt",
							ContentType: "text/plain",
							Reader:      bytes.NewBufferString("Hello World"),
						},
					},
				},
			}, nil
		},
	})

	p.InteractionCommands.MustAddInteractionCommand(&subway.InteractionCommandable{
		Name:        "modal_test",
		Description: "Test modals with submit",
		Handler: func(ctx context.Context, sub *subway.Subway, interaction discord.Interaction) (*discord.InteractionResponse, error) {
			println("options:", interaction.Data.CustomID)
			for i, k := range interaction.Data.Options {
				println(i, k.Focused, k.Name, k.Type, k.Value)
				for j, l := range k.Options {
					println("-", j, l.Name, l.Value)
				}
			}

			err := interaction.SendResponse(sub.EmptySession, discord.InteractionCallbackTypeModal, &discord.InteractionCallbackData{
				CustomID: "test",
				Title:    "modal test!",
				Components: []*discord.InteractionComponent{
					{
						Type: discord.InteractionComponentTypeActionRow,
						Components: []*discord.InteractionComponent{
							{
								Label:    "name",
								Type:     discord.InteractionComponentTypeTextInput,
								Style:    discord.InteractionComponentStyleShort,
								CustomID: "name",
							}},
					}, {
						Type: discord.InteractionComponentTypeActionRow,
						Components: []*discord.InteractionComponent{
							{
								Label:    "pronouns",
								Type:     discord.InteractionComponentTypeTextInput,
								Style:    discord.InteractionComponentStyleParagraph,
								CustomID: "pronouns",
							},
						},
					},
				},
			})
			if err != nil {
				println(err.Error())
			}

			return nil, err
		},
	})

	// TODO: add COMPONENT_TYPE handler.
	// - Register queued listeners. Channel with timeout. Return DEFERRED_UPDATE_MESSAGE_RESPONSE.
	// - Check for global or cog InteractionHandler
	p.InteractionCommands.MustAddInteractionCommand(&subway.InteractionCommandable{
		Name:        "button_test",
		Description: "Test buttons with callback",
		Handler: func(ctx context.Context, sub *subway.Subway, interaction discord.Interaction) (*discord.InteractionResponse, error) {
			return &discord.InteractionResponse{
				Type: discord.InteractionCallbackTypeChannelMessageSource,
				Data: &discord.InteractionCallbackData{
					Content: "Buttons!",
					Components: []*discord.InteractionComponent{
						{
							Type: discord.InteractionComponentTypeActionRow,
							Components: []*discord.InteractionComponent{
								{
									Type:     discord.InteractionComponentTypeButton,
									Style:    discord.InteractionComponentStylePrimary,
									Label:    "Click Me",
									CustomID: "test",
								},
							},
						},
					},
				},
			}, nil
		},
	})

	p.InteractionCommands.MustAddInteractionCommand(&subway.InteractionCommandable{
		Name:        "argument_test",
		Description: "Test arguments",
		ArgumentParameter: []subway.ArgumentParameter{
			{
				Required:     true,
				Name:         "name",
				ArgumentType: subway.ArgumentTypeString,
				Description:  "your name",
			}, {
				Required:     false,
				Name:         "pronouns",
				ArgumentType: subway.ArgumentTypeString,
				Description:  "your pronouns",
			},
		},
		Handler: func(ctx context.Context, sub *subway.Subway, interaction discord.Interaction) (*discord.InteractionResponse, error) {
			name := subway.MustGetArgument(ctx, "name").MustString()
			pronouns := subway.MustGetArgument(ctx, "pronouns").MustString()

			return &discord.InteractionResponse{
				Type: discord.InteractionCallbackTypeChannelMessageSource,
				Data: &discord.InteractionCallbackData{
					Content: name + " " + pronouns,
				},
			}, nil
		},
	})

	p.InteractionCommands.MustAddInteractionCommand(&subway.InteractionCommandable{
		Name:        "context_test",
		Description: "Test custom context values",
		Handler: func(ctx context.Context, sub *subway.Subway, interaction discord.Interaction) (*discord.InteractionResponse, error) {
			text := ctx.Value("test").(string)

			return &discord.InteractionResponse{
				Type: discord.InteractionCallbackTypeChannelMessageSource,
				Data: &discord.InteractionCallbackData{
					Content: text,
				},
			}, nil
		},
	})

	p.InteractionCommands.MustAddInteractionCommand(&subway.InteractionCommandable{
		Name:        "relay_test",
		Description: "Test daemon event relay",
		Handler: func(ctx context.Context, sub *subway.Subway, interaction discord.Interaction) (*discord.InteractionResponse, error) {
			manager := internal.GetManagerNameFromContext(ctx)

			interaction.Member.GuildID = interaction.GuildID

			mem, err := jsoniter.Marshal(*interaction.Member)
			if err != nil {
				return &discord.InteractionResponse{
					Type: discord.InteractionCallbackTypeChannelMessageSource,
					Data: &discord.InteractionCallbackData{
						Content: err.Error(),
					},
				}, nil
			}

			res, err := sub.SandwichClient.RelayMessage(ctx, &sandwich.RelayMessageRequest{
				Manager: manager,
				Type:    "WELCOMER_INVOKE_WELCOMER",
				Data:    mem,
			})
			if err != nil {
				return &discord.InteractionResponse{
					Type: discord.InteractionCallbackTypeChannelMessageSource,
					Data: &discord.InteractionCallbackData{
						Content: err.Error(),
					},
				}, nil
			}

			return &discord.InteractionResponse{
				Type: discord.InteractionCallbackTypeChannelMessageSource,
				Data: &discord.InteractionCallbackData{
					Content: fmt.Sprintf("%t %s", res.Ok, res.Error),
				},
			}, nil
		},
	})

	return nil
}
