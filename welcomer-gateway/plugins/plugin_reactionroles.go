package plugins

import (
	"errors"
	"fmt"
	"slices"
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
		if guildMember.User == nil || channel.GuildID == nil {
			return nil
		}

		return r.OnReact(eventCtx, *channel.GuildID, messageID, emoji, guildMember, new(true))
	})

	r.EventHandler.RegisterOnMessageReactionRemoveEvent(func(eventCtx *sandwich.EventContext, channel *discord.Channel, messageID discord.Snowflake, emoji discord.Emoji, user *discord.User) error {
		if user == nil || channel.GuildID == nil {
			return nil
		}

		return r.OnReact(eventCtx, *channel.GuildID, messageID, emoji, discord.GuildMember{User: user}, new(false))
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

	// If a reaction role is deleted, set the message_id in the database to 0.
	r.EventHandler.RegisterOnMessageDeleteEvent(func(eventCtx *sandwich.EventContext, channel *discord.Channel, messageID discord.Snowflake) error {
		if channel.GuildID == nil {
			return nil
		}

		reactionRole, err := welcomer.Queries.GetReactionRoleSettingByMessageId(eventCtx.Context, database.GetReactionRoleSettingByMessageIdParams{
			MessageID: int64(messageID),
			GuildID:   int64(*channel.GuildID),
		})
		if err != nil {
			if !errors.Is(err, pgx.ErrNoRows) {
				welcomer.Logger.Warn().Err(err).Int64("guild_id", int64(*channel.GuildID)).Int64("message_id", int64(messageID)).Msg("Failed to get reaction role setting by message ID for message delete event")
			}

			return nil
		}

		_, err = welcomer.Queries.UpdateReactionRoleSettingMessageId(eventCtx.Context, database.UpdateReactionRoleSettingMessageIdParams{
			ReactionRoleID: reactionRole.ReactionRoleID,
			GuildID:        int64(*channel.GuildID),
			MessageID:      0,
		})
		if err != nil {
			welcomer.Logger.Warn().Err(err).Int64("guild_id", int64(*channel.GuildID)).Int64("message_id", int64(messageID)).Msg("Failed to update reaction role setting message ID to 0 for message delete event")

			return nil
		}

		return nil
	})

	// Call OnInvokeReactionRoles when CustomEventInvokeReactionRoles is triggered.
	r.EventHandler.RegisterEvent(core.CustomEventInvokeReactionRoles, nil, (welcomer.OnInvokeReactionRolesFuncType)(r.OnInvokeReactionRoles))

	return nil
}

func (r *ReactionRolesCog) OnReact(eventCtx *sandwich.EventContext, guildID, messageID discord.Snowflake, emoji discord.Emoji, member discord.GuildMember, assign *bool) error {
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
		welcomer.Logger.Warn().Int64("guild_id", int64(guildID)).Msg("Reaction role is not enabled")

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
		welcomer.Logger.Warn().Int64("guild_id", int64(guildID)).Msg("Reaction role emoji not found")

		return nil
	}

	member.GuildID = &guildID

	return r.OnInvokeReactionRoles(eventCtx, core.CustomEventInvokeReactionRolesStructure{
		Interaction:      nil,
		Member:           &member,
		ReactionRoleUUID: reactionRole.ReactionRoleID,
		RoleID:           roleID,
		Assign:           assign,
	})
}

