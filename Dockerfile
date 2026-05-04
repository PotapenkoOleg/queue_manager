# https://hub.docker.com/_/golang
# https://hub.docker.com/hardened-images/catalog/dhi/debian-base
FROM golang:1.26 AS builder

WORKDIR /usr/src/app

# pre-copy/cache go.mod for pre-downloading dependencies and only redownloading them in subsequent builds if they change
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -v -o /usr/local/bin/app ./...

FROM dhi/debian-base:latest

# Switch to the existing nonroot user
USER nonroot

RUN apt-get update && apt-get

WORKDIR /app

COPY --from=builder /usr/local/bin/app /app

CMD ["app"]


