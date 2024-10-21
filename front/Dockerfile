FROM golang:1.23 AS builder

WORKDIR /app
COPY . /app
RUN go install github.com/a-h/templ/cmd/templ@latest
RUN templ generate
RUN CGO_ENABLED=0 GOOS=linux go build -o front .

FROM alpine:3.20
COPY --from=builder /app /
EXPOSE 3000
ENTRYPOINT ["./front"]
