# URL Shortener with Analytics

Простой сервис сокращения ссылок на Go + Gin, с поддержкой аналитики переходов и возможностью агрегации по дням, месяцам и User-Agent.

## Функциональность

POST `/shorten` — создать новую короткую ссылку

GET `/s/{short_url}` — перенаправление на оригинальный URL

GET `/analytics/{short_url}` — получить аналитику переходов

### Аналитика включает:

Общее количество переходов

Группировку по дням, месяцам, или User-Agent

Фильтрацию по периоду времени (`from` / `to`)

## Архитектура
```
/cmd
  /app/main.go          — запуск сервера
/internal
  /models               — структуры данных (URL, Click, Analytics)
  /repository            — работа с PostgreSQL
  /service               — бизнес-логика
  /handler               — HTTP-эндпоинты (Gin)
```

#### Слои:

- Handler — принимает HTTP-запросы, вызывает сервис.

- Service — инкапсулирует бизнес-логику.

- Repository — выполняет SQL-запросы к PostgreSQL (через dbpg).

## Зависимости

- Go 1.22+

- Gin

- PostgreSQL

- wb-go/wbf (dbpg, retry, ginext, zlog)