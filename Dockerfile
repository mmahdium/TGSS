# https://docs.docker.com/guides/golang/build-images/

# syntax=docker/dockerfile:1

# Build the application from source
FROM golang:1.26.3-trixie AS build-stage

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY cmd /app/cmd/
COPY internal /app/internal/

RUN GOOS=linux go build -ldflags="-s -w" -tags=release -o /TGSS ./cmd/server/main.go 

# Deploy the application binary into a lean image
# TODO: Move to alpine
FROM gcr.io/distroless/base-debian13 AS build-release-stage

WORKDIR /

COPY --from=build-stage /TGSS /TGSS


EXPOSE 3000

ENV GIN_MODE=release

ENTRYPOINT ["/TGSS"]