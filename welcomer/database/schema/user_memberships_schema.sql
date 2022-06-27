CREATE TABLE IF NOT EXISTS user_memberships (
    membership_uuid uuid NOT NULL UNIQUE PRIMARY KEY,
    origin_membership_id uuid NULL,
    created_at timestamp NOT NULL,
    updated_at timestamp NOT NULL,
    status integer NOT NULL,
    transaction_uuid uuid NULL,
    user_id bigint NOT NULL,
    guild_id bigint NULL,
    FOREIGN KEY (transaction_uuid) REFERENCES user_transactions (transaction_uuid) ON DELETE RESTRICT ON UPDATE CASCADE,
    FOREIGN KEY (origin_membership_id) REFERENCES user_memberships (membership_uuid) ON DELETE RESTRICT ON UPDATE CASCADE
);

CREATE INDEX IF NOT EXISTS user_memberships_guild_id ON user_memberships (guild_id);

CREATE INDEX IF NOT EXISTS user_memberships_user_id ON user_memberships (user_id);

