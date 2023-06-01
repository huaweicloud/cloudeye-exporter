module github.com/huaweicloud/cloudeye-exporter

go 1.16

require (
	github.com/huaweicloud/huaweicloud-sdk-go-v3 v0.1.35
	github.com/prometheus/client_golang v1.12.2
	go.uber.org/zap v1.24.0
	gopkg.in/natefinch/lumberjack.v2 v2.2.1
	gopkg.in/yaml.v3 v3.0.1
)

replace (
	github.com/gogo/protobuf => github.com/gogo/protobuf v1.3.2
	golang.org/x/crypto => golang.org/x/crypto v0.0.0-20220525230936-793ad666bf5e
)
