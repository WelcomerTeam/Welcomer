echo "Docker build and push"
docker build --tag ghcr.io/welcomerteam/welcomer-backend:latest -f welcomer-backend/Dockerfile .
docker push ghcr.io/welcomerteam/welcomer-backend:latest
