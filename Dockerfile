FROM golang:1.25

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o /usr/local/bin/app ./cmd/main.go

EXPOSE 8080

ENTRYPOINT ["/usr/local/bin/app"]

# # Stage 1 – build
# FROM golang:1.25-alpine AS builder

# # Устанавливаем необходимые пакеты, если они нужны (например, для cgo)
# RUN apk add --no-cache git ca-certificates

# # Рабочий каталог внутри контейнера
# WORKDIR /app

# # Копируем файлы зависимостей и скачиваем их
# COPY go.mod go.sum ./
# RUN go mod download

# # Копируем остальной исходный код и компилируем приложение
# COPY . .
# RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
#     go build -trimpath -o /usr/local/bin/app ./cmd/main.go  # ← замените на ваш пакет

# # Stage 2 – runtime
# FROM alpine:3.20

# # Копируем сертификаты (если нужно HTTPS)
# COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# # Добавляем исполняемый файл из builder‑stage
# COPY --from=builder /usr/local/bin/app /usr/local/bin/app

# # Опционально: создаём пользователя без root‑прав (безопаснее)
# RUN addgroup -S app && adduser -S app -G app
# USER app

# # Указываем порт, который будет слушать ваш сервис
# EXPOSE 8080

# # Команда запуска
# ENTRYPOINT ["/usr/local/bin/app"]