module github.com/WelcomerTeam/Welcomer/welcomer-gateway

go 1.23

require (
	github.com/WelcomerTeam/Discord v0.0.0-20240830112951-b06f039734f5
	github.com/WelcomerTeam/Sandwich v0.0.0-20240830113518-2b681113998e
	github.com/WelcomerTeam/Sandwich-Daemon v0.0.0-20241222231050-c5fbbb3ebb11
	github.com/WelcomerTeam/Welcomer/welcomer-core v0.0.0
	github.com/WelcomerTeam/Welcomer/welcomer-utils v0.0.0
	github.com/gofrs/uuid v4.4.0+incompatible
	github.com/jackc/pgx/v4 v4.18.3
	github.com/joho/godotenv v1.5.1
	github.com/rs/zerolog v1.33.0
	github.com/savsgio/gotils v0.0.0-20240704082632-aef3928b8a38
	google.golang.org/grpc v1.69.2
)

require (
	github.com/Knetic/govaluate v3.0.1-0.20171022003610-9aa49832a739+incompatible // indirect
	github.com/WelcomerTeam/Mustachvulate v1.2.1-0.20231218130351-adad26f1e96e // indirect
	github.com/WelcomerTeam/Subway v0.0.0-20240809224607-b8332ec15045 // indirect
	github.com/asaskevich/govalidator v0.0.0-20230301143203-a9d515a09cc2 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/davidmytton/url-verifier v1.0.1 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/go-redis/redis/v8 v8.11.5 // indirect
	github.com/jackc/chunkreader/v2 v2.0.1 // indirect
	github.com/jackc/pgconn v1.14.3 // indirect
	github.com/jackc/pgio v1.0.0 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgproto3/v2 v2.3.3 // indirect
	github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761 // indirect
	github.com/jackc/pgtype v1.14.4 // indirect
	github.com/jackc/puddle v1.3.0 // indirect
	github.com/klauspost/compress v1.17.11 // indirect
	github.com/lithammer/fuzzysearch v1.1.8 // indirect
	github.com/mattn/go-colorable v0.1.14 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/munnerz/goautoneg v0.0.0-20191010083416-a7dc8b61c822 // indirect
	github.com/nats-io/nats.go v1.39.0 // indirect
	github.com/nats-io/nkeys v0.4.9 // indirect
	github.com/nats-io/nuid v1.0.1 // indirect
	github.com/pierrec/lz4/v4 v4.1.22 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/plutov/paypal/v4 v4.11.0 // indirect
	github.com/prometheus/client_golang v1.20.5 // indirect
	github.com/prometheus/client_model v0.6.1 // indirect
	github.com/prometheus/common v0.62.0 // indirect
	github.com/prometheus/procfs v0.15.1 // indirect
	github.com/segmentio/kafka-go v0.4.47 // indirect
	golang.org/x/crypto v0.33.0 // indirect
	golang.org/x/net v0.34.0 // indirect
	golang.org/x/sys v0.30.0 // indirect
	golang.org/x/text v0.22.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20241219192143-6b3ec007d9bb // indirect
	google.golang.org/protobuf v1.36.2 // indirect
)

replace github.com/WelcomerTeam/Welcomer/welcomer-core v0.0.0 => ../welcomer-core

replace github.com/WelcomerTeam/Welcomer/welcomer-utils v0.0.0 => ../welcomer-utils
