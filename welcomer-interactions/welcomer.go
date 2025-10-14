package welcomer

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"runtime/debug"
	"strings"
	"time"

	"github.com/WelcomerTeam/Discord/discord"
	subway "github.com/WelcomerTeam/Subway/subway"
	"github.com/WelcomerTeam/Welcomer/welcomer-core"
	"github.com/WelcomerTeam/Welcomer/welcomer-core/database"
	plugins "github.com/WelcomerTeam/Welcomer/welcomer-interactions/plugins"
	"github.com/jackc/pgtype"
)

func NewWelcomer(ctx context.Context, options subway.SubwayOptions) *subway.Subway {
	sub, err := subway.NewSubway(ctx, options)
	if err != nil {
		panic(fmt.Errorf("subway.NewSubway(%v): %v", options, err))
	}

	sub.Commands.ErrorHandler = func(ctx context.Context, sub *subway.Subway, interaction discord.Interaction, err error) (*discord.InteractionResponse, error) {
		s, _ := json.Marshal(interaction)

		welcomer.Logger.Error().Err(err).Bytes("json", s).Msg("Exception executing interaction")
		println(string(debug.Stack()))

		return nil, nil
	}

	sub.MustRegisterCog(plugins.NewWelcomerCog())
	sub.MustRegisterCog(plugins.NewRulesCog())
	sub.MustRegisterCog(plugins.NewBorderwallCog())
	sub.MustRegisterCog(plugins.NewAutoRolesCog())
	sub.MustRegisterCog(plugins.NewLeaverCog())
	sub.MustRegisterCog(plugins.NewFreeRolesCog())
	sub.MustRegisterCog(plugins.NewTimeRolesCog())
	sub.MustRegisterCog(plugins.NewTempChannelsCog())
	sub.MustRegisterCog(plugins.NewMiscellaneousCog())
	sub.MustRegisterCog(plugins.NewDebugCog())
	sub.MustRegisterCog(plugins.NewMembershipCog())
	sub.MustRegisterCog(plugins.NewPrideCog())

	sub.OnAfterInteraction = func(ctx context.Context, sub *subway.Subway, interaction discord.Interaction, resp *discord.InteractionResponse, interactionError error) error {
		var guildID int64
		if interaction.GuildID != nil {
			guildID = int64(*interaction.GuildID)
		}

		var channelID sql.NullInt64
		if interaction.ChannelID != nil {
			channelID.Int64 = int64(*interaction.ChannelID)
			channelID.Valid = true
		}

		commandTree := subway.GetCommandTreeFromContext(ctx)
		command := subway.GetInteractionCommandFromContext(ctx)
		interactionCommandName := strings.Join(append([]string{command.Name}, commandTree...), " ")

		var userID int64
		if interaction.Member != nil {
			userID = int64(interaction.Member.User.ID)
		} else {
			userID = int64(interaction.User.ID)
		}

		var usage *database.ScienceCommandUsages

		err = welcomer.RetryWithFallback(
			func() error {
				usage, err = welcomer.Queries.CreateCommandUsage(ctx, database.CreateCommandUsageParams{
					GuildID:         guildID,
					UserID:          userID,
					ChannelID:       channelID,
					Command:         interactionCommandName,
					Errored:         interactionError != nil,
					ExecutionTimeMs: time.Since(interaction.ID.Time()).Milliseconds(),
				})
				return err
			},
			func() error {
				return welcomer.EnsureGuild(ctx, discord.Snowflake(guildID))
			},
			nil,
		)
		if err != nil {
			welcomer.Logger.Warn().Err(err).Msg("Failed to create command usage")
		}

		if interactionError != nil {
			interactionJSON, err := json.Marshal(interaction)
			if err != nil {
				welcomer.Logger.Warn().Err(err).Msg("Failed to marshal interaction")
			}

			err = welcomer.RetryWithFallback(
				func() error {
					_, err = welcomer.Queries.CreateCommandError(ctx, database.CreateCommandErrorParams{
						CommandUuid: usage.CommandUuid,
						Trace:       interactionError.Error(),
						Data:        pgtype.JSONB{Bytes: interactionJSON, Status: pgtype.Present},
					})
					return err
				},
				func() error {
					return welcomer.EnsureGuild(ctx, discord.Snowflake(guildID))
				},
				nil,
			)
			if err != nil {
				welcomer.Logger.Warn().Err(err).Msg("Failed to create command error")
			}

			welcomer.Logger.Error().
				Str("command", usage.Command).
				Int64("guild_id", usage.GuildID).
				Int64("user_id", usage.UserID).
				Int64("execution", usage.ExecutionTimeMs).
				Err(err).
				Msg("Command executed with errors")
		} else {
			welcomer.Logger.Info().
				Str("command", usage.Command).
				Int64("guild_id", usage.GuildID).
				Int64("user_id", usage.UserID).
				Int64("execution", usage.ExecutionTimeMs).
				Msg("Command executed")
		}

		return nil
	}

	return sub
}
