package plugins

import (
	"context"

	"github.com/WelcomerTeam/Discord/discord"
	subway "github.com/WelcomerTeam/Subway/subway"
	"github.com/WelcomerTeam/Welcomer/welcomer-core"
)

const (
	pollsSetupMenuTitleKey        = "title"
	pollsSetupMenuDescriptionKey  = "description"
	pollsSetupMenuAccentColorKey  = "accent_color"
	pollsSetupMenuThumbnailURLKey = "thumbnail_url"

	pollsSetupMenuOptionsKey                 = "options"
	pollsSetupMenuDurationKey                = "duration"
	pollsSetupMenuMaximumOptionsKey          = "maximum_options"
	pollsSetupMenuToggleAnononymousVotingKey = "toggle_anonymous_voting"

	pollsSetupMenuRolesAllowedKey         = "roles_allowed"
	pollsSetupMenuRolesAllowedIncludedKey = "roles_allowed_included"
	pollsSetupMenuRolesAllowedExcludedKey = "roles_allowed_excluded"

	pollsSetupMenuManageResubmissionsKey = "manage_resubmissions"
	pollsSetupMenuNoResubmissionsKey     = "no_resubmissions"
	pollsSetupMenuAllowAdditionsOnlyKey  = "allow_additions_only"
	pollsSetupMenuAllowResubmissionsKey  = "allow_resubmissions"

	pollsSetupMenuShowResultsKey                   = "show_results"
	pollsSetupMenuResultsAlwaysVisibleKey          = "results_always_visible"
	pollsSetupMenuResultsVisibleAfterVotingKey     = "results_visible_after_voting"
	pollsSetupMenuResultsVisibleAfterVotingEndsKey = "results_visible_after_voting_ends"

	pollsSetupMenuPingKey                    = "ping"
	pollsSetupMenuPingEveryoneKey            = "ping_everyone"
	pollsSetupMenuPingHereKey                = "ping_here"
	pollsSetupMenuPingRolesAllowedToEnterKey = "ping_roles_allowed_to_enter"
	pollsSetupMenuPingAdditionalRolesKey     = "additional_roles_to_ping"
)

func NewPollsCog() *PollsCog {
	return &PollsCog{
		InteractionCommands: subway.SetupInteractionCommandable(&subway.InteractionCommandable{}),
	}
}

type PollsCog struct {
	InteractionCommands *subway.InteractionCommandable
}

// Assert types.

var (
	_ subway.Cog                        = (*PollsCog)(nil)
	_ subway.CogWithInteractionCommands = (*PollsCog)(nil)
)

func (cog *PollsCog) CogInfo() *subway.CogInfo {
	return &subway.CogInfo{
		Name:        "Polls",
		Description: "Provides the functionality for the 'Polls' feature",
	}
}

func (cog *PollsCog) GetInteractionCommandable() *subway.InteractionCommandable {
	return cog.InteractionCommands
}

func (cog *PollsCog) RegisterCog(sub *subway.Subway) error {
	pollsGroup := subway.NewSubcommandGroup(
		"polls",
		"Poll commands",
	)

	pollsGroup.MustAddInteractionCommand(&subway.InteractionCommandable{
		Name:        "new",
		Description: "Makes a new poll",

		Type: subway.InteractionCommandableTypeSubcommand,

		DMPermission:            new(false),
		DefaultMemberPermission: new(discord.Int64(welcomer.PermissionElevated)),

		Handler: func(ctx context.Context, sub *subway.Subway, interaction discord.Interaction) (*discord.InteractionResponse, error) {
			return welcomer.RequireGuildElevation(sub, interaction, func() (*discord.InteractionResponse, error) {
				return nil, nil
			})
		},
	})

	return nil
}
