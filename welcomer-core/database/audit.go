package database

//go:generate go-enum -f=$GOFILE --marshal

// ENUM(unknown, borderwall_requests, custom_bots, guild_settings_autoroles, guild_settings_borderwall, guild_settings_freeroles, guild_settings_leaver, guild_settings_rules, guild_settings_tempchannels, guild_settings_timeroles, guild_settings_welcomer, guild_settings_welcomer_dms, guild_settings_welcomer_images, guild_settings_welcomer_text, guilds, users, welcomer_images, guild_features)
type AuditType int32
