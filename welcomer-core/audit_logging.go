package welcomer

import (
	"context"
	"database/sql"

	"github.com/WelcomerTeam/Discord/discord"
	"github.com/WelcomerTeam/Welcomer/welcomer-core/database"
	"github.com/jackc/pgtype"
)

func AuditChange(ctx context.Context, guildID discord.Snowflake, userID discord.Snowflake, oldValue, newValue interface{}, auditType database.AuditType) {
	changesAsJSON, hasChanges, err := CompareStructsAsJSON(oldValue, newValue)
	if err != nil {
		Logger.Warn().Err(err).Msg("Failed to compare structs for audit logging")

		return
	}

	if !hasChanges {
		return
	}

	_, err = Queries.InsertAuditLog(ctx, database.InsertAuditLogParams{
		GuildID: sql.NullInt64{
			Int64: int64(guildID),
			Valid: guildID != 0,
		},
		UserID:    int64(userID),
		AuditType: int32(auditType),
		Changes:   pgtype.JSONB{Bytes: changesAsJSON, Status: pgtype.Present},
	})
	if err != nil {
		Logger.Warn().Err(err).Msg("Failed to insert audit log")
	}
}
