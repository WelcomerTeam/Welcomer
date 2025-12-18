echo "Docker build and push"
docker build --tag ghcr.io/welcomerteam/welcomer-images-next:latest -f welcomer-images-next/Dockerfile .
docker push ghcr.io/welcomerteam/welcomer-images-next:latest
