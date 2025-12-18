echo "Docker build and push"
docker build --tag ghcr.io/welcomerteam/welcomer-images-proxy:latest -f welcomer-images-proxy/Dockerfile .
docker push ghcr.io/welcomerteam/welcomer-images-proxy:latest
