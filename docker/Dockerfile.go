# Multi-stage Go build
FROM golang:1.23-alpine AS builder

ARG SERVICE=api

WORKDIR /app

# Cache dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy source
COPY . .

# Build the specified service
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/service ./cmd/${SERVICE}

# Runtime image
FROM alpine:3.20

RUN apk --no-cache add ca-certificates tzdata

COPY --from=builder /app/service /usr/local/bin/service
COPY --from=builder /app/migrations /app/migrations

EXPOSE 8080 8081

CMD ["/usr/local/bin/service"]
