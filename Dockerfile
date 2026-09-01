# syntax=docker/dockerfile:1

FROM golang:1.26-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /bin/vicomp ./cmd/vicomp

# --- dev: live-reloading via air (used by docker compose) ---
FROM golang:1.26-alpine AS dev
RUN apk add --no-cache git
RUN go install github.com/air-verse/air@v1.67.4
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
EXPOSE 8080
CMD ["air", "-c", ".air.toml"]

# --- prod: minimal runtime image (default target for a plain docker build) ---
FROM alpine:3.20 AS prod
RUN adduser -D -u 10001 app
COPY --from=build /bin/vicomp /bin/vicomp
COPY --from=build /src/migrations /migrations
COPY --from=build /src/templates /templates
COPY --from=build /src/static /static
USER app
EXPOSE 8080
ENTRYPOINT ["/bin/vicomp"]
