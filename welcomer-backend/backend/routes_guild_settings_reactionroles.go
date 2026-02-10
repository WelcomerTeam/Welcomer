package backend

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	discord "github.com/WelcomerTeam/Discord/discord"
	sandwich "github.com/WelcomerTeam/Sandwich-Daemon/proto"
	"github.com/WelcomerTeam/Welcomer/welcomer-core"
	"github.com/WelcomerTeam/Welcomer/welcomer-core/database"
	"github.com/gin-gonic/gin"
	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v4"
)

const (
	MaximumOptionsEmojis   = 20
	MaximumOptionsButtons  = 25
	MaximumOptionsDropdown = 25
)

// Route GET /api/guild/:guildID/reactionroles.
func getGuildSettingsReactionRoles(ctx *gin.Context) {
	requireOAuthAuthorization(ctx, func(ctx *gin.Context) {
		requireGuildElevation(ctx, func(ctx *gin.Context) {
			guildID := tryGetGuildID(ctx)

			reactionroles, err := welcomer.Queries.GetReactionRolesGuildSettings(ctx, int64(guildID))
			if err != nil {
				if errors.Is(err, pgx.ErrNoRows) {
					reactionroles = &database.GuildSettingsReactionRoles{
						GuildID:       int64(guildID),
						ToggleEnabled: welcomer.DefaultReactionRoles.ToggleEnabled,
						ReactionRoles: welcomer.DefaultReactionRoles.ReactionRoles,
					}
				} else {
					welcomer.Logger.Warn().Err(err).Int64("guild_id", int64(guildID)).Msg("Failed to get guild reaction roles settings")

					ctx.JSON(http.StatusInternalServerError, NewBaseResponse(NewGenericErrorWithLineNumber(), nil))

					return
				}
			}

			partial := GuildSettingsReactionRolesSettingsToPartial(reactionroles)

			ctx.JSON(http.StatusOK, BaseResponse{
				Ok:   true,
				Data: partial,
			})
		})
	})
}

// Route POST /api/guild/:guildID/reactionroles.
func setGuildSettingsReactionRoles(ctx *gin.Context) {
	requireOAuthAuthorization(ctx, func(ctx *gin.Context) {
		requireGuildElevation(ctx, func(ctx *gin.Context) {
			partial := &GuildSettingsReactionRoles{}

			var err error

			err = ctx.BindJSON(partial)
			if err != nil {
				ctx.JSON(http.StatusBadRequest, BaseResponse{
					Ok:    false,
					Error: err.Error(),
				})

				return
			}

			for i, reactionRole := range partial.ReactionRoles {
				if reactionRole.ID == "" {
					partial.ReactionRoles[i].ID = uuid.Must(gen.NewV7()).String()
				}
			}

			guildID := tryGetGuildID(ctx)

			errGroup := doValidateReactionRoles(ctx, guildID, partial)
			if errGroup != nil && !errGroup.Empty() {
				ctx.JSON(http.StatusBadRequest, BaseResponse{
					Ok:    false,
					Error: errGroup.ErrorWithDelimiter("\n"),
				})

				return
			}

			oldReactionRoleSettings, err := welcomer.Queries.GetReactionRolesGuildSettings(ctx, int64(guildID))
			if err != nil && !errors.Is(err, pgx.ErrNoRows) {
				welcomer.Logger.Warn().Err(err).Int64("guild_id", int64(guildID)).Msg("Failed to get existing guild reaction roles settings for update")

				ctx.JSON(http.StatusInternalServerError, NewBaseResponse(NewGenericErrorWithLineNumber(), nil))

				return
			}

			user := tryGetUser(ctx)

			if partial.ToggleEnabled {
				eg, messageIDs := processReactionRolesSettingsChange(ctx, GuildSettingsReactionRolesSettingsToPartial(oldReactionRoleSettings), partial)

				// Update message IDs
				for id, messageID := range messageIDs {
					for i, k := range partial.ReactionRoles {
						if k.ID == id {
							partial.ReactionRoles[i].MessageID = messageID
						}
					}
				}

				if eg != nil && !eg.Empty() {
					welcomer.Logger.Warn().Err(eg).Int64("guild_id", int64(guildID)).Int64("user_id", int64(user.ID)).Msg("Failed to process reaction roles settings change")

					ctx.JSON(http.StatusInternalServerError, BaseResponse{
						Ok:    false,
						Error: eg.ErrorWithDelimiter("\n"),
					})

					return
				}
			}

			reactionroles := PartialToGuildSettingsReactionRolesSettings(int64(guildID), partial)
			databaseReactionRolesGuildSettings := database.CreateOrUpdateReactionRolesGuildSettingsParams(*reactionroles)

			welcomer.Logger.Info().Int64("guild_id", int64(guildID)).Interface("obj", *reactionroles).Int64("user_id", int64(user.ID)).Msg("Creating or updating reaction roles guild settings")

			err = welcomer.RetryWithFallback(
				func() error {
					_, err = welcomer.CreateOrUpdateReactionRolesGuildSettingsWithAudit(ctx, databaseReactionRolesGuildSettings, user.ID)

					return err
				},
				func() error {
					return welcomer.EnsureGuild(ctx, guildID)
				},
				nil,
			)
			if err != nil {
				welcomer.Logger.Warn().Err(err).Int64("guild_id", int64(guildID)).Int64("user_id", int64(user.ID)).Msg("Failed to create or update guild reaction roles settings")

				ctx.JSON(http.StatusInternalServerError, NewBaseResponse(NewGenericErrorWithLineNumber(), nil))

				return
			}

			getGuildSettingsReactionRoles(ctx)
		})
	})
}

