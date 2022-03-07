echo Updating dependencies
go get -u github.com/WelcomerTeam/Discord/discord
go get -u github.com/WelcomerTeam/Sandwich-Daemon/protobuf
go get -u github.com/WelcomerTeam/Sandwich-Daemon/structs
go get -u github.com/WelcomerTeam/Sandwich/messaging
go get -u github.com/WelcomerTeam/Sandwich/sandwich
go get -u github.com/json-iterator/go
go get -u github.com/rs/zerolog
go get -u google.golang.org/grpc
go get -u google.golang.org/grpc/credentials/insecure
go get -u gopkg.in/natefinch/lumberjack.v2
go get -u github.com/jackc/pgx/v4/pgxpool
go get -u github.com/jackc/pgtype/ext/gofrs-uuid

go mod tidy
