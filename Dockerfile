FROM golang:1.24-alpine

WORKDIR /app

# Установка зависимостей для postgres
RUN apk add --no-cache postgresql-client

# Копирование и загрузка зависимостей
COPY go.mod go.sum ./
RUN go mod download

# Копирование исходного кода
COPY . .

# Сборка приложения
RUN go build -o main ./cmd/main.go

# Открываем порты для HTTP и Prometheus
EXPOSE 8080 9000

# Запуск приложения
CMD ["./main"] 