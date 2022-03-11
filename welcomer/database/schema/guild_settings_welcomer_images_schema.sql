CREATE TABLE IF NOT EXISTS guild_settings_welcomer_images (
    guild_id bigint NOT NULL UNIQUE PRIMARY KEY,
    toggle_enabled boolean DEFAULT 'false',
    toggle_image_border boolean DEFAULT 'false',
    background_name text NULL,
    colour_text integer DEFAULT '16777216',
    colour_text_border integer DEFAULT '0',
    colour_image_border integer DEFAULT '16777216',
    colour_profile_border integer DEFAULT '16777216',
    image_alignment integer DEFAULT '0',
    image_theme integer DEFAULT '0',
    image_message text NULL,
    image_profile_border_type integer DEFAULT '0',
    FOREIGN KEY (guild_id) REFERENCES guilds (guild_id) ON DELETE CASCADE ON UPDATE CASCADE
);

