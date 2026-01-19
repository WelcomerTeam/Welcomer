FROM golang:1.25 AS build_base
COPY . /tmp
WORKDIR /tmp/welcomer-ingest
RUN cd /tmp/welcomer-ingest
RUN go mod tidy; exit 0
RUN apt update -y \
    && apt install -y git build-essential cmake zlib1g-dev
RUN CGO_ENABLED=1 GOOS=linux GOARCH=amd64 LD_LIBRARY_PATH='/usr/local/lib' \
    go build -a --trimpath -o ./out/ingest ./cmd/main.go

FROM alpine:3
RUN apk add ca-certificates libc6-compat
COPY --from=build_base /usr/local/lib /usr/local/lib
COPY --from=build_base /tmp/welcomer-ingest/out/ingest /app/ingest
CMD ["/app/ingest"]

LABEL org.opencontainers.image.source=https://github.com/WelcomerTeam/Welcomer