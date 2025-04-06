package welcomer

import (
	"context"
	"net/url"

	subway "github.com/WelcomerTeam/Subway/subway"
)

type InteractionsContextKey int

const (
	ManagerNameKey InteractionsContextKey = iota
)

// ManagerName context handler.
func AddManagerNameToContext(ctx context.Context, v string) context.Context {
	return context.WithValue(ctx, ManagerNameKey, v)
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

	value, _ := ctx.Value(ManagerNameKey).(string)

	return value
}
