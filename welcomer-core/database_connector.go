package welcomer

import (
	"context"
	"errors"

	discord "github.com/WelcomerTeam/Discord/discord"
	"github.com/WelcomerTeam/Welcomer/welcomer-core/database"
	"github.com/jackc/pgx/v4"
)

// CreateOrUpdateUserWithAudit wraps Queries.CreateOrUpdateUser and logs an audit
// entry when changes are detected. It will attempt to fetch the existing user
// row (if any), call AuditChange comparing old and new, then perform the DB
// operation and return the resulting row.
func CreateOrUpdateUserWithAudit(ctx context.Context, params database.CreateOrUpdateUserParams, actor discord.Snowflake) (*database.Users, error) {
	var old database.Users

	// Attempt to fetch existing user by primary key.
	existing, err := Queries.GetUser(ctx, params.UserID)
	if err == nil {
		old = *existing
	}

	// Build the new value as the DB querier would return (partial). We call
	// the underlying CreateOrUpdate and then compare the returned row with
	// old to produce the audit log.
	newRow, err := Queries.CreateOrUpdateUser(ctx, params)
	if err != nil {
		return nil, err
	}

	AuditChange(ctx, discord.Snowflake(params.UserID), actor, old, *newRow, database.AuditTypeUsers)

	return newRow, nil
}

func CreateOrUpdateWelcomerGuildSettingsWithAudit(ctx context.Context, params database.CreateOrUpdateWelcomerGuildSettingsParams, actor discord.Snowflake) (*database.GuildSettingsWelcomer, error) {
	var old database.GuildSettingsWelcomer
	if existing, err := Queries.GetWelcomerGuildSettings(ctx, params.GuildID); err == nil {
		old = *existing
	}

	newRow, err := Queries.CreateOrUpdateWelcomerGuildSettings(ctx, params)
	if err != nil {
		return nil, err
	}

	AuditChange(ctx, discord.Snowflake(params.GuildID), actor, old, *newRow, database.AuditTypeGuildSettingsWelcomer)

	return newRow, nil
}

// Generic wrappers for guild settings create/update operations.
func CreateOrUpdateAutoRolesGuildSettingsWithAudit(ctx context.Context, params database.CreateOrUpdateAutoRolesGuildSettingsParams, actor discord.Snowflake) (*database.GuildSettingsAutoroles, error) {
	var old database.GuildSettingsAutoroles
	if existing, err := Queries.GetAutoRolesGuildSettings(ctx, params.GuildID); err == nil {
		old = *existing
	}

	newRow, err := Queries.CreateOrUpdateAutoRolesGuildSettings(ctx, params)
	if err != nil {
		return nil, err
	}

	AuditChange(ctx, discord.Snowflake(params.GuildID), actor, old, *newRow, database.AuditTypeGuildSettingsAutoroles)

	return newRow, nil
}

func CreateOrUpdateBorderwallGuildSettingsWithAudit(ctx context.Context, params database.CreateOrUpdateBorderwallGuildSettingsParams, actor discord.Snowflake) (*database.GuildSettingsBorderwall, error) {
	var old database.GuildSettingsBorderwall
	if existing, err := Queries.GetBorderwallGuildSettings(ctx, params.GuildID); err == nil {
		old = *existing
	}

	newRow, err := Queries.CreateOrUpdateBorderwallGuildSettings(ctx, params)
	if err != nil {
		return nil, err
	}

	AuditChange(ctx, discord.Snowflake(params.GuildID), actor, old, *newRow, database.AuditTypeGuildSettingsBorderwall)

	return newRow, nil
}

