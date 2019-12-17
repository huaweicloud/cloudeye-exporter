FROM golang:1.13-alpine as build
RUN apk --no-cache add ca-certificates && \
    rm -Rf /var/cache/apk/*
WORKDIR /app
COPY . .
RUN GOPROXY=https://goproxy.cn CGO_ENABLED=0 go build -o cloudeye-exporter main.go 

FROM alpine:3.8
RUN addgroup -S app && adduser -S app -G app
USER app
EXPOSE 8080
ENTRYPOINT [ "/usr/local/bin/cloudeye-exporter" ]
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=build /app/cloudeye-exporter /usr/local/bin/cloudeye-exporter