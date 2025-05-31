go generate service/service.go &&
go build -v -x -o welcomer-images.local cmd/main.go &&
./welcomer-images.local --debug