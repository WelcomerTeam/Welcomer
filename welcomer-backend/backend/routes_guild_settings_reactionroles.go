package backend

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"slices"

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

			reactionroles, err := welcomer.Queries.GetReactionRoleSettingByGuildId(ctx, int64(guildID))
			if err != nil {
				if errors.Is(err, pgx.ErrNoRows) {
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
				if reactionRole.ReactionRoleID.IsNil() {
					partial.ReactionRoles[i].ReactionRoleID = uuid.Must(gen.NewV7())
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

			oldReactionRoleSettings, err := welcomer.Queries.GetReactionRoleSettingByGuildId(ctx, int64(guildID))
			if err != nil && !errors.Is(err, pgx.ErrNoRows) {
				welcomer.Logger.Warn().Err(err).Int64("guild_id", int64(guildID)).Msg("Failed to get existing guild reaction roles settings for update")

				ctx.JSON(http.StatusInternalServerError, NewBaseResponse(NewGenericErrorWithLineNumber(), nil))

				return
			}

			user := tryGetUser(ctx)

			reactionroles := PartialToGuildSettingsReactionRolesSettings(int64(guildID), partial)

			err = welcomer.RetryWithFallback(
				func() error {
					eg := welcomer.CreateOrUpdateReactionRolesGuildSettingsWithAudit(ctx, guildID, reactionroles, user.ID)

					if eg != nil {
						return eg.AsStandardError()
					}

					return nil
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

			eg := processReactionRolesSettingsChange(ctx, GuildSettingsReactionRolesSettingsToPartial(oldReactionRoleSettings), partial)
			if eg != nil && !eg.Empty() {
				welcomer.Logger.Warn().Err(eg).Int64("guild_id", int64(guildID)).Int64("user_id", int64(user.ID)).Msg("Failed to process reaction roles settings change")

				ctx.JSON(http.StatusInternalServerError, BaseResponse{
					Ok:    false,
					Error: eg.ErrorWithDelimiter("\n"),
				})

				return
			}

			welcomer.Logger.Info().Int64("guild_id", int64(guildID)).Interface("obj", reactionroles).Int64("user_id", int64(user.ID)).Msg("Creating or updating reaction roles guild settings")

			getGuildSettingsReactionRoles(ctx)
		})
	})
}

type reactionRoleConfigurations struct {
	Old      *welcomer.GuildSettingsReactionRole
	New      *welcomer.GuildSettingsReactionRole
	NewIndex int
}

func processReactionRolesSettingsChange(ctx *gin.Context, old, new *GuildSettingsReactionRoles) *welcomer.ErrorGroup {
	eg := welcomer.NewErrorGroup()

	configurationChanges := make(map[uuid.UUID]reactionRoleConfigurations)

	for _, oldConfig := range old.ReactionRoles {
		_, ok := configurationChanges[oldConfig.ReactionRoleID]
		if !ok {
			configurationChanges[oldConfig.ReactionRoleID] = reactionRoleConfigurations{}
		}

		configuration := configurationChanges[oldConfig.ReactionRoleID]
		configuration.Old = &oldConfig
		configurationChanges[oldConfig.ReactionRoleID] = configuration
	}

	for i, newConfig := range new.ReactionRoles {
		_, ok := configurationChanges[newConfig.ReactionRoleID]
		if !ok {
			configurationChanges[newConfig.ReactionRoleID] = reactionRoleConfigurations{}
		}

		configuration := configurationChanges[newConfig.ReactionRoleID]
		configuration.New = &newConfig
		configuration.NewIndex = i
		configurationChanges[newConfig.ReactionRoleID] = configuration
	}

	for _, config := range configurationChanges {
		if (config.Old != nil && config.Old.IsSystemMessage) || (config.New != nil && config.New.IsSystemMessage) {
			processReactionRolesSettingsChangeSystemMessage(ctx, eg, config.Old, config.New)
		} else {
			processReactionRolesSettingsChangeNonSystemMessage(ctx, eg, config.Old, config.New)
		}
	}

	return eg
}

