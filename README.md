# Welcomer
The welcomer discord bot! This repository contains all the inner workings

## Tracing

The backend and interactions services support OpenTelemetry tracing. Tracing is
disabled by default; enable OTLP export with `OTEL_TRACES_EXPORTER=otlp` or an
`OTEL_EXPORTER_OTLP_ENDPOINT` (or `OTEL_EXPORTER_OTLP_TRACES_ENDPOINT`). The
default OTLP transport is gRPC; set `OTEL_EXPORTER_OTLP_PROTOCOL=http/protobuf`
for HTTP. Standard OTLP endpoint, authentication, TLS, sampling, and resource
environment variables are supported. The default service names are
`welcomer-backend` and `welcomer-interactions`; `OTEL_SERVICE_NAME` overrides
them. Set `OTEL_SDK_DISABLED=true` or `OTEL_TRACES_EXPORTER=none` to disable
export.

When tracing is enabled, new traces are sampled by default, while an incoming
trace's sampling decision is normally respected. Set
`WELCOMER_TRACE_SAMPLE_PERCENT` to a number from `0` to `100` to control the
percentage of traces sampled by each service (for example, `25` for 25%).
Decimal percentages are supported. `0` samples none and `100` samples all.
This setting takes precedence over `OTEL_TRACES_SAMPLER` and
`OTEL_TRACES_SAMPLER_ARG` and makes its decision from the trace ID even for
incoming traces, overriding an upstream sampling flag. Requests sharing the
same trace ID therefore share the same sampling decision. If the setting is
unset or empty, standard OpenTelemetry sampler settings and defaults apply.
Enabling sampling does not itself enable the exporter, and sampling does not
guarantee delivery if export fails.

Server spans include the route, method, response status, and, when available,
`discord.user.id` and `discord.guild.id`. Interaction command spans also include
`discord.command` with subcommand names, but not argument values. User and
guild IDs are personal data; configure access and retention in the collector
accordingly. Interaction metadata is omitted for rejected requests and payloads
larger than 1 MiB. Request bodies, query strings, and credentials are not
exported. The current Subway dependency does not propagate HTTP request context
to interaction handlers, so the interaction server span covers execution but
handler-created child spans cannot currently inherit it. Subway can return HTTP
204 for a failed command, so the server span's HTTP status alone does not
reliably indicate command success. On SIGINT or SIGTERM, both services wait
briefly for active requests and flush queued spans before exiting.
