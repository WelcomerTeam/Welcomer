FROM chromedp/headless-shell:latest

RUN apt-get update && \
    apt-get install -y ca-certificates && \
    update-ca-certificates

COPY welcomer-images-proxy/service/welcomer-ca.crt /usr/local/share/ca-certificates/welcomer-ca.crt
RUN update-ca-certificates

CMD ["/headless-shell/run.sh"]

LABEL org.opencontainers.image.source=https://github.com/WelcomerTeam/Welcomer