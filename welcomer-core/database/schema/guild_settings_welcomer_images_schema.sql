CREATE TABLE IF NOT EXISTS guild_settings_welcomer_images (
    guild_id bigint NOT NULL UNIQUE PRIMARY KEY,
    toggle_enabled boolean NOT NULL,
    toggle_image_border boolean NOT NULL,
    toggle_show_avatar boolean NOT NULL,
    background_name text NOT NULL,
    colour_text text NOT NULL,
    colour_text_border text NOT NULL,
    colour_image_border text NOT NULL,
    colour_profile_border text NOT NULL,
    image_alignment integer NOT NULL,
    image_theme integer NOT NULL,
    image_message text NOT NULL,
    image_profile_border_type integer NOT NULL,
    FOREIGN KEY (guild_id) REFERENCES guilds (guild_id) ON DELETE CASCADE ON UPDATE CASCADE
);
