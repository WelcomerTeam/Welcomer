package welcomer

import (
	"github.com/WelcomerTeam/Discord/discord"
	sandwich "github.com/WelcomerTeam/Sandwich/sandwich"
	"github.com/gofrs/uuid"
)

const (
	CustomEventInvokeWelcomer = "WELCOMER_INVOKE_WELCOMER"
	CustomEventInvokeLeaver   = "WELCOMER_INVOKE_LEAVER"

	CustomEventInvokeTempChannels       = "WELCOMER_INVOKE_TEMPCHANNELS"
	CustomEventInvokeTempChannelsRemove = "WELCOMER_INVOKE_TEMPCHANNELS_REMOVE"

	CustomEventInvokeBorderwall           = "WELCOMER_INVOKE_BORDERWALL"
	CustomEventInvokeBorderwallCompletion = "WELCOMER_INVOKE_BORDERWALL_COMPLETION"

	CustomEventInvokeReactionRoles = "WELCOMER_INVOKE_REACTION_ROLES"
)

type OnInvokeWelcomerFuncType func(eventCtx *sandwich.EventContext, event CustomEventInvokeWelcomerStructure) error

type CustomEventInvokeWelcomerStructure struct {
	Interaction  *discord.Interaction
	Member       discord.GuildMember
	IgnoreDedupe bool
}

type OnInvokeLeaverFuncType func(eventCtx *sandwich.EventContext, event CustomEventInvokeLeaverStructure) error

type CustomEventInvokeLeaverStructure struct {
	Interaction *discord.Interaction
	User        discord.User
	GuildID     discord.Snowflake
}

type OnInvokeTempChannelsFuncType func(eventCtx *sandwich.EventContext, event CustomEventInvokeTempChannelsStructure) error

type CustomEventInvokeTempChannelsStructure struct {
	Interaction *discord.Interaction
	Member      discord.GuildMember
}

type OnInvokeTempChannelsRemoveFuncType func(eventCtx *sandwich.EventContext, event CustomEventInvokeTempChannelsRemoveStructure) error

type CustomEventInvokeTempChannelsRemoveStructure struct {
	Interaction *discord.Interaction
	Member      discord.GuildMember
}

type OnInvokeBorderwallCompletionFuncType func(eventCtx *sandwich.EventContext, event CustomEventInvokeBorderwallCompletionStructure) error

type CustomEventInvokeBorderwallCompletionStructure struct {
	Member discord.GuildMember
}

type OnInvokeBorderwallFuncType func(eventCtx *sandwich.EventContext, event CustomEventInvokeBorderwallStructure) error

type CustomEventInvokeBorderwallStructure struct {
	Member discord.GuildMember
}

type OnInvokeReactionRolesFuncType func(eventCtx *sandwich.EventContext, event CustomEventInvokeReactionRolesStructure) error

type CustomEventInvokeReactionRolesStructure struct {
	Interaction *discord.Interaction
	Member      *discord.GuildMember

	ReactionRoleUUID uuid.UUID
	RoleID           discord.Snowflake
	Assign           *bool
}
