# FROM golang:1.22 AS builder
# WORKDIR /app
# COPY go.mod go.sum ./
# RUN go mod download
# COPY *.go ./

# RUN CGO_ENABLED=0 GOOS=linux go build -o main -ldflags '-extldflags "-static"' .

# FROM builder AS run-tests

# FROM scratch
FROM golang:1.22.5
LABEL org.opencontainers.image.source=https://github.com/arljohnston/go-http
WORKDIR /app
COPY ./*.go .
RUN go mod init main
RUN go mod tidy
RUN go build -o main
# RUN go test -v ./...
ENTRYPOINT ["/app/main"]
# COPY --from=builder /app /
# CMD ["/main"]
