FROM golang:1.26-alpine AS builder
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o task-manager-api ./cmd/api/main.go

FROM alpine
WORKDIR /app

COPY --from=builder /app/task-manager-api /app/task-manager-api

EXPOSE 8080
CMD ["/app/task-manager-api"]
