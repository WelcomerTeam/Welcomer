echo "Docker build and push"
docker build --tag ghcr.io/welcomerteam/welcomer-interactions:latest -f welcomer-interactions/Dockerfile .
docker push ghcr.io/welcomerteam/welcomer-interactions:latest
