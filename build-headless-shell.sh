echo "Docker build and push"
docker build --tag ghcr.io/welcomerteam/headless-shell:latest -f welcomer-headless-shell/Dockerfile .
docker push ghcr.io/welcomerteam/headless-shell:latest