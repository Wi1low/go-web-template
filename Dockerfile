# syntax=docker/dockerfile:1

FROM golang:1.25-alpine AS builder
WORKDIR /src

# Cache module downloads independently of source changes.
COPY go.mod go.sum ./
RUN go mod download

COPY . .
# docs/ is committed; rebuild it before shipping schema changes (make swagger).
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/app ./cmd/api

FROM alpine:3.20
RUN adduser -D -u 10001 app && apk add --no-cache ca-certificates tzdata
WORKDIR /app
COPY --from=builder /out/app ./app
COPY cmd/etc ./cmd/etc
USER app

EXPOSE 8080 9091
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s \
  CMD wget -qO- http://127.0.0.1:8080/ping || exit 1
ENTRYPOINT ["./app"]
