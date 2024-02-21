CREATE TABLE IF NOT EXISTS borderwall_requests (
    request_uuid uuid NOT NULL UNIQUE PRIMARY KEY,
    created_at timestamp NOT NULL,
    updated_at timestamp NOT NULL,
    guild_id bigint NOT NULL,
    user_id bigint NOT NULL,
    is_verified boolean NOT NULL,
    verified_at timestamp,
    ip_address inet,
    recaptcha_score real,
    ipintel_score real,
    ua_family text,
    ua_family_version text,
    ua_os text,
    ua_os_version text
);

CREATE INDEX IF NOT EXISTS borderwall_requests_guild_id_user_id ON borderwall_requests (guild_id, user_id);

CREATE INDEX IF NOT EXISTS borderwall_requests_user_id ON borderwall_requests (user_id);

CREATE INDEX IF NOT EXISTS borderwall_requests_ip_address ON borderwall_requests (ip_address);