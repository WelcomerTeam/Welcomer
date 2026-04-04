package plugins

import (
	"context"
	"errors"
	"math"
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

var userCache = map[discord.Snowflake]*discord.User{}

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
	TotalCommonEggs    = 10
	TotalRareEggs      = 11
	TotalEpicEggs      = 8
	TotalLegendaryEggs = 4
	TotalMythicEggs    = 1

	RarityCommon    = int(math.Ceil(175000 / float64(TotalCommonEggs)))
	RarityRare      = int(math.Ceil(50000 / float64(TotalRareEggs)))
	RarityEpic      = int(math.Ceil(22500 / float64(TotalEpicEggs)))
	RarityLegendary = int(math.Ceil(2500 / float64(TotalLegendaryEggs)))
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
	{
		ID:                "quail",
		TestingEmojiID:    0,
		ProductionEmojiID: 1489016355314204933,
		Name:              "Quail",
		Description:       "A small, delicate egg laid by quails. Often used in gourmet dishes.",
		Rarity:            RarityCommon,
	},
	{
		ID:                "ostrich",
		TestingEmojiID:    0,
		ProductionEmojiID: 1489016353842266132,
		Name:              "Ostrich",
		Description:       "A massive egg laid by an ostrich. It's the largest of all bird eggs.",
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
	{
		ID:                "wooden",
		TestingEmojiID:    0,
		ProductionEmojiID: 1489016361416917163,
		Name:              "Wooden",
		Description:       "A wooden egg, carved from a single piece of wood. It may give you splinters and it is very throwable.",
		Rarity:            RarityRare,
	},
	{
		ID:                "rgb",
		TestingEmojiID:    0,
		ProductionEmojiID: 1489016356526620672,
		Name:              "RGB",
		Description:       "I used some RGB lights to make this egg. The egg is a lie.",
		Rarity:            RarityRare,
	},
	{
		ID:                "jelly",
		TestingEmojiID:    0,
		ProductionEmojiID: 1489016351493197854,
		Name:              "Jelly",
		Description:       "Lime flavoured. Do not drop it, it will make a mess.",
		Rarity:            RarityRare,
	},
	{
		ID:                "wireframe",
		TestingEmojiID:    0,
		ProductionEmojiID: 1489016360133595306,
		Name:              "Wireframe",
		Description:       "I'm not very good at 3D modelling. The tris on this one could be made better, but it is still an egg.",
		Rarity:            RarityRare,
	},
	{
		ID:                "blueprint",
		TestingEmojiID:    0,
		ProductionEmojiID: 1489016349073215620,
		Name:              "Blueprint",
		Description:       "We are currently designing the next generation of eggs. This is a sneak peak into the future of protein-based lifeforms.",
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
	{
		ID:                "blurry",
		TestingEmojiID:    0,
		ProductionEmojiID: 1489016350209871883,
		Name:              "Blurry",
		Description:       "I'm not sure if my eyes are just bad or if I need to focus more.",
		Rarity:            RarityEpic,
	},
	{
		ID:                "stickers",
		TestingEmojiID:    0,
		ProductionEmojiID: 1489016358879494255,
		Name:              "Stickered",
		Description:       "I have covered this egg in stickers of gudetama. It is very cute, but I would not advise eating it.",
		Rarity:            RarityEpic,
	},
	{
		ID:                "loading",
		TestingEmojiID:    0,
		ProductionEmojiID: 1489016352655151305,
		Name:              "Loading",
		Description:       "Usually eggs load whilst in the chicken, but this one seems to be taking a while.",
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
	{
		ID:                "shadow",
		TestingEmojiID:    0,
		ProductionEmojiID: 1489016357696704724,
		Name:              "Shadow",
		Description:       "Nobody is sure where this egg came from. Those close to it feel like they are being sucked into it, however it is still an egg.",
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

		ArgumentParameter: []subway.ArgumentParameter{
			{
				ArgumentType: subway.ArgumentTypeString,
				Name:         "audience",
				Description:  "Filter leaderboard against guild or global.",
				Choices: []discord.ApplicationCommandOptionChoice{
					{Name: "Guild", Value: welcomer.StringToJsonLiteral("guild")},
					{Name: "Global", Value: welcomer.StringToJsonLiteral("global")},
				},
			},
		},

		Handler: func(ctx context.Context, sub *subway.Subway, interaction discord.Interaction) (*discord.InteractionResponse, error) {
			audience := subway.MustGetArgument(ctx, "audience").MustString()

			if audience != "guild" {
				leaderboard, err := welcomer.Queries.GetCollectedEasterEggs(ctx)
				if err != nil && !errors.Is(err, pgx.ErrNoRows) {
					return nil, err
				}

				var components []discord.InteractionComponent

				components = append(components, []discord.InteractionComponent{
					{
						Type:    discord.InteractionComponentTypeTextDisplay,
						Content: "Global Egg Catching Leaderboard",
					},
					{
						Type: discord.InteractionComponentTypeSeparator,
					},
				}...)

				session, err := welcomer.AcquireSession(ctx, welcomer.GetManagerNameFromContext(ctx))
				if err != nil {
					return nil, err
				}

				for i, entry := range leaderboard {
					user, ok := userCache[discord.Snowflake(entry.UserID)]
					if !ok {
						user, err := welcomer.FetchUserWithDiscordFallback(ctx, session, discord.Snowflake(entry.UserID))
						if err != nil {
							welcomer.Logger.Error().Err(err).
								Int64("user_id", entry.UserID).
								Msg("Failed to fetch user for easter egg leaderboard")
						}

						if user != nil {
							userCache[discord.Snowflake(entry.UserID)] = user
						}
					}

					var displayName string

					if user != nil {
						displayName = welcomer.GetUserDisplayName(user) + " "
					}

					displayName += "<@" + welcomer.Itoa(int64(entry.UserID)) + ">"

					components = append(components, discord.InteractionComponent{
						Type:    discord.InteractionComponentTypeTextDisplay,
						Content: welcomer.Itoa(int64(i+1)) + ". " + displayName + " - **" + welcomer.Itoa(int64(entry.Count)) + "** egg" + welcomer.If(entry.Count != 1, "s", ""),
					})
				}

				if len(leaderboard) > 0 {
					components = append(components, discord.InteractionComponent{
						Type: discord.InteractionComponentTypeSeparator,
					})

					components = append(components, discord.InteractionComponent{
						Type:    discord.InteractionComponentTypeTextDisplay,
						Content: "Total eggs caught globally: **" + welcomer.Itoa(int64(leaderboard[0].Total)) + "**",
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
			} else {
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
						Content: "Guild Egg Catching Leaderboard",
					},
					{
						Type: discord.InteractionComponentTypeSeparator,
					},
				}...)

				session, err := welcomer.AcquireSession(ctx, welcomer.GetManagerNameFromContext(ctx))
				if err != nil {
					return nil, err
				}

				for i, entry := range leaderboard {
					user, ok := userCache[discord.Snowflake(entry.UserID)]
					if !ok {
						user, err := welcomer.FetchUserWithDiscordFallback(ctx, session, discord.Snowflake(entry.UserID))
						if err != nil {
							welcomer.Logger.Error().Err(err).
								Int64("user_id", entry.UserID).
								Msg("Failed to fetch user for easter egg leaderboard")
						}

						if user != nil {
							userCache[discord.Snowflake(entry.UserID)] = user
						}
					}

					var displayName string

					if user != nil {
						displayName = welcomer.GetUserDisplayName(user) + " "
					}

					displayName += "<@" + welcomer.Itoa(int64(entry.UserID)) + ">"

					components = append(components, discord.InteractionComponent{
						Type:    discord.InteractionComponentTypeTextDisplay,
						Content: welcomer.Itoa(int64(i+1)) + ". " + displayName + " - **" + welcomer.Itoa(int64(entry.Count)) + "** egg" + welcomer.If(entry.Count != 1, "s", ""),
					})
				}

				if len(leaderboard) > 0 {
					components = append(components, discord.InteractionComponent{
						Type: discord.InteractionComponentTypeSeparator,
					})

					components = append(components, discord.InteractionComponent{
						Type:    discord.InteractionComponentTypeTextDisplay,
						Content: "Total eggs caught on this guild: **" + welcomer.Itoa(int64(leaderboard[0].Total)) + "**",
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
			}
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

			countCommon := 0
			countRare := 0
			countEpic := 0
			countLegendary := 0

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

				rarityString, _ := getRarityString(egg.Rarity)

				switch egg.Rarity {
				case RarityCommon:
					countCommon++
				case RarityRare:
					countRare++
				case RarityEpic:
					countEpic++
				case RarityLegendary:
					countLegendary++
				}

				components = append(components, discord.InteractionComponent{
					Type:    discord.InteractionComponentTypeTextDisplay,
					Content: welcomer.Itoa(int64(eggCollection.Count)) + "x <:" + egg.ID + ":" + emojiID.String() + "> **" + egg.Name + "** - **" + rarityString + "**\n",
					// "-# " + egg.Description,
				})
			}

			components = append(components, discord.InteractionComponent{
				Type: discord.InteractionComponentTypeSeparator,
			})

			components = append(components, discord.InteractionComponent{
				Type: discord.InteractionComponentTypeTextDisplay,
				Content: "Total eggs in collection: **" + welcomer.Itoa(int64(len(collection))) + "/" + welcomer.Itoa(int64(TotalCommonEggs+TotalRareEggs+TotalEpicEggs+TotalLegendaryEggs+TotalMythicEggs)) + "**\n" +
					"Common: **" + welcomer.Itoa(int64(countCommon)) + "/" + welcomer.Itoa(int64(TotalCommonEggs)) + "**\n" +
					"Rare: **" + welcomer.Itoa(int64(countRare)) + "/" + welcomer.Itoa(int64(TotalRareEggs)) + "**\n" +
					"Epic: **" + welcomer.Itoa(int64(countEpic)) + "/" + welcomer.Itoa(int64(TotalEpicEggs)) + "**\n" +
					"Legendary: **" + welcomer.Itoa(int64(countLegendary)) + "/" + welcomer.Itoa(int64(TotalLegendaryEggs)) + "**",
			})

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

		welcomer.Logger.Info().
			Str("egg_id", egg.ID).
			Str("egg_name", egg.Name).
			Str("rarity", rarity).
			Int64("user_id", int64(interaction.GetUser().ID)).
			Str("wmuserid", interaction.Data.CustomID[10:]).
			Msg("Egg caught")

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
