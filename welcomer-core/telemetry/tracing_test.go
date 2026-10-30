package telemetry

import (
	"context"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"strconv"
	"sync/atomic"
	"testing"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
	collectortrace "go.opentelemetry.io/proto/otlp/collector/trace/v1"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/proto"
)

type testTraceCollector struct {
	collectortrace.UnimplementedTraceServiceServer
	requests chan *collectortrace.ExportTraceServiceRequest
}

func (c *testTraceCollector) Export(
	_ context.Context,
	request *collectortrace.ExportTraceServiceRequest,
) (*collectortrace.ExportTraceServiceResponse, error) {
	c.requests <- request
	return &collectortrace.ExportTraceServiceResponse{}, nil
}

func TestOTLPGRPCExport(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	server := grpc.NewServer()
	collector := &testTraceCollector{
		requests: make(chan *collectortrace.ExportTraceServiceRequest, 1),
	}
	collectortrace.RegisterTraceServiceServer(server, collector)
	go func() { _ = server.Serve(listener) }()
	defer server.Stop()
	t.Setenv("OTEL_SDK_DISABLED", "false")
	t.Setenv("OTEL_TRACES_EXPORTER", "otlp")
	t.Setenv("OTEL_EXPORTER_OTLP_ENDPOINT", "http://"+listener.Addr().String())
	t.Setenv("OTEL_EXPORTER_OTLP_TRACES_ENDPOINT", "")
	t.Setenv("OTEL_EXPORTER_OTLP_PROTOCOL", "grpc")
	t.Setenv("OTEL_EXPORTER_OTLP_TRACES_PROTOCOL", "")
	t.Setenv("OTEL_TRACES_SAMPLER", "always_on")
	t.Setenv("WELCOMER_TRACE_SAMPLE_PERCENT", "")
	oldProvider, oldPropagator := otel.GetTracerProvider(), otel.GetTextMapPropagator()
	defer otel.SetTracerProvider(oldProvider)
	defer otel.SetTextMapPropagator(oldPropagator)
	shutdown, err := Setup(context.Background(), "welcomer-test")
	if err != nil {
		t.Fatal(err)
	}
	_, span := otel.Tracer("test").Start(context.Background(), "grpc-span")
	span.End()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := shutdown(ctx); err != nil {
		t.Fatal(err)
	}
	select {
	case request := <-collector.requests:
		resources := request.GetResourceSpans()
		if len(resources) != 1 || len(resources[0].GetScopeSpans()) != 1 ||
			len(resources[0].GetScopeSpans()[0].GetSpans()) != 1 ||
			resources[0].GetScopeSpans()[0].GetSpans()[0].GetName() != "grpc-span" {
			t.Errorf("unexpected gRPC export: %v", request)
		}
	case <-ctx.Done():
		t.Fatal("no gRPC trace export received")
	}
}

