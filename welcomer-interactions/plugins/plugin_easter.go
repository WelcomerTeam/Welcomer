package plugins

import (
	"context"
	"errors"
	"math/rand"
	"os"

	"github.com/WelcomerTeam/Discord/discord"
	subway "github.com/WelcomerTeam/Subway/subway"
	"github.com/WelcomerTeam/Welcomer/welcomer-core"
	"github.com/WelcomerTeam/Welcomer/welcomer-core/database"
	"github.com/jackc/pgx/v4"
)

var IsEasterEnabled = os.Getenv("EASTER_EGG_ENABLED") == "true"

func NewEasterCog() *EasterCog {
	return &EasterCog{
		InteractionCommands: subway.SetupInteractionCommandable(&subway.InteractionCommandable{}),
	}
}

type EasterCog struct {
	InteractionCommands *subway.InteractionCommandable
}

// Assert types.

var (
	_ subway.Cog                        = (*EasterCog)(nil)
	_ subway.CogWithInteractionCommands = (*EasterCog)(nil)
)

type EasterEgg struct {
	ID                string
	TestingEmojiID    discord.Snowflake
	ProductionEmojiID discord.Snowflake

	Name        string
	Description string
	Rarity      int
}

var (
	RarityCommon    = 21875 // 175000 / 8
	RarityRare      = 8333  // 50000 / 6
	RarityEpic      = 4500  // 22500 / 5
	RarityLegendary = 833   // 2500 / 3
	RarityMythic    = 1
)

