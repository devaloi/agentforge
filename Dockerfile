# Build stage
FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o bin/agentforge ./cmd/agentforge/

# Runtime stage
FROM alpine:3.21

RUN apk add --no-cache ca-certificates

WORKDIR /app

COPY --from=builder /app/bin/agentforge /usr/local/bin/agentforge
COPY --from=builder /app/config/ ./config/

ENV OPENAI_API_KEY=""
ENV ANTHROPIC_API_KEY=""
ENV OLLAMA_BASE_URL=""

ENTRYPOINT ["agentforge"]
CMD ["--help"]
