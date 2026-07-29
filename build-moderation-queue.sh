echo "Docker build and push"
docker build --tag ghcr.io/welcomerteam/welcomer-moderation-queue:latest -f moderation-queue/Dockerfile .
docker push ghcr.io/welcomerteam/welcomer-moderation-queue:latest
