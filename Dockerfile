FROM golang:alpine AS base
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY *.go ./
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/open-dpp-api

FROM alpine AS production
COPY --from=base /app/open-dpp-api /iter/open-dpp-api

EXPOSE 7000

USER iter:iter

ENTRYPOINT [ "/iter/open-dpp-api" ]
