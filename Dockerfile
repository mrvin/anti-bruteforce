## Build
FROM golang:1.22.4-alpine AS build

MAINTAINER mrvin <v.v.vinogradovv@gmail.com>

RUN apk add --update make

WORKDIR  /app

# Copy the code into the container.
COPY main.go Makefile ./
COPY internal internal
COPY pkg pkg

# Copy and download dependency using go mod.
COPY go.mod go.sum ./
RUN go mod download

RUN make build

## Deploy
FROM scratch

WORKDIR /

COPY --from=build ["/app/bin/anti-bruteforce", "/"]

ENTRYPOINT ["/anti-bruteforce"]
