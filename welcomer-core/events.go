package welcomer

import (
	"github.com/WelcomerTeam/Discord/discord"
	sandwich "github.com/WelcomerTeam/Sandwich/sandwich"
)

const (
	CustomEventInvokeWelcomer = "WELCOMER_INVOKE_WELCOMER"
)

type OnInvokeWelcomerFuncType func(eventCtx *sandwich.EventContext, member discord.GuildMember) error
