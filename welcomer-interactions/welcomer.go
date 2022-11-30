package welcomer

import (
	"context"
	"fmt"
	"runtime/debug"

	"github.com/WelcomerTeam/Discord/discord"
	subway "github.com/WelcomerTeam/Subway/subway"
	plugins "github.com/WelcomerTeam/Welcomer/welcomer-interactions/plugins"
	jsoniter "github.com/json-iterator/go"
)

func NewWelcomer(ctx context.Context, options subway.SubwayOptions) *subway.Subway {
	sub, err := subway.NewSubway(ctx, options)
	if err != nil {
		panic(fmt.Errorf("failed to create subway client. sub.NewClient(%w): %w", options, err))
	}

	sub.Commands.ErrorHandler = func(ctx context.Context, sub *subway.Subway, interaction discord.Interaction, err error) (*discord.InteractionResponse, error) {
		s, _ := jsoniter.MarshalToString(interaction)

		sub.Logger.Error().Err(err).Str("json", s).Msg("Exception executing interaction")
		println(string(debug.Stack()))

		return nil, nil
	}

	sub.MustRegisterCog(plugins.NewGeneralCog())

	return sub
}