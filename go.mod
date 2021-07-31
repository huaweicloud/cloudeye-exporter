module github.com/huaweicloud/cloudeye-exporter

go 1.14

require (
	github.com/huaweicloud/golangsdk v0.0.0-20200703083934-0708c52f1c75
	github.com/prometheus/client_golang v1.7.1
	github.com/prometheus/common v0.17.0
	gopkg.in/yaml.v2 v2.3.0
)

replace (
	github.com/go-kit/kit => github.com/go-kit/kit v0.9.0
	github.com/gogo/protobuf => github.com/gogo/protobuf v1.3.2
	golang.org/x/crypto => golang.org/x/crypto v0.0.0-20200604202706-70a84ac30bf9
	golang.org/x/text => golang.org/x/text v0.3.5
)
