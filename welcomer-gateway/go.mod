module github.com/WelcomerTeam/Welcomer/welcomer-gateway

go 1.22

toolchain go1.22.0

require (
	github.com/WelcomerTeam/Discord v0.0.0-20240514163306-3aed427237a5
	github.com/WelcomerTeam/Sandwich v0.0.0-20240601213226-0e512a04534a
	github.com/WelcomerTeam/Sandwich-Daemon v0.0.0-20240531214409-1966ff46a177
	github.com/WelcomerTeam/Welcomer/welcomer-core v0.0.0
	github.com/WelcomerTeam/Welcomer/welcomer-utils v0.0.0
	github.com/gofrs/uuid v4.4.0+incompatible
	github.com/jackc/pgx/v4 v4.18.3
	github.com/joho/godotenv v1.5.1
	github.com/rs/zerolog v1.33.0
	github.com/savsgio/gotils v0.0.0-20240303185622-093b76447511
	google.golang.org/grpc v1.64.0
)

require (
	github.com/Knetic/govaluate v3.0.1-0.20171022003610-9aa49832a739+incompatible // indirect
	github.com/WelcomerTeam/Mustachvulate v1.2.1-0.20231218130351-adad26f1e96e // indirect
	github.com/WelcomerTeam/Subway v0.0.0-20240515214206-eef3c3b554ac // indirect
	github.com/asaskevich/govalidator v0.0.0-20230301143203-a9d515a09cc2 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/davidmytton/url-verifier v1.0.1 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/go-redis/redis/v8 v8.11.5 // indirect
	github.com/hashicorp/golang-lru v1.0.2 // indirect
	github.com/jackc/chunkreader/v2 v2.0.1 // indirect
	github.com/jackc/pgconn v1.14.3 // indirect
	github.com/jackc/pgio v1.0.0 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgproto3/v2 v2.3.3 // indirect
	github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761 // indirect
	github.com/jackc/pgtype v1.14.3 // indirect
	github.com/jackc/puddle v1.3.0 // indirect
	github.com/klauspost/compress v1.17.9 // indirect
	github.com/lithammer/fuzzysearch v1.1.8 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/nats-io/nats.go v1.36.0 // indirect
	github.com/nats-io/nkeys v0.4.7 // indirect
	github.com/nats-io/nuid v1.0.1 // indirect
	github.com/pierrec/lz4/v4 v4.1.21 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/prometheus/client_golang v1.19.1 // indirect
	github.com/prometheus/client_model v0.6.1 // indirect
	github.com/prometheus/common v0.54.0 // indirect
	github.com/prometheus/procfs v0.15.1 // indirect
	github.com/segmentio/kafka-go v0.4.47 // indirect
	github.com/ua-parser/uap-go v0.0.0-20240611065828-3a4781585db6 // indirect
	golang.org/x/crypto v0.24.0 // indirect
	golang.org/x/net v0.26.0 // indirect
	golang.org/x/sys v0.21.0 // indirect
	golang.org/x/text v0.16.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20240617180043-68d350f18fd4 // indirect
	google.golang.org/protobuf v1.34.2 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
)

replace github.com/WelcomerTeam/Welcomer/welcomer-core v0.0.0 => ../welcomer-core

replace github.com/WelcomerTeam/Welcomer/welcomer-utils v0.0.0 => ../welcomer-utils
