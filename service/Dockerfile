FROM golang:1.25-alpine AS builder

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/account-service ./cmd/account-service

FROM ubuntu:24.04

RUN apt-get update \
    && apt-get install -y --no-install-recommends bash ca-certificates curl unzip \
    && /bin/bash -o pipefail -c 'curl -fsSL https://cli.kiro.dev/install | bash' \
    && install -m 0755 /root/.local/bin/kiro-cli* /usr/local/bin/ \
    && rm -rf /var/lib/apt/lists/*

RUN groupadd --system app && useradd --system --gid app --home-dir /app --shell /usr/sbin/nologin app

WORKDIR /app
COPY --from=builder /out/account-service /app/account-service

USER app
EXPOSE 8000

ENTRYPOINT ["/app/account-service"]
