// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.14.0

package database

import (
	"context"
	"database/sql"

	"github.com/gofrs/uuid"
)

type Querier interface {
	CreateAutoModerationGuildSettings(ctx context.Context, guildID int64) (*GuildSettingsAutomoderation, error)
	CreateBorderwallGuildSettings(ctx context.Context, guildID int64) (*GuildSettingsBorderwall, error)
	CreateBorderwallRequest(ctx context.Context, arg *CreateBorderwallRequestParams) (*BorderwallRequests, error)
	CreateCommandError(ctx context.Context, arg *CreateCommandErrorParams) (*ScienceCommandErrors, error)
	CreateCommandUsage(ctx context.Context, arg *CreateCommandUsageParams) (*ScienceCommandUsages, error)
	CreateFreeRolesGuildSettings(ctx context.Context, guildID int64) (*GuildSettingsFreeroles, error)
	CreateGuild(ctx context.Context, arg *CreateGuildParams) (*Guilds, error)
	CreateGuildPunishment(ctx context.Context, arg *CreateGuildPunishmentParams) (*GuildPunishments, error)
	CreateGuildWelcomerImage(ctx context.Context, arg *CreateGuildWelcomerImageParams) (*GuildWelcomerImages, error)
	CreateModerationGuildSettings(ctx context.Context, guildID int64) (*GuildSettingsModeration, error)
	CreateNewMembership(ctx context.Context, arg *CreateNewMembershipParams) (*UserMemberships, error)
	CreatePatreonUser(ctx context.Context, arg *CreatePatreonUserParams) (*PatreonUsers, error)
	CreateRulesGuildSettings(ctx context.Context, guildID int64) (*GuildSettingsRules, error)
	CreateScienceEvent(ctx context.Context, arg *CreateScienceEventParams) (*ScienceEvents, error)
	CreateScienceGuildEvent(ctx context.Context, arg *CreateScienceGuildEventParams) (*ScienceGuildEvents, error)
	CreateScienceGuildModerationEvent(ctx context.Context, arg *CreateScienceGuildModerationEventParams) (*ScienceGuildModerationEvents, error)
	CreateTempChannelsGuildSettings(ctx context.Context, guildID int64) (*GuildSettingsTempchannels, error)
	CreateUser(ctx context.Context, arg *CreateUserParams) (*Users, error)
	CreateUserTransaction(ctx context.Context, arg *CreateUserTransactionParams) (*UserTransactions, error)
	CreateWelcomerDMsGuildSettings(ctx context.Context, guildID int64) (*GuildSettingsWelcomerDms, error)
	CreateWelcomerImagesGuildSettings(ctx context.Context, guildID int64) (*GuildSettingsWelcomerImages, error)
	CreateWelcomerTextGuildSettings(ctx context.Context, guildID int64) (*GuildSettingsWelcomerText, error)
	DeleteGuildWelcomerImage(ctx context.Context, imageUuid uuid.UUID) (int64, error)
	DeletePatreonUser(ctx context.Context, patreonUserID int64) (int64, error)
	GetAutoModerationGuildSettings(ctx context.Context, guildID int64) (*GuildSettingsAutomoderation, error)
	GetBorderwallGuildSettings(ctx context.Context, guildID int64) (*GuildSettingsBorderwall, error)
	GetBorderwallRequest(ctx context.Context, requestUuid uuid.UUID) (*BorderwallRequests, error)
	GetBorderwallRequestsByGuildIDUserID(ctx context.Context, arg *GetBorderwallRequestsByGuildIDUserIDParams) ([]*BorderwallRequests, error)
	GetCommandError(ctx context.Context, commandUuid uuid.UUID) (*GetCommandErrorRow, error)
	GetCommandUsage(ctx context.Context, commandUuid uuid.UUID) (*ScienceCommandUsages, error)
	GetFreeRolesGuildSettings(ctx context.Context, guildID int64) (*GuildSettingsFreeroles, error)
	GetGuild(ctx context.Context, guildID int64) (*Guilds, error)
	GetGuildPunishment(ctx context.Context, punishmentUuid uuid.UUID) (*GuildPunishments, error)
	GetGuildPunishmentsByGuildIDUserID(ctx context.Context, arg *GetGuildPunishmentsByGuildIDUserIDParams) ([]*GuildPunishments, error)
	GetGuildWelcomerImage(ctx context.Context, imageUuid uuid.UUID) (*GuildWelcomerImages, error)
	GetModerationGuildSettings(ctx context.Context, guildID int64) (*GuildSettingsModeration, error)
	GetPatreonUser(ctx context.Context, patreonUserID int64) (*PatreonUsers, error)
	GetPatreonUsersByUserID(ctx context.Context, userID int64) ([]*PatreonUsers, error)
	GetRulesGuildSettings(ctx context.Context, guildID int64) (*GuildSettingsRules, error)
	GetScienceEvent(ctx context.Context, eventUuid uuid.UUID) (*ScienceEvents, error)
	GetScienceGuildEvent(ctx context.Context, guildEventUuid uuid.UUID) (*ScienceGuildEvents, error)
	GetScienceGuildModerationEvent(ctx context.Context, moderationEventID uuid.UUID) (*ScienceGuildModerationEvents, error)
	GetTempChannelsGuildSettings(ctx context.Context, guildID int64) (*GuildSettingsTempchannels, error)
	GetUser(ctx context.Context, userID int64) (*Users, error)
	GetUserMembership(ctx context.Context, membershipUuid uuid.UUID) (*UserMemberships, error)
	GetUserMembershipsByGuildID(ctx context.Context, guildID sql.NullInt64) ([]*UserMemberships, error)
	GetUserMembershipsByUserID(ctx context.Context, userID int64) ([]*UserMemberships, error)
	GetUserTransaction(ctx context.Context, transactionUuid uuid.UUID) (*UserTransactions, error)
	GetUserTransactionsByUserID(ctx context.Context, userID int64) ([]*UserTransactions, error)
	GetWelcomerDMsGuildSettings(ctx context.Context, guildID int64) (*GuildSettingsWelcomerDms, error)
	GetWelcomerImagesGuildSettings(ctx context.Context, guildID int64) (*GuildSettingsWelcomerImages, error)
	GetWelcomerTextGuildSettings(ctx context.Context, guildID int64) (*GuildSettingsWelcomerText, error)
	UpdateAutoModerationGuildSettings(ctx context.Context, arg *UpdateAutoModerationGuildSettingsParams) (int64, error)
	UpdateBorderwallGuildSettings(ctx context.Context, arg *UpdateBorderwallGuildSettingsParams) (int64, error)
	UpdateFreeRolesGuildSettings(ctx context.Context, arg *UpdateFreeRolesGuildSettingsParams) (int64, error)
	UpdateGuild(ctx context.Context, arg *UpdateGuildParams) (int64, error)
	UpdateGuildPunishment(ctx context.Context, arg *UpdateGuildPunishmentParams) (int64, error)
	UpdateModerationGuildSettings(ctx context.Context, arg *UpdateModerationGuildSettingsParams) (int64, error)
	UpdatePatreonUser(ctx context.Context, arg *UpdatePatreonUserParams) (int64, error)
	UpdateRuleGuildSettings(ctx context.Context, arg *UpdateRuleGuildSettingsParams) (int64, error)
	UpdateTempChannelsGuildSettings(ctx context.Context, arg *UpdateTempChannelsGuildSettingsParams) (int64, error)
	UpdateUser(ctx context.Context, arg *UpdateUserParams) (int64, error)
	UpdateUserMembership(ctx context.Context, arg *UpdateUserMembershipParams) (int64, error)
	UpdateUserTransaction(ctx context.Context, arg *UpdateUserTransactionParams) (int64, error)
	UpdateWelcomerDMsGuildSettings(ctx context.Context, arg *UpdateWelcomerDMsGuildSettingsParams) (int64, error)
	UpdateWelcomerImagesGuildSettings(ctx context.Context, arg *UpdateWelcomerImagesGuildSettingsParams) (int64, error)
	UpdateWelcomerTextGuildSettings(ctx context.Context, arg *UpdateWelcomerTextGuildSettingsParams) (int64, error)
}

var _ Querier = (*Queries)(nil)
