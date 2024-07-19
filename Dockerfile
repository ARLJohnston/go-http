FROM golang AS builder
WORKDIR /app
COPY . .
COPY . ../go.mod
RUN ls
RUN CGO_ENABLED=0 GOOS=linux go build -o hello-world -ldflags '-extldflags "-static"' .

FROM builder AS run-tests
RUN go test -v ./...

FROM scratch
LABEL org.opencontainers.image.source https://github.com/arljohnston/ghcr-go-test
COPY --from=builder /app /
CMD ["/hello-world"]
