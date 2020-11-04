FROM golang:1.15.3-alpine as build
ADD . /src
ARG ARCH="amd64"
ARG OS="linux"
ENV CGO_ENABLED 0

RUN addgroup -S cloudeye && adduser -S cloudeye -G cloudeye
RUN apk --no-cache add ca-certificates git && \
    rm -Rf /var/cache/apk/*
RUN cd /src && GOARCH=${ARCH} GOOS=${OS} go build -o cloudeye-exporter

FROM scratch
WORKDIR /app
COPY --from=build /etc/passwd /etc/passwd
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=build /src/cloudeye-exporter /app/

USER cloudeye
EXPOSE 8087

ENTRYPOINT [ "/app/cloudeye-exporter" ]
