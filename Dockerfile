## Build
FROM golang:1.25.5-alpine AS dev

LABEL maintainer="mrvin v.v.vinogradovv@gmail.com"

RUN apk add --update make && apk add tzdata

WORKDIR  /app

# Copy the code into the container.
COPY cmd/anti-bruteforce cmd/anti-bruteforce
COPY internal internal
COPY pkg pkg
COPY Makefile ./

# Copy and download dependency using go mod.
COPY go.mod go.sum ./
RUN go mod download

RUN make build

RUN mkdir /var/log/anti-bruteforce/

ENV TZ=Europe/Moscow

EXPOSE 50051

ENTRYPOINT ["/app/bin/anti-bruteforce"]

## Deploy
FROM scratch AS prod

LABEL maintainer="mrvin v.v.vinogradovv@gmail.com"

WORKDIR /

COPY --from=dev ["/var/log/anti-bruteforce/", "/var/log/anti-bruteforce/"]
COPY --from=dev ["/usr/share/zoneinfo", "/usr/share/zoneinfo"]
COPY --from=dev ["/app/bin/anti-bruteforce", "/usr/local/bin/anti-bruteforce"]

ENV TZ=Europe/Moscow

EXPOSE 50051

ENTRYPOINT ["/usr/local/bin/anti-bruteforce"]
