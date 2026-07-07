CREATE TABLE IF NOT EXISTS moderation_checkup (
    checkup_uuid uuid NOT NULL UNIQUE PRIMARY KEY,
    guild_id bigint NOT NULL,
    user_id bigint NOT NULL,
    data_type INT NOT NULL,
    started_at timestamp NOT NULL,
    completed_at timestamp,
    
    dom JSONB,
    inv JSONB,
    score_change FLOAT,
    score_safe FLOAT,
    score_question FLOAT,
    score_explicit FLOAT,
    is_blocked BOOLEAN
)