FROM golang:1.23 AS build_base
COPY . /tmp
WORKDIR /tmp/welcomer-gateway
RUN cd /tmp/welcomer-gateway
RUN go mod tidy; exit 0
RUN CGO_ENABLED=1 GOOS=linux GOARCH=amd64 LD_LIBRARY_PATH='/usr/local/lib' \
    go build -a --trimpath -o ./out/gateway ./cmd/main.go

FROM alpine:3
RUN apk add ca-certificates libc6-compat
COPY --from=build_base /usr/local/lib /usr/local/lib
COPY --from=build_base /tmp/welcomer-gateway/out/gateway /app/gateway
CMD ["/app/gateway"]

LABEL org.opencontainers.image.source https://github.com/WelcomerTeam/Welcomer