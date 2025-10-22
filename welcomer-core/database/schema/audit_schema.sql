CREATE TABLE IF NOT EXISTS audit_logs (
    audit_uuid uuid NOT NULL UNIQUE PRIMARY KEY,
    created_at timestamp NOT NULL,
    guild_id BIGINT,
    user_id BIGINT NOT NULL,
    audit_type INT NOT NULL,
    changes_bytes BYTEA NOT NULL
);