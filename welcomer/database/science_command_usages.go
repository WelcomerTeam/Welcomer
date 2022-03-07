package database

import (
	"context"
	"time"

	"github.com/WelcomerTeam/Discord/discord"
	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v4/pgxpool"
)

type ScienceCommandUsages struct {
	*pgxpool.Pool
}

type ScienceCommandUsage struct {
	CommandUUID     uuid.UUID
	DateTime        time.Time
	GuildID         *discord.Snowflake
	UserID          discord.Snowflake
	ChannelID       *discord.Snowflake
	Command         string
	Errored         bool
	ExecutionTimeMs int64
}

func newScienceCommandUsages(pool *pgxpool.Pool) *ScienceCommandUsages {
	return &ScienceCommandUsages{pool}
}

func (t *ScienceCommandUsages) Schema() string {
	return `
CREATE TABLE IF NOT EXISTS "science_command_usages" (
	"command_uuid" uuid NOT NULL UNIQUE PRIMARY KEY,
	"datetime" timestamp NOT NULL,
	"guild_id" bigint NULL,
	"user_id" bigint NOT NULL,
	"channel_id" bigint NOT NULL,
	"command" text NOT NULL,
	"errored" boolean NOT NULL,
	"execution_time_ms" integer NOT NULL
);

CREATE INDEX IF NOT EXISTS "science_command_usages_user_id" ON "science_command_usages" ("user_id");
CREATE INDEX IF NOT EXISTS "science_command_usages_guild_id" ON "science_command_usages" ("guild_id");
CREATE INDEX IF NOT EXISTS "science_command_usages_command" ON "science_command_usages" ("command");
`
}

func (t *ScienceCommandUsages) Create(ctx context.Context, guildID *discord.Snowflake, userID discord.Snowflake, channelID *discord.Snowflake, commandTree string, errored bool, executionTime int64) (commandUUID uuid.UUID, err error) {
	query := `INSERT INTO "science_command_usages" ("command_uuid", "datetime", "guild_id", "user_id", "channel_id", "command", "errored", "execution_time_ms") VALUES (uuid_generate_v4(), now(), $1, $2, $3, $4, $5, $6) RETURNING "command_uuid";`

	row := t.QueryRow(ctx, query, guildID, userID, channelID, commandTree, errored, executionTime)
	err = row.Scan(&commandUUID)

	return
}

func (t *ScienceCommandUsages) Get(ctx context.Context, commandUUID uuid.UUID) (scienceCommandUsage *ScienceCommandUsage, err error) {
	scienceCommandUsage = &ScienceCommandUsage{}

	query := `SELECT command_uuid, datetime, guild_id, user_id, channel_id, command, errored, execution_time_ms FROM "science_command_usages" WHERE command_uuid = $1`
	row := t.QueryRow(ctx, query, commandUUID)
	err = row.Scan(
		&scienceCommandUsage.CommandUUID,
		&scienceCommandUsage.DateTime,
		&scienceCommandUsage.GuildID,
		&scienceCommandUsage.UserID,
		&scienceCommandUsage.ChannelID,
		&scienceCommandUsage.Command,
		&scienceCommandUsage.Errored,
		&scienceCommandUsage.ExecutionTimeMs,
	)

	return
}
