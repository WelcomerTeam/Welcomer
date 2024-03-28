echo "Docker build and push"
docker build --tag ghcr.io/welcomerteam/welcomer-images:latest -f welcomer-images/Dockerfile .
docker push ghcr.io/welcomerteam/welcomer-images:latest
