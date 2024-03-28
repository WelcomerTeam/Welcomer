echo "Docker build and push"
docker build --tag ghcr.io/welcomerteam/welcomer-gateway:latest -f welcomer-gateway/Dockerfile .
docker push ghcr.io/welcomerteam/welcomer-gateway:latest
