package plugins

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/WelcomerTeam/Discord/discord"
	subway "github.com/WelcomerTeam/Subway/subway"
	"github.com/WelcomerTeam/Welcomer/welcomer-core"
	"github.com/WelcomerTeam/Welcomer/welcomer-core/database"
)

var presets map[string]string = map[string]string{
	"agender":           "#000000,#B9B9B9,#FFFFFF,#B9F484,#FFFFFF,#B9B9B9,#000000",
	"aromantic":         "#3DA542,#A8D47A,#FFFFFF,#A9A9A9,#000000",
	"asexual":           "#000000,#A4A4A4,#FFFFFF,#810081",
	"bisexual":          "#D60270,#D60270,#9B4F96,#0038A8,#0038A8",
	"demisexual":        "#000000,#A4A4A4,#FFFFFF,#4A0057",
	"genderfluid":       "#FF75A2,#FFFFFF,#BE18D6,#000000,#333EBD",
	"genderqueer":       "#B57EDC,#FFFFFF,#4A8123",
	"lesbian":           "#D52D00,#FF9A56,#FFFFFF,#D362A4,#A30262",
	"nonbinary":         "#FFF430,#FFFFFF,#9C59D1,#000000",
	"pansexual":         "#FF218C,#FFD800,#21B1FF",
	"rainbow-inclusive": "#FFF430,#9C59D1,#784F17,#000000,#FFFFFF,#F5A9B8,#5BCEFA,#E40303,#FF8C00,#FFED00,#008026,#004DFF,#750787",
	"rainbow":           "#FF0018,#FFA52C,#FFFF41,#008018,#0000F9,#86007D",
	"transgender":       "#5BCEFA,#F5A9B8,#FFFFFF,#F5A9B8,#5BCEFA",
}

func NewPrideCog() *PrideCog {
	return &PrideCog{
		InteractionCommands: subway.SetupInteractionCommandable(&subway.InteractionCommandable{}),
	}
}

type PrideCog struct {
	InteractionCommands *subway.InteractionCommandable
}

// Assert types.
var (
	_ subway.Cog                        = (*PrideCog)(nil)
	_ subway.CogWithInteractionCommands = (*PrideCog)(nil)
)

func (p *PrideCog) CogInfo() *subway.CogInfo {
	return &subway.CogInfo{
		Name:        "Pride",
		Description: "A pride plugin for Welcomer.",
	}
}

func (p *PrideCog) GetInteractionCommandable() *subway.InteractionCommandable {
	return p.InteractionCommands
}

func toTitleCase(s string) string {
	s = strings.ReplaceAll(s, "-", " ")
	words := strings.Fields(s)

	for i, word := range words {
		if len(word) > 0 {
			words[i] = strings.ToUpper(string(word[0])) + strings.ToLower(word[1:])
		}
	}

	return strings.Join(words, " ")
}

