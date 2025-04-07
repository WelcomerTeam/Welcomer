package backend

import (
	"encoding/json"
	"net/http"
	"os"
	"sync"
	"time"

	sandwich "github.com/WelcomerTeam/Sandwich-Daemon/structs"
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

	url := os.Getenv("SANDWICH_STATUS_ENDPOINT")

	req, err := http.NewRequestWithContext(ctx.Request.Context(), http.MethodGet, url, nil)
	if req == nil || err != nil {
		welcomer.Logger.Error().Err(err).Str("url", url).Msg("Failed to create request for status API")

		ctx.JSON(http.StatusInternalServerError, BaseResponse{
			Ok: false,
		})

		return
	}

	resp, err := http.DefaultClient.Do(req)
	if resp == nil || err != nil {
		welcomer.Logger.Error().Err(err).Str("url", url).Msg("Failed to fetch status from API")

		ctx.JSON(http.StatusInternalServerError, BaseResponse{
			Ok: false,
		})

		return
	}

	defer resp.Body.Close()

	var baseResponse sandwich.BaseRestResponse

	err = json.NewDecoder(resp.Body).Decode(&baseResponse)
	if err != nil {
		welcomer.Logger.Error().Err(err).Msg("Failed to decode status base response")

		ctx.JSON(http.StatusInternalServerError, BaseResponse{
			Ok: false,
		})

		return
	}

	statusResponseJson, err := json.Marshal(baseResponse.Data)
	if err != nil {
		welcomer.Logger.Error().Err(err).Msg("Failed to marshal base response data")

		ctx.JSON(http.StatusInternalServerError, BaseResponse{
			Ok: false,
		})

		return
	}

	var sandwichStatusResponse sandwich.StatusEndpointResponse

	err = json.Unmarshal(statusResponseJson, &sandwichStatusResponse)
	if err != nil {
		welcomer.Logger.Error().Err(err).Msg("Failed to unmarshal status response")

		ctx.JSON(http.StatusInternalServerError, BaseResponse{
			Ok: false,
		})

		return
	}

	newManagers := make([]GetStatusResponseManager, 0, len(sandwichStatusResponse.Managers))

	for _, manager := range sandwichStatusResponse.Managers {
		newShards := make([]GetStatusResponseShard, 0)

		if len(manager.ShardGroups) > 0 {
			shardGroup := manager.ShardGroups[len(manager.ShardGroups)-1]

			for _, shard := range shardGroup.Shards {
				newShards = append(newShards, GetStatusResponseShard{
					ShardID: shard[0],
					Status:  shard[1],
					Latency: shard[2],
					Guilds:  shard[3],
					Uptime:  shard[4],
				})
			}
		}

		newManagers = append(newManagers, GetStatusResponseManager{
			Name:   manager.DisplayName,
			Shards: newShards,
		})
	}

	statusResponseMu.Lock()
	statusResponse.UpdatedAt = time.Now()
	statusResponse.Managers = newManagers
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
