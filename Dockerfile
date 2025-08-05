# ---------- Stage 1: Build ----------
FROM golang:1.23.4-alpine AS builder

WORKDIR /app

# Копируем только go.mod и go.sum, чтобы кешировать зависимости
COPY go.mod go.sum ./
RUN go mod download

# Копируем остальной код
COPY . .

# Сборка бинарника из ./cmd/server
RUN go build -o main ./cmd/server

# ---------- Stage 2: Final ----------
FROM alpine:3.18

WORKDIR /root/
COPY --from=builder /app/main .
COPY config ./config
CMD ["./main", "--config_path=./config/prod.yaml"]
