# Stage 1: сборка приложения
FROM golang:1.23.5-alpine AS builder

# Устанавливаем рабочую директорию внутри контейнера
WORKDIR /app

# Копируем go.mod и go.sum и загружаем зависимости
COPY go.mod go.sum ./
RUN go mod download

# Копируем исходный код
COPY . .

RUN GO111MODULE=on CGO_ENABLED=0 go install github.com/swaggo/swag/v2/cmd/swag@latest

RUN swag init -g cmd/app/main.go --output docs/swagger

# Добавляем зависимость в go.mod
RUN go get github.com/swaggo/swag/v2@latest

# Собираем бинарник. Используем опцию -trimpath для уменьшения размера.
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -o /app/bin/app ./cmd/app/main.go

# Stage 2: финальный образ
FROM alpine:latest

# Создаем рабочую директорию для приложения
WORKDIR /app

# Копируем бинарник из сборочного образа
COPY --from=builder /app/bin/app /app/app

# Копируем файл миграций (если они требуются на старте)
COPY --from=builder /app/migrations /app/migrations

# Устанавливаем переменные окружения (можно переопределить через docker-compose)
ENV PORT=8080

# Пробрасываем порт
EXPOSE 8080

# Запускаем приложение
CMD ["/app/app"]