func TestOTLPHTTPExport(t *testing.T) {
	requests := make(chan *collectortrace.ExportTraceServiceRequest, 1)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/traces" {
			t.Errorf("unexpected export path %q", r.URL.Path)
		}
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Errorf("read export: %v", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		request := new(collectortrace.ExportTraceServiceRequest)
		if err := proto.Unmarshal(body, request); err != nil {
			t.Errorf("decode export: %v", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		requests <- request
		w.Header().Set("Content-Type", "application/x-protobuf")
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	t.Setenv("OTEL_SDK_DISABLED", "false")
	t.Setenv("OTEL_TRACES_EXPORTER", "otlp")
	t.Setenv("OTEL_EXPORTER_OTLP_ENDPOINT", server.URL)
	t.Setenv("OTEL_EXPORTER_OTLP_TRACES_ENDPOINT", "")
	t.Setenv("OTEL_EXPORTER_OTLP_PROTOCOL", "http/protobuf")
	t.Setenv("OTEL_EXPORTER_OTLP_TRACES_PROTOCOL", "")
	t.Setenv("OTEL_SERVICE_NAME", "")
	t.Setenv("OTEL_RESOURCE_ATTRIBUTES", "")
	t.Setenv("OTEL_TRACES_SAMPLER", "always_on")
	t.Setenv("WELCOMER_TRACE_SAMPLE_PERCENT", "")

	oldProvider, oldPropagator := otel.GetTracerProvider(), otel.GetTextMapPropagator()
	defer otel.SetTracerProvider(oldProvider)
	defer otel.SetTextMapPropagator(oldPropagator)
	shutdown, err := Setup(context.Background(), "welcomer-test")
	if err != nil {
		t.Fatal(err)
	}
	_, span := otel.Tracer("test").Start(context.Background(), "test-span")
	span.End()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := shutdown(ctx); err != nil {
		t.Fatal(err)
	}
	select {
	case request := <-requests:
		resources := request.GetResourceSpans()
		if len(resources) != 1 {
			t.Fatalf("got %d resource spans", len(resources))
		}
		var service string
		for _, attr := range resources[0].GetResource().GetAttributes() {
			if attr.GetKey() == "service.name" {
				service = attr.GetValue().GetStringValue()
			}
		}
		if service != "welcomer-test" {
			t.Errorf("service.name = %q", service)
		}
		spans := resources[0].GetScopeSpans()
		if len(spans) != 1 || len(spans[0].GetSpans()) != 1 ||
			spans[0].GetSpans()[0].GetName() != "test-span" {
			t.Errorf("unexpected exported spans: %v", spans)
		}
	case <-ctx.Done():
		t.Fatal("no trace export received")
	}
}

func TestInvalidExporterConfiguration(t *testing.T) {
	t.Setenv("OTEL_SDK_DISABLED", "false")
	t.Setenv("OTEL_TRACES_EXPORTER", "otlp")
	t.Setenv("OTEL_EXPORTER_OTLP_TRACES_PROTOCOL", "unsupported")
	t.Setenv("WELCOMER_TRACE_SAMPLE_PERCENT", "")
	if _, err := Setup(context.Background(), "test"); err == nil {
		t.Fatal("expected invalid protocol to fail")
	}
}

func TestConfiguredSamplePercent(t *testing.T) {
	tests := []struct {
		percent string
		want    int
	}{
		{"0", 0},
		{"12.5", 32},
		{"25", 64},
		{"50", 128},
		{"100", 256},
	}
	for _, tt := range tests {
		t.Run(tt.percent, func(t *testing.T) {
			var exported atomic.Int64
			server := httptest.NewServer(
				http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					body, err := io.ReadAll(r.Body)
					if err != nil {
						t.Errorf("read export: %v", err)
						w.WriteHeader(http.StatusBadRequest)
						return
					}
					var request collectortrace.ExportTraceServiceRequest
					if err := proto.Unmarshal(body, &request); err != nil {
						t.Errorf("decode export: %v", err)
						w.WriteHeader(http.StatusBadRequest)
						return
					}
					for _, resource := range request.GetResourceSpans() {
						for _, scope := range resource.GetScopeSpans() {
							exported.Add(int64(len(scope.GetSpans())))
						}
					}
					w.Header().Set("Content-Type", "application/x-protobuf")
				}),
			)
			defer server.Close()
			t.Setenv("OTEL_SDK_DISABLED", "false")
			t.Setenv("OTEL_TRACES_EXPORTER", "otlp")
			t.Setenv("OTEL_EXPORTER_OTLP_ENDPOINT", server.URL)
			t.Setenv("OTEL_EXPORTER_OTLP_TRACES_ENDPOINT", "")
			t.Setenv("OTEL_EXPORTER_OTLP_PROTOCOL", "http/protobuf")
			t.Setenv("OTEL_EXPORTER_OTLP_TRACES_PROTOCOL", "")
			t.Setenv("OTEL_TRACES_SAMPLER", "always_off")
			t.Setenv("WELCOMER_TRACE_SAMPLE_PERCENT", tt.percent)
			oldProvider, oldPropagator := otel.GetTracerProvider(), otel.GetTextMapPropagator()
			defer otel.SetTracerProvider(oldProvider)
			defer otel.SetTextMapPropagator(oldPropagator)
			shutdown, err := Setup(context.Background(), "sample-test")
			if err != nil {
				t.Fatal(err)
			}
			recorded := 0
			for i := 0; i < 256; i++ {
				var id trace.TraceID
				id[0], id[8] = 1, byte(i)
				flags := trace.TraceFlags(0)
				if i%2 == 0 {
					flags = trace.FlagsSampled
				}
				parent := trace.NewSpanContext(trace.SpanContextConfig{
					TraceID: id, SpanID: trace.SpanID{1}, TraceFlags: flags, Remote: true,
				})
				ctx := trace.ContextWithRemoteSpanContext(context.Background(), parent)
				req := httptest.NewRequest(http.MethodGet, "/", nil).WithContext(ctx)
				_, span := StartServerSpan(req, "request")
				if span.IsRecording() {
					recorded++
				}
				FinishServerSpan(span, http.StatusOK)
			}
			shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			if err := shutdown(shutdownCtx); err != nil {
				t.Fatal(err)
			}
			if recorded != tt.want || exported.Load() != int64(tt.want) {
				t.Fatalf(
					"recorded %s and exported %d spans, want %d",
					strconv.Itoa(recorded),
					exported.Load(),
					tt.want,
				)
			}
		})
	}
}

