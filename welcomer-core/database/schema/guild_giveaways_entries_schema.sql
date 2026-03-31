CREATE TABLE IF NOT EXISTS guild_giveaways_entries (
    guild_giveaway_entry_uuid uuid NOT NULL UNIQUE PRIMARY KEY,
    giveaway_uuid uuid NOT NULL,
    user_id bigint NOT NULL,
    created_at timestamp NOT NULL,
    FOREIGN KEY (giveaway_uuid) REFERENCES guild_giveaways (giveaway_uuid) ON DELETE CASCADE ON UPDATE CASCADE
);

CREATE UNIQUE INDEX IF NOT EXISTS guild_giveaways_entries_giveaway_uuid_user_id ON guild_giveaways_entries (giveaway_uuid, user_id);
CREATE INDEX IF NOT EXISTS guild_giveaways_entries_giveaway_uuid ON guild_giveaways_entries (giveaway_uuid);