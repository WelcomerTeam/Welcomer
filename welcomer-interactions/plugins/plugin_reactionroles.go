package plugins

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/WelcomerTeam/Discord/discord"
	sandwich "github.com/WelcomerTeam/Sandwich-Daemon/proto"
	subway "github.com/WelcomerTeam/Subway/subway"
	"github.com/WelcomerTeam/Welcomer/welcomer-core"
	core "github.com/WelcomerTeam/Welcomer/welcomer-core"
	"github.com/WelcomerTeam/Welcomer/welcomer-core/database"
	"github.com/gofrs/uuid"
)

func NewReactionRolesCog() *ReactionRolesCog {
	return &ReactionRolesCog{}
}

type ReactionRolesCog struct{}

var _ subway.Cog = (*ReactionRolesCog)(nil)

func (c *ReactionRolesCog) CogInfo() *subway.CogInfo {
	return &subway.CogInfo{
		Name:        "Reaction Roles",
		Description: "Provides the cog for the 'Reaction Roles' feature.",
	}
}

func (r *ReactionRolesCog) RegisterCog(sub *subway.Subway) error {
	reactionRoleListener := &subway.ComponentListener{
		Channel:            nil,
		InitialInteraction: discord.Interaction{},
		Handler: func(ctx context.Context, sub *subway.Subway, interaction discord.Interaction) (*discord.InteractionResponse, error) {
			if interaction.GuildID == nil || interaction.Data.CustomID == "" {
				return nil, nil
			}

			customIDSplit := strings.Split(interaction.Data.CustomID, ":")
			if len(customIDSplit) != 3 {
				return nil, nil
			}

			if customIDSplit[0] != "reaction_role" {
				return nil, nil
			}

			reactionRoleUUID, err := uuid.FromString(customIDSplit[1])
			if err != nil {
				return nil, err
			}

			roleID, err := welcomer.Atoi(customIDSplit[2])
			if err != nil {
				return nil, err
			}

			reactionRole, err := welcomer.Queries.GetReactionRoleSettingById(ctx, database.GetReactionRoleSettingByIdParams{
				ReactionRoleID: reactionRoleUUID,
				GuildID:        int64(*interaction.GuildID),
			})
			if err != nil {
				return nil, err
			}

			if !reactionRole.ToggleEnabled {
				return &discord.InteractionResponse{
					Type: discord.InteractionCallbackTypeChannelMessageSource,
					Data: &discord.InteractionCallbackData{
						Embeds: []discord.Embed{
							{
								Description: "This reaction role is not available at the moment",
							},
						},
						Flags: uint32(discord.MessageFlagEphemeral),
					},
				}, nil
			}
			reactionRoleSettings := welcomer.UnmarshalReactionRolesJSON(reactionRole.Roles.Bytes)
			var found bool
			for _, setting := range reactionRoleSettings {
				if setting.RoleID == discord.Snowflake(roleID) {
					found = true
					break
				}
			}
			if !found {
				return &discord.InteractionResponse{
					Type: discord.InteractionCallbackTypeChannelMessageSource,
					Data: &discord.InteractionCallbackData{
						Embeds: []discord.Embed{
							{
								Description: "This reaction role is not available at the moment",
							},
						},
						Flags: uint32(discord.MessageFlagEphemeral),
					},
				}, nil
			}

			// GuildID may be missing, fill it in.
			interaction.Member.GuildID = interaction.GuildID

			data, err := json.Marshal(core.CustomEventInvokeReactionRolesStructure{
				Interaction:      &interaction,
				Member:           interaction.Member,
				ReactionRoleUUID: reactionRoleUUID,
				RoleID:           discord.Snowflake(roleID),
			})
			if err != nil {
				return nil, err
			}

			_, err = sub.SandwichClient.RelayMessage(ctx, &sandwich.RelayMessageRequest{
				Identifier: core.GetManagerNameFromContext(ctx),
				Type:       core.CustomEventInvokeReactionRoles,
				Data:       data,
			})
			if err != nil {
				return nil, err
			}
			return &discord.InteractionResponse{
				Type: discord.InteractionCallbackTypeDeferredChannelMessageSource,
			}, nil
		},
	}

	sub.ComponentListenersMu.Lock()
	sub.ComponentListeners["reaction_role:*"] = reactionRoleListener
	sub.ComponentListenersMu.Unlock()

	return nil
}
