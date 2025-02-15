
FROM golang:1.22.5-alpine AS builder
WORKDIR /app
COPY . .
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o urlshortener ./cmd/main.go


FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/urlshortener .
EXPOSE 8080
CMD ["./urlshortener"]
