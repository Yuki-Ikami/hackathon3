FROM golang:1.18 as build

WORKDIR /app
COPY . .
RUN go mod download
RUN go build -o /app
CMD ["/app"]

EXPOSE 8080