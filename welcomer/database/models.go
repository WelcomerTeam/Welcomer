// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.15.0

package database

import (
	"database/sql"
	"time"

	"github.com/gofrs/uuid"
	"github.com/jackc/pgtype"
)

type BorderwallRequests struct {
	RequestUuid uuid.UUID    `json:"request_uuid"`
	CreatedAt   time.Time    `json:"created_at"`
	UpdatedAt   time.Time    `json:"updated_at"`
	GuildID     int64        `json:"guild_id"`
	UserID      int64        `json:"user_id"`
	IsVerified  bool         `json:"is_verified"`
	VerifiedAt  sql.NullTime `json:"verified_at"`
}

type GuildSettingsAutomoderation struct {
	GuildID                     int64        `json:"guild_id"`
	ToggleMasscaps              sql.NullBool `json:"toggle_masscaps"`
	ToggleMassmentions          sql.NullBool `json:"toggle_massmentions"`
	ToggleUrls                  sql.NullBool `json:"toggle_urls"`
	ToggleInvites               sql.NullBool `json:"toggle_invites"`
	ToggleRegex                 sql.NullBool `json:"toggle_regex"`
	TogglePhishing              sql.NullBool `json:"toggle_phishing"`
	ToggleMassemojis            sql.NullBool `json:"toggle_massemojis"`
	ThresholdMassCapsPercentage int32        `json:"threshold_mass_caps_percentage"`
	ThresholdMassMentions       int32        `json:"threshold_mass_mentions"`
	ThresholdMassEmojis         int32        `json:"threshold_mass_emojis"`
	BlacklistedUrls             []string     `json:"blacklisted_urls"`
	WhitelistedUrls             []string     `json:"whitelisted_urls"`
	RegexPatterns               []string     `json:"regex_patterns"`
	ToggleSmartmod              bool         `json:"toggle_smartmod"`
}

type GuildSettingsBorderwall struct {
	GuildID         int64        `json:"guild_id"`
	ToggleEnabled   sql.NullBool `json:"toggle_enabled"`
	MessageVerify   pgtype.JSONB `json:"message_verify"`
	MessageVerified pgtype.JSONB `json:"message_verified"`
	RolesOnJoin     []int64      `json:"roles_on_join"`
	RolesOnVerify   []int64      `json:"roles_on_verify"`
}

type GuildSettingsFreeroles struct {
	GuildID       int64        `json:"guild_id"`
	ToggleEnabled sql.NullBool `json:"toggle_enabled"`
	Roles         []int64      `json:"roles"`
}

type GuildSettingsModeration struct {
	GuildID               int64        `json:"guild_id"`
	ToggleReasonsRequired sql.NullBool `json:"toggle_reasons_required"`
}

type GuildSettingsRules struct {
	GuildID          int64        `json:"guild_id"`
	ToggleEnabled    sql.NullBool `json:"toggle_enabled"`
	ToggleDmsEnabled sql.NullBool `json:"toggle_dms_enabled"`
	Rules            []string     `json:"rules"`
}

type GuildSettingsTempchannels struct {
	GuildID          int64         `json:"guild_id"`
	ToggleEnabled    sql.NullBool  `json:"toggle_enabled"`
	ToggleAutopurge  sql.NullBool  `json:"toggle_autopurge"`
	ChannelLobby     sql.NullInt64 `json:"channel_lobby"`
	ChannelCategory  sql.NullInt64 `json:"channel_category"`
	DefaultUserCount sql.NullInt32 `json:"default_user_count"`
}

type GuildSettingsWelcomerDms struct {
	GuildID             int64        `json:"guild_id"`
	ToggleEnabled       sql.NullBool `json:"toggle_enabled"`
	ToggleUseTextFormat sql.NullBool `json:"toggle_use_text_format"`
	ToggleIncludeImage  sql.NullBool `json:"toggle_include_image"`
	MessageFormat       pgtype.JSONB `json:"message_format"`
}

