FROM golang:1.23.2-alpine AS builder

WORKDIR /app
COPY . .
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o pontogo ./app/cmd/main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/pontogo .
COPY --from=builder /app/.env_example .env

ENTRYPOINT ["./pontogo"] 