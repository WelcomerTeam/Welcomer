package plugins

import (
	"errors"

	"github.com/WelcomerTeam/Discord/discord"
	sandwich "github.com/WelcomerTeam/Sandwich/sandwich"
	"github.com/WelcomerTeam/Welcomer/welcomer-core"
	"github.com/WelcomerTeam/Welcomer/welcomer-core/database"
	utils "github.com/WelcomerTeam/Welcomer/welcomer-utils"
	"github.com/jackc/pgx/v4"
)

type AutoRolesCog struct {
	EventHandler *sandwich.Handlers
}

// Assert types.

var (
	_ sandwich.Cog           = (*AutoRolesCog)(nil)
	_ sandwich.CogWithEvents = (*AutoRolesCog)(nil)
)

func NewAutoRolesCog() *AutoRolesCog {
	return &AutoRolesCog{
		EventHandler: sandwich.SetupHandler(nil),
	}
}

func (p *AutoRolesCog) CogInfo() *sandwich.CogInfo {
	return &sandwich.CogInfo{
		Name:        "AutoRoles",
		Description: "Provides the functionality for the 'AutoRoles' feature",
	}
}

func (p *AutoRolesCog) GetEventHandlers() *sandwich.Handlers {
	return p.EventHandler
}

func (p *AutoRolesCog) RegisterCog(bot *sandwich.Bot) error {

	// Trigger OnInvokeAutoRoles when ON_GUILD_MEMBER_ADD event is received.
	p.EventHandler.RegisterOnGuildMemberAddEvent(func(eventCtx *sandwich.EventContext, member discord.GuildMember) error {
		return p.OnInvokeAutoRoles(eventCtx, member)
	})

	return nil
}

func (p *AutoRolesCog) OnInvokeAutoRoles(eventCtx *sandwich.EventContext, member discord.GuildMember) (err error) {
	queries := welcomer.GetQueriesFromContext(eventCtx.Context)

	// Fetch guild settings.

	guildSettingsAutoRoles, err := queries.GetAutoRolesGuildSettings(eventCtx.Context, int64(eventCtx.Guild.ID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			guildSettingsAutoRoles = &database.GuildSettingsAutoroles{
				GuildID:       int64(eventCtx.Guild.ID),
				ToggleEnabled: database.DefaultAutoroles.ToggleEnabled,
				Roles:         database.DefaultAutoroles.Roles,
			}
		} else {
			eventCtx.Logger.Error().Err(err).
				Int64("guild_id", int64(eventCtx.Guild.ID)).
				Msg("Failed to get autorole settings")

			return err
		}
	}

	// Quit if not enabled or no roles are set.
	if !guildSettingsAutoRoles.ToggleEnabled || len(guildSettingsAutoRoles.Roles) == 0 {
		return nil
	}

	assignableRoles, err := welcomer.FilterAssignableRoles(eventCtx.Context, eventCtx.Sandwich.SandwichClient, eventCtx.Logger, int64(*member.GuildID), int64(eventCtx.Identifier.ID), guildSettingsAutoRoles.Roles)
	if err != nil {
		eventCtx.Logger.Error().Err(err).
			Int64("guild_id", int64(*member.GuildID)).
			Msg("Failed to filter assignable roles for autoroles")

		return err
	}

	if len(assignableRoles) == 0 {
		eventCtx.Logger.Warn().
			Int64("guild_id", int64(*member.GuildID)).
			Msg("No roles to assign for autoroles")

		return nil
	}

	err = member.AddRoles(eventCtx.Session, assignableRoles, utils.ToPointer("Automatically assigned with AutoRoles"), true)
	if err != nil {
		eventCtx.Logger.Error().Err(err).
			Int64("guild_id", int64(*member.GuildID)).
			Int64("member_id", int64(member.User.ID)).
			Msg("Failed to add roles to member for autoroles")

		return err
	}

	return nil
}