func (p *PrideCog) RegisterCog(sub *subway.Subway) error {
	prideGroup := subway.NewSubcommandGroup(
		"pride",
		"Show off your pride on servers you join with a customizable background or just join in style.",
	)

	prideGroup.MustAddInteractionCommand(&subway.InteractionCommandable{
		Name:        "clear",
		Description: "Clear your pride background.",

		Type: subway.InteractionCommandableTypeSubcommand,

		Handler: func(ctx context.Context, sub *subway.Subway, interaction discord.Interaction) (*discord.InteractionResponse, error) {
			var user discord.User
			if interaction.Member != nil {
				user = *interaction.Member.User
			} else {
				user = *interaction.User
			}

			if _, err := welcomer.Queries.CreateOrUpdateUser(ctx, database.CreateOrUpdateUserParams{
				UserID:        int64(user.ID),
				Name:          welcomer.GetUserDisplayName(&user),
				Discriminator: user.Discriminator,
				AvatarHash:    user.Avatar,
				Background:    "",
			}); err != nil {
				return nil, err
			}

			return &discord.InteractionResponse{
				Type: discord.InteractionCallbackTypeChannelMessageSource,
				Data: &discord.InteractionCallbackData{
					Embeds: welcomer.NewEmbed("Your background has been cleared.", welcomer.EmbedColourSuccess),
					Flags:  uint32(discord.MessageFlagEphemeral),
				},
			}, nil
		},
	})

	prideGroup.MustAddInteractionCommand(&subway.InteractionCommandable{
		Name:        "set",
		Description: "Set your pride background.",

		Type: subway.InteractionCommandableTypeSubcommand,

		AutocompleteHandler: func(ctx context.Context, sub *subway.Subway, interaction discord.Interaction) ([]discord.ApplicationCommandOptionChoice, error) {
			choices := make([]discord.ApplicationCommandOptionChoice, 0, len(presets)+1)

			choices = append(choices, discord.ApplicationCommandOptionChoice{
				Name:  "Custom",
				Value: welcomer.StringToJsonLiteral("custom"),
			})

			for name := range presets {
				choices = append(choices, discord.ApplicationCommandOptionChoice{
					Name:  name,
					Value: welcomer.StringToJsonLiteral(name),
				})
			}

			// Sort presets alphabetically by name
			names := make([]string, 0, len(presets))
			for name := range presets {
				names = append(names, name)
			}

			sort.Strings(names)

			choices = choices[:1] // keep "Custom" at the top
			for _, name := range names {
				choices = append(choices, discord.ApplicationCommandOptionChoice{
					Name:  toTitleCase(name),
					Value: welcomer.StringToJsonLiteral(name),
				})
			}

			if background := subway.MustGetArgument(ctx, "background").MustString(); background != "" {
				filtered := make([]discord.ApplicationCommandOptionChoice, 0, len(choices))
				for _, choice := range choices {
					if strings.Contains(strings.ToLower(choice.Name), strings.ToLower(background)) {
						filtered = append(filtered, choice)
					}
				}
				choices = filtered
			}

			return choices, nil
		},

		ArgumentParameter: []subway.ArgumentParameter{
			{
				Required:     true,
				ArgumentType: subway.ArgumentTypeString,
				Name:         "direction",
				Description:  "The direction of the stripes (horizontal or vertical).",
				Choices: []discord.ApplicationCommandOptionChoice{
					{
						Name:  "Horizontal",
						Value: welcomer.StringToJsonLiteral("h"),
					},
					{
						Name:  "Vertical",
						Value: welcomer.StringToJsonLiteral("v"),
					},
				},
			},
			{
				Required:     true,
				ArgumentType: subway.ArgumentTypeString,
				Name:         "background",
				Description:  "The background you want to set.",
				Autocomplete: &welcomer.True,
			},
			{
				Required:     false,
				ArgumentType: subway.ArgumentTypeString,
				Name:         "custom",
				Description:  "A comma-separated list of hex codes.",
			},
		},

		Handler: func(ctx context.Context, sub *subway.Subway, interaction discord.Interaction) (*discord.InteractionResponse, error) {
			var user discord.User
			if interaction.Member != nil {
				user = *interaction.Member.User
			} else {
				user = *interaction.User
			}

			background := subway.MustGetArgument(ctx, "background").MustString()
			direction := subway.MustGetArgument(ctx, "direction").MustString()
			custom := subway.MustGetArgument(ctx, "custom").MustString()

			var val string

			if background == "custom" {
				if custom == "" {
					return &discord.InteractionResponse{
						Type: discord.InteractionCallbackTypeChannelMessageSource,
						Data: &discord.InteractionCallbackData{
							Embeds: welcomer.NewEmbed("Custom background must be provided when background is set to custom.", welcomer.EmbedColourError),
							Flags:  uint32(discord.MessageFlagEphemeral),
						},
					}, nil
				} else {
					vals := []string{}

					if strings.Contains(custom, ",") {
						vals = strings.Split(custom, ",")
					} else {
						vals = []string{custom}
					}

					for i, v := range vals {
						v = strings.TrimPrefix(v, "#")
						if len(v) != 6 {
							return &discord.InteractionResponse{
								Type: discord.InteractionCallbackTypeChannelMessageSource,
								Data: &discord.InteractionCallbackData{
									Embeds: welcomer.NewEmbed(fmt.Sprintf("Invalid hex code `%s` at index %d. Must be 6 characters long.", v, i), welcomer.EmbedColourError),
									Flags:  uint32(discord.MessageFlagEphemeral),
								},
							}, nil
						}

						_, err := strconv.ParseInt(v, 16, 32)
						if err != nil {
							return &discord.InteractionResponse{
								Type: discord.InteractionCallbackTypeChannelMessageSource,
								Data: &discord.InteractionCallbackData{
									Embeds: welcomer.NewEmbed(fmt.Sprintf("Invalid hex code `%s` at index %d.", v, i), welcomer.EmbedColourError),
									Flags:  uint32(discord.MessageFlagEphemeral),
								},
							}, nil
						}

						if val != "" {
							val += ","
						}

						val += "#" + v
					}
				}
			} else {
				var ok bool

				val, ok = presets[background]
				if !ok {
					return &discord.InteractionResponse{
						Type: discord.InteractionCallbackTypeChannelMessageSource,
						Data: &discord.InteractionCallbackData{
							Embeds: welcomer.NewEmbed("Invalid background preset. Please choose a valid preset or use 'custom'.", welcomer.EmbedColourError),
							Flags:  uint32(discord.MessageFlagEphemeral),
						},
					}, nil
				}
			}

			newBackground := "stripes:" + direction + ":" + strings.ReplaceAll(val, "#", "")

			if _, err := welcomer.Queries.CreateOrUpdateUser(ctx, database.CreateOrUpdateUserParams{
				UserID:        int64(user.ID),
				Name:          welcomer.GetUserDisplayName(&user),
				Discriminator: user.Discriminator,
				AvatarHash:    user.Avatar,
				Background:    newBackground,
			}); err != nil {
				welcomer.Logger.Error().Err(err).Msg("Failed to update user background")

				return nil, err
			}

			generateOptionsRaw := welcomer.GenerateImageOptionsRaw{
				ShowAvatar:         true,
				AvatarURL:          welcomer.GetUserAvatar(&user),
				Background:         newBackground,
				TextColor:          0xFFFFFFFF,
				ProfileBorderColor: 0xFFFFFFFF,
				GuildID:            int64(*interaction.GuildID),
				ImageBorderColor:   0xFFFFFFFF,
				Theme:              int32(welcomer.ImageThemeDefault),
				ImageBorderWidth:   16,
				ProfileBorderWidth: 8,
				ProfileBorderCurve: int32(welcomer.ImageProfileBorderTypeCircular),
				TextStroke:         true,
				AllowAnimated:      false,
			}

			var files []discord.File

			optionsJSON, err := json.Marshal(generateOptionsRaw)
			if err == nil {
				resp, err := http.DefaultClient.Post(os.Getenv("IMAGE_ADDRESS")+"/generate", "application/json", bytes.NewBuffer(optionsJSON))
				if err != nil {
					welcomer.Logger.Error().Err(err).Msg("Failed to generate pride background image")
				} else if resp.StatusCode == http.StatusOK {
					defer resp.Body.Close()

					body, err := io.ReadAll(resp.Body)
					if err != nil {
						welcomer.Logger.Error().Err(err).Msg("Failed to read pride background image response")
					} else {
						files = append(files, discord.File{
							Name:        "background.png",
							ContentType: "image/png",
							Reader:      bytes.NewBuffer(body),
						})
					}
				}
			}

			embeds := welcomer.NewEmbed("Your background has been set.", welcomer.EmbedColourSuccess)
			embeds[0].SetImage(discord.NewEmbedImage("attachment://background.png"))

			err = interaction.SendResponse(ctx, sub.EmptySession, discord.InteractionCallbackTypeChannelMessageSource, &discord.InteractionCallbackData{
				Embeds: embeds,
				Flags:  uint32(discord.MessageFlagEphemeral),
				Files:  files,
			})
			if err != nil {
				welcomer.Logger.Error().Err(err).Msg("Failed to send pride background set response")

				return nil, err
			}

			return &discord.InteractionResponse{
				Type: discord.InteractionCallbackTypeDeferredChannelMessageSource,
			}, nil
		},
	})

	p.InteractionCommands.MustAddInteractionCommand(prideGroup)

	return nil
}