type reactionRoleConfigurationsKey struct {
	ChannelID discord.Snowflake
	MessageID discord.Snowflake
}

type reactionRoleConfigurations struct {
	Old      *welcomer.GuildSettingsReactionRole
	New      *welcomer.GuildSettingsReactionRole
	NewIndex int
}

func processReactionRolesSettingsChange(ctx *gin.Context, old, new *GuildSettingsReactionRoles) (*welcomer.ErrorGroup, map[string]discord.Snowflake) {
	eg := welcomer.NewErrorGroup()

	configurationChanges := make(map[reactionRoleConfigurationsKey]reactionRoleConfigurations)

	for _, oldConfig := range old.ReactionRoles {
		key := reactionRoleConfigurationsKey{oldConfig.ChannelID, oldConfig.MessageID}
		_, ok := configurationChanges[key]
		if !ok {
			configurationChanges[key] = reactionRoleConfigurations{}
		}

		configuration := configurationChanges[key]
		configuration.Old = &oldConfig
		configurationChanges[key] = configuration
	}

	for i, newConfig := range new.ReactionRoles {
		key := reactionRoleConfigurationsKey{newConfig.ChannelID, newConfig.MessageID}
		_, ok := configurationChanges[key]
		if !ok {
			configurationChanges[key] = reactionRoleConfigurations{}
		}

		configuration := configurationChanges[key]
		configuration.New = &newConfig
		configuration.NewIndex = i
		configurationChanges[key] = configuration
	}

	messageIDs := make(map[string]discord.Snowflake)

	for _, config := range configurationChanges {
		if (config.Old != nil && config.Old.IsSystemMessage) || (config.New != nil && config.New.IsSystemMessage) {
			messageID := processReactionRolesSettingsChangeSystemMessage(ctx, eg, config.Old, config.New)
			messageIDs[config.New.ID] = messageID
		} else {
			processReactionRolesSettingsChangeNonSystemMessage(ctx, eg, config.Old, config.New)
		}
	}

	return eg, messageIDs
}

