CREATE TABLE IF NOT EXISTS guild_polls_entries (
    guild_poll_entry_uuid uuid NOT NULL UNIQUE PRIMARY KEY,
    poll_uuid uuid NOT NULL,
    user_id bigint NOT NULL,
    option_index integer NOT NULL,
    created_at timestamp NOT NULL,
    FOREIGN KEY (poll_uuid) REFERENCES guild_polls (poll_uuid) ON DELETE CASCADE ON UPDATE CASCADE
);

CREATE UNIQUE INDEX IF NOT EXISTS guild_polls_entries_poll_uuid_user_id ON guild_polls_entries (poll_uuid, user_id);
CREATE INDEX IF NOT EXISTS guild_polls_entries_poll_uuid ON guild_polls_entries (poll_uuid);