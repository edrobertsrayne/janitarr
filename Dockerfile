# Build stage
FROM golang:1.23-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git

# Install templ
RUN go install github.com/a-h/templ/cmd/templ@latest

# Set GOTOOLCHAIN to allow auto-downloading newer Go versions
ENV GOTOOLCHAIN=auto

WORKDIR /build

# Download dependencies first (better caching)
COPY go.mod go.sum ./
RUN go mod download

# Copy source and generate templates
COPY . .
RUN templ generate
RUN go build -ldflags="-s -w" -o janitarr ./src

# Runtime stage
FROM alpine:latest

RUN apk add --no-cache su-exec shadow wget

COPY --from=builder /build/janitarr /usr/local/bin/janitarr
COPY docker-entrypoint.sh /docker-entrypoint.sh

RUN chmod +x /docker-entrypoint.sh

ENV PUID=1000 \
    PGID=1000 \
    JANITARR_PORT=3434 \
    JANITARR_DB_PATH=/data/janitarr.db

EXPOSE 3434

VOLUME ["/data"]

HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
  CMD wget --quiet --tries=1 --spider http://localhost:${JANITARR_PORT:-3434}/health || exit 1

ENTRYPOINT ["/docker-entrypoint.sh"]
CMD ["janitarr", "start", "--host", "0.0.0.0"]
