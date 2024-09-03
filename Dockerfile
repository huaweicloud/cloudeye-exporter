# 使用基础镜像
FROM golang:1.19 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 go build -mod=vendor -v -o cloudeye-exporter

FROM ubuntu:latest

COPY --from=builder /app/cloudeye-exporter /usr/local/bin/cloudeye-exporter

COPY . /root/

# 设置工作目录
WORKDIR /root

CMD ["cloudeye-exporter", "-config", "clouds.yml"]