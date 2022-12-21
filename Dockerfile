# syntax=docker/dockerfile:1

FROM golang:alpine

WORKDIR /app

RUN mkdir static
COPY static/* ./static/
COPY go.mod ./
COPY go.sum ./
RUN go mod download

COPY *.go ./

RUN go build -o /nzbsearch

EXPOSE 8080

CMD [ "/nzbsearch" ]
