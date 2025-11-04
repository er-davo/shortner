FROM golang:alpine AS builder

RUN apk add --no-cache gcc musl-dev

WORKDIR /url-shortener

COPY app/go.mod /url-shortener/

RUN go mod download

COPY app/ /url-shortener/

RUN go build -o build/main cmd/main.go

FROM alpine:latest AS runner

WORKDIR /app

COPY --from=builder /url-shortener/build/main /app/

COPY /config/config.yaml /app/config.yaml
COPY public/ /app/public/
COPY /migrations /app/migrations

ENV CONFIG_PATH=/app/config.yaml
ENV APP_MIGRATION_DIR=/app/migrations

CMD [ "/app/main" ]