FROM golang:1.20.6-alpine

WORKDIR /app

COPY . /app
RUN go mod download

COPY *.go ./

RUN CGO_ENABLED=0 GOOS=linux go build ./cmd/web

EXPOSE 4000

CMD [ "./web" ]

