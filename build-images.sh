echo "Docker build and push"
docker build --tag ghcr.io/welcomerteam/utils.images:latest -f utils.images/Dockerfile .
docker push ghcr.io/welcomerteam/utils.images:latest
