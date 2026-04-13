FROM golang:1.24 as builder

ENV CGO_ENABLED=0

RUN mkdir "/opt/build"

WORKDIR /opt/build

COPY . /opt/build

RUN go build -o /tmp/weather-comparison


FROM debian:13-slim

RUN apt update && apt -y install ca-certificates

LABEL org.opencontainers.image.source=https://github.com/chilledornaments/weather-comparison

RUN mkdir /opt/weather

WORKDIR /opt/weather

COPY --from=builder /tmp/weather-comparison /opt/weather/weather-comparison

ENTRYPOINT ["./weather-comparison"]
