FROM golang:1.25-alpine AS builder

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/account-service ./cmd/account-service

FROM alpine:3.22

RUN apk add --no-cache bash ca-certificates curl \
    && /bin/bash -o pipefail -c 'curl -fsSL https://cli.kiro.dev/install | bash' \
    && install -m 0755 /root/.local/bin/kiro-cli* /usr/local/bin/

RUN addgroup -S app && adduser -S app -G app

WORKDIR /app
COPY --from=builder /out/account-service /app/account-service

USER app
EXPOSE 8000

ENTRYPOINT ["/app/account-service"]
