# KonsumZcy (member store) build.
# Run scripts/stage.sh first to populate .kzcy/ (libs + engine source). Build
# context = this repo root; see docker-compose.yml.
FROM golang:1.25-alpine AS builder
WORKDIR /app
RUN apk add --no-cache git gcc musl-dev

COPY .kzcy/libs/ /libs/
COPY .kzcy/services/KonsumZcy/KonsumZcy/ ./

# Rewrite the relative replace paths (../../../libs/) to the in-image location.
RUN sed -i 's|../../../libs/|/libs/|g' go.mod

# izcy-engine VPS blocks Google's module proxy + checksum DB.
ENV GOPROXY=https://goproxy.io,direct
ENV GOSUMDB=off
RUN go mod download && go build -a -installsuffix cgo -o konsumzcy ./cmd/server/

FROM alpine:3.19
RUN apk add --no-cache ca-certificates tzdata wget
WORKDIR /app
COPY --from=builder /app/konsumzcy .
ENV GIN_MODE=release
EXPOSE 8084
HEALTHCHECK --interval=30s --timeout=3s CMD wget -qO- http://127.0.0.1:8084/health || exit 1
CMD ["./konsumzcy"]
