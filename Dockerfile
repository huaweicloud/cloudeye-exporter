FROM golang:1.19-alpine AS build

WORKDIR /src

RUN apk add --no-cache ca-certificates=20230506-r0

COPY go.mod go.sum ./
RUN go mod download

COPY . .

ARG TARGETOS=linux
ARG TARGETARCH=amd64

ENV CGO_ENABLED=0
ENV GOOS=$TARGETOS
ENV GOARCH=$TARGETARCH

RUN go build -trimpath -ldflags="-s -w" -o /out/cloudeye-exporter .

FROM gcr.io/distroless/static-debian11:nonroot

WORKDIR /app

COPY --from=build /out/cloudeye-exporter /usr/local/bin/cloudeye-exporter
COPY --from=build /src/metric.yml /app/metric.yml
COPY --from=build /src/logs.yml /app/logs.yml
COPY --from=build /src/endpoints.yml /app/endpoints.yml
COPY --from=build /src/i18n.json /app/i18n.json
COPY --from=build /src/unit_standard_transform.json /app/unit_standard_transform.json

USER nonroot:nonroot

EXPOSE 8087

ENTRYPOINT ["/usr/local/bin/cloudeye-exporter"]
CMD ["-config", "/app/clouds.yml"]
