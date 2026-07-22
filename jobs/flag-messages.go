package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/WelcomerTeam/Welcomer/welcomer-core"
	"github.com/WelcomerTeam/Welcomer/welcomer-core/database"
	"github.com/jackc/pgtype"
	_ "github.com/joho/godotenv/autoload"	
)

type moderationTarget struct {
	tableName string
	dataType  database.AuditType
	selectSQL string
	updateSQL string
}

type moderationTargetRow struct {
	guildID int64
	payload []byte
}

var nsfwFilterTerms = []string{
	"nsfw",
	"nude",
	"nudes",
	"porn",
	"porno",
	"xxx",
	"sex",
	"sexy",
	"erotic",
	"hentai",
	"lewd",
	"explicit",
	"boob",
	"boobs",
	"ass",
	"pussy",
	"dick",
	"cock",
	"cum",
	"milf",
	"onlyfans",
}

func main() {
	postgresURL := flag.String("postgresURL", os.Getenv("POSTGRES_URL"), "Postgres connection URL")

	flag.Parse()

	ctx, cancel := context.WithCancel(context.Background())

	welcomer.SetupDatabase(ctx, *postgresURL)

	entrypoint(ctx)

	cancel()
}

func entrypoint(ctx context.Context) {
	targets := []moderationTarget{
		{
			tableName: "guild_settings_welcomer_dms",
			dataType:  database.AuditTypeGuildSettingsWelcomerDms,
			selectSQL: `SELECT guild_id, message_format
FROM guild_settings_welcomer_dms
WHERE moderation_checkup_uuid IS NULL
ORDER BY guild_id`,
			updateSQL: `UPDATE guild_settings_welcomer_dms
SET moderation_checkup_uuid = $1
WHERE guild_id = $2`,
		},
		{
			tableName: "guild_settings_welcomer_text",
			dataType:  database.AuditTypeGuildSettingsWelcomerText,
			selectSQL: `SELECT guild_id, message_format
FROM guild_settings_welcomer_text
WHERE moderation_checkup_uuid IS NULL
ORDER BY guild_id`,
			updateSQL: `UPDATE guild_settings_welcomer_text
SET moderation_checkup_uuid = $1
WHERE guild_id = $2`,
		},
		{
			tableName: "guild_settings_borderwall",
			dataType:  database.AuditTypeGuildSettingsBorderwall,
			selectSQL: `SELECT guild_id, message_verify, message_verified
FROM guild_settings_borderwall
WHERE moderation_checkup_uuid IS NULL
ORDER BY guild_id`,
			updateSQL: `UPDATE guild_settings_borderwall
SET moderation_checkup_uuid = $1
WHERE guild_id = $2`,
		},
	}

	totalMatched := 0
	for _, target := range targets {
		matched, err := processTarget(ctx, target)
		if err != nil {
			panic(err)
		}

		totalMatched += matched
	}

	fmt.Printf("processed %d matching rows\n", totalMatched)
}

func processTarget(ctx context.Context, target moderationTarget) (int, error) {
	rows, err := welcomer.Pool.Query(ctx, target.selectSQL)
	if err != nil {
		return 0, fmt.Errorf("query %s rows: %w", target.tableName, err)
	}
	defer rows.Close()

	matched := 0
	for rows.Next() {
		var row moderationTargetRow
		if target.tableName == "guild_settings_borderwall" {
			var messageVerify pgtype.JSONB
			var messageVerified pgtype.JSONB
			if err := rows.Scan(&row.guildID, &messageVerify, &messageVerified); err != nil {
				return matched, fmt.Errorf("scan %s row: %w", target.tableName, err)
			}

			row.payload = combineJSONBPayload(messageVerify, messageVerified)
		} else {
			var messageFormat pgtype.JSONB
			if err := rows.Scan(&row.guildID, &messageFormat); err != nil {
				return matched, fmt.Errorf("scan %s row: %w", target.tableName, err)
			}

			row.payload = append([]byte(nil), messageFormat.Bytes...)
		}

		if len(row.payload) == 0 {
			continue
		}

		if !matchesNSFWFilter(row.payload) {
			continue
		}

		matched++
		fmt.Printf("matched table=%s guild_id=%d\n", target.tableName, row.guildID)

		if err := enqueueAndLink(ctx, target, row); err != nil {
			return matched, err
		}
	}

	if err := rows.Err(); err != nil {
		return matched, fmt.Errorf("iterate %s rows: %w", target.tableName, err)
	}

	return matched, nil
}

func enqueueAndLink(ctx context.Context, target moderationTarget, row moderationTargetRow) error {
	tx, err := welcomer.Pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx for %s guild_id=%d: %w", target.tableName, row.guildID, err)
	}
	defer tx.Rollback(ctx)

	txQueries := database.New(tx)
	queue, err := txQueries.CreateModerationCheckupQueue(ctx, database.CreateModerationCheckupQueueParams{
		GuildID:  row.guildID,
		UserID:   0,
		DataType: int32(target.dataType),
		Value:    string(row.payload),
	})
	if err != nil {
		return fmt.Errorf("create moderation queue for %s guild_id=%d: %w", target.tableName, row.guildID, err)
	}

	_, err = txQueries.CreateModerationCheckup(ctx, database.CreateModerationCheckupParams{
		CheckupUuid: queue.CheckupQueueUuid,
		GuildID:     row.guildID,
		UserID:      0,
		DataType:    int32(target.dataType),
	})
	if err != nil {
		return fmt.Errorf("create moderation checkup for %s guild_id=%d: %w", target.tableName, row.guildID, err)
	}

	_, err = tx.Exec(ctx, target.updateSQL, queue.CheckupQueueUuid, row.guildID)
	if err != nil {
		return fmt.Errorf("link moderation checkup for %s guild_id=%d: %w", target.tableName, row.guildID, err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit moderation checkup for %s guild_id=%d: %w", target.tableName, row.guildID, err)
	}

	return nil
}

func matchesNSFWFilter(payload []byte) bool {
	text := strings.ToLower(string(payload))
	for _, term := range nsfwFilterTerms {
		if strings.Contains(text, term) {
			return true
		}
	}

	return false
}

func combineJSONBPayload(messageVerify, messageVerified pgtype.JSONB) []byte {
	parts := make([]json.RawMessage, 0, 2)

	if messageVerify.Status == pgtype.Present {
		parts = append(parts, json.RawMessage(messageVerify.Bytes))
	}

	if messageVerified.Status == pgtype.Present {
		parts = append(parts, json.RawMessage(messageVerified.Bytes))
	}

	if len(parts) == 0 {
		return nil
	}

	combined, err := json.Marshal(parts)
	if err != nil {
		return nil
	}

	return combined
}
