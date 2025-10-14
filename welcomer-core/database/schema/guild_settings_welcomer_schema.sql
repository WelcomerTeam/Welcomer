CREATE TABLE IF NOT EXISTS guild_settings_welcomer (
    guild_id bigint NOT NULL UNIQUE PRIMARY KEY,
    auto_delete_welcome_messages boolean NOT NULL,
    welcome_message_lifetime integer NOT NULL,
    auto_delete_welcome_messages_on_leave boolean NOT NULL,
    FOREIGN KEY (guild_id) REFERENCES guilds (guild_id) ON DELETE CASCADE ON UPDATE CASCADE
)