func CreateOrUpdateFreeRolesGuildSettingsWithAudit(ctx context.Context, params database.CreateOrUpdateFreeRolesGuildSettingsParams, actor discord.Snowflake) (*database.GuildSettingsFreeroles, error) {
	var old database.GuildSettingsFreeroles
	if existing, err := Queries.GetFreeRolesGuildSettings(ctx, params.GuildID); err == nil {
		old = *existing
	}

	newRow, err := Queries.CreateOrUpdateFreeRolesGuildSettings(ctx, params)
	if err != nil {
		return nil, err
	}

	AuditChange(ctx, discord.Snowflake(params.GuildID), actor, old, *newRow, database.AuditTypeGuildSettingsFreeroles)

	return newRow, nil
}

func CreateOrUpdateLeaverGuildSettingsWithAudit(ctx context.Context, params database.CreateOrUpdateLeaverGuildSettingsParams, actor discord.Snowflake) (*database.GuildSettingsLeaver, error) {
	var old database.GuildSettingsLeaver
	if existing, err := Queries.GetLeaverGuildSettings(ctx, params.GuildID); err == nil {
		old = *existing
	}

	newRow, err := Queries.CreateOrUpdateLeaverGuildSettings(ctx, params)
	if err != nil {
		return nil, err
	}

	AuditChange(ctx, discord.Snowflake(params.GuildID), actor, old, *newRow, database.AuditTypeGuildSettingsLeaver)

	return newRow, nil
}

func CreateOrUpdateRulesGuildSettingsWithAudit(ctx context.Context, params database.CreateOrUpdateRulesGuildSettingsParams, actor discord.Snowflake) (*database.GuildSettingsRules, error) {
	var old database.GuildSettingsRules
	if existing, err := Queries.GetRulesGuildSettings(ctx, params.GuildID); err == nil {
		old = *existing
	}

	newRow, err := Queries.CreateOrUpdateRulesGuildSettings(ctx, params)
	if err != nil {
		return nil, err
	}

	AuditChange(ctx, discord.Snowflake(params.GuildID), actor, old, *newRow, database.AuditTypeGuildSettingsRules)

	return newRow, nil
}

func CreateOrUpdateTempChannelsGuildSettingsWithAudit(ctx context.Context, params database.CreateOrUpdateTempChannelsGuildSettingsParams, actor discord.Snowflake) (*database.GuildSettingsTempchannels, error) {
	var old database.GuildSettingsTempchannels
	if existing, err := Queries.GetTempChannelsGuildSettings(ctx, params.GuildID); err == nil {
		old = *existing
	}

	newRow, err := Queries.CreateOrUpdateTempChannelsGuildSettings(ctx, params)
	if err != nil {
		return nil, err
	}

	AuditChange(ctx, discord.Snowflake(params.GuildID), actor, old, *newRow, database.AuditTypeGuildSettingsTempchannels)

	return newRow, nil
}

func CreateOrUpdateTimeRolesGuildSettingsWithAudit(ctx context.Context, params database.CreateOrUpdateTimeRolesGuildSettingsParams, actor discord.Snowflake) (*database.GuildSettingsTimeroles, error) {
	var old database.GuildSettingsTimeroles
	if existing, err := Queries.GetTimeRolesGuildSettings(ctx, params.GuildID); err == nil {
		old = *existing
	}

	newRow, err := Queries.CreateOrUpdateTimeRolesGuildSettings(ctx, params)
	if err != nil {
		return nil, err
	}

	AuditChange(ctx, discord.Snowflake(params.GuildID), actor, old, *newRow, database.AuditTypeGuildSettingsTimeroles)

	return newRow, nil
}

func CreateOrUpdateWelcomerTextGuildSettingsWithAudit(ctx context.Context, params database.CreateOrUpdateWelcomerTextGuildSettingsParams, actor discord.Snowflake) (*database.GuildSettingsWelcomerText, error) {
	var old database.GuildSettingsWelcomerText
	if existing, err := Queries.GetWelcomerTextGuildSettings(ctx, params.GuildID); err == nil {
		old = *existing
	}

	newRow, err := Queries.CreateOrUpdateWelcomerTextGuildSettings(ctx, params)
	if err != nil {
		return nil, err
	}

	AuditChange(ctx, discord.Snowflake(params.GuildID), actor, old, *newRow, database.AuditTypeGuildSettingsWelcomerText)

	return newRow, nil
}

