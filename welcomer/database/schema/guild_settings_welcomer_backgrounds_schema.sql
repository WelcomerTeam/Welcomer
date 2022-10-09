CREATE TABLE IF NOT EXISTS guild_settings_welcomer_backgrounds (
	image_uuid uuid NOT NULL UNIQUE PRIMARY KEY,
	created_at timestamp NOT NULL,
	guild_id bigint NOT NULL,
	filename text NOT NULL,
	filesize int NOT NULL,
	filetype text NOT NULL,
	FOREIGN KEY (guild_id) REFERENCES guilds (guild_id) ON DELETE CASCADE ON UPDATE CASCADE
);
