module github.com/huaweicloud/cloudeye-exporter

go 1.24

require (
	github.com/agiledragon/gomonkey/v2 v2.11.0
	github.com/golang/snappy v1.0.0
	github.com/huaweicloud/huaweicloud-sdk-go-v3 v0.1.110
	github.com/prometheus/client_golang v1.23.2
	github.com/stretchr/testify v1.11.1
	go.uber.org/zap v1.27.0
	google.golang.org/protobuf v1.36.8
	gopkg.in/natefinch/lumberjack.v2 v2.2.1
	gopkg.in/yaml.v3 v3.0.1
)

require (
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/munnerz/goautoneg v0.0.0-20191010083416-a7dc8b61c822 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/prometheus/client_model v0.6.2 // indirect
	github.com/prometheus/common v0.66.1 // indirect
	github.com/prometheus/procfs v0.16.1 // indirect
	github.com/tjfoc/gmsm v1.4.1 // indirect
	go.mongodb.org/mongo-driver v1.12.0 // indirect
	go.uber.org/multierr v1.10.0 // indirect
	go.yaml.in/yaml/v2 v2.4.2 // indirect
	golang.org/x/crypto v0.23.0 // indirect
	golang.org/x/sys v0.35.0 // indirect
	gopkg.in/ini.v1 v1.67.0 // indirect
)

replace (
	github.com/gogo/protobuf => github.com/gogo/protobuf v1.3.2
	golang.org/x/crypto => golang.org/x/crypto v0.0.0-20220525230936-793ad666bf5e
	golang.org/x/net => golang.org/x/net v0.10.0
)
