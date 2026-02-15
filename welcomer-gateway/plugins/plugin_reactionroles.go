package plugins

import (
	"errors"
	"fmt"
	"time"

	"github.com/WelcomerTeam/Discord/discord"
	sandwich_daemon "github.com/WelcomerTeam/Sandwich-Daemon"
	pb "github.com/WelcomerTeam/Sandwich-Daemon/proto"
	sandwich "github.com/WelcomerTeam/Sandwich/sandwich"
	"github.com/WelcomerTeam/Welcomer/welcomer-core"
	core "github.com/WelcomerTeam/Welcomer/welcomer-core"
	"github.com/WelcomerTeam/Welcomer/welcomer-core/database"
	"github.com/jackc/pgx/v4"
)

type ReactionRolesCog struct {
	EventHandler *sandwich.Handlers
}

// Assert types.

var (
	_ sandwich.Cog           = (*ReactionRolesCog)(nil)
	_ sandwich.CogWithEvents = (*ReactionRolesCog)(nil)
)

func NewReactionRolesCog() *ReactionRolesCog {
	return &ReactionRolesCog{
		EventHandler: sandwich.SetupHandler(nil),
	}
}

func (c *ReactionRolesCog) CogInfo() *sandwich.CogInfo {
	return &sandwich.CogInfo{
		Name:        "Reaction Roles",
		Description: "Provides the functionality for the 'Reaction Roles' feature",
	}
}

func (c *ReactionRolesCog) GetEventHandlers() *sandwich.Handlers {
	return c.EventHandler
}

func (r *ReactionRolesCog) RegisterCog(bot *sandwich.Bot) error {
	// Register reaction add/remove handler.

	r.EventHandler.RegisterOnMessageReactionAddEvent(func(eventCtx *sandwich.EventContext, channel *discord.Channel, messageID discord.Snowflake, emoji discord.Emoji, guildMember discord.GuildMember) error {
		println("REACTION ADD")
		if guildMember.User == nil || channel.GuildID == nil {
			return nil
		}

		return r.OnReact(eventCtx, *channel.GuildID, messageID, emoji, guildMember)
	})

	r.EventHandler.RegisterOnMessageReactionRemoveEvent(func(eventCtx *sandwich.EventContext, channel *discord.Channel, messageID discord.Snowflake, emoji discord.Emoji, user *discord.User) error {
		println("REACTION REMOVE")
		if user == nil || channel.GuildID == nil {
			return nil
		}

		return r.OnReact(eventCtx, *channel.GuildID, messageID, emoji, discord.GuildMember{User: user})
	})

	r.EventHandler.RegisterEventHandler(core.CustomEventInvokeReactionRoles, func(eventCtx *sandwich.EventContext, payload sandwich_daemon.ProducedPayload) error {
		var invokeReactionRolesPayload core.CustomEventInvokeReactionRolesStructure
		if err := eventCtx.DecodeContent(payload, &invokeReactionRolesPayload); err != nil {
			return fmt.Errorf("failed to unmarshal payload: %w", err)
		}

		if invokeReactionRolesPayload.Member.GuildID != nil {
			eventCtx.Guild = sandwich.NewGuild(*invokeReactionRolesPayload.Member.GuildID)
		}

		eventCtx.EventHandler.EventsMu.RLock()
		defer eventCtx.EventHandler.EventsMu.RUnlock()

		for _, event := range eventCtx.EventHandler.Events {
			if f, ok := event.(welcomer.OnInvokeReactionRolesFuncType); ok {
				return eventCtx.Handlers.WrapFuncType(eventCtx, f(eventCtx, invokeReactionRolesPayload))
			}
		}

		return nil
	})

	// Call OnInvokeReactionRoles when CustomEventInvokeReactionRoles is triggered.
	r.EventHandler.RegisterEvent(core.CustomEventInvokeReactionRoles, nil, (welcomer.OnInvokeReactionRolesFuncType)(r.OnInvokeReactionRoles))

	return nil
}

func (r *ReactionRolesCog) OnReact(eventCtx *sandwich.EventContext, guildID, messageID discord.Snowflake, emoji discord.Emoji, member discord.GuildMember) error {
	reactionRole, err := welcomer.Queries.GetReactionRoleSettingByMessageId(eventCtx.Context, database.GetReactionRoleSettingByMessageIdParams{
		MessageID: int64(messageID),
		GuildID:   int64(guildID),
	})
	if err != nil {
		if !errors.Is(err, pgx.ErrNoRows) {
			welcomer.Logger.Warn().Err(err).Int64("guild_id", int64(guildID)).Int64("message_id", int64(messageID)).Msg("Failed to get reaction role setting by message ID")
		}

		return nil
	}

	if !reactionRole.ToggleEnabled {
		return nil
	}

	reactionRoleSettings := welcomer.UnmarshalReactionRolesJSON(reactionRole.Roles.Bytes)

	var found bool
	var roleID discord.Snowflake

	for _, setting := range reactionRoleSettings {
		if setting.Emoji == emoji.Name || setting.Emoji == emoji.ID.String() {
			roleID = setting.RoleID
			found = true

			break
		}
	}

	if !found {
		return nil
	}

	return r.OnInvokeReactionRoles(eventCtx, core.CustomEventInvokeReactionRolesStructure{
		Interaction:      nil,
		Member:           &member,
		ReactionRoleUUID: reactionRole.ReactionRoleID,
		RoleID:           roleID,
	})
}

