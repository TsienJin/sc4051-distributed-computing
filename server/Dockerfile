FROM golang:1.23-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o server ./cmd/server.go

FROM alpine:3.20.3

WORKDIR /app
COPY --from=builder /app/server .
RUN chmod +x /app/server

CMD ["/app/server"]