func CreateOrUpdateWelcomerImagesGuildSettingsWithAudit(ctx context.Context, params database.CreateOrUpdateWelcomerImagesGuildSettingsParams, actor discord.Snowflake) (*database.GuildSettingsWelcomerImages, error) {
	var old database.GuildSettingsWelcomerImages
	if existing, err := Queries.GetWelcomerImagesGuildSettings(ctx, params.GuildID); err == nil {
		old = *existing
	}

	newRow, err := Queries.CreateOrUpdateWelcomerImagesGuildSettings(ctx, params)
	if err != nil {
		return nil, err
	}

	AuditChange(ctx, discord.Snowflake(params.GuildID), actor, old, *newRow, database.AuditTypeGuildSettingsWelcomerImages)

	return newRow, nil
}

func CreateOrUpdateWelcomerDMsGuildSettingsWithAudit(ctx context.Context, params database.CreateOrUpdateWelcomerDMsGuildSettingsParams, actor discord.Snowflake) (*database.GuildSettingsWelcomerDms, error) {
	var old database.GuildSettingsWelcomerDms

	if existing, err := Queries.GetWelcomerDMsGuildSettings(ctx, params.GuildID); err == nil {
		old = *existing
	}

	newRow, err := Queries.CreateOrUpdateWelcomerDMsGuildSettings(ctx, params)
	if err != nil {
		return nil, err
	}

	AuditChange(ctx, discord.Snowflake(params.GuildID), actor, old, *newRow, database.AuditTypeGuildSettingsWelcomerDms)

	return newRow, nil
}

// Simple create wrappers for non-guild-specific objects.
func CreateWelcomerImagesWithAudit(ctx context.Context, params database.CreateWelcomerImagesParams, actor discord.Snowflake) (*database.WelcomerImages, error) {
	var old database.WelcomerImages

	newRow, err := Queries.CreateWelcomerImages(ctx, params)
	if err != nil {
		return nil, err
	}

	AuditChange(ctx, 0, actor, old, *newRow, database.AuditTypeWelcomerImages)

	return newRow, nil
}

func CreateBorderwallRequestWithAudit(ctx context.Context, params database.CreateBorderwallRequestParams, actor discord.Snowflake) (*database.BorderwallRequests, error) {
	var old database.BorderwallRequests

	newRow, err := Queries.CreateBorderwallRequest(ctx, params)
	if err != nil {
		return nil, err
	}

	AuditChange(ctx, discord.Snowflake(params.GuildID), actor, old, *newRow, database.AuditTypeBorderwallRequests)

	return newRow, nil
}

func CreateUserWithAudit(ctx context.Context, params database.CreateUserParams, actor discord.Snowflake) (*database.Users, error) {
	var old database.Users

	newRow, err := Queries.CreateUser(ctx, params)
	if err != nil {
		return nil, err
	}

	AuditChange(ctx, discord.Snowflake(params.UserID), actor, old, *newRow, database.AuditTypeUsers)

	return newRow, nil
}

func CreateCustomBotWithAudit(ctx context.Context, params database.CreateCustomBotParams, actor discord.Snowflake) (*database.CustomBots, error) {
	var old database.CustomBots

	newRow, err := Queries.CreateCustomBot(ctx, params)
	if err != nil {
		return nil, err
	}

	AuditChange(ctx, discord.Snowflake(params.GuildID), actor, old, *newRow, database.AuditTypeCustomBots)

	return newRow, nil
}