func processReactionRolesSettingsChangeSystemMessage(ctx *gin.Context, eg *welcomer.ErrorGroup, old, new *welcomer.GuildSettingsReactionRole) discord.Snowflake {
	if old != nil && new == nil {
		err := disableReactionRoleMessage(ctx, old.ChannelID, old.MessageID)
		if err != nil {
			eg.Add(fmt.Errorf("failed to disable message for removed system message reaction role configuration: %v", err))
		}
	}

	if !hasConfigurationChanged(old, new) {
		println("No configuration change detected for system message reaction role, skipping processing")
		return new.MessageID
	}

	var message *discord.Message
	var err error

	if hasConfigurationChangedMessage(old, new) && ((old == nil || new == nil) || old.ChannelID != new.ChannelID) {
		// If channel has changed, send new message.

		if new != nil {
			message, err = createReactionRoleMessage(ctx, eg, new)
			if err != nil {
				welcomer.Logger.Warn().Err(err).Int64("guild_id", int64(tryGetGuildID(ctx))).Int64("channel_id", int64(new.ChannelID)).Msg("Failed to create message for updated system message reaction role configuration")
			} else {
				if new != nil {
					new.MessageID = message.ID
				}
			}
		}

		if old != nil {
			// Disable old message.
			err = disableReactionRoleMessage(ctx, old.ChannelID, old.MessageID)
			if err != nil {
				eg.Add(fmt.Errorf("failed to disable message for removed system message reaction role configuration: %v", err))
			}
		}
	} else if hasConfigurationChangedRoles(old, new) || old.Message != new.Message || old.Enabled != new.Enabled {
		// If roles, message or enabled has changed, update existing message.

		if !new.Enabled {
			err = disableReactionRoleMessage(ctx, new.ChannelID, new.MessageID)
			if err != nil {
				eg.Add(fmt.Errorf("failed to disable message for disabled system message reaction role configuration: %v", err))
			}

			return new.MessageID
		}

		if message == nil {
			message, err = discord.GetChannelMessage(ctx, backend.BotSession, new.ChannelID, new.MessageID)
			if err != nil {
				welcomer.Logger.Warn().Err(err).Int64("guild_id", int64(tryGetGuildID(ctx))).Int64("channel_id", int64(new.ChannelID)).Int64("message_id", int64(new.MessageID)).Msg("Failed to get message for updated system message reaction role configuration")

				message, err = createReactionRoleMessage(ctx, eg, new)
				if err != nil {
					welcomer.Logger.Warn().Err(err).Int64("guild_id", int64(tryGetGuildID(ctx))).Int64("channel_id", int64(new.ChannelID)).Msg("Failed to create message for updated system message reaction role configuration")
				}

				return new.MessageID
			}
		}

		// If message cannot be retrieved or created, there is not much we can do to update the configuration, so just return and log the error.
		if message == nil {
			return new.MessageID
		}

		// Message has changed or button/dropdown configuration has changed.
		if old.Message != new.Message || hasConfigurationChangedRoles(old, new) || (old.Type != new.Type) || (!old.Enabled && new.Enabled) {
			if messageParams := setupMessageParamsForReactionRoleConfiguration(new); messageParams != nil {
				_, err = message.Edit(ctx, backend.BotSession, *messageParams)
				if err != nil {
					welcomer.Logger.Warn().Err(err).Int64("guild_id", int64(tryGetGuildID(ctx))).Int64("channel_id", int64(new.ChannelID)).Int64("message_id", int64(message.ID)).Msg("Failed to edit message for updated system message reaction role configuration")

					eg.Add(fmt.Errorf("failed to edit message for updated system message reaction role configuration: %v", err))

					return new.MessageID
				}
			}
		}

		if hasConfigurationChangedRoles(old, new) && new.Type == welcomer.ReactionRoleTypeEmoji {
			for _, reaction := range message.Reactions {
				found := false

				for _, option := range new.Roles {
					if option.Emoji == reaction.Emoji.ID.String() || option.Emoji == reaction.Emoji.Name {
						found = true

						break
					}
				}

				// Remove reactions that are no longer valid for the updated configuration.
				if !found {
					var emoji string

					if reaction.Emoji.ID != 0 {
						emoji = "_:" + reaction.Emoji.ID.String()
					} else {
						emoji = reaction.Emoji.Name
					}

					err = message.ClearReaction(ctx, backend.BotSession, emoji)
					if err != nil {
						welcomer.Logger.Warn().Err(err).Int64("guild_id", int64(tryGetGuildID(ctx))).Int64("channel_id", int64(new.ChannelID)).Int64("message_id", int64(message.ID)).Str("emoji", emoji).Msg("Failed to remove reaction from message for updated system message reaction role configuration")

						eg.Add(fmt.Errorf("failed to remove reaction '%s' from message for updated system message reaction role configuration: %v", reaction.Emoji, err))
					}
				}
			}

			for _, option := range new.Roles {
				if option.Emoji != "" {
					var emoji string

					if _, err := welcomer.Atoi(option.Emoji); err == nil {
						emoji = "_:" + option.Emoji
					} else {
						emoji = option.Emoji
					}

					err = message.AddReaction(ctx, backend.BotSession, emoji)
					if err != nil {
						welcomer.Logger.Warn().Err(err).Int64("guild_id", int64(tryGetGuildID(ctx))).Int64("channel_id", int64(new.ChannelID)).Int64("message_id", int64(message.ID)).Str("emoji", option.Emoji).Msg("Failed to add reaction to message for updated system message reaction role configuration")

						eg.Add(fmt.Errorf("failed to add reaction '%s' to message for updated system message reaction role configuration: %v", option.Emoji, err))
					}
				}
			}
		}
	}

	return new.MessageID
}

