FROM golang:1.26.3-alpine3.23 AS base
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY *.go ./
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/open-dpp-api

FROM alpine:3.23.5 AS production
RUN addgroup -g 1000 itergroup && \
    adduser -u 1000 -G itergroup -s /bin/sh -D iteruser
COPY --from=base --chown=iteruser:itergroup /app/open-dpp-api /iter/open-dpp-api

EXPOSE 7000

USER iteruser:itergroup

ENTRYPOINT [ "/iter/open-dpp-api" ]
