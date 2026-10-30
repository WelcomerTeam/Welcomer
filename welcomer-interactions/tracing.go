package welcomer

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"github.com/WelcomerTeam/Discord/discord"
	"github.com/WelcomerTeam/Welcomer/welcomer-core"
	"github.com/WelcomerTeam/Welcomer/welcomer-core/telemetry"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

const maxTraceInteractionBody = 1024 * 1024

type traceBody struct {
	io.Reader
	io.Closer
}

type traceCapture struct {
	bytes.Buffer
	truncated bool
}

func (c *traceCapture) Write(p []byte) (int, error) {
	n := len(p)

	remaining := maxTraceInteractionBody - c.Len()
	if n > remaining {
		c.truncated = true
		p = p[:remaining]
	}

	_, _ = c.Buffer.Write(p)

	return n, nil
}

type traceResponseWriter struct {
	http.ResponseWriter
	status int
}

func (w *traceResponseWriter) WriteHeader(status int) {
	if w.status == 0 {
		w.status = status
		w.ResponseWriter.WriteHeader(status)
	}
}

func (w *traceResponseWriter) Write(p []byte) (int, error) {
	if w.status == 0 {
		w.WriteHeader(http.StatusOK)
	}

	return w.ResponseWriter.Write(p)
}

func (w *traceResponseWriter) Unwrap() http.ResponseWriter {
	return w.ResponseWriter
}

// TraceHTTP traces requests without collecting URLs, headers, or interaction payloads.
func TraceHTTP(mux *http.ServeMux) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, route := mux.Handler(r)
		if route == "" {
			route = "unmatched"
		}

		ctx, span := telemetry.StartServerSpan(r, r.Method+" "+route)
		span.SetAttributes(attribute.String("http.route", route))

		r = r.WithContext(ctx)

		var captured traceCapture
		if span.IsRecording() && route == "/" && r.Body != nil {
			r.Body = traceBody{Reader: io.TeeReader(r.Body, &captured), Closer: r.Body}
		}

		response := &traceResponseWriter{ResponseWriter: w}

		defer func() {
			if recovered := recover(); recovered != nil {
				telemetry.FinishServerSpan(span, http.StatusInternalServerError)
				panic(recovered)
			}

			status := response.status
			if status == 0 {
				status = http.StatusOK
			}

			if span.IsRecording() && route == "/" && status < http.StatusBadRequest {
				if captured.truncated {
					welcomer.Logger.Warn().Msg("Interaction body exceeds tracing metadata limit")
				} else {
					setInteractionAttributes(span, captured.Bytes())
				}
			}

			telemetry.FinishServerSpan(span, status)
		}()

		mux.ServeHTTP(response, r)
	})
}

func setInteractionAttributes(span trace.Span, body []byte) {
	var interaction discord.Interaction
	if err := json.Unmarshal(body, &interaction); err != nil {
		welcomer.Logger.Warn().Err(err).Msg("Failed to parse interaction tracing metadata")
		return
	}

	if user := interaction.GetUser(); user != nil {
		span.SetAttributes(attribute.String("discord.user.id", user.ID.String()))
	}

	if interaction.GuildID != nil {
		span.SetAttributes(attribute.String("discord.guild.id", interaction.GuildID.String()))
	}

	if interaction.Data != nil && (interaction.Type == discord.InteractionTypeApplicationCommand ||
		interaction.Type == discord.InteractionTypeApplicationCommandAutocomplete) {
		command := []string{interaction.Data.Name}

		options := interaction.Data.Options
		for len(options) > 0 {
			found := false

			for _, option := range options {
				if option.Type == discord.ApplicationCommandOptionTypeSubCommand ||
					option.Type == discord.ApplicationCommandOptionTypeSubCommandGroup {
					command = append(command, option.Name)
					options = option.Options
					found = true

					break
				}
			}

			if !found {
				break
			}
		}

		span.SetAttributes(attribute.String("discord.command", strings.Join(command, " ")))
	}
}