func processReactionRolesSettingsChangeNonSystemMessage(ctx *gin.Context, eg *welcomer.ErrorGroup, old, new *welcomer.GuildSettingsReactionRole) {
	return
}

func createReactionRoleMessage(ctx *gin.Context, eg *welcomer.ErrorGroup, new *welcomer.GuildSettingsReactionRole) (message *discord.Message, err error) {
	if messageParams := setupMessageParamsForReactionRoleConfiguration(new); messageParams != nil {
		channel := discord.Channel{ID: new.ChannelID}

		message, err = channel.Send(ctx, backend.BotSession, *messageParams)
		if err != nil {
			welcomer.Logger.Warn().Err(err).Int64("guild_id", int64(tryGetGuildID(ctx))).Int64("channel_id", int64(new.ChannelID)).Msg("Failed to send message for updated system message reaction role configuration")

			eg.Add(fmt.Errorf("failed to send message for updated system message reaction role configuration: %v", err))

			return nil, err
		}

		new.MessageID = message.ID

		// TODO: update database with new message ID if they have changed

		if new.Type == welcomer.ReactionRoleTypeEmoji {
			for _, option := range new.Roles {
				if option.Emoji != "" {
					var emoji string

					if _, err := welcomer.Atoi(option.Emoji); err == nil {
						emoji = "_:" + option.Emoji
					} else {
						emoji = option.Emoji
					}

					err = message.AddReaction(ctx, backend.BotSession, emoji)
					if err != nil {
						welcomer.Logger.Warn().Err(err).Int64("guild_id", int64(tryGetGuildID(ctx))).Int64("channel_id", int64(new.ChannelID)).Int64("message_id", int64(message.ID)).Str("emoji", option.Emoji).Msg("Failed to add reaction to message for updated system message reaction role configuration")

						eg.Add(fmt.Errorf("failed to add reaction '%s' to message for updated system message reaction role configuration: %v", option.Emoji, err))
					}
				}
			}
		}
	}

	return message, nil
}

func disableReactionRoleMessage(ctx *gin.Context, channelID, messageID discord.Snowflake) error {
	if messageID == 0 {
		return nil
	}

	message, err := discord.GetChannelMessage(ctx, backend.BotSession, channelID, messageID)
	if err != nil {
		welcomer.Logger.Warn().Err(err).Int64("guild_id", int64(tryGetGuildID(ctx))).Int64("channel_id", int64(channelID)).Int64("message_id", int64(messageID)).Msg("Failed to get message for removed system message reaction role configuration")

		return nil
	}

	_, err = message.Edit(ctx, backend.BotSession, discord.MessageParams{
		Components: []discord.InteractionComponent{
			{
				Type: discord.InteractionComponentTypeActionRow,
				Components: []discord.InteractionComponent{
					{
						Type:     discord.InteractionComponentTypeButton,
						CustomID: "disabled",
						Disabled: true,
						Label:    "This reaction role is no longer available",
						Style:    discord.InteractionComponentStyleSecondary,
					},
				},
			},
		},
	})
	if err != nil {
		welcomer.Logger.Warn().Err(err).Int64("guild_id", int64(tryGetGuildID(ctx))).Int64("channel_id", int64(channelID)).Int64("message_id", int64(messageID)).Msg("Failed to edit message for removed system message reaction role configuration")

		return err
	}

	return nil
}

func setupMessageParamsForReactionRoleConfiguration(config *welcomer.GuildSettingsReactionRole) *discord.MessageParams {
	var messageParams discord.MessageParams

	err := json.Unmarshal([]byte(config.Message), &messageParams)
	if err != nil {
		welcomer.Logger.Warn().Err(err).Int64("guild_id", int64(tryGetGuildID(nil))).Int64("channel_id", int64(config.ChannelID)).Int64("message_id", int64(config.MessageID)).Msg("Failed to unmarshal message params for reaction role configuration")

		return nil
	}

	if config.Type == welcomer.ReactionRoleTypeButtons || config.Type == welcomer.ReactionRoleTypeDropdown {
		messageParams.Components = createMessageComponentsForReactionRole(config)
	} else {
		messageParams.Components = []discord.InteractionComponent{}
	}

	messageParams.StickerIDs = nil

	return &messageParams
}

