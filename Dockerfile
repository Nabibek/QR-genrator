# Build stage
FROM golang:1.25-alpine AS builder

WORKDIR /app

# Копируем go.mod и go.sum
COPY go.mod go.sum ./

# Скачиваем зависимости
RUN go mod download

# Копируем исходный код
COPY . .

# Компилируем приложение
    RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .
# Runtime stage
FROM alpine:latest

# Устанавливаем необходимые зависимости
RUN apk --no-cache add ca-certificates libc6-compat

WORKDIR /app

# Копируем скомпилированное приложение из builder stage
COPY --from=builder /app/main .

# Копируем static файлы
COPY --from=builder /app/static ./static

# Expose порт
EXPOSE 8081

# Команда запуска
CMD ["./main"]
