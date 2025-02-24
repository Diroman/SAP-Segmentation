# erp/Dockerfile
FROM golang:1.21-alpine AS builder

WORKDIR /app

RUN go mod init github.com/user/erp  # Замените на ваш модуль, если нужен
RUN go get github.com/gorilla/mux

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o erp-server ./main.go

FROM alpine:3.18
WORKDIR /app
COPY --from=builder /app/erp-server .
RUN apk add --no-cache netcat-openbsd
EXPOSE 8080
ENTRYPOINT ["./erp-server"]