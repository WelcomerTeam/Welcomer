package welcomer

import (
	"context"
	"net/url"

	"github.com/WelcomerTeam/Discord/discord"
	subway "github.com/WelcomerTeam/Subway/subway"
)

type ContextKey int

const (
	ManagerNameContextKey ContextKey = iota
	SessionContextKey
)

func AddManagerNameToContext(ctx context.Context, managerName string) context.Context {
	return context.WithValue(ctx, ManagerNameContextKey, managerName)
}

func AddSessionToContext(ctx context.Context, session *discord.Session) context.Context {
	return context.WithValue(ctx, SessionContextKey, session)
}

func GetSessionFromContext(ctx context.Context) (*discord.Session, bool) {
	s, ok := ctx.Value(SessionContextKey).(*discord.Session)

	return s, ok
}

func TryGetURLFromContext(ctx context.Context) url.URL {
	defer func() {
		if r := recover(); r != nil {
			return
		}
	}()

	url := subway.GetURLFromContext(ctx)

	return url
}

func GetManagerNameFromContext(ctx context.Context) string {
	url := TryGetURLFromContext(ctx)
	query := url.Query()

	manager := query.Get("manager")
	if manager != "" {
		return manager
	}

	manager, _ = ctx.Value(ManagerNameContextKey).(string)
	if manager != "" {
		return manager
	}

	return DefaultManagerName
}
