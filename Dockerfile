# DEV runner & autocompiller on file change
FROM golang:1.24.0-bookworm AS develop

ARG TZ=Europe/Kyiv

RUN apt-get update && apt-get install -y tzdata && \
    ln -fs /usr/share/zoneinfo/$TZ /etc/localtime && \
    dpkg-reconfigure -f noninteractive tzdata

ARG USER=gopher
ARG GROUPNAME=gopher
ARG UID=1000
ARG GID=1000

RUN addgroup \
    --gid "$GID" \
    "$GROUPNAME" \
&&  adduser \
    --disabled-password \
    --gecos "" \
    --ingroup "$GROUPNAME" \
    --uid "$UID" \
    $USER

USER gopher
WORKDIR /app

RUN go install github.com/air-verse/air@v1.61.7

CMD ["air", "-c", ".air.toml"]

# PROD builder
FROM golang:1.24.0-alpine3.21 AS prod_builder

RUN apk add --no-cache build-base

RUN mkdir /workdir
WORKDIR /workdir

ENV CGO_ENABLED=0

COPY . .

RUN go build  \
    -ldflags="-s -w" \
    -o /usr/local/bin/web \
    ./cmd/web/web.go

RUN if ldd /usr/local/bin/web; then echo "web: binary is not static"; exit 1; fi

# PROD runner
FROM alpine:3.21 AS prod_runner

RUN apk add --no-cache tzdata
RUN apk add --no-cache wget

ENV TZ=Europe/Kyiv

ARG USER=gopher
ARG GROUPNAME=gopher
ARG UID=1000
ARG GID=1000

RUN addgroup \
    --gid "$GID" \
    "$GROUPNAME" \
&&  adduser \
    --disabled-password \
    --gecos "" \
    --ingroup "$GROUPNAME" \
    --uid "$UID" \
    $USER

WORKDIR /home/gopher
USER gopher

COPY --link --from=prod_builder /usr/local/bin/web /usr/local/bin/web

ENTRYPOINT ["web"]
