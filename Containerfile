FROM docker.io/library/golang:1.27.0-trixie AS builder

LABEL org.opencontainers.image.title="scrapedoctl" \
      org.opencontainers.image.description="Scrape.do CLI and MCP server" \
      org.opencontainers.image.source="https://github.com/ioplane/scrapedoctl" \
      org.opencontainers.image.vendor="ioplane" \
      org.opencontainers.image.licenses="MIT" \
      org.opencontainers.image.base.name="docker.io/library/golang:1.27.0-trixie"

WORKDIR /src

COPY go.mod go.sum ./
RUN ["go", "mod", "download"]
RUN ["go", "mod", "verify"]

COPY cmd ./cmd
COPY internal ./internal
COPY pkg ./pkg

ENV CGO_ENABLED=0

RUN ["go", "build", "-trimpath", "-ldflags=-s -w", "-o", "/out/scrapedoctl", "./cmd/scrapedoctl"]
RUN ["mkdir", "-p", "/out/home/.scrapedoctl"]

FROM scratch AS production

LABEL org.opencontainers.image.title="scrapedoctl" \
      org.opencontainers.image.description="Scrape.do CLI and MCP server" \
      org.opencontainers.image.source="https://github.com/ioplane/scrapedoctl" \
      org.opencontainers.image.vendor="ioplane" \
      org.opencontainers.image.licenses="MIT"

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=builder --chown=65532:65532 /out/home /home/scrape
COPY --from=builder --chown=65532:65532 /out/scrapedoctl /bin/scrapedoctl

ENV HOME=/home/scrape

USER 65532:65532
VOLUME ["/home/scrape/.scrapedoctl"]
ENTRYPOINT ["/bin/scrapedoctl"]
