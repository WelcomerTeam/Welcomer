CREATE TABLE IF NOT EXISTS interaction_commands (
    application_id bigint NOT NULL,
    command text NOT NULL,
    interaction_id bigint NOT NULL,
    created_at timestamp NOT NULL
);

CREATE UNIQUE INDEX IF NOT EXISTS interaction_commands_unique ON bot_interaction_commands (application_id, command, interaction_id);