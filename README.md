# news

## done

1. [x] htmx+tailwind on frontend
2. [x] crud
3. [x] data validation
4. [x] pagination
5. [x] search
6. [x] resent posts (posts already sorted by creation time)
7. [x] development environment setup (air)
8. [x] makefile
9. [x] tests

## requirements

- `go` >= 1.24.0
- `make`
- `docker` (tested on 28.1.1 linux/amd64)

## how to run

- `git clone https://github.com/mineroot/news`
- `make test` (may take a while on the first run to download docker image)
- `docker compose -f compose.yaml -f compose.prod.yaml up --build -d --wait`
- http://127.0.0.1:8080/

## development

- optionally, create `.env` to override `default.env`, e.g.`echo "APP_ENV=dev\nAPP_LOG_LEVEL=debug" > .env`
- `docker compose up --build -d`
- http://127.0.0.1:8080/
- on every `*.go` file change docker container compile and start new binary