func processReactionRolesSettingsChangeSystemMessage(ctx *gin.Context, eg *welcomer.ErrorGroup, old, new *welcomer.GuildSettingsReactionRole) {
	// if old != nil && new == nil {
	// 	err := disableReactionRoleMessage(ctx, old.ChannelID, old.MessageID)
	// 	if err != nil {
	// 		eg.Add(fmt.Errorf("failed to disable message for removed system message reaction role configuration: %v", err))
	// 	}
	// }

	if !hasConfigurationChanged(old, new) {
		return
	}

	var message *discord.Message
	var err error

	if hasConfigurationChangedMessage(old, new) && ((old == nil || new == nil) || old.ChannelID != new.ChannelID) {
		// If channel has changed, send new message.

		if new != nil && new.Enabled {
			message, err = createReactionRoleMessage(ctx, eg, new)
			if err != nil {
				welcomer.Logger.Warn().Err(err).Int64("guild_id", int64(tryGetGuildID(ctx))).Int64("channel_id", int64(new.ChannelID)).Msg("Failed to create message for updated system message reaction role configuration")
			}

			if message != nil {
				new.MessageID = message.ID
			} else {
				new.MessageID = 0
			}
		}

		if old != nil && !old.ChannelID.IsNil() && !old.MessageID.IsNil() {
			// Disable old message.
			err = disableReactionRoleMessage(ctx, old.ChannelID, old.MessageID)
			if err != nil {
				eg.Add(fmt.Errorf("failed to disable message for removed system message reaction role configuration: %v", err))
			}
		}
	} else if hasConfigurationChangedRoles(old, new) || old.Message != new.Message || old.Enabled != new.Enabled {
		// If roles, message or enabled has changed, update existing message.

		if old != nil && !new.Enabled && !old.ChannelID.IsNil() && !old.MessageID.IsNil() {
			err = disableReactionRoleMessage(ctx, old.ChannelID, old.MessageID)
			if err != nil {
				eg.Add(fmt.Errorf("failed to disable message for disabled system message reaction role configuration: %v", err))
			}

			return
		}

		if message == nil {
			message, err = discord.GetChannelMessage(ctx, backend.BotSession, new.ChannelID, new.MessageID)
			if err != nil {
				welcomer.Logger.Warn().Err(err).Int64("guild_id", int64(tryGetGuildID(ctx))).Int64("channel_id", int64(new.ChannelID)).Int64("message_id", int64(new.MessageID)).Msg("Failed to get message for updated system message reaction role configuration")

				message, err = createReactionRoleMessage(ctx, eg, new)
				if err != nil {
					welcomer.Logger.Warn().Err(err).Int64("guild_id", int64(tryGetGuildID(ctx))).Int64("channel_id", int64(new.ChannelID)).Msg("Failed to create message for updated system message reaction role configuration")
				}

				return
			}
		}

		// If message cannot be retrieved or created, there is not much we can do to update the configuration, so just return and log the error.
		if message == nil {
			return
		}

		// Message has changed or button/dropdown configuration has changed.
		if old.Message != new.Message || hasConfigurationChangedRoles(old, new) || (old.Type != new.Type) || (!old.Enabled && new.Enabled) {
			if messageParams := setupMessageParamsForReactionRoleConfiguration(new); messageParams != nil {
				_, err = message.Edit(ctx, backend.BotSession, *messageParams)
				if err != nil {
					welcomer.Logger.Warn().Err(err).Int64("guild_id", int64(tryGetGuildID(ctx))).Int64("channel_id", int64(new.ChannelID)).Int64("message_id", int64(message.ID)).Msg("Failed to edit message for updated system message reaction role configuration")

					eg.Add(fmt.Errorf("failed to edit message for updated system message reaction role configuration: %v", err))

					return
				}
			}
		}

		if hasConfigurationChangedRoles(old, new) && new.Type == welcomer.ReactionRoleTypeEmoji {
			removeUnusedMessageReactions(ctx, eg, message, new)
			addMessageReactions(ctx, eg, message, new)
		}
	}
}

