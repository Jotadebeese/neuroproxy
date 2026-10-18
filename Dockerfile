# Stage 1: Build Go Self-Healing Gateway Binary
FROM golang:1.23-alpine AS builder

WORKDIR /app
COPY go.mod ./
COPY main.go ./
COPY cmd/ ./cmd/
COPY internal/ ./internal/
COPY frontend/out/ ./frontend/out/

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o /app/neuroproxy ./cmd/neuroproxy

# Stage 2: Minimal Production Image for Google Cloud Run
FROM alpine:3.20
RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app
COPY --from=builder /app/neuroproxy /app/neuroproxy
COPY --from=builder /app/frontend/out /app/frontend/out

ENV PORT=8080
ENV STATIC_DIR=/app/frontend/out
EXPOSE 8080

CMD ["/app/neuroproxy"]