func (r *ReactionRolesCog) OnInvokeReactionRoles(eventCtx *sandwich.EventContext, event core.CustomEventInvokeReactionRolesStructure) error {
	if err := r.OnInvokeReactionRolesInner(eventCtx, event); err != nil {
		welcomer.Logger.Warn().Err(err).
			Int64("guild_id", int64(*event.Member.GuildID)).
			Int64("user_id", int64(event.Member.User.ID)).
			Int64("role_id", int64(event.RoleID)).
			Msg("Failed to invoke reaction roles")

		if event.Interaction != nil {
			_, err := event.Interaction.EditOriginalResponse(eventCtx.Context, eventCtx.Session,
				discord.WebhookMessageParams{
					Embeds: welcomer.NewEmbed(fmt.Sprintf("There was an issue assigning your reaction role. Please try again later.`", event.RoleID), welcomer.EmbedColourSuccess),
					Flags:  discord.MessageFlagEphemeral,
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

		return err
	}

	return nil
}

func (r *ReactionRolesCog) OnInvokeReactionRolesInner(eventCtx *sandwich.EventContext, event core.CustomEventInvokeReactionRolesStructure) error {
	startedAt := time.Now()

	if len(event.Member.Roles) == 0 {
		_, err := welcomer.SandwichClient.RequestGuildChunk(eventCtx.Context, &pb.RequestGuildChunkRequest{
			GuildId:     int64(*event.Member.GuildID),
			AlwaysChunk: false,
		})
		if err != nil {
			welcomer.Logger.Error().Err(err).
				Int64("guild_id", int64(*event.Member.GuildID)).
				Int64("user_id", int64(event.Member.User.ID)).
				Msg("Failed to request guild chunk for reaction roles")

			return fmt.Errorf("failed to request guild chunk: %w", err)
		}

		membersPb, err := welcomer.SandwichClient.FetchGuildMember(eventCtx.Context, &pb.FetchGuildMemberRequest{
			GuildId: int64(*event.Member.GuildID),
			UserIds: []int64{int64(event.Member.User.ID)},
		})
		if err != nil {
			welcomer.Logger.Error().Err(err).
				Int64("guild_id", int64(*event.Member.GuildID)).
				Int64("user_id", int64(event.Member.User.ID)).
				Msg("Failed to fetch guild member for reaction roles")

			return fmt.Errorf("failed to fetch guild member: %w", err)
		}

		memberPb, ok := membersPb.GuildMembers[int64(event.Member.User.ID)]
		if !ok {
			welcomer.Logger.Error().
				Int64("guild_id", int64(*event.Member.GuildID)).
				Int64("user_id", int64(event.Member.User.ID)).
				Msg("Guild member not found in response for reaction roles")

			return fmt.Errorf("guild member not found in response")
		}

		memberPb.GuildID = int64(*event.Member.GuildID)
		event.Member = pb.PBToGuildMember(memberPb)
	}

	hasRole := slices.Contains(event.Member.Roles, event.RoleID)

	if !hasRole && (event.Assign == nil || *event.Assign) {
		welcomer.Logger.Info().
			Int64("guild_id", int64(*event.Member.GuildID)).
			Int64("user_id", int64(event.Member.User.ID)).
			Int64("role_id", int64(event.RoleID)).
			Msg("Assigning reaction role to user")

		err := event.Member.AddRoles(eventCtx.Context, eventCtx.Session, []discord.Snowflake{event.RoleID}, new("Automatically assigned with Reaction Roles"), true)
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
						Flags:  discord.MessageFlagEphemeral,
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
					Flags:  discord.MessageFlagEphemeral,
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
	} else if hasRole && (event.Assign == nil || !*event.Assign) {
		welcomer.Logger.Info().
			Int64("guild_id", int64(*event.Member.GuildID)).
			Int64("user_id", int64(event.Member.User.ID)).
			Int64("role_id", int64(event.RoleID)).
			Msg("Unassigning reaction role to user")

		err := event.Member.RemoveRoles(eventCtx.Context, eventCtx.Session, []discord.Snowflake{event.RoleID}, new("Automatically removed with Reaction Roles"), true)
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
						Flags:  discord.MessageFlagEphemeral,
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
			_, err := event.Interaction.EditOriginalResponse(eventCtx.Context, eventCtx.Session,
				discord.WebhookMessageParams{
					Embeds: welcomer.NewEmbed(fmt.Sprintf("Role <@&%d> removed successfully!", event.RoleID), welcomer.EmbedColourSuccess),
					Flags:  discord.MessageFlagEphemeral,
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
	} else {
		welcomer.Logger.Info().
			Int64("guild_id", int64(*event.Member.GuildID)).
			Int64("user_id", int64(event.Member.User.ID)).
			Int64("role_id", int64(event.RoleID)).
			Bool("assign", welcomer.IfFunc(
				event.Assign != nil,
				func() bool {
					return *event.Assign
				},
				func() bool {
					return false
				},
			)).
			Bool("has_role", hasRole).
			Msg("Nothing to do with reaction role")
	}

	return nil
}
