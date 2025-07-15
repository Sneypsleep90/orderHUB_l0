FROM golang:1.23.4 AS builder


WORKDIR /app


COPY go.mod go.sum ./


RUN go mod download


COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o orderHub_L0 ./cmd/server/main.go

FROM alpine:3.20


RUN apk add --no-cache ca-certificates


WORKDIR /app

COPY --from=builder /app/orderHub_L0 .

COPY .env .


EXPOSE 8081

CMD ["./orderHub_L0"]