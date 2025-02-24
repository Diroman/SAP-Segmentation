FROM golang:1.22-alpine AS builder

WORKDIR /app

RUN apk add --no-cache make git

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Сборка приложения с использованием Makefile
RUN make build

FROM alpine:3.18

COPY --from=builder /app/sap_segmentationd /app/
WORKDIR /app

RUN apk --no-cache add tzdata && \
    mkdir -p /log && \
    chmod 755 /log

ENTRYPOINT ["./sap_segmentationd"]