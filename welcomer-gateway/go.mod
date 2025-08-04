module github.com/WelcomerTeam/Welcomer/welcomer-gateway

go 1.23.0

toolchain go1.23.1

require (
	github.com/WelcomerTeam/Discord v0.0.0-20250802100629-f340fe10f307
	github.com/WelcomerTeam/Sandwich v0.0.0-20250704091340-1979653e6d21
	github.com/WelcomerTeam/Sandwich-Daemon v0.0.0-20250802101748-9052bf06297e
	github.com/WelcomerTeam/Welcomer/welcomer-core v0.0.0
	github.com/gofrs/uuid v4.4.0+incompatible
	github.com/jackc/pgx/v4 v4.18.3
	github.com/joho/godotenv v1.5.1
	github.com/nats-io/nats.go v1.44.0
	github.com/savsgio/gotils v0.0.0-20250408102913-196191ec6287
	google.golang.org/grpc v1.74.2
)

require (
	github.com/Knetic/govaluate v3.0.1-0.20171022003610-9aa49832a739+incompatible // indirect
	github.com/WelcomerTeam/Mustachvulate v1.2.1-0.20231218130351-adad26f1e96e // indirect
	github.com/WelcomerTeam/Subway v0.0.0-20250728235017-02b54da8d689 // indirect
	github.com/WelcomerTeam/czlib v0.0.0-20210907121728-d7ed7721c904 // indirect
	github.com/asaskevich/govalidator v0.0.0-20230301143203-a9d515a09cc2 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/coder/websocket v1.8.13 // indirect
	github.com/davidmytton/url-verifier v1.0.1 // indirect
	github.com/jackc/chunkreader/v2 v2.0.1 // indirect
	github.com/jackc/pgconn v1.14.3 // indirect
	github.com/jackc/pgio v1.0.0 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgproto3/v2 v2.3.3 // indirect
	github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761 // indirect
	github.com/jackc/pgtype v1.14.4 // indirect
	github.com/jackc/puddle v1.3.0 // indirect
	github.com/klauspost/compress v1.18.0 // indirect
	github.com/mattn/go-colorable v0.1.14 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/mhmtszr/concurrent-swiss-map v1.0.8 // indirect
	github.com/munnerz/goautoneg v0.0.0-20191010083416-a7dc8b61c822 // indirect
	github.com/nats-io/nkeys v0.4.11 // indirect
	github.com/nats-io/nuid v1.0.1 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/plutov/paypal/v4 v4.12.0 // indirect
	github.com/prometheus/client_golang v1.23.0 // indirect
	github.com/prometheus/client_model v0.6.2 // indirect
	github.com/prometheus/common v0.65.0 // indirect
	github.com/prometheus/procfs v0.17.0 // indirect
	github.com/rs/zerolog v1.34.0 // indirect
	golang.org/x/crypto v0.40.0 // indirect
	golang.org/x/net v0.42.0 // indirect
	golang.org/x/sys v0.34.0 // indirect
	golang.org/x/text v0.27.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20250728155136-f173205681a0 // indirect
	google.golang.org/protobuf v1.36.6 // indirect
)

replace github.com/WelcomerTeam/Welcomer/welcomer-core v0.0.0 => ../welcomer-core
