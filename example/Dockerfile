FROM golang:1.26.7-alpine AS build
WORKDIR /src
COPY . .
ENV GOEXPERIMENT=jsonv2 GOWORK=off
RUN go build -o /demo ./example

FROM alpine:3.20
COPY --from=build /demo /demo
EXPOSE 8765
ENTRYPOINT ["/demo"]
