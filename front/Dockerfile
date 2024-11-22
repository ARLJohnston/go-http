FROM --platform=$BUILDPLATFORM golang:1.23 AS builder
ARG TARGETOS
ARG TARGETARCH
WORKDIR /app
COPY . /app
RUN go install github.com/a-h/templ/cmd/templ@latest
RUN templ generate
RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -o front .

FROM alpine:3.20
LABEL org.opencontainers.image.source=https://github.com/arljohnston/go-microservice-front
COPY --from=builder /app /
EXPOSE 3000
ENTRYPOINT ["./front"]
