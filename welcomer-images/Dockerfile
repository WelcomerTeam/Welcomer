FROM golang:1.22 AS build_base
COPY . /tmp
WORKDIR /tmp/welcomer_images
RUN cd /tmp/welcomer_images
RUN go mod tidy; exit 0
RUN CGO_ENABLED=1 GOOS=linux GOARCH=amd64 LD_LIBRARY_PATH='/usr/local/lib' \
    go build -a --trimpath -o ./out/images ./cmd/main.go

FROM alpine:3
RUN apk add ca-certificates libc6-compat
COPY --from=build_base /usr/local/lib /usr/local/lib
COPY --from=build_base /tmp/welcomer_images/out/images /app/images
CMD ["/app/images"]

LABEL org.opencontainers.image.source https://github.com/WelcomerTeam/Welcomer