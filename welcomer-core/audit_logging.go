package welcomer

import (
	"context"
	"database/sql"

	"github.com/WelcomerTeam/Discord/discord"
	"github.com/WelcomerTeam/Welcomer/welcomer-core/database"
)

func AuditChange[T comparable](ctx context.Context, guildID discord.Snowflake, userID discord.Snowflake, oldValue, newValue T, auditType database.AuditType) {
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
		UserID:       int64(userID),
		AuditType:    int32(auditType),
		ChangesBytes: changesAsJSON,
	})
	if err != nil {
		Logger.Warn().Err(err).Msg("Failed to insert audit log")
	}
}
