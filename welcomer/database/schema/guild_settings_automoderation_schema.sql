CREATE TABLE IF NOT EXISTS guild_settings_automoderation (
    guild_id bigint NOT NULL UNIQUE PRIMARY KEY,
    toggle_masscaps boolean DEFAULT 'false',
    toggle_massmentions boolean DEFAULT 'false',
    toggle_urls boolean DEFAULT 'false',
    toggle_invites boolean DEFAULT 'false',
    toggle_regex boolean DEFAULT 'false',
    toggle_phishing boolean DEFAULT 'false',
    toggle_massemojis boolean DEFAULT 'false',
    threshold_mass_caps_percentage integer NOT NULL DEFAULT '75',
    threshold_mass_mentions integer NOT NULL DEFAULT '5',
    threshold_mass_emojis integer NOT NULL DEFAULT '10',
    blacklisted_urls text[],
    whitelisted_urls text[],
    regex_patterns text[],
    toggle_smartmod boolean NOT NULL DEFAULT 'false',
    FOREIGN KEY (guild_id) REFERENCES guilds (guild_id) ON DELETE CASCADE ON UPDATE CASCADE
);

