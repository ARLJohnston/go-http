FROM golang:1.23 AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . /app
RUN CGO_ENABLED=0 GOOS=linux go build -o /entrypoint

FROM alpine:3.20
WORKDIR /
COPY --from=builder /entrypoint /entrypoint
EXPOSE 3000
ENTRYPOINT ["/entrypoint"]