var EasterEggs = []EasterEgg{
	// Common
	{
		ID:                "white",
		TestingEmojiID:    1488655431722205345,
		ProductionEmojiID: 1488654789549232140,
		Name:              "White",
		Description:       "A clean, classic egg that looks closest to a white chicken egg fresh from the carton.",
		Rarity:            RarityCommon,
	},
	{
		ID:                "speckled",
		TestingEmojiID:    1488655424227119204,
		ProductionEmojiID: 1488654783521882152,
		Name:              "Speckled",
		Description:       "Produced by specific chicken breeds, this egg has a charming speckled shell that adds character to its appearance.",
		Rarity:            RarityCommon,
	},
	{
		ID:                "pink",
		TestingEmojiID:    1488655421400023072,
		ProductionEmojiID: 1488654781449900214,
		Name:              "Pink",
		Description:       "Produced by Buff Orpington chickens, this egg can have a soft pink shell that adds a touch of pastel charm to any Easter collection.",
		Rarity:            RarityCommon,
	},
	{
		ID:                "green",
		TestingEmojiID:    1488655409815224420,
		ProductionEmojiID: 1488654772864286866,
		Name:              "Green",
		Description:       "Whilst not a natural color for chicken eggs, this vibrant green egg adds a fun and festive twist to the Easter lineup. Do not have this with ham.",
		Rarity:            RarityCommon,
	},
	{
		ID:                "dark_chocolate",
		TestingEmojiID:    1488655404656230552,
		ProductionEmojiID: 1488654767948566638,
		Name:              "Dark Chocolate",
		Description:       "No. This is not a chocolate egg. This is an egg hatched by Black Copper Maran chickens.",
		Rarity:            RarityCommon,
	},
	{
		ID:                "cream",
		TestingEmojiID:    1488655403406463097,
		ProductionEmojiID: 1488654766459322458,
		Name:              "Cream",
		Description:       "A soft off-white egg that most closely resembles a cream-colored chicken egg.",
		Rarity:            RarityCommon,
	},
	{
		ID:                "brown",
		TestingEmojiID:    1488655401128951858,
		ProductionEmojiID: 1488654763913379840,
		Name:              "Brown",
		Description:       "A classic brown egg, commonly found in many households and farms.",
		Rarity:            RarityCommon,
	},
	{
		ID:                "blue",
		TestingEmojiID:    1488655400055210055,
		ProductionEmojiID: 1488654762382594058,
		Name:              "Blue",
		Description:       "A rare blue egg, often laid by Araucana chickens, adding a unique touch to any collection.",
		Rarity:            RarityCommon,
	},

	// Rare
	{
		ID:                "well_done",
		TestingEmojiID:    1488655430715707462,
		ProductionEmojiID: 1488654788592799744,
		Name:              "Well Done",
		Description:       "This egg has been cooked to perfection. Cook an egg in boiling water for 6 to 8 minutes.",
		Rarity:            RarityRare,
	},
	{
		ID:                "missing_texture",
		TestingEmojiID:    1488655417709039666,
		ProductionEmojiID: 1488654777951850516,
		Name:              "Missing Texture",
		Description:       "Did you remember to download counter strike source? At least the model is there, just the texture is missing.",
		Rarity:            RarityRare,
	},
	{
		ID:                "mini",
		TestingEmojiID:    1488655415943237632,
		ProductionEmojiID: 1488654776551084143,
		Name:              "Mini",
		Description:       "Mini eggs are small, bite-sized confections that are often coated in a colourful candy shell.",
		Rarity:            RarityRare,
	},
	{
		ID:                "minecraft",
		TestingEmojiID:    1488655412503904277,
		ProductionEmojiID: 1488654775275880559,
		Name:              "Minecraft",
		Description:       "The chicken that laid this one looked a little different from the rest. It was a cube, and the egg it laid was a cube too.",
		Rarity:            RarityRare,
	},
	{
		ID:                "instagram",
		TestingEmojiID:    1488655411304464434,
		ProductionEmojiID: 1488654773891764236,
		Name:              "Instagram",
		Description:       "https://en.wikipedia.org/wiki/Instagram_egg",
		Rarity:            RarityRare,
	},
	{
		ID:                "golden",
		TestingEmojiID:    1488655408628371588,
		ProductionEmojiID: 1488654771404542004,
		Name:              "Golden",
		Description:       "我值得拥有财富，金钱不断流向我，我的生活充满富足与繁荣。",
		Rarity:            RarityRare,
	},

	// Epic
	{
		ID:                "welcoming",
		TestingEmojiID:    1488655429650354266,
		ProductionEmojiID: 1488654787430846524,
		Name:              "Welcoming",
		Description:       "I recognise this egg! It is here to welcome new members to the server, and to give them a warm, eggy hug on their first day. It is a symbol of hospitality and friendliness.",
		Rarity:            RarityEpic,
	},
	{
		ID:                "suspicious",
		TestingEmojiID:    1488655425631944744,
		ProductionEmojiID: 1488654784624984134,
		Name:              "Suspicious",
		Description:       "This egg seems a bit sus. I cannot tell if it is safe to eat.",
		Rarity:            RarityEpic,
	},
	{
		ID:                "onyx",
		TestingEmojiID:    1488655420472950844,
		ProductionEmojiID: 1488654780573417492,
		Name:              "Onyx",
		Description:       "A sleek, shadowy egg with polished-stone vibes and a suspiciously dramatic aura.",
		Rarity:            RarityEpic,
	},
	{
		ID:                "discord",
		TestingEmojiID:    1488655406141014046,
		ProductionEmojiID: 1488654769198203040,
		Name:              "Discord",
		Description:       "I'm not sure what laid this egg, or what is inside it.",
		Rarity:            RarityEpic,
	},
	{
		ID:                "cracked",
		TestingEmojiID:    1488655402194436278,
		ProductionEmojiID: 1488654765117280506,
		Name:              "Cracked",
		Description:       "This egg has been through a lot, but it is still whole and beautiful.",
		Rarity:            RarityEpic,
	},

	// Legendary
	{
		ID:                "thonking",
		TestingEmojiID:    1488655428697985125,
		ProductionEmojiID: 1488654786109640814,
		Name:              "Thonking",
		Description:       "This egg is deep in thought. I am not sure how that makes sense, but it is a very thonk-worthy egg.",
		Rarity:            RarityLegendary,
	},
	{
		ID:                "rock_shaped",
		TestingEmojiID:    1488655422456987648,
		ProductionEmojiID: 1488654782473437295,
		Name:              "Rock Shaped",
		Description:       "Ok this is definitely not an egg. This is just a rock that looks like an egg. It is very hard and not edible at all.",
		Rarity:            RarityLegendary,
	},
	{
		ID:                "galaxy",
		TestingEmojiID:    1488655407512817775,
		ProductionEmojiID: 1488654770196578434,
		Name:              "Galaxy",
		Description:       "Space? Stars? Planets? This egg has it all. It is a miniature universe in egg form.",
		Rarity:            RarityLegendary,
	},

	// Mythic
	{
		ID:                "mystery",
		TestingEmojiID:    1488655419290423357,
		ProductionEmojiID: 1488654779222589522,
		Name:              "Mystery",
		Description:       "Nobody is entirely sure what is inside this egg, and that is exactly how it likes it.",
		Rarity:            RarityMythic,
	},
}

