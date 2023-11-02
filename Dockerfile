FROM golang:1.21-alpine

WORKDIR /app

COPY . /app
RUN go mod download

COPY *.go ./

RUN CGO_ENABLED=0 GOOS=linux go build ./cmd/web

EXPOSE 8080

CMD [ "./web" ]

