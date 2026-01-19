echo "Docker build and push"
docker build --tag ghcr.io/welcomerteam/welcomer-ingest:latest -f welcomer-ingest/Dockerfile .
docker push ghcr.io/welcomerteam/welcomer-ingest:latest