func processReactionRolesSettingsChangeNonSystemMessage(ctx *gin.Context, eg *welcomer.ErrorGroup, old, new *welcomer.GuildSettingsReactionRole) {
	if !new.Enabled || new.ChannelID.IsNil() || new.MessageID.IsNil() {
		return
	}

	message, err := discord.GetChannelMessage(ctx, backend.BotSession, new.ChannelID, new.MessageID)
	if err != nil {
		welcomer.Logger.Warn().Err(err).Int64("guild_id", int64(tryGetGuildID(ctx))).Int64("channel_id", int64(new.ChannelID)).Int64("message_id", int64(new.MessageID)).Msg("Failed to get message for updated system message reaction role configuration")

		eg.Add(fmt.Errorf("failed to get message for reaction role configuration: %v", err))

		return
	}

	switch new.Type {
	case welcomer.ReactionRoleTypeEmoji:
		removeUnusedMessageReactions(ctx, eg, message, new)
		addMessageReactions(ctx, eg, message, new)
	default:
		eg.Add(fmt.Errorf("unsupported reaction role type for non-system message reaction role configuration: %s", new.Type))
	}
}

func removeUnusedMessageReactions(ctx *gin.Context, eg *welcomer.ErrorGroup, message *discord.Message, new *welcomer.GuildSettingsReactionRole) {
	for _, reaction := range message.Reactions {
		// Remove reactions that are no longer valid for the updated configuration.
		if found := slices.ContainsFunc(new.Roles, func(option welcomer.ReactionRoleOption) bool {
			return option.Emoji == reaction.Emoji.Name || option.Emoji == reaction.Emoji.ID.String()
		}); !found {
			var emoji string

			if reaction.Emoji.ID != 0 {
				emoji = "_:" + reaction.Emoji.ID.String()
			} else {
				emoji = reaction.Emoji.Name
			}

			err := message.ClearReaction(ctx, backend.BotSession, emoji)
			if err != nil {
				welcomer.Logger.Warn().Err(err).Int64("guild_id", int64(tryGetGuildID(ctx))).Int64("channel_id", int64(new.ChannelID)).Int64("message_id", int64(message.ID)).Str("emoji", emoji).Msg("Failed to remove reaction from message for updated system message reaction role configuration")

				eg.Add(fmt.Errorf("failed to remove reaction '%s:%s' from message for updated system message reaction role configuration: %v", reaction.Emoji.Name, reaction.Emoji.ID, err))
			}
		}
	}
}