// CreateGuildWithAudit wraps CreateGuild and audits the creation of a guild row.
func CreateGuildWithAudit(ctx context.Context, params database.CreateGuildParams, actor discord.Snowflake) (*database.Guilds, error) {
	var old database.Guilds

	newRow, err := Queries.CreateGuild(ctx, params)
	if err != nil {
		return nil, err
	}

	AuditChange(ctx, discord.Snowflake(params.GuildID), actor, old, *newRow, database.AuditTypeGuilds)

	return newRow, nil
}

func UpdateCustomBotWithAudit(ctx context.Context, params database.UpdateCustomBotParams, actor, guildID discord.Snowflake) (*database.CustomBots, error) {
	var old database.CustomBots
	if existing, err := Queries.GetCustomBotById(ctx, database.GetCustomBotByIdParams{
		CustomBotUuid: params.CustomBotUuid,
		GuildID:       int64(guildID),
	}); err == nil {
		old = database.CustomBots{
			CustomBotUuid:     existing.CustomBotUuid,
			GuildID:           existing.GuildID,
			PublicKey:         existing.PublicKey,
			Token:             "",
			CreatedAt:         existing.CreatedAt,
			IsActive:          existing.IsActive,
			ApplicationID:     existing.ApplicationID,
			ApplicationName:   existing.ApplicationName,
			ApplicationAvatar: existing.ApplicationAvatar,
			Environment:       existing.Environment,
		}
	}

	newRow, err := Queries.UpdateCustomBot(ctx, params)
	if err != nil {
		return nil, err
	}

	AuditChange(ctx, guildID, actor, old, *newRow, database.AuditTypeCustomBots)

	return newRow, nil
}

func UpdateGuildWithAudit(ctx context.Context, params database.UpdateGuildParams, actor discord.Snowflake) (*database.Guilds, error) {
	var old database.Guilds
	if existing, err := Queries.GetGuild(ctx, params.GuildID); err == nil {
		old = *existing
	}

	newRow, err := Queries.UpdateGuild(ctx, params)
	if err != nil {
		return nil, err
	}

	AuditChange(ctx, discord.Snowflake(params.GuildID), actor, old, *newRow, database.AuditTypeGuilds)

	return newRow, nil
}

func AddGuildFeatureWithAudit(ctx context.Context, params database.AddGuildFeatureParams, actor discord.Snowflake) error {
	var oldFeatures []string
	if existing, err := Queries.GetGuildFeatures(ctx, params.GuildID); err == nil {
		oldFeatures = existing
	}

	err := Queries.AddGuildFeature(ctx, params)
	if err != nil {
		return err
	}

	newFeatures, err := Queries.GetGuildFeatures(ctx, params.GuildID)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return err
	}

	AuditChange(ctx, discord.Snowflake(params.GuildID), actor, oldFeatures, newFeatures, database.AuditTypeGuildFeatures)

	return nil
}

func RemoveGuildFeatureWithAudit(ctx context.Context, params database.RemoveGuildFeatureParams, actor discord.Snowflake) error {
	var oldFeatures []string
	if existing, err := Queries.GetGuildFeatures(ctx, params.GuildID); err == nil {
		oldFeatures = existing
	}

	err := Queries.RemoveGuildFeature(ctx, params)
	if err != nil {
		return err
	}

	newFeatures, err := Queries.GetGuildFeatures(ctx, params.GuildID)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return err
	}

	AuditChange(ctx, discord.Snowflake(params.GuildID), actor, oldFeatures, newFeatures, database.AuditTypeGuildFeatures)

	return nil
}

func UpdateBioWithAudit(ctx context.Context, params database.UpdateGuildBioParams, actor discord.Snowflake) (*database.Guilds, error) {
	var old string
	if existing, err := Queries.GetGuild(ctx, params.GuildID); err == nil {
		old = existing.Bio
	}

	newRow, err := Queries.UpdateGuildBio(ctx, params)
	if err != nil {
		return nil, err
	}

	AuditChange(ctx, discord.Snowflake(params.GuildID), actor, old, params.Bio, database.AuditTypeBio)

	return newRow, nil
}
