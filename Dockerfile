FROM golang:1.25-alpine AS builder

WORKDIR /app

# Install build dependencies for CGO (sqlite)
RUN apk add --no-cache gcc musl-dev

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=1 go build -o url-shortener .

FROM alpine:latest

WORKDIR /app

# Install ca-certificates for HTTPS and libc for CGO binary
RUN apk add --no-cache ca-certificates libc6-compat

COPY --from=builder /app/url-shortener .
COPY --from=builder /app/static ./static

EXPOSE 4343

CMD ["./url-shortener"]