func addMessageReactions(ctx *gin.Context, eg *welcomer.ErrorGroup, message *discord.Message, new *welcomer.GuildSettingsReactionRole) {
	for _, option := range new.Roles {
		if option.Emoji != "" {
			var emoji string

			if _, err := welcomer.Atoi(option.Emoji); err == nil {
				emoji = "_:" + option.Emoji
			} else {
				emoji = option.Emoji
			}

			err := message.AddReaction(ctx, backend.BotSession, emoji)
			if err != nil {
				welcomer.Logger.Warn().Err(err).Int64("guild_id", int64(tryGetGuildID(ctx))).Int64("channel_id", int64(new.ChannelID)).Int64("message_id", int64(message.ID)).Str("emoji", option.Emoji).Msg("Failed to add reaction to message for updated system message reaction role configuration")

				eg.Add(fmt.Errorf("failed to add reaction '%s' to message for updated system message reaction role configuration: %v", option.Emoji, err))
			}
		}
	}
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

		_, err = welcomer.Queries.UpdateReactionRoleSettingMessageId(ctx, database.UpdateReactionRoleSettingMessageIdParams{
			ReactionRoleID: new.ReactionRoleID,
			GuildID:        int64(tryGetGuildID(ctx)),
			MessageID:      int64(message.ID),
		})
		if err != nil {
			welcomer.Logger.Warn().Err(err).
				Int64("guild_id", int64(tryGetGuildID(ctx))).
				Int64("channel_id", int64(new.ChannelID)).
				Int64("message_id", int64(message.ID)).
				Msg("Failed to update message ID for updated system message reaction role configuration")
		}

		if new.Type == welcomer.ReactionRoleTypeEmoji {
			addMessageReactions(ctx, eg, message, new)
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
				CustomID: fmt.Sprintf("reaction_role:%s:%s", config.ReactionRoleID.String(), option.RoleID),
				Emoji:    emoji,
				Label:    option.Name,
				Style:    discord.InteractionComponentStyleSecondary,
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
				Value:       fmt.Sprintf("reaction_role:%s:%s", config.ReactionRoleID.String(), option.RoleID),
			}

			options = append(options, selectOption)
		}

		return []discord.InteractionComponent{
			{
				Type: discord.InteractionComponentTypeActionRow,
				Components: []discord.InteractionComponent{
					{
						Type:     discord.InteractionComponentTypeStringSelect,
						CustomID: fmt.Sprintf("reaction_role:%s", config.ReactionRoleID.String()),
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

// Validates if string is a valid emoji. This can be a snowflake or a unicode emoji.
func isValidEmoji(emoji string) bool {
	if _, err := welcomer.Atoi(emoji); err == nil {
		return true
	}

	for _, rune := range emoji {
		switch {

		// emoji blocks
		case rune >= 0x1F600 && rune <= 0x1F64F:
		case rune >= 0x1F300 && rune <= 0x1F5FF:
		case rune >= 0x1F680 && rune <= 0x1F6FF:
		case rune >= 0x1F900 && rune <= 0x1F9FF:
		case rune >= 0x1FA70 && rune <= 0x1FAFF:
		case rune >= 0x1F1E6 && rune <= 0x1F1FF:
		case rune >= 0x2600 && rune <= 0x26FF:
		case rune >= 0x2700 && rune <= 0x27BF:

		// emoji modifiers
		case rune >= 0x1F3FB && rune <= 0x1F3FF: // skin tones

		// joiners / selectors
		case rune == 0x200D: // ZWJ
		case rune == 0xFE0F: // variation selector

		default:
			return false
		}
	}

	return len(emoji) > 0
}

// Validate reaction role settings.
func doValidateReactionRoles(ctx context.Context, guildID discord.Snowflake, partial *GuildSettingsReactionRoles) *welcomer.ErrorGroup {
	errorGroup := welcomer.NewErrorGroup()

	// Pre-validate emojis, even if reaction role is disabled.
	for i, rr := range partial.ReactionRoles {
		for j, option := range rr.Roles {
			if option.Emoji != "" && !isValidEmoji(option.Emoji) {
				errorGroup.Add(fmt.Errorf("reaction role %d: option %d: emoji is not a valid unicode emoji or custom emoji ID", i+1, j+1))
			}
		}
	}

	for reactionRoleIndex, reactionRole := range partial.ReactionRoles {
		if reactionRole.ChannelID != 0 {
			validGuild, err := welcomer.CheckChannelGuild(ctx, welcomer.SandwichClient, guildID, reactionRole.ChannelID)
			if err != nil {
				welcomer.Logger.Warn().Err(err).Int64("guild_id", int64(guildID)).Int64("channel_id", int64(reactionRole.ChannelID)).Msg("Failed to check channel guild for reaction role settings validation")

				errorGroup.Add(fmt.Errorf("reaction role %d: failed to validate channel ID: %v", reactionRoleIndex+1, err))

				continue
			}

			if !validGuild {
				welcomer.Logger.Warn().Int64("guild_id", int64(guildID)).Int64("channel_id", int64(reactionRole.ChannelID)).Msg("Channel ID does not belong to this guild for reaction role settings validation")

				errorGroup.Add(fmt.Errorf("reaction role %d: channel ID does not belong to this guild", reactionRoleIndex+1))

				continue
			}
		}

		if !reactionRole.Enabled {
			continue
		}

		if reactionRole.IsSystemMessage {
			if reactionRole.ChannelID == 0 {
				errorGroup.Add(fmt.Errorf("reaction role %d: channel ID must be specified for system message reaction roles", reactionRoleIndex+1))
			}

			if reactionRole.Message == "" {
				errorGroup.Add(fmt.Errorf("reaction role %d: message must be specified for system message reaction roles", reactionRoleIndex+1))
			} else {
				var messageParams discord.MessageParams

				if err := json.Unmarshal([]byte(reactionRole.Message), &messageParams); err != nil {
					errorGroup.Add(fmt.Errorf("reaction role %d: message must be a valid JSON object: %v", reactionRoleIndex+1, err))
				} else if welcomer.IsMessageParamsEmpty(messageParams) {
					errorGroup.Add(fmt.Errorf("reaction role %d: message cannot be empty for system message reaction roles", reactionRoleIndex+1))
				}
			}
		} else {
			if reactionRole.ChannelID == 0 {
				errorGroup.Add(fmt.Errorf("reaction role %d: channel ID must be specified for non-system message reaction roles", reactionRoleIndex+1))
			}

			if reactionRole.MessageID == 0 {
				errorGroup.Add(fmt.Errorf("reaction role %d: message ID must be specified for non-system message reaction roles", reactionRoleIndex+1))
			}
		}

		if len(reactionRole.Roles) == 0 {
			errorGroup.Add(fmt.Errorf("reaction role %d: at least one role option must be specified for enabled reaction roles", reactionRoleIndex+1))
		}

		switch reactionRole.Type {
		case welcomer.ReactionRoleTypeButtons:
			if len(reactionRole.Roles) > MaximumOptionsButtons {
				errorGroup.Add(fmt.Errorf("reaction role %d: cannot have more than %d options for buttons type", reactionRoleIndex+1, MaximumOptionsButtons))
			}

			for j, option := range reactionRole.Roles {
				if option.Name == "" {
					errorGroup.Add(fmt.Errorf("reaction role %d: option %d: name is required", reactionRoleIndex+1, j+1))
				}

				if len(option.Name) > 50 {
					errorGroup.Add(fmt.Errorf("reaction role %d: option %d: name cannot be longer than 50 characters", reactionRoleIndex+1, j+1))
				}
			}
		case welcomer.ReactionRoleTypeDropdown:
			if len(reactionRole.Roles) > MaximumOptionsDropdown {
				errorGroup.Add(fmt.Errorf("reaction role %d: cannot have more than %d options for dropdown type", reactionRoleIndex+1, MaximumOptionsDropdown))
			}

			for j, option := range reactionRole.Roles {
				if option.Name == "" {
					errorGroup.Add(fmt.Errorf("reaction role %d: option %d: name is required for dropdown type", reactionRoleIndex+1, j+1))
				}

				if len(option.Name) > 50 {
					errorGroup.Add(fmt.Errorf("reaction role %d: option %d: name cannot be longer than 50 characters", reactionRoleIndex+1, j+1))
				}

				if option.Description != "" && len(option.Description) > 100 {
					errorGroup.Add(fmt.Errorf("reaction role %d: option %d: description cannot be longer than 100 characters", reactionRoleIndex+1, j+1))
				}
			}
		case welcomer.ReactionRoleTypeEmoji:
			if len(reactionRole.Roles) > MaximumOptionsEmojis {
				errorGroup.Add(fmt.Errorf("reaction role %d: cannot have more than %d options for emoji type", reactionRoleIndex+1, MaximumOptionsEmojis))
			}

			for j, option := range reactionRole.Roles {
				if option.Emoji == "" {
					errorGroup.Add(fmt.Errorf("reaction role %d: option %d: emoji is required for emoji type", reactionRoleIndex+1, j+1))
				}

				// Check for duplicate emojis
				for k := j + 1; k < len(reactionRole.Roles); k++ {
					if option.Emoji == reactionRole.Roles[k].Emoji {
						errorGroup.Add(fmt.Errorf("reaction role %d: option %d: emoji has already been used", reactionRoleIndex+1, k))
						break
					}
				}
			}
		default:
			errorGroup.Add(fmt.Errorf("reaction role %d: invalid reaction role type", reactionRoleIndex+1))
		}

		// Check for duplicate roles
		for i, option := range reactionRole.Roles {
			for j := i + 1; j < len(reactionRole.Roles); j++ {
				if option.RoleID == reactionRole.Roles[j].RoleID {
					errorGroup.Add(fmt.Errorf("reaction role %d: options %d and %d: duplicate option name '%s'", i+1, i+1, j+1, option.Name))
				}
			}
		}
	}

	if errorGroup.Empty() {
		return nil
	}

	return errorGroup
}

func registerGuildSettingsReactionRolesRoutes(g *gin.Engine) {
	g.GET("/api/guild/:guildID/reactionroles", getGuildSettingsReactionRoles)
	g.POST("/api/guild/:guildID/reactionroles", setGuildSettingsReactionRoles)

	g.POST("/api/guild/:guildID/checkmessage/:channelID/:messageID", checkGuildSettingsReactionRolesMessage)
}
