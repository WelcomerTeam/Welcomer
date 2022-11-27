CREATE TABLE IF NOT EXISTS user_memberships (
    membership_uuid uuid NOT NULL UNIQUE PRIMARY KEY,
    created_at timestamp NOT NULL,
    updated_at timestamp NOT NULL,
    started_at timestamp NOT NULL,
    expires_at timestamp NOT NULL,
    status integer NOT NULL,
    membership_type integer NOT NULL,
    transaction_uuid uuid NOT NULL,
    user_id bigint NOT NULL,
    guild_id bigint NOT NULL,
    FOREIGN KEY (transaction_uuid) REFERENCES user_transactions (transaction_uuid) ON DELETE RESTRICT ON UPDATE CASCADE
);

CREATE INDEX IF NOT EXISTS user_memberships_guild_id ON user_memberships (guild_id);

CREATE INDEX IF NOT EXISTS user_memberships_user_id ON user_memberships (user_id);

