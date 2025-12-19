package welcomer

import (
	"github.com/WelcomerTeam/Discord/discord"
	sandwich "github.com/WelcomerTeam/Sandwich/sandwich"
)

const (
	CustomEventInvokeWelcomer = "WELCOMER_INVOKE_WELCOMER"
	CustomEventInvokeLeaver   = "WELCOMER_INVOKE_LEAVER"

	CustomEventInvokeTempChannels       = "WELCOMER_INVOKE_TEMPCHANNELS"
	CustomEventInvokeTempChannelsRemove = "WELCOMER_INVOKE_TEMPCHANNELS_REMOVE"

	CustomEventInvokeBorderwall           = "WELCOMER_INVOKE_BORDERWALL"
	CustomEventInvokeBorderwallCompletion = "WELCOMER_INVOKE_BORDERWALL_COMPLETION"
)

type OnInvokeWelcomerFuncType func(eventCtx *sandwich.EventContext, member CustomEventInvokeWelcomerStructure) error

type CustomEventInvokeWelcomerStructure struct {
	Interaction  *discord.Interaction
	Member       discord.GuildMember
	IgnoreDedupe bool
}

type OnInvokeLeaverFuncType func(eventCtx *sandwich.EventContext, member CustomEventInvokeLeaverStructure) error

type CustomEventInvokeLeaverStructure struct {
	Interaction *discord.Interaction
	User        discord.User
	GuildID     discord.Snowflake
}

type OnInvokeTempChannelsFuncType func(eventCtx *sandwich.EventContext, member CustomEventInvokeTempChannelsStructure) error

type CustomEventInvokeTempChannelsStructure struct {
	Interaction *discord.Interaction
	Member      discord.GuildMember
}

type OnInvokeTempChannelsRemoveFuncType func(eventCtx *sandwich.EventContext, member CustomEventInvokeTempChannelsRemoveStructure) error

type CustomEventInvokeTempChannelsRemoveStructure struct {
	Interaction *discord.Interaction
	Member      discord.GuildMember
}

type OnInvokeBorderwallCompletionFuncType func(eventCtx *sandwich.EventContext, member CustomEventInvokeBorderwallCompletionStructure) error

type CustomEventInvokeBorderwallCompletionStructure struct {
	Member discord.GuildMember
}

type OnInvokeBorderwallFuncType func(eventCtx *sandwich.EventContext, member CustomEventInvokeBorderwallStructure) error

type CustomEventInvokeBorderwallStructure struct {
	Member discord.GuildMember
}