type GuildSettingsWelcomerImages struct {
	GuildID                int64          `json:"guild_id"`
	ToggleEnabled          sql.NullBool   `json:"toggle_enabled"`
	ToggleImageBorder      sql.NullBool   `json:"toggle_image_border"`
	BackgroundName         sql.NullString `json:"background_name"`
	ColourText             sql.NullString `json:"colour_text"`
	ColourTextBorder       sql.NullString `json:"colour_text_border"`
	ColourImageBorder      sql.NullString `json:"colour_image_border"`
	ColourProfileBorder    sql.NullString `json:"colour_profile_border"`
	ImageAlignment         sql.NullInt32  `json:"image_alignment"`
	ImageTheme             sql.NullInt32  `json:"image_theme"`
	ImageMessage           sql.NullString `json:"image_message"`
	ImageProfileBorderType sql.NullInt32  `json:"image_profile_border_type"`
}

type GuildSettingsWelcomerText struct {
	GuildID       int64         `json:"guild_id"`
	ToggleEnabled sql.NullBool  `json:"toggle_enabled"`
	Channel       sql.NullInt64 `json:"channel"`
	MessageFormat pgtype.JSONB  `json:"message_format"`
}

type Guilds struct {
	GuildID          int64          `json:"guild_id"`
	CreatedAt        time.Time      `json:"created_at"`
	UpdatedAt        time.Time      `json:"updated_at"`
	Name             string         `json:"name"`
	EmbedColour      int32          `json:"embed_colour"`
	SiteSplashUrl    sql.NullString `json:"site_splash_url"`
	SiteStaffVisible sql.NullBool   `json:"site_staff_visible"`
	SiteGuildVisible sql.NullBool   `json:"site_guild_visible"`
	SiteAllowInvites sql.NullBool   `json:"site_allow_invites"`
}

type PatreonUsers struct {
	PatreonUserID int64          `json:"patreon_user_id"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	UserID        int64          `json:"user_id"`
	FullName      string         `json:"full_name"`
	Email         sql.NullString `json:"email"`
	ThumbUrl      sql.NullString `json:"thumb_url"`
}

type ScienceCommandErrors struct {
	CommandUuid uuid.UUID    `json:"command_uuid"`
	Trace       string       `json:"trace"`
	Data        pgtype.JSONB `json:"data"`
}

type ScienceCommandUsages struct {
	CommandUuid     uuid.UUID     `json:"command_uuid"`
	CreatedAt       time.Time     `json:"created_at"`
	UpdatedAt       time.Time     `json:"updated_at"`
	GuildID         sql.NullInt64 `json:"guild_id"`
	UserID          int64         `json:"user_id"`
	ChannelID       sql.NullInt64 `json:"channel_id"`
	Command         string        `json:"command"`
	Errored         bool          `json:"errored"`
	ExecutionTimeMs int64         `json:"execution_time_ms"`
}

type ScienceEvents struct {
	EventUuid uuid.UUID    `json:"event_uuid"`
	CreatedAt time.Time    `json:"created_at"`
	EventType int32        `json:"event_type"`
	Data      pgtype.JSONB `json:"data"`
}

type ScienceGuildEvents struct {
	GuildEventUuid uuid.UUID    `json:"guild_event_uuid"`
	GuildID        int64        `json:"guild_id"`
	CreatedAt      time.Time    `json:"created_at"`
	EventType      int32        `json:"event_type"`
	Data           pgtype.JSONB `json:"data"`
}

type UserMemberships struct {
	MembershipUuid  uuid.UUID     `json:"membership_uuid"`
	CreatedAt       time.Time     `json:"created_at"`
	UpdatedAt       time.Time     `json:"updated_at"`
	StartedAt       time.Time     `json:"started_at"`
	ExpiresAt       time.Time     `json:"expires_at"`
	Status          int32         `json:"status"`
	MembershipType  int32         `json:"membership_type"`
	TransactionUuid uuid.NullUUID `json:"transaction_uuid"`
	UserID          int64         `json:"user_id"`
	GuildID         sql.NullInt64 `json:"guild_id"`
}

type UserTransactions struct {
	TransactionUuid   uuid.UUID `json:"transaction_uuid"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
	UserID            int64     `json:"user_id"`
	PlatformType      int32     `json:"platform_type"`
	TransactionID     string    `json:"transaction_id"`
	TransactionStatus int32     `json:"transaction_status"`
	CurrencyCode      string    `json:"currency_code"`
	Amount            int32     `json:"amount"`
}

type Users struct {
	UserID        int64          `json:"user_id"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	Name          string         `json:"name"`
	Discriminator string         `json:"discriminator"`
	AvatarHash    sql.NullString `json:"avatar_hash"`
}
