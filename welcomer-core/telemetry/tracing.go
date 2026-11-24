package telemetry

import (
	"context"
	"errors"
	"fmt"
	"math"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
)

func Enabled() bool {
	if strings.EqualFold(os.Getenv("OTEL_SDK_DISABLED"), "true") {
		return false
	}
	exporterName := os.Getenv("OTEL_TRACES_EXPORTER")
	return exporterName != "none" && !(exporterName == "" &&
		os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT") == "" &&
		os.Getenv("OTEL_EXPORTER_OTLP_TRACES_ENDPOINT") == "")
}

// Setup enables OTLP tracing when an exporter or endpoint is explicitly configured.
func Setup(ctx context.Context, serviceName string) (func(context.Context) error, error) {
	if !Enabled() {
		return func(context.Context) error { return nil }, nil
	}
	exporterName := os.Getenv("OTEL_TRACES_EXPORTER")
	if exporterName != "" && exporterName != "otlp" {
		return nil, fmt.Errorf("unsupported OTEL_TRACES_EXPORTER %q", exporterName)
	}

	var sampler sdktrace.Sampler
	percentValue := strings.TrimSpace(os.Getenv("WELCOMER_TRACE_SAMPLE_PERCENT"))
	if percentValue != "" {
		percent, err := strconv.ParseFloat(percentValue, 64)
		if err != nil || math.IsNaN(percent) || math.IsInf(percent, 0) || percent < 0 ||
			percent > 100 {
			return nil, fmt.Errorf(
				"WELCOMER_TRACE_SAMPLE_PERCENT must be a number between 0 and 100, got %q",
				percentValue,
			)
		}
		sampler = sdktrace.TraceIDRatioBased(percent / 100)
	}

	protocol := os.Getenv("OTEL_EXPORTER_OTLP_TRACES_PROTOCOL")
	if protocol == "" {
		protocol = os.Getenv("OTEL_EXPORTER_OTLP_PROTOCOL")
	}
	if protocol == "" {
		protocol = "grpc"
	}

	var exporter sdktrace.SpanExporter
	var err error
	switch protocol {
	case "grpc":
		exporter, err = otlptracegrpc.New(ctx)
	case "http/protobuf":
		exporter, err = otlptracehttp.New(ctx)
	default:
		return nil, fmt.Errorf("unsupported OTLP traces protocol %q", protocol)
	}
	if err != nil {
		return nil, fmt.Errorf("create OTLP trace exporter: %w", err)
	}

	detected, err := resource.New(ctx, resource.WithTelemetrySDK(), resource.WithFromEnv())
	if err != nil {
		return nil, errors.Join(
			fmt.Errorf("detect telemetry resource: %w", err),
			exporter.Shutdown(ctx),
		)
	}
	res, err := resource.Merge(
		resource.NewSchemaless(attribute.String("service.name", serviceName)),
		detected,
	)
	if err != nil {
		return nil, errors.Join(
			fmt.Errorf("merge telemetry resource: %w", err),
			exporter.Shutdown(ctx),
		)
	}

	options := []sdktrace.TracerProviderOption{
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
	}
	if sampler != nil {
		options = append(options, sdktrace.WithSampler(sampler))
	}
	provider := sdktrace.NewTracerProvider(options...)
	otel.SetTracerProvider(provider)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{}, propagation.Baggage{},
	))

	return provider.Shutdown, nil
}

// StartServerSpan extracts the W3C trace context without recording request URLs or headers.
func StartServerSpan(r *http.Request, name string) (context.Context, trace.Span) {
	ctx := otel.GetTextMapPropagator().Extract(r.Context(), propagation.HeaderCarrier(r.Header))
	return otel.Tracer("github.com/WelcomerTeam/Welcomer").Start(
		ctx, name,
		trace.WithSpanKind(trace.SpanKindServer),
		trace.WithAttributes(attribute.String("http.request.method", r.Method)),
	)
}

func FinishServerSpan(span trace.Span, status int) {
	span.SetAttributes(attribute.Int("http.response.status_code", status))
	if status >= http.StatusInternalServerError {
		span.SetStatus(codes.Error, http.StatusText(status))
	}
	span.End()
}

// Serve stops accepting requests and waits for active handlers when ctx is cancelled.
func Serve(ctx context.Context, server *http.Server) error {
	return serve(ctx, server, server.ListenAndServe)
}

func serve(ctx context.Context, server *http.Server, run func() error) error {
	done := make(chan struct{})
	shutdownResult := make(chan error, 1)
	go func() {
		select {
		case <-ctx.Done():
			shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			err := server.Shutdown(shutdownCtx)
			if err != nil {
				err = errors.Join(err, server.Close())
			}
			shutdownResult <- err
		case <-done:
			shutdownResult <- nil
		}
	}()
	err := run()
	close(done)
	shutdownErr := <-shutdownResult
	if errors.Is(err, http.ErrServerClosed) {
		err = nil
	}
	return errors.Join(err, shutdownErr)
}
