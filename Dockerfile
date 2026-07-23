FROM golang:1.22-alpine AS builder

WORKDIR /app
COPY go.mod ./
COPY . .
RUN go mod tidy
RUN CGO_ENABLED=0 GOOS=linux go build -o server ./cmd/server

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/server .
COPY --from=builder /app/web ./web
COPY .env .env

EXPOSE 8081
CMD ["./server"]
