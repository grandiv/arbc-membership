# PromoZcy (voucher / campaign engine) build.
# Run scripts/stage.sh first to populate .kzcy/. See docker-compose.yml.
FROM golang:1.25-alpine AS builder
WORKDIR /app
RUN apk add --no-cache git gcc musl-dev

COPY .kzcy/libs/ /libs/
COPY .kzcy/services/PromoZcy/PromoZcy/ ./

RUN sed -i 's|../../../libs/|/libs/|g' go.mod

ENV GOPROXY=https://goproxy.io,direct
ENV GOSUMDB=off
RUN go mod download && go build -a -installsuffix cgo -o promozcy ./cmd/server/

FROM alpine:3.19
RUN apk add --no-cache ca-certificates tzdata wget
WORKDIR /app
COPY --from=builder /app/promozcy .
ENV GIN_MODE=release
EXPOSE 8082
HEALTHCHECK --interval=30s --timeout=3s CMD wget -qO- http://127.0.0.1:8082/health || exit 1
CMD ["./promozcy"]
