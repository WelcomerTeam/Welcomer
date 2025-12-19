FROM golang:1.25 AS build_base
COPY . /tmp
WORKDIR /tmp/welcomer-images-next
RUN cd /tmp/welcomer-images-next
RUN go mod tidy; exit 0
RUN apt update -y \
    && apt install -y git build-essential cmake zlib1g-dev
RUN CGO_ENABLED=1 GOOS=linux GOARCH=amd64 LD_LIBRARY_PATH='/usr/local/lib' \
    go build -a --trimpath -o ./out/images-next ./cmd/main.go

FROM alpine:3
RUN apk add ca-certificates libc6-compat
COPY --from=build_base /usr/local/lib /usr/local/lib
COPY --from=build_base /tmp/welcomer-images-next/out/images-next /app/images-next
CMD ["/app/images-next"]

LABEL org.opencontainers.image.source=https://github.com/WelcomerTeam/Welcomer