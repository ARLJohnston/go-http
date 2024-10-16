FROM golang:1.23 AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY *.go ./

RUN CGO_ENABLED=0 GOOS=linux go build -o main -ldflags '-extldflags "-static"' .

# FROM builder AS run-tests

FROM scratch
LABEL org.opencontainers.image.source=https://github.com/arljohnston/go-http
COPY --from=builder /app /
CMD ["/main"]
