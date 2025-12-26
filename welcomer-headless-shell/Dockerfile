FROM chromedp/headless-shell:latest

RUN apt-get update && \
    apt-get install -y ca-certificates \
    fontconfig \
    fonts-liberation \
    fonts-dejavu \
    fonts-noto \
    fonts-noto-color-emoji && \
    update-ca-certificates

COPY welcomer-headless-shell/fonts/*.ttf /usr/share/fonts/truetype/custom/
RUN fc-cache -f -v

RUN update-ca-certificates

CMD ["/headless-shell/run.sh"]

LABEL org.opencontainers.image.source=https://github.com/WelcomerTeam/Welcomer