CREATE TABLE IF NOT EXISTS guild_giveaways_winners (
    giveaway_winner_uuid uuid NOT NULL UNIQUE PRIMARY KEY,
    giveaway_uuid uuid NOT NULL,
    user_id bigint NOT NULL,
    prize text NOT NULL,
    message_id bigint NOT NULL,
    FOREIGN KEY (giveaway_uuid) REFERENCES guild_giveaways (giveaway_uuid) ON DELETE CASCADE ON UPDATE CASCADE
);

CREATE INDEX IF NOT EXISTS guild_giveaways_winners_giveaway_uuid ON guild_giveaways_winners (giveaway_uuid);