FROM golang:1.14.10-alpine as build

WORKDIR /src
ENV CGO_ENABLED=0
COPY . .
ARG ARCH="amd64"
ARG OS="linux"

RUN addgroup -S cloudeye && adduser -S cloudeye -G cloudeye
RUN apk --no-cache add ca-certificates git && \
    rm -Rf /var/cache/apk/*
RUN GOARCH=${ARCH} GOOS=${OS} go build -o /go/bin/cloudeye-exporter .

FROM scratch

COPY --from=build /etc/passwd /etc/passwd
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=build /go/bin/cloudeye-exporter /

USER cloudeye
EXPOSE 8087

ENTRYPOINT [ "/cloudeye-exporter" ]
