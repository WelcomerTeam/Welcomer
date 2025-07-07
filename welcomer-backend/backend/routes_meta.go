package backend

import (
	"encoding/json"
	"net/http"
	"slices"
	"sync"
	"time"

	sandwich_protobuf "github.com/WelcomerTeam/Sandwich-Daemon/proto"
	"github.com/WelcomerTeam/Welcomer/welcomer-core"
	"github.com/gin-gonic/gin"
	"go.uber.org/atomic"
)

type GetStatusResponse struct {
	UpdatedAt time.Time                  `json:"updated_at"`
	Managers  []GetStatusResponseManager `json:"managers"`
}

type GetStatusResponseManager struct {
	Name   string                   `json:"name"`
	Shards []GetStatusResponseShard `json:"shards"`
}

type GetStatusResponseShard struct {
	ShardID int `json:"shard_id"`
	Status  int `json:"status"`
	Latency int `json:"latency"`
	Guilds  int `json:"guilds"`
	Uptime  int `json:"uptime"`
}

var statusLastFetchedAt *atomic.Time = atomic.NewTime(time.Time{})

var statusResponse GetStatusResponse

// How long to keep the status response in memory.
var statusResponseLifetime time.Duration = 10 * time.Second

var statusResponseMu sync.RWMutex

// Route GET /api/status.
func getStatus(ctx *gin.Context) {
	if time.Since(statusLastFetchedAt.Load()) < statusResponseLifetime {
		statusResponseMu.RLock()
		defer statusResponseMu.RUnlock()

		ctx.JSON(http.StatusOK, BaseResponse{
			Ok:   true,
			Data: statusResponse,
		})

		return
	}

	applicationsPb, err := welcomer.SandwichClient.FetchApplication(ctx, &sandwich_protobuf.ApplicationIdentifier{})
	if err != nil {
		statusResponseMu.RLock()
		defer statusResponseMu.RUnlock()

		ctx.JSON(http.StatusOK, BaseResponse{
			Ok:   true,
			Data: statusResponse,
		})

		return
	}

	applications := applicationsPb.GetApplications()
	newApplications := make([]GetStatusResponseManager, 0, len(applications))

	for _, application := range applications {
		applicationValues := welcomer.ApplicationValues{}

		// Unmarshal the application values.
		err = json.Unmarshal(application.GetValues(), &applicationValues)
		if err != nil {
			welcomer.Logger.Error().Err(err).Msg("Failed to unmarshal application values")
		}

		// if applicationValues.IsCustomBot {
		// 	// Skip custom bots
		// 	continue
		// }

		newShards := make([]GetStatusResponseShard, 0)

		for _, shard := range application.Shards {
			newShards = append(newShards, GetStatusResponseShard{
				ShardID: int(shard.GetId()),
				Status:  int(shard.GetStatus()),
				Latency: int(shard.GetGatewayLatency()),
				Guilds:  int(shard.GetGuilds()),
				Uptime:  int(time.Since(time.Unix(shard.GetStartedAt(), 0)).Seconds()),
			})
		}

		slices.SortFunc(newShards, func(a, b GetStatusResponseShard) int {
			return a.ShardID - b.ShardID
		})

		newApplications = append(newApplications, GetStatusResponseManager{
			Name:   application.GetDisplayName(),
			Shards: newShards,
		})
	}

	statusResponseMu.Lock()
	statusResponse.UpdatedAt = time.Now()
	statusResponse.Managers = newApplications
	statusResponseMu.Unlock()

	statusLastFetchedAt.Store(statusResponse.UpdatedAt)

	welcomer.Logger.Info().Msg("Fetched status from API")

	ctx.JSON(http.StatusOK, BaseResponse{
		Ok:   true,
		Data: statusResponse,
	})
}

func registerMetaRoutes(g *gin.Engine) {
	g.GET("/api/status", getStatus)
}
