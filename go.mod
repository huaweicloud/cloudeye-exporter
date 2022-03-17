module github.com/huaweicloud/cloudeye-exporter

go 1.14

require (
	github.com/cihub/seelog v0.0.0-20191126193922-f561c5e57575
	github.com/huaweicloud/huaweicloud-sdk-go-v3 v0.0.65
	github.com/prometheus/client_golang v1.7.1
	github.com/prometheus/common v0.17.0 // indirect
	gopkg.in/yaml.v2 v2.3.0
)

replace (
	github.com/go-kit/kit => github.com/go-kit/kit v0.9.0
	github.com/gogo/protobuf => github.com/gogo/protobuf v1.3.2
	golang.org/x/crypto => golang.org/x/crypto v0.0.0-20200604202706-70a84ac30bf9
	golang.org/x/text => golang.org/x/text v0.3.5
)
