FROM golang:1.18 as build

WORKDIR /go/src/app
COPY . .

RUN go mod download
RUN go build -o /go/bin/app

EXPOSE 8080
CMD ["/go/bin/app"]
