module github.com/WelcomerTeam/Welcomer/welcomer-interactions

go 1.21

require (
	github.com/WelcomerTeam/Discord v0.0.0-20230919203812-a65cb654c4a8
	github.com/WelcomerTeam/Sandwich v0.0.0-20230914001140-a7a1fd53a02f
	github.com/WelcomerTeam/Sandwich-Daemon v0.0.0-20230916083319-1035b332fc77
	github.com/WelcomerTeam/Subway v0.0.0-20231115212023-2cd9c14a4fb0
	github.com/WelcomerTeam/Welcomer/welcomer-core v0.0.0
	github.com/jackc/pgx/v4 v4.18.1
	github.com/joho/godotenv v1.5.1
	github.com/json-iterator/go v1.1.12
	github.com/rs/zerolog v1.31.0
	google.golang.org/grpc v1.59.0
	gopkg.in/natefinch/lumberjack.v2 v2.2.1
)

require (
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/cespare/xxhash/v2 v2.2.0 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/go-redis/redis/v8 v8.11.5 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang/protobuf v1.5.3 // indirect
	github.com/jackc/chunkreader/v2 v2.0.1 // indirect
	github.com/jackc/pgconn v1.14.1 // indirect
	github.com/jackc/pgio v1.0.0 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgproto3/v2 v2.3.2 // indirect
	github.com/jackc/pgservicefile v0.0.0-20221227161230-091c0ba34f0a // indirect
	github.com/jackc/pgtype v1.14.0 // indirect
	github.com/jackc/puddle v1.3.0 // indirect
	github.com/klauspost/compress v1.17.3 // indirect
	github.com/lithammer/fuzzysearch v1.1.8 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/matttproud/golang_protobuf_extensions/v2 v2.0.0 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/nats-io/nats.go v1.31.0 // indirect
	github.com/nats-io/nkeys v0.4.6 // indirect
	github.com/nats-io/nuid v1.0.1 // indirect
	github.com/nats-io/stan.go v0.10.4 // indirect
	github.com/pierrec/lz4/v4 v4.1.18 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/prometheus/client_golang v1.17.0 // indirect
	github.com/prometheus/client_model v0.5.0 // indirect
	github.com/prometheus/common v0.45.0 // indirect
	github.com/prometheus/procfs v0.12.0 // indirect
	github.com/savsgio/gotils v0.0.0-20230208104028-c358bd845dee // indirect
	github.com/segmentio/kafka-go v0.4.46 // indirect
	golang.org/x/crypto v0.16.0 // indirect
	golang.org/x/net v0.19.0 // indirect
	golang.org/x/sys v0.15.0 // indirect
	golang.org/x/text v0.14.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20231127180814-3a041ad873d4 // indirect
	google.golang.org/protobuf v1.31.0 // indirect
)

replace github.com/WelcomerTeam/Welcomer/welcomer-core v0.0.0 => ../welcomer-core
