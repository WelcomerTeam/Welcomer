package backend

import (
	"github.com/WelcomerTeam/Welcomer/welcomer-core/telemetry"
	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel/attribute"
)

func traceRequests() gin.HandlerFunc {
	return func(c *gin.Context) {
		route := c.FullPath()
		if route == "" {
			route = "unmatched"
		}

		ctx, span := telemetry.StartServerSpan(c.Request, c.Request.Method+" "+route)
		c.Request = c.Request.WithContext(ctx)

		span.SetAttributes(attribute.String("http.route", route))
		defer func() {
			telemetry.FinishServerSpan(span, c.Writer.Status())
		}()

		c.Next()
	}
}
