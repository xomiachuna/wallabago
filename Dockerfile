FROM golang:1.24.1
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY *.go ./
RUN GOOS=linux CGO_ENABLED=0 go build -o /app/wallabago
CMD ["/app/wallabago"]
