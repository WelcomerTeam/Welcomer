// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.26.0

package database

import (
	"context"

	"github.com/gofrs/uuid"
	"github.com/jackc/pgtype"
)

type Querier interface {
	CreateAutoRolesGuildSettings(ctx context.Context, arg *CreateAutoRolesGuildSettingsParams) (*GuildSettingsAutoroles, error)
	CreateBorderwallGuildSettings(ctx context.Context, arg *CreateBorderwallGuildSettingsParams) (*GuildSettingsBorderwall, error)
	CreateBorderwallRequest(ctx context.Context, arg *CreateBorderwallRequestParams) (*BorderwallRequests, error)
	CreateCommandError(ctx context.Context, arg *CreateCommandErrorParams) (*ScienceCommandErrors, error)
	CreateCommandUsage(ctx context.Context, arg *CreateCommandUsageParams) (*ScienceCommandUsages, error)
	CreateFreeRolesGuildSettings(ctx context.Context, arg *CreateFreeRolesGuildSettingsParams) (*GuildSettingsFreeroles, error)
	CreateGuild(ctx context.Context, arg *CreateGuildParams) (*Guilds, error)
	CreateGuildInvites(ctx context.Context, arg *CreateGuildInvitesParams) (*GuildInvites, error)
	CreateLeaverGuildSettings(ctx context.Context, arg *CreateLeaverGuildSettingsParams) (*GuildSettingsLeaver, error)
	CreateNewMembership(ctx context.Context, arg *CreateNewMembershipParams) (*UserMemberships, error)
	CreateOrUpdateAutoRolesGuildSettings(ctx context.Context, arg *CreateOrUpdateAutoRolesGuildSettingsParams) (*GuildSettingsAutoroles, error)
	CreateOrUpdateBorderwallGuildSettings(ctx context.Context, arg *CreateOrUpdateBorderwallGuildSettingsParams) (*GuildSettingsBorderwall, error)
	CreateOrUpdateFreeRolesGuildSettings(ctx context.Context, arg *CreateOrUpdateFreeRolesGuildSettingsParams) (*GuildSettingsFreeroles, error)
	CreateOrUpdateGuild(ctx context.Context, arg *CreateOrUpdateGuildParams) (*Guilds, error)
	CreateOrUpdateGuildInvites(ctx context.Context, arg *CreateOrUpdateGuildInvitesParams) (*GuildInvites, error)
	CreateOrUpdateLeaverGuildSettings(ctx context.Context, arg *CreateOrUpdateLeaverGuildSettingsParams) (*GuildSettingsLeaver, error)
	CreateOrUpdateNewMembership(ctx context.Context, arg *CreateOrUpdateNewMembershipParams) (*UserMemberships, error)
	CreateOrUpdatePatreonUser(ctx context.Context, arg *CreateOrUpdatePatreonUserParams) (*PatreonUsers, error)
	CreateOrUpdateRulesGuildSettings(ctx context.Context, arg *CreateOrUpdateRulesGuildSettingsParams) (*GuildSettingsRules, error)
	CreateOrUpdateTempChannelsGuildSettings(ctx context.Context, arg *CreateOrUpdateTempChannelsGuildSettingsParams) (*GuildSettingsTempchannels, error)
	CreateOrUpdateTimeRolesGuildSettings(ctx context.Context, arg *CreateOrUpdateTimeRolesGuildSettingsParams) (*GuildSettingsTimeroles, error)
	CreateOrUpdateUser(ctx context.Context, arg *CreateOrUpdateUserParams) (*Users, error)
	CreateOrUpdateUserTransaction(ctx context.Context, arg *CreateOrUpdateUserTransactionParams) (*UserTransactions, error)
	CreateOrUpdateWelcomerDMsGuildSettings(ctx context.Context, arg *CreateOrUpdateWelcomerDMsGuildSettingsParams) (*GuildSettingsWelcomerDms, error)
	CreateOrUpdateWelcomerImagesGuildSettings(ctx context.Context, arg *CreateOrUpdateWelcomerImagesGuildSettingsParams) (*GuildSettingsWelcomerImages, error)
	CreateOrUpdateWelcomerTextGuildSettings(ctx context.Context, arg *CreateOrUpdateWelcomerTextGuildSettingsParams) (*GuildSettingsWelcomerText, error)
	CreatePatreonUser(ctx context.Context, arg *CreatePatreonUserParams) (*PatreonUsers, error)
	CreateRulesGuildSettings(ctx context.Context, arg *CreateRulesGuildSettingsParams) (*GuildSettingsRules, error)
	CreateScienceEvent(ctx context.Context, arg *CreateScienceEventParams) (*ScienceEvents, error)
	CreateScienceGuildEvent(ctx context.Context, arg *CreateScienceGuildEventParams) (*ScienceGuildEvents, error)
	CreateTempChannelsGuildSettings(ctx context.Context, arg *CreateTempChannelsGuildSettingsParams) (*GuildSettingsTempchannels, error)
	CreateTimeRolesGuildSettings(ctx context.Context, arg *CreateTimeRolesGuildSettingsParams) (*GuildSettingsTimeroles, error)
	CreateUser(ctx context.Context, arg *CreateUserParams) (*Users, error)
	CreateUserTransaction(ctx context.Context, arg *CreateUserTransactionParams) (*UserTransactions, error)
	CreateWelcomerDMsGuildSettings(ctx context.Context, arg *CreateWelcomerDMsGuildSettingsParams) (*GuildSettingsWelcomerDms, error)
	CreateWelcomerImages(ctx context.Context, arg *CreateWelcomerImagesParams) (*WelcomerImages, error)
	CreateWelcomerImagesGuildSettings(ctx context.Context, arg *CreateWelcomerImagesGuildSettingsParams) (*GuildSettingsWelcomerImages, error)
	CreateWelcomerTextGuildSettings(ctx context.Context, arg *CreateWelcomerTextGuildSettingsParams) (*GuildSettingsWelcomerText, error)
	DeleteGuildInvites(ctx context.Context, arg *DeleteGuildInvitesParams) (int64, error)
	DeletePatreonUser(ctx context.Context, patreonUserID int64) (int64, error)
	DeleteWelcomerImage(ctx context.Context, welcomerImageUuid uuid.UUID) (int64, error)
	GetAutoRolesGuildSettings(ctx context.Context, guildID int64) (*GuildSettingsAutoroles, error)
	GetBorderwallGuildSettings(ctx context.Context, guildID int64) (*GuildSettingsBorderwall, error)
	GetBorderwallRequest(ctx context.Context, requestUuid uuid.UUID) (*BorderwallRequests, error)
	GetBorderwallRequestsByGuildIDUserID(ctx context.Context, arg *GetBorderwallRequestsByGuildIDUserIDParams) ([]*BorderwallRequests, error)
	GetBorderwallRequestsByIPAddress(ctx context.Context, ipAddress pgtype.Inet) ([]*BorderwallRequests, error)
	GetCommandError(ctx context.Context, commandUuid uuid.UUID) (*GetCommandErrorRow, error)
	GetCommandUsage(ctx context.Context, commandUuid uuid.UUID) (*ScienceCommandUsages, error)
	GetFreeRolesGuildSettings(ctx context.Context, guildID int64) (*GuildSettingsFreeroles, error)
	GetGuild(ctx context.Context, guildID int64) (*Guilds, error)
	GetGuildInvites(ctx context.Context, guildID int64) ([]*GuildInvites, error)
	GetLeaverGuildSettings(ctx context.Context, guildID int64) (*GuildSettingsLeaver, error)
	GetPatreonUser(ctx context.Context, patreonUserID int64) (*PatreonUsers, error)
	GetPatreonUsersByUserID(ctx context.Context, userID int64) ([]*PatreonUsers, error)
	GetRulesGuildSettings(ctx context.Context, guildID int64) (*GuildSettingsRules, error)
	GetScienceEvent(ctx context.Context, eventUuid uuid.UUID) (*ScienceEvents, error)
	GetScienceGuildEvent(ctx context.Context, guildEventUuid uuid.UUID) (*ScienceGuildEvents, error)
	GetTempChannelsGuildSettings(ctx context.Context, guildID int64) (*GuildSettingsTempchannels, error)
	GetTimeRolesGuildSettings(ctx context.Context, guildID int64) (*GuildSettingsTimeroles, error)
	GetUser(ctx context.Context, userID int64) (*Users, error)
	GetUserMembership(ctx context.Context, membershipUuid uuid.UUID) (*GetUserMembershipRow, error)
	GetUserMembershipsByGuildID(ctx context.Context, guildID int64) ([]*GetUserMembershipsByGuildIDRow, error)
	GetUserMembershipsByUserID(ctx context.Context, userID int64) ([]*GetUserMembershipsByUserIDRow, error)
	GetUserTransaction(ctx context.Context, transactionUuid uuid.UUID) (*UserTransactions, error)
	GetUserTransactionsByTransactionID(ctx context.Context, transactionID string) ([]*UserTransactions, error)
	GetUserTransactionsByUserID(ctx context.Context, userID int64) ([]*UserTransactions, error)
	GetWelcomerDMsGuildSettings(ctx context.Context, guildID int64) (*GuildSettingsWelcomerDms, error)
	GetWelcomerImages(ctx context.Context, welcomerImageUuid uuid.UUID) (*WelcomerImages, error)
	GetWelcomerImagesByGuildId(ctx context.Context, guildID int64) ([]*WelcomerImages, error)
	GetWelcomerImagesGuildSettings(ctx context.Context, guildID int64) (*GuildSettingsWelcomerImages, error)
	GetWelcomerTextGuildSettings(ctx context.Context, guildID int64) (*GuildSettingsWelcomerText, error)
	UpdateAutoRolesGuildSettings(ctx context.Context, arg *UpdateAutoRolesGuildSettingsParams) (int64, error)
	UpdateBorderwallGuildSettings(ctx context.Context, arg *UpdateBorderwallGuildSettingsParams) (int64, error)
	UpdateBorderwallRequest(ctx context.Context, arg *UpdateBorderwallRequestParams) (int64, error)
	UpdateFreeRolesGuildSettings(ctx context.Context, arg *UpdateFreeRolesGuildSettingsParams) (int64, error)
	UpdateGuild(ctx context.Context, arg *UpdateGuildParams) (int64, error)
	UpdateLeaverGuildSettings(ctx context.Context, arg *UpdateLeaverGuildSettingsParams) (int64, error)
	UpdatePatreonUser(ctx context.Context, arg *UpdatePatreonUserParams) (int64, error)
	UpdateRuleGuildSettings(ctx context.Context, arg *UpdateRuleGuildSettingsParams) (int64, error)
	UpdateTempChannelsGuildSettings(ctx context.Context, arg *UpdateTempChannelsGuildSettingsParams) (int64, error)
	UpdateTimeRolesGuildSettings(ctx context.Context, arg *UpdateTimeRolesGuildSettingsParams) (int64, error)
	UpdateUser(ctx context.Context, arg *UpdateUserParams) (int64, error)
	UpdateUserMembership(ctx context.Context, arg *UpdateUserMembershipParams) (int64, error)
	UpdateUserTransaction(ctx context.Context, arg *UpdateUserTransactionParams) (int64, error)
	UpdateWelcomerDMsGuildSettings(ctx context.Context, arg *UpdateWelcomerDMsGuildSettingsParams) (int64, error)
	UpdateWelcomerImagesGuildSettings(ctx context.Context, arg *UpdateWelcomerImagesGuildSettingsParams) (int64, error)
	UpdateWelcomerTextGuildSettings(ctx context.Context, arg *UpdateWelcomerTextGuildSettingsParams) (int64, error)
}

var _ Querier = (*Queries)(nil)
