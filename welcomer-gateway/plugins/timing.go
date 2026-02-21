package plugins

import (
	"time"

	"github.com/WelcomerTeam/Welcomer/welcomer-core"
)

func notifyTiming(startTime time.Time, shard [3]int32, name string) {
	if time.Now().Sub(startTime) > time.Millisecond*10 {
		welcomer.Logger.Warn().Int32("shard", shard[1]).Str("name", name).Dur("duration", time.Now().Sub(startTime)).Msg("Timing warning: operation took longer than 10ms")
	}
}
