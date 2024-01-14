package welcomer

import (
	"github.com/WelcomerTeam/Discord/discord"
	sandwich "github.com/WelcomerTeam/Sandwich/sandwich"
)

const (
	CustomEventInvokeWelcomer     = "WELCOMER_INVOKE_WELCOMER"
	CustomEventInvokeTempChannels = "WELCOMER_INVOKE_TEMPCHANNELS"
)

type OnInvokeWelcomerFuncType func(eventCtx *sandwich.EventContext, member CustomEventInvokeWelcomerStructure) error

type CustomEventInvokeWelcomerStructure struct {
	Interaction *discord.Interaction
	Member      *discord.GuildMember
}

type OnInvokeTempChannelsFuncType func(eventCtx *sandwich.EventContext, member CustomEventInvokeTempChannelsStructure) error

type CustomEventInvokeTempChannelsStructure struct {
	Interaction *discord.Interaction
	Member      *discord.GuildMember
}