func createMessageComponentsForReactionRole(config *welcomer.GuildSettingsReactionRole) []discord.InteractionComponent {
	switch config.Type {
	case welcomer.ReactionRoleTypeButtons:
		var components []discord.InteractionComponent

		for i, option := range config.Roles {
			if i%5 == 0 {
				components = append(components, discord.InteractionComponent{
					Type: discord.InteractionComponentTypeActionRow,
				})
			}

			var emoji *discord.Emoji

			if option.Emoji != "" {
				if v, err := welcomer.Atoi(option.Emoji); err == nil {
					emoji = &discord.Emoji{
						ID:   discord.Snowflake(v),
						Name: "_",
					}
				} else {
					emoji = &discord.Emoji{
						Name: option.Emoji,
					}
				}
			}

			component := discord.InteractionComponent{
				CustomID: fmt.Sprintf("reaction_role|%d|%d|%d", config.ChannelID, config.MessageID, option.RoleID),
				Emoji:    emoji,
				Label:    option.Name,
				Style:    discord.InteractionComponentStylePrimary,
				Type:     discord.InteractionComponentTypeButton,
			}

			components[len(components)-1].Components = append(components[len(components)-1].Components, component)
		}

		return components
	case welcomer.ReactionRoleTypeDropdown:
		var options []discord.ApplicationSelectOption

		for _, option := range config.Roles {

			var emoji *discord.Emoji

			if option.Emoji != "" {
				if v, err := welcomer.Atoi(option.Emoji); err == nil {
					emoji = &discord.Emoji{
						ID:   discord.Snowflake(v),
						Name: "_",
					}
				} else {
					emoji = &discord.Emoji{
						Name: option.Emoji,
					}
				}
			}

			selectOption := discord.ApplicationSelectOption{
				Description: option.Description,
				Emoji:       emoji,
				Label:       option.Name,
				Value:       fmt.Sprintf("%d", option.RoleID),
			}

			options = append(options, selectOption)
		}

		return []discord.InteractionComponent{
			{
				Type: discord.InteractionComponentTypeActionRow,
				Components: []discord.InteractionComponent{
					{
						Type:     discord.InteractionComponentTypeStringSelect,
						CustomID: fmt.Sprintf("reaction_role|%d|%d", config.ChannelID, config.MessageID),
						Options:  options,
					},
				},
			},
		}
	default:
		return nil
	}
}

func hasConfigurationChangedMessage(old, new *welcomer.GuildSettingsReactionRole) bool {
	if old == nil || new == nil {
		return true
	}

	if old.ChannelID != new.ChannelID {
		return true
	}

	if old.MessageID != new.MessageID {
		return true
	}

	if old.Message != new.Message {
		return true
	}

	if old.Type != new.Type {
		return true
	}

	return false
}

func hasConfigurationChangedRoles(old, new *welcomer.GuildSettingsReactionRole) bool {
	if len(old.Roles) != len(new.Roles) {
		return true
	}

	for i := range old.Roles {
		if old.Roles[i] != new.Roles[i] {
			return true
		}

		if old.Roles[i].Name != new.Roles[i].Name {
			return true
		}

		if old.Roles[i].Emoji != new.Roles[i].Emoji {
			return true
		}

		if old.Roles[i].Description != new.Roles[i].Description {
			return true
		}
	}

	return false
}

func hasConfigurationChanged(old, new *welcomer.GuildSettingsReactionRole) bool {
	if old == nil || new == nil {
		return true
	}

	if old.Enabled != new.Enabled {
		return true
	}

	if hasConfigurationChangedMessage(old, new) {
		return true
	}

	if hasConfigurationChangedRoles(old, new) {
		return true
	}

	return false
}

func tryGetSnowflakeFromCtx(ctx *gin.Context, key string) discord.Snowflake {
	rawSnowflake := ctx.Param(key)

	snowflakeInt, err := welcomer.Atoi(rawSnowflake)
	if err != nil {
		return 0
	}

	return discord.Snowflake(snowflakeInt)
}

