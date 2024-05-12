CREATE TABLE IF NOT EXISTS guild_invites (
    invite_code text NOT NULL UNIQUE PRIMARY KEY,
    guild_id bigint NOT NULL,
    created_by bigint NOT NULL,
    created_at timestamp NOT NULL,
    uses bigint NOT NULL
);

CREATE INDEX IF NOT EXISTS guild_invites_guild_id ON guild_invites (guild_id);

CREATE INDEX IF NOT EXISTS guild_invites_created_by ON guild_invites (created_by);