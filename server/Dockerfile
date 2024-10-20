FROM golang:1.23 AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY *.go ./

RUN CGO_ENABLED=0 GOOS=linux go build -o main .
FROM alpine:3.20
COPY --from=builder /app /
EXPOSE 50051
CMD ["/main"]