// Route GET /api/guild/:guildID/checkmessage/:channelID/:messageID.
func checkGuildSettingsReactionRolesMessage(ctx *gin.Context) {
	requireOAuthAuthorization(ctx, func(ctx *gin.Context) {
		requireGuildElevation(ctx, func(ctx *gin.Context) {
			guildID := tryGetGuildID(ctx)
			channelID := tryGetSnowflakeFromCtx(ctx, "channelID")
			messageID := tryGetSnowflakeFromCtx(ctx, "messageID")

			channel, err := welcomer.SandwichClient.FetchGuildChannel(ctx, &sandwich.FetchGuildChannelRequest{
				GuildId:    int64(guildID),
				ChannelIds: []int64{int64(channelID)},
			})
			if err != nil {
				welcomer.Logger.Warn().Err(err).Int64("guild_id", int64(guildID)).Int64("channel_id", int64(channelID)).Msg("Failed to fetch channel for reaction roles")

				ctx.JSON(http.StatusInternalServerError, NewGenericErrorWithLineNumber())

				return
			}

			if len(channel.Channels) == 0 {
				welcomer.Logger.Warn().Int64("guild_id", int64(guildID)).Int64("channel_id", int64(channelID)).Msg("Channel not found for reaction roles")

				ctx.JSON(http.StatusBadRequest, BaseResponse{
					Ok:    false,
					Error: "The channel specified was not found in this server.",
				})

				return
			}

			message, err := discord.GetChannelMessage(ctx, backend.BotSession, channelID, messageID)
			if err != nil {
				welcomer.Logger.Warn().Err(err).Int64("guild_id", int64(guildID)).Int64("channel_id", int64(channelID)).Int64("message_id", int64(messageID)).Msg("Failed to get message for reaction roles")

				ctx.JSON(http.StatusBadRequest, BaseResponse{
					Ok:    false,
					Error: "Failed to get message. Make sure the message exists and the bot has access to it.",
				})

				return
			}

			emoji := "🎉"

			err = message.AddReaction(ctx, backend.BotSession, emoji)
			if err != nil {
				welcomer.Logger.Warn().Err(err).Int64("guild_id", int64(guildID)).Int64("channel_id", int64(channelID)).Int64("message_id", int64(messageID)).Msg("Failed to add reaction to message for reaction roles")

				ctx.JSON(http.StatusBadRequest, BaseResponse{
					Ok:    false,
					Error: "Failed to add reaction to message. Make sure the bot has permission to add reactions in the specified channel.",
				})

				return
			}

			_ = discord.DeleteOwnReaction(ctx, backend.BotSession, message.ChannelID, message.ID, emoji)

			ctx.JSON(http.StatusOK, BaseResponse{
				Ok: true,
			})
		})
	})
}

