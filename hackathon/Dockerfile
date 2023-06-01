FROM golang:1.18 as build

WORKDIR /go/src/app
COPY . .
EXPOSE 8080
CMD ["go", "run", "main.go"]