FROM golang:alpine AS build-env
ADD . /src
RUN cd /src && go build -o cloudeye-exporter

FROM alpine
WORKDIR /app
COPY --from=build-env /src/cloudeye-exporter /app/
ENTRYPOINT ./cloudeye-exporter
