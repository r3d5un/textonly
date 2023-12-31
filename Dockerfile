# BASE IMAGE
FROM golang:1.21-alpine as base

WORKDIR /app

COPY . /app
RUN go mod download

COPY *.go ./

RUN CGO_ENABLED=0 GOOS=linux go build ./cmd/web

# RUNNER
FROM alpine:latest
WORKDIR /app
COPY --from=base /app/web .
COPY ./cmd/web/config/config.yaml ./config.yaml

EXPOSE 8080
CMD [ "./web" ]

