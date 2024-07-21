FROM golang:1.22.4 AS builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o app .

FROM alpine:3.20

COPY --from=builder /app/app /app/app
COPY --from=builder /app/db /app/db

WORKDIR /app

CMD ["./app"]
