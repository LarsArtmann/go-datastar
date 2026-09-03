# Build from the REPOSITORY ROOT (the example is part of the root Go module,
# so go.mod must be in the build context):
#
# docker build -f example/Dockerfile -t go-datastar-example .
# docker compose -f example/docker-compose.yml up
FROM golang:1.26.7-alpine AS build
ENV GOEXPERIMENT=jsonv2 GOWORK=off
WORKDIR /src
COPY go.mod go.sum ./
COPY static/go.mod ./static/
RUN go mod download
COPY . .
RUN go build -o /demo ./example

FROM alpine:3.20
COPY --from=build /demo /demo
EXPOSE 8765
ENTRYPOINT ["/demo"]
