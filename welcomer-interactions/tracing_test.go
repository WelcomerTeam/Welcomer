package welcomer

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/hex"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	subway "github.com/WelcomerTeam/Subway/subway"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
)

func TestInteractionTracing(t *testing.T) {
	exporter := tracetest.NewInMemoryExporter()
	provider := sdktrace.NewTracerProvider(sdktrace.WithSyncer(exporter))
	oldProvider, oldPropagator := otel.GetTracerProvider(), otel.GetTextMapPropagator()

	otel.SetTracerProvider(provider)
	otel.SetTextMapPropagator(propagation.TraceContext{})

	defer otel.SetTracerProvider(oldProvider)
	defer otel.SetTextMapPropagator(oldPropagator)

	body := `{"type":2,"token":"secret-token","guild_id":"456","member":{"user":{"id":"123"}},"data":{"name":"settings","options":[{"type":2,"name":"group","options":[{"type":1,"name":"show","options":[{"type":3,"name":"value","value":"secret-value"}]}]}]}}`
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		got, err := io.ReadAll(r.Body)
		if err != nil || string(got) != body {
			t.Errorf("handler received altered body: %v", err)
		}

		w.WriteHeader(http.StatusNoContent)
	})

	request := httptest.NewRequest(
		http.MethodPost,
		"/?manager=secret-manager",
		strings.NewReader(body),
	)
	request.Header.Set("traceparent", "00-0123456789abcdef0123456789abcdef-0123456789abcdef-01")

	response := httptest.NewRecorder()
	TraceHTTP(mux).ServeHTTP(response, request)

	if response.Code != http.StatusNoContent {
		t.Fatalf("status = %d", response.Code)
	}

	spans := exporter.GetSpans()
	if len(spans) != 1 {
		t.Fatalf("got %d spans", len(spans))
	}

	span := spans[0]
	if span.Parent.SpanID().String() != "0123456789abcdef" {
		t.Fatalf("unexpected parent %s", span.Parent.SpanID())
	}

	attrs := make(map[string]string)
	for _, attr := range span.Attributes {
		attrs[string(attr.Key)] = attr.Value.Emit()
	}

	if attrs["discord.user.id"] != "123" || attrs["discord.guild.id"] != "456" ||
		attrs["discord.command"] != "settings group show" {
		t.Fatalf("unexpected attributes: %v", attrs)
	}

	for _, value := range attrs {
		if strings.Contains(value, "secret") {
			t.Fatalf("secret recorded in span: %v", attrs)
		}
	}
}

func TestRejectedInteractionHasNoMetadata(t *testing.T) {
	exporter := tracetest.NewInMemoryExporter()
	provider := sdktrace.NewTracerProvider(sdktrace.WithSyncer(exporter))
	oldProvider := otel.GetTracerProvider()

	otel.SetTracerProvider(provider)
	defer otel.SetTracerProvider(oldProvider)

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.ReadAll(r.Body)

		http.Error(w, "invalid signature", http.StatusBadRequest)
	})

	body := `{"type":2,"user":{"id":"123"},"data":{"name":"test"}}`
	TraceHTTP(
		mux,
	).ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body)))

	for _, attr := range exporter.GetSpans()[0].Attributes {
		if strings.HasPrefix(string(attr.Key), "discord.") {
			t.Fatalf("rejected interaction has metadata: %v", attr)
		}
	}
}

type testPublicKeys []ed25519.PublicKey

func (keys testPublicKeys) GetPublicKeys(*http.Request) []ed25519.PublicKey {
	return keys
}

func TestTraceActualSubwaySignatureValidation(t *testing.T) {
	exporter := tracetest.NewInMemoryExporter()
	provider := sdktrace.NewTracerProvider(sdktrace.WithSyncer(exporter))
	oldProvider := otel.GetTracerProvider()

	otel.SetTracerProvider(provider)
	defer otel.SetTracerProvider(oldProvider)

	publicKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	app, err := subway.NewSubway(ctx, subway.SubwayOptions{
		Logger: slog.New(
			slog.NewTextHandler(io.Discard, nil),
		),
		PublicKeysHandler: testPublicKeys{publicKey},
	})
	if err != nil {
		t.Fatal(err)
	}

	mux := app.PrepareMux("", nil)
	body := `{"type":2,"guild_id":"456","member":{"user":{"id":"123"}},"data":{"name":"missing-command"}}`
	request := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	request.Header.Set("X-Signature-Timestamp", "123")
	request.Header.Set(
		"X-Signature-Ed25519",
		hex.EncodeToString(ed25519.Sign(privateKey, []byte("123"+body))),
	)

	response := httptest.NewRecorder()
	TraceHTTP(mux).ServeHTTP(response, request)

	if response.Code != http.StatusNoContent {
		t.Fatalf("valid signed request status = %d", response.Code)
	}

	attrs := make(map[string]string)
	for _, attr := range exporter.GetSpans()[0].Attributes {
		attrs[string(attr.Key)] = attr.Value.Emit()
	}

	if attrs["discord.user.id"] != "123" || attrs["discord.guild.id"] != "456" ||
		attrs["discord.command"] != "missing-command" {
		t.Fatalf("missing signed interaction metadata: %v", attrs)
	}
}

func TestDirectMessageMetadata(t *testing.T) {
	exporter := tracetest.NewInMemoryExporter()
	provider := sdktrace.NewTracerProvider(sdktrace.WithSyncer(exporter))
	_, span := provider.Tracer("test").Start(context.Background(), "interaction")
	setInteractionAttributes(
		span,
		[]byte(`{"type":3,"user":{"id":"789"},"data":{"custom_id":"private-value"}}`),
	)
	span.End()

	attrs := make(map[string]string)
	for _, attr := range exporter.GetSpans()[0].Attributes {
		attrs[string(attr.Key)] = attr.Value.Emit()
	}

	if attrs["discord.user.id"] != "789" || attrs["discord.guild.id"] != "" ||
		attrs["discord.command"] != "" {
		t.Fatalf("unexpected direct message metadata: %v", attrs)
	}
}

func TestOversizedInteractionMetadataIsOmitted(t *testing.T) {
	exporter := tracetest.NewInMemoryExporter()
	provider := sdktrace.NewTracerProvider(sdktrace.WithSyncer(exporter))
	oldProvider := otel.GetTracerProvider()

	otel.SetTracerProvider(provider)
	defer otel.SetTracerProvider(oldProvider)

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.Copy(io.Discard, r.Body)

		w.WriteHeader(http.StatusNoContent)
	})

	body := `{"type":2,"user":{"id":"123"},"data":{"name":"test"},"padding":"` +
		strings.Repeat("x", maxTraceInteractionBody) + `"}`
	response := httptest.NewRecorder()
	TraceHTTP(
		mux,
	).ServeHTTP(response, httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body)))

	if response.Code != http.StatusNoContent {
		t.Fatalf("status = %d", response.Code)
	}

	for _, attr := range exporter.GetSpans()[0].Attributes {
		if strings.HasPrefix(string(attr.Key), "discord.") {
			t.Fatalf("oversized interaction has metadata: %v", attr)
		}
	}
}