// Validate reaction role settings.
func doValidateReactionRoles(ctx context.Context, guildID discord.Snowflake, partial *GuildSettingsReactionRoles) *welcomer.ErrorGroup {
	if !partial.ToggleEnabled {
		return nil
	}

	eg := welcomer.NewErrorGroup()

	for i, rr := range partial.ReactionRoles {
		if !rr.Enabled {
			continue
		}

		if rr.ChannelID != 0 {
			validGuild, err := welcomer.CheckChannelGuild(ctx, welcomer.SandwichClient, guildID, rr.ChannelID)
			if err != nil {
				welcomer.Logger.Warn().Err(err).Int64("guild_id", int64(guildID)).Int64("channel_id", int64(rr.ChannelID)).Msg("Failed to check channel guild for reaction role settings validation")

				eg.Add(fmt.Errorf("reaction role %d: failed to validate channel ID: %v", i+1, err))

				continue
			}

			if !validGuild {
				welcomer.Logger.Warn().Int64("guild_id", int64(guildID)).Int64("channel_id", int64(rr.ChannelID)).Msg("Channel ID does not belong to this guild for reaction role settings validation")

				eg.Add(fmt.Errorf("reaction role %d: channel ID does not belong to this guild", i+1))

				continue
			}
		}

		if rr.IsSystemMessage {
			if rr.ChannelID == 0 {
				eg.Add(fmt.Errorf("reaction role %d: channel ID must be specified for system message reaction roles", i+1))
			}

			if rr.Message == "" {
				eg.Add(fmt.Errorf("reaction role %d: message must be specified for system message reaction roles", i+1))
			} else {
				var messageParams discord.MessageParams

				if err := json.Unmarshal([]byte(rr.Message), &messageParams); err != nil {
					eg.Add(fmt.Errorf("reaction role %d: message must be a valid JSON object: %v", i+1, err))
				} else if welcomer.IsMessageParamsEmpty(messageParams) {
					eg.Add(fmt.Errorf("reaction role %d: message cannot be empty for system message reaction roles", i+1))
				}
			}
		} else {
			if rr.ChannelID == 0 {
				eg.Add(fmt.Errorf("reaction role %d: channel ID must be specified for non-system message reaction roles", i+1))
			}

			if rr.MessageID == 0 {
				eg.Add(fmt.Errorf("reaction role %d: message ID must be specified for non-system message reaction roles", i+1))
			}
		}

		if len(rr.Roles) == 0 {
			eg.Add(fmt.Errorf("reaction role %d: at least one role option must be specified for enabled reaction roles", i+1))
		}

		switch rr.Type {
		case welcomer.ReactionRoleTypeButtons:
			if len(rr.Roles) > MaximumOptionsButtons {
				eg.Add(fmt.Errorf("reaction role %d: cannot have more than %d options for buttons type", i+1, MaximumOptionsButtons))
			}

			for j, option := range rr.Roles {
				if option.Name == "" {
					eg.Add(fmt.Errorf("reaction role %d: option %d: name is required", i+1, j+1))
				}

				if len(option.Name) > 50 {
					eg.Add(fmt.Errorf("reaction role %d: option %d: name cannot be longer than 50 characters", i+1, j+1))
				}
			}
		case welcomer.ReactionRoleTypeDropdown:
			if len(rr.Roles) > MaximumOptionsDropdown {
				eg.Add(fmt.Errorf("reaction role %d: cannot have more than %d options for dropdown type", i+1, MaximumOptionsDropdown))
			}

			for j, option := range rr.Roles {
				if option.Name == "" {
					eg.Add(fmt.Errorf("reaction role %d: option %d: name is required for dropdown type", i+1, j+1))
				}

				if len(option.Name) > 50 {
					eg.Add(fmt.Errorf("reaction role %d: option %d: name cannot be longer than 50 characters", i+1, j+1))
				}

				if option.Description != "" && len(option.Description) > 100 {
					eg.Add(fmt.Errorf("reaction role %d: option %d: description cannot be longer than 100 characters", i+1, j+1))
				}
			}
		case welcomer.ReactionRoleTypeEmoji:
			if len(rr.Roles) > MaximumOptionsEmojis {
				eg.Add(fmt.Errorf("reaction role %d: cannot have more than %d options for emoji type", i+1, MaximumOptionsEmojis))
			}

			for j, option := range rr.Roles {
				if option.Emoji == "" {
					eg.Add(fmt.Errorf("reaction role %d: option %d: emoji is required for emoji type", i+1, j+1))
				}

				// Check for duplicate emojis
				for k := j + 1; k < len(rr.Roles); k++ {
					if option.Emoji == rr.Roles[k].Emoji {
						eg.Add(fmt.Errorf("reaction role %d: option %d: emoji has already been used", i+1, k))
						break
					}
				}
			}
		default:
			eg.Add(fmt.Errorf("reaction role %d: invalid reaction role type", i+1))
		}

		// Check for duplicate roles
		for i, option := range rr.Roles {
			for j := i + 1; j < len(rr.Roles); j++ {
				if option.RoleID == rr.Roles[j].RoleID {
					eg.Add(fmt.Errorf("reaction role %d: options %d and %d: duplicate option name '%s'", i+1, i+1, j+1, option.Name))
				}
			}
		}
	}

	if eg.Empty() {
		return nil
	}

	return eg
}

func registerGuildSettingsReactionRolesRoutes(g *gin.Engine) {
	g.GET("/api/guild/:guildID/reactionroles", getGuildSettingsReactionRoles)
	g.POST("/api/guild/:guildID/reactionroles", setGuildSettingsReactionRoles)

	g.POST("/api/guild/:guildID/checkmessage/:channelID/:messageID", checkGuildSettingsReactionRolesMessage)
}
