echo "Docker build and push"
docker build --tag ghcr.io/welcomerteam/sandwich:latest -f sandwich/Dockerfile .
docker push ghcr.io/welcomerteam/sandwich:latest
