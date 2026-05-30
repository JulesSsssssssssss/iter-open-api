FROM golang:alpine AS base
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY *.go ./
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/open-dpp-api

EXPOSE 7000
CMD [ "/app/open-dpp-api" ]
