# binary build image
FROM golang:1.24.0-bullseye AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -trimpath -ldflags "-s -w" -o app ./cmd

# deploy image
FROM debian:stretch-slim as deploy
RUN apt-get update

COPY --from=builder /app/app .

CMD ["./app"]

### dev image
FROM golang:1.24.0 as dev-auth
WORKDIR /app
EXPOSE 80
CMD ["go", "run", "./cmd", "80"]