func TestInvalidSamplePercent(t *testing.T) {
	t.Setenv("OTEL_SDK_DISABLED", "false")
	t.Setenv("OTEL_TRACES_EXPORTER", "otlp")
	for _, value := range []string{"-1", "100.01", "NaN", "Inf", "-Inf", "abc", "25%"} {
		t.Run(value, func(t *testing.T) {
			t.Setenv("WELCOMER_TRACE_SAMPLE_PERCENT", value)
			if _, err := Setup(context.Background(), "test"); err == nil {
				t.Fatalf("expected invalid percentage %q to fail", value)
			}
		})
	}
}

func TestDefaultSamplingRespectsRemoteParent(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/x-protobuf")
	}))
	defer server.Close()
	t.Setenv("OTEL_SDK_DISABLED", "false")
	t.Setenv("OTEL_TRACES_EXPORTER", "otlp")
	t.Setenv("OTEL_EXPORTER_OTLP_ENDPOINT", server.URL)
	t.Setenv("OTEL_EXPORTER_OTLP_TRACES_ENDPOINT", "")
	t.Setenv("OTEL_EXPORTER_OTLP_PROTOCOL", "http/protobuf")
	t.Setenv("OTEL_EXPORTER_OTLP_TRACES_PROTOCOL", "")
	t.Setenv("OTEL_TRACES_SAMPLER", "")
	t.Setenv("OTEL_TRACES_SAMPLER_ARG", "")
	t.Setenv("WELCOMER_TRACE_SAMPLE_PERCENT", "")
	oldProvider, oldPropagator := otel.GetTracerProvider(), otel.GetTextMapPropagator()
	defer otel.SetTracerProvider(oldProvider)
	defer otel.SetTextMapPropagator(oldPropagator)
	shutdown, err := Setup(context.Background(), "sample-test")
	if err != nil {
		t.Fatal(err)
	}
	_, root := StartServerSpan(httptest.NewRequest(http.MethodGet, "/", nil), "root")
	if !root.IsRecording() {
		t.Error("new trace was not sampled by default")
	}
	root.End()
	for _, sampled := range []bool{false, true} {
		flags := trace.TraceFlags(0)
		if sampled {
			flags = trace.FlagsSampled
		}
		parent := trace.NewSpanContext(trace.SpanContextConfig{
			TraceID: trace.TraceID{1}, SpanID: trace.SpanID{1}, TraceFlags: flags, Remote: true,
		})
		ctx := trace.ContextWithRemoteSpanContext(context.Background(), parent)
		_, span := StartServerSpan(httptest.NewRequest(http.MethodGet, "/", nil).WithContext(ctx), "child")
		if span.IsRecording() != sampled {
			t.Errorf("remote parent sampled=%t, child recording=%t", sampled, span.IsRecording())
		}
		span.End()
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := shutdown(ctx); err != nil {
		t.Fatal(err)
	}
}

func TestTracingIsDisabledWithoutConfiguration(t *testing.T) {
	t.Setenv("OTEL_SDK_DISABLED", "false")
	t.Setenv("OTEL_TRACES_EXPORTER", "")
	t.Setenv("OTEL_EXPORTER_OTLP_ENDPOINT", "")
	t.Setenv("OTEL_EXPORTER_OTLP_TRACES_ENDPOINT", "")
	if Enabled() {
		t.Fatal("tracing should be disabled")
	}
	shutdown, err := Setup(context.Background(), "test")
	if err != nil {
		t.Fatal(err)
	}
	if err := shutdown(context.Background()); err != nil {
		t.Fatal(err)
	}
}

func TestServeWaitsForActiveRequests(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	started, release := make(chan struct{}), make(chan struct{})
	server := &http.Server{Handler: http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		close(started)
		<-release
		w.WriteHeader(http.StatusNoContent)
	})}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	serverResult := make(chan error, 1)
	go func() { serverResult <- serve(ctx, server, func() error { return server.Serve(listener) }) }()
	requestResult := make(chan error, 1)
	go func() {
		resp, err := http.Get("http://" + listener.Addr().String())
		if err == nil {
			_ = resp.Body.Close()
		}
		requestResult <- err
	}()
	select {
	case <-started:
	case <-time.After(5 * time.Second):
		t.Fatal("request did not start")
	}
	cancel()
	select {
	case err := <-serverResult:
		t.Fatalf("server exited before request completed: %v", err)
	case <-time.After(50 * time.Millisecond):
	}
	close(release)
	if err := <-requestResult; err != nil {
		t.Fatal(err)
	}
	if err := <-serverResult; err != nil {
		t.Fatal(err)
	}
}
