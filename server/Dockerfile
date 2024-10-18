FROM golang:1.23 as builder

WORKDIR /app

ADD . /app

RUN go build -o bin/server server/main.go

# Use alpine rather than scratch so can attach for debugging as needed
FROM alpine:latest

COPY --from=builder /app/bin/server .

EXPOSE 8080

CMD ["./server"]