func (r *ReactionRolesCog) OnInvokeReactionRoles(eventCtx *sandwich.EventContext, event core.CustomEventInvokeReactionRolesStructure) error {
	startedAt := time.Now()

	if len(event.Member.Roles) == 0 {
		membersPb, err := welcomer.SandwichClient.FetchGuildMember(eventCtx.Context, &pb.FetchGuildMemberRequest{
			GuildId: int64(*event.Member.GuildID),
			UserIds: []int64{int64(event.Member.User.ID)},
		})
		if err != nil {
			return fmt.Errorf("failed to fetch guild member: %w", err)
		}

		memberPb, ok := membersPb.GuildMembers[int64(event.Member.User.ID)]
		if !ok {
			return fmt.Errorf("guild member not found in response")
		}

		event.Member = pb.PBToGuildMember(memberPb)
	}

	var hasRole bool

	println(event.RoleID, len(event.Member.Roles))

	for _, role := range event.Member.Roles {
		if role == event.RoleID {
			hasRole = true

			break
		}
	}

	println("has role:", hasRole)

	if !hasRole {
		err := event.Member.AddRoles(eventCtx.Context, eventCtx.Session, []discord.Snowflake{event.RoleID}, welcomer.ToPointer("Automatically assigned with Reaction Roles"), true)
		if err != nil {
			welcomer.Logger.Error().Err(err).
				Int64("guild_id", int64(*event.Member.GuildID)).
				Int64("user_id", int64(event.Member.User.ID)).
				Int64("role_id", int64(event.RoleID)).
				Msg("Failed to add role to member for reaction roles")

			if event.Interaction != nil {
				_, err = event.Interaction.EditOriginalResponse(eventCtx.Context, eventCtx.Session,
					discord.WebhookMessageParams{
						Embeds: welcomer.NewEmbed(fmt.Sprintf("There was an error adding the <@&%d> role to you.", event.RoleID), welcomer.EmbedColourError),
					},
				)
				if err != nil {
					welcomer.Logger.Warn().Err(err).
						Int64("guild_id", int64(*event.Member.GuildID)).
						Int64("user_id", int64(event.Member.User.ID)).
						Int64("role_id", int64(event.RoleID)).
						Msg("Failed to send interaction response for failed add role in reaction roles")
				}
			}

			return nil
		}

		welcomer.PusherGuildScience.Push(
			eventCtx.Context,
			eventCtx.Guild.ID,
			event.Member.User.ID,
			database.ScienceGuildEventTypeReactionRoleGiven,
			welcomer.GuildScienceReactionRoleGivenRemoved{
				TimeToResolveMs:  time.Since(startedAt).Milliseconds(),
				RoleID:           event.RoleID,
				ReactionRoleUUID: event.ReactionRoleUUID,
			},
		)

		if event.Interaction != nil {
			_, err = event.Interaction.EditOriginalResponse(eventCtx.Context, eventCtx.Session,
				discord.WebhookMessageParams{
					Embeds: welcomer.NewEmbed(fmt.Sprintf("Role <@&%d> added successfully!", event.RoleID), welcomer.EmbedColourSuccess),
				},
			)
			if err != nil {
				welcomer.Logger.Warn().Err(err).
					Int64("guild_id", int64(*event.Member.GuildID)).
					Int64("user_id", int64(event.Member.User.ID)).
					Int64("role_id", int64(event.RoleID)).
					Msg("Failed to send interaction response for successful add role in reaction roles")
			}
		}

		return nil
	}

	err := event.Member.RemoveRoles(eventCtx.Context, eventCtx.Session, []discord.Snowflake{event.RoleID}, welcomer.ToPointer("Automatically removed with Reaction Roles"), true)
	if err != nil {
		welcomer.Logger.Error().Err(err).
			Int64("guild_id", int64(*event.Member.GuildID)).
			Int64("user_id", int64(event.Member.User.ID)).
			Int64("role_id", int64(event.RoleID)).
			Msg("Failed to remove role from member for reaction roles")

		if event.Interaction != nil {
			_, err = event.Interaction.EditOriginalResponse(eventCtx.Context, eventCtx.Session,
				discord.WebhookMessageParams{
					Embeds: welcomer.NewEmbed(fmt.Sprintf("There was an error removing the <@&%d> role from you.", event.RoleID), welcomer.EmbedColourError),
				},
			)
			if err != nil {
				welcomer.Logger.Warn().Err(err).
					Int64("guild_id", int64(*event.Member.GuildID)).
					Int64("user_id", int64(event.Member.User.ID)).
					Int64("role_id", int64(event.RoleID)).
					Msg("Failed to send interaction response for failed remove role in reaction roles")
			}
		}

		return nil
	}

	welcomer.PusherGuildScience.Push(
		eventCtx.Context,
		eventCtx.Guild.ID,
		event.Member.User.ID,
		database.ScienceGuildEventTypeReactionRoleRemoved,
		welcomer.GuildScienceReactionRoleGivenRemoved{
			TimeToResolveMs:  time.Since(startedAt).Milliseconds(),
			RoleID:           event.RoleID,
			ReactionRoleUUID: event.ReactionRoleUUID,
		},
	)

	if event.Interaction != nil {
		_, err = event.Interaction.EditOriginalResponse(eventCtx.Context, eventCtx.Session,
			discord.WebhookMessageParams{
				Embeds: welcomer.NewEmbed(fmt.Sprintf("Role <@&%d> removed successfully!", event.RoleID), welcomer.EmbedColourSuccess),
			},
		)
		if err != nil {
			welcomer.Logger.Warn().Err(err).
				Int64("guild_id", int64(*event.Member.GuildID)).
				Int64("user_id", int64(event.Member.User.ID)).
				Int64("role_id", int64(event.RoleID)).
				Msg("Failed to send interaction response for successful remove role in reaction roles")
		}
	}

	return nil
}
