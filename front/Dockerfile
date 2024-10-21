FROM golang:1.23

WORKDIR /app
COPY . /app
RUN go install github.com/a-h/templ/cmd/templ@latest
RUN templ generate
RUN CGO_ENABLED=0 GOOS=linux go build -o main .
EXPOSE 3000
ENTRYPOINT ["./main"]
