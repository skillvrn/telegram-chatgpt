FROM golang:1.17 as builder
WORKDIR /app
COPY go.mod go.sum ./
COPY bot.go .
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o bot

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/bot .
CMD ["./bot"]
