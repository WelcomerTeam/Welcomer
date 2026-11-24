package backend

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/WelcomerTeam/Discord/discord"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	"go.opentelemetry.io/otel/trace"
	"golang.org/x/oauth2"
)

func TestTraceRequestsWithAuthenticatedIDs(t *testing.T) {
	exporter := tracetest.NewInMemoryExporter()
	provider := sdktrace.NewTracerProvider(sdktrace.WithSyncer(exporter))
	oldProvider, oldPropagator := otel.GetTracerProvider(), otel.GetTextMapPropagator()

	otel.SetTracerProvider(provider)
	otel.SetTextMapPropagator(propagation.TraceContext{})

	defer otel.SetTracerProvider(oldProvider)
	defer otel.SetTextMapPropagator(oldPropagator)

	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.ContextWithFallback = true
	router.Use(traceRequests())
	router.Use(
		sessions.Sessions("session", cookie.NewStore([]byte("01234567890123456789012345678901"))),
	)
	router.GET("/login", func(c *gin.Context) {
		session := sessions.Default(c)
		SetUserSession(session, SessionUser{ID: discord.Snowflake(123)})
		SetTokenSession(session, oauth2.Token{AccessToken: "secret"})

		if err := session.Save(); err != nil {
			t.Fatal(err)
		}

		c.Status(http.StatusNoContent)
	})
	router.GET("/guild/:guildID", func(c *gin.Context) {
		requireOAuthAuthorization(c, func(c *gin.Context) {
			requireGuildIDKey(c, func(c *gin.Context) {
				if !trace.SpanFromContext(c).SpanContext().IsValid() {
					t.Error("Gin context did not propagate the request span")
				}

				c.Status(http.StatusNoContent)
			})
		})
	})

	login := httptest.NewRecorder()
	router.ServeHTTP(login, httptest.NewRequest(http.MethodGet, "/login", nil))
	request := httptest.NewRequest(http.MethodGet, "/guild/456?token=secret", nil)
	request.AddCookie(login.Result().Cookies()[0])
	request.Header.Set("traceparent", "00-0123456789abcdef0123456789abcdef-0123456789abcdef-01")

	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)

	if response.Code != http.StatusNoContent {
		t.Fatalf("status = %d", response.Code)
	}

	spans := exporter.GetSpans()

	span := spans[len(spans)-1]
	if span.Name != "GET /guild/:guildID" || span.Parent.SpanID().String() != "0123456789abcdef" {
		t.Fatalf("unexpected span name or parent: %s %s", span.Name, span.Parent.SpanID())
	}

	attrs := make(map[string]string)
	for _, attr := range span.Attributes {
		attrs[string(attr.Key)] = attr.Value.Emit()
	}

	if attrs["discord.user.id"] != "123" || attrs["discord.guild.id"] != "456" {
		t.Fatalf("missing IDs: %v", attrs)
	}

	for _, value := range attrs {
		if value == "secret" || value == "/guild/456?token=secret" {
			t.Fatal("secret or raw URL recorded in span")
		}
	}
}
