FROM golang:1.18 as build

WORKDIR /go/src/app
COPY . .

CMD ["go", "run", "main.go"]

EXPOSE 8080