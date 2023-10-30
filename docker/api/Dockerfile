FROM golang:1.21-alpine AS builder

RUN apk update && apk upgrade && apk add --no-cache ca-certificates
RUN update-ca-certificates

WORKDIR /app
COPY src/go.mod src/go.sum ./
RUN go mod download && go mod verify

COPY src .
RUN CGO_ENABLED=0 GOOS=linux go build -a -ldflags '-extldflags "-static"' -o main cmd/api/main.go

FROM scratch
WORKDIR /
COPY --from=builder /app/main /app
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
CMD ["/app"]