func (e *EasterCog) CogInfo() *subway.CogInfo {
	return &subway.CogInfo{
		Name:        "Easter",
		Description: "Easter-related commands.",
	}
}

func (e *EasterCog) GetInteractionCommandable() *subway.InteractionCommandable {
	return e.InteractionCommands
}

func (e *EasterCog) RegisterCog(sub *subway.Subway) error {
	easterGroup := subway.NewSubcommandGroup(
		"easter",
		"Provides commands for easter event",
	)

	easterGroup.MustAddInteractionCommand(&subway.InteractionCommandable{
		Name:        "leaderboard",
		Description: "Shows the leaderboard of users with the most eggs in their collection on this server",

		Type: subway.InteractionCommandableTypeSubcommand,

		Handler: func(ctx context.Context, sub *subway.Subway, interaction discord.Interaction) (*discord.InteractionResponse, error) {
			leaderboard, err := welcomer.Queries.GetCollectedEasterEggsByGuildID(ctx, int64(interaction.Guild.ID))
			if err != nil && !errors.Is(err, pgx.ErrNoRows) {
				return nil, err
			}

			if len(leaderboard) == 0 {
				return &discord.InteractionResponse{
					Type: discord.InteractionCallbackTypeChannelMessageSource,
					Data: &discord.InteractionCallbackData{
						Embeds: welcomer.NewEmbed("No one has caught any eggs yet! Be the first one to catch an egg and top the leaderboard!\n-# You can find them on Welcome messages.", welcomer.EmbedColourInfo),
						Flags:  uint32(discord.MessageFlagEphemeral),
					},
				}, nil
			}

			var components []discord.InteractionComponent

			components = append(components, []discord.InteractionComponent{
				{
					Type:    discord.InteractionComponentTypeTextDisplay,
					Content: "Egg Catching Leaderboard",
				},
				{
					Type: discord.InteractionComponentTypeSeparator,
				},
			}...)

			for i, entry := range leaderboard {
				components = append(components, discord.InteractionComponent{
					Type:    discord.InteractionComponentTypeTextDisplay,
					Content: welcomer.Itoa(int64(i+1)) + ". <@" + welcomer.Itoa(int64(entry.UserID)) + "> - **" + welcomer.Itoa(int64(entry.Column1)) + "** egg" + welcomer.If(entry.Column1 != 1, "s", ""),
				})
			}

			return &discord.InteractionResponse{
				Type: discord.InteractionCallbackTypeChannelMessageSource,
				Data: &discord.InteractionCallbackData{
					Components: []discord.InteractionComponent{
						{
							Type:       discord.InteractionComponentTypeContainer,
							Components: components,
						},
					},
					Flags: uint32(discord.MessageFlagEphemeral + discord.MessageFlagIsComponentsV2),
				},
			}, nil
		},
	})

	easterGroup.MustAddInteractionCommand(&subway.InteractionCommandable{
		Name:        "collection",
		Description: "Lists your egg collection",

		Type: subway.InteractionCommandableTypeSubcommand,

		Handler: func(ctx context.Context, sub *subway.Subway, interaction discord.Interaction) (*discord.InteractionResponse, error) {
			collection, err := welcomer.Queries.GetEasterEggsByUserID(ctx, int64(interaction.GetUser().ID))
			if err != nil && !errors.Is(err, pgx.ErrNoRows) {
				return nil, err
			}

			if len(collection) == 0 {
				return &discord.InteractionResponse{
					Type: discord.InteractionCallbackTypeChannelMessageSource,
					Data: &discord.InteractionCallbackData{
						Embeds: welcomer.NewEmbed("You don't have any eggs in your collection yet! Go catch some eggs!\n-# You can find them on Welcome messages.", welcomer.EmbedColourInfo),
						Flags:  uint32(discord.MessageFlagEphemeral),
					},
				}, nil
			}

			var components []discord.InteractionComponent

			for _, eggCollection := range collection {
				var emojiID discord.Snowflake

				egg, found := getEggByID(eggCollection.ClaimedEgg)
				if !found {
					continue
				}

				if isProduction {
					emojiID = egg.ProductionEmojiID
				} else {
					emojiID = egg.TestingEmojiID
				}

				components = append(components, discord.InteractionComponent{
					Type: discord.InteractionComponentTypeTextDisplay,
					Content: welcomer.Itoa(int64(eggCollection.Count)) + "x <:" + egg.ID + ":" + emojiID.String() + "> **" + egg.Name + "**\n" +
						"-# " + egg.Description,
				})
			}

			return &discord.InteractionResponse{
				Type: discord.InteractionCallbackTypeChannelMessageSource,
				Data: &discord.InteractionCallbackData{
					Components: []discord.InteractionComponent{
						{
							Type:       discord.InteractionComponentTypeContainer,
							Components: components,
						},
					},
					Flags: uint32(discord.MessageFlagEphemeral + discord.MessageFlagIsComponentsV2),
				},
			}, nil
		},
	})

	sub.RegisterComponentListener("catch_egg:*", func(ctx context.Context, sub *subway.Subway, interaction discord.Interaction) (*discord.InteractionResponse, error) {
		if !IsEasterEnabled {
			return &discord.InteractionResponse{
				Type: discord.InteractionCallbackTypeChannelMessageSource,
				Data: &discord.InteractionCallbackData{
					Embeds: welcomer.NewEmbed("Sorry, this easter event is no longer active!", welcomer.EmbedColourWarn),
					Flags:  uint32(discord.MessageFlagEphemeral),
				},
			}, nil
		}

		egg := getRandomEgg()

		var emojiID discord.Snowflake

		if isProduction {
			emojiID = egg.ProductionEmojiID
		} else {
			emojiID = egg.TestingEmojiID
		}

		_, err := welcomer.Queries.InsertEasterEgg(ctx, database.InsertEasterEggParams{
			GuildID:    int64(interaction.Guild.ID),
			ChannelID:  int64(interaction.Channel.ID),
			MessageID:  int64(interaction.Message.ID),
			UserID:     int64(interaction.GetUser().ID),
			ClaimedEgg: egg.ID,
			WmUserID:   interaction.Data.CustomID[10:],
		})
		if err != nil && errors.Is(err, pgx.ErrNoRows) {
			return &discord.InteractionResponse{
				Type: discord.InteractionCallbackTypeChannelMessageSource,
				Data: &discord.InteractionCallbackData{
					Embeds: welcomer.NewEmbed("This egg has already been caught!", welcomer.EmbedColourWarn),
					Flags:  uint32(discord.MessageFlagEphemeral),
				},
			}, nil
		} else if err != nil {
			welcomer.Logger.Error().Err(err).Msg("Failed to insert easter egg")

			return nil, err
		}

		rarity, an := getRarityString(egg.Rarity)

		return &discord.InteractionResponse{
			Type: discord.InteractionCallbackTypeChannelMessageSource,
			Data: &discord.InteractionCallbackData{
				Components: []discord.InteractionComponent{
					{
						Type: discord.InteractionComponentTypeContainer,
						Components: []discord.InteractionComponent{
							{
								Type: discord.InteractionComponentTypeTextDisplay,
								Content: "You caught a" + welcomer.If(an, "n", "") + " **" + rarity + "** egg!\n\n" +
									"# <:" + egg.ID + ":" + emojiID.String() + ">\n\n" +
									"**" + egg.Name + "**\n" +
									"-# " + egg.Description,
							},
						},
					},
				},
				Flags: uint32(discord.MessageFlagEphemeral + discord.MessageFlagIsComponentsV2),
			},
		}, nil
	})

	e.InteractionCommands.MustAddInteractionCommand(easterGroup)

	return nil
}

var (
	totalRarity  int
	isProduction bool
)

func init() {
	isProduction = welcomer.GetEnvironmentType() != welcomer.EnvironmentTypeDevelopment

	for _, egg := range EasterEggs {
		totalRarity += egg.Rarity
	}
}

func getRandomEgg() EasterEgg {
	randomNum := rand.Intn(totalRarity)

	for _, egg := range EasterEggs {
		if randomNum < egg.Rarity {
			return egg
		}
		randomNum -= egg.Rarity
	}

	// This should never happen, but return the first egg just in case.
	return EasterEggs[0]
}

func getEggByID(id string) (EasterEgg, bool) {
	for _, egg := range EasterEggs {
		if egg.ID == id {
			return egg, true
		}
	}
	return EasterEgg{}, false
}

func getRarityString(rarity int) (string, bool) {
	switch rarity {
	case RarityCommon:
		return "Common", false
	case RarityRare:
		return "Rare", false
	case RarityEpic:
		return "Epic", true
	case RarityLegendary:
		return "Legendary", false
	case RarityMythic:
		return "Mythic", false
	default:
		return "Unknown Rarity", false
	}
}
