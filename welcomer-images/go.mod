module github.com/WelcomerTeam/Welcomer/welcomer-images

go 1.23

require (
	github.com/WelcomerTeam/Discord v0.0.0-20240830112951-b06f039734f5
	github.com/WelcomerTeam/RealRock v0.0.0-20220122233305-f97b8c8cbc15
	github.com/WelcomerTeam/Welcomer/welcomer-core v0.0.0
	github.com/WelcomerTeam/Welcomer/welcomer-utils v0.0.0
	github.com/disintegration/imaging v1.6.2
	github.com/fogleman/gg v1.3.0
	github.com/gin-contrib/logger v1.2.2
	github.com/gin-gonic/gin v1.10.0
	github.com/gofrs/uuid v4.4.0+incompatible
	github.com/jackc/pgx/v4 v4.18.3
	github.com/joho/godotenv v1.5.1
	github.com/prometheus/client_golang v1.20.5
	github.com/rs/zerolog v1.33.0
	github.com/savsgio/gotils v0.0.0-20240704082632-aef3928b8a38
	github.com/ultimate-guitar/go-imagequant v0.0.0-20201216103743-29e607cca148
	golang.org/x/image v0.23.0
	google.golang.org/grpc v1.69.2
	google.golang.org/protobuf v1.36.0
)

require (
	github.com/asaskevich/govalidator v0.0.0-20230301143203-a9d515a09cc2 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/bytedance/sonic v1.12.6 // indirect
	github.com/bytedance/sonic/loader v0.2.1 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/cloudwego/base64x v0.1.4 // indirect
	github.com/cloudwego/iasm v0.2.0 // indirect
	github.com/davidmytton/url-verifier v1.0.1 // indirect
	github.com/gabriel-vasile/mimetype v1.4.7 // indirect
	github.com/gin-contrib/sse v0.1.0 // indirect
	github.com/go-playground/locales v0.14.1 // indirect
	github.com/go-playground/universal-translator v0.18.1 // indirect
	github.com/go-playground/validator/v10 v10.23.0 // indirect
	github.com/goccy/go-json v0.10.4 // indirect
	github.com/golang/freetype v0.0.0-20170609003504-e2365dfdc4a0 // indirect
	github.com/hashicorp/golang-lru v1.0.2 // indirect
	github.com/jackc/chunkreader/v2 v2.0.1 // indirect
	github.com/jackc/pgconn v1.14.3 // indirect
	github.com/jackc/pgio v1.0.0 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgproto3/v2 v2.3.3 // indirect
	github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761 // indirect
	github.com/jackc/pgtype v1.14.4 // indirect
	github.com/jackc/puddle v1.3.0 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/klauspost/compress v1.17.11 // indirect
	github.com/klauspost/cpuid/v2 v2.2.9 // indirect
	github.com/kr/text v0.2.0 // indirect
	github.com/leodido/go-urn v1.4.0 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/munnerz/goautoneg v0.0.0-20191010083416-a7dc8b61c822 // indirect
	github.com/pelletier/go-toml/v2 v2.2.3 // indirect
	github.com/prometheus/client_model v0.6.1 // indirect
	github.com/prometheus/common v0.61.0 // indirect
	github.com/prometheus/procfs v0.15.1 // indirect
	github.com/twitchyliquid64/golang-asm v0.15.1 // indirect
	github.com/ua-parser/uap-go v0.0.0-20241012191800-bbb40edc15aa // indirect
	github.com/ugorji/go/codec v1.2.12 // indirect
	golang.org/x/arch v0.12.0 // indirect
	golang.org/x/crypto v0.31.0 // indirect
	golang.org/x/net v0.33.0 // indirect
	golang.org/x/sys v0.28.0 // indirect
	golang.org/x/text v0.21.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20241219192143-6b3ec007d9bb // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace github.com/WelcomerTeam/Welcomer/welcomer-core v0.0.0 => ../welcomer-core

replace github.com/WelcomerTeam/Welcomer/welcomer-utils v0.0.0 => ../welcomer-utils
