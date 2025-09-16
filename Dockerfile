# Многоэтапная сборка Docker-образа для планировщика задач

# Этап сборки
FROM golang:1.24-alpine AS builder

# Установка зависимостей для CGO
RUN apk add --no-cache gcc musl-dev

# Установка рабочей директории
WORKDIR /app

# Копирование файлов go.mod и go.sum
COPY go.mod go.sum ./

# Загрузка зависимостей
RUN go mod download

# Копирование исходного кода
COPY . .

# Компиляция приложения для Linux
RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -o go_final_project .

# Финальный этап
FROM alpine:latest

# Установка зависимостей для работы с SQLite
RUN apk add --no-cache ca-certificates

# Установка рабочей директории
WORKDIR /root/

# Копирование скомпилированного бинарного файла из этапа сборки
COPY --from=builder /app/go_final_project .

# Копирование веб-файлов
COPY --from=builder /app/web ./web

# Установка порта по умолчанию
EXPOSE 7540

# Установка переменных окружения по умолчанию
ENV TODO_PORT=7540
ENV TODO_DBFILE=scheduler.db
ENV TODO_PASSWORD=

# Команда запуска приложения
CMD ["./go_final_project"]