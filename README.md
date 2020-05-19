# cloudeye-exporter

Prometheus cloudeye exporter for [Huaweicloud](https://www.huaweicloud.com/).

## Download
```
$ git clone https://github.com/huaweicloud/cloudeye-exporter
```

## (Option) Building The Discovery with Exact steps on clean Ubuntu 16.04 
```
$ wget https://dl.google.com/go/go1.12.5.linux-amd64.tar.gz
$ sudo tar -C /usr/local -xzf go1.12.5.linux-amd64.tar.gz
$ export PATH=$PATH:/usr/local/go/bin # You should put in your .profile or .bashrc
$ go version # to verify it runs and version #

$ go get github.com/huaweicloud/cloudeye-exporter
$ cd ~/go/src/github.com/huaweicloud/cloudeye-exporter
$ go build
```

## Usage
```
 ./cloudeye-exporter  -config=clouds.yml
```

The default port is 8087, default config file location is ./clouds.yml.

Visit metrics in http://localhost:8087/metrics?services=SYS.VPC,SYS.ELB


## Help
```
Usage of ./cloudeye-exporter:
  -config string
        Path to the cloud configuration file (default "./clouds.yml")
  -debug
        If debug the code.
 
```

## Example of config file(clouds.yml)
The "URL" value can be get from [Identity and Access Management (IAM) endpoint list](https://developer.huaweicloud.com/en-us/endpoint).
```
global:
  prefix: "huaweicloud"
  port: ":8087"
  metric_path: "/metrics"
  retrieve_offset: "0"
  cloudeye_timestamp: false
  ignore_empty_datapoints: false

auth:
  auth_url: "https://iam.cn-north-1.myhwclouds.com/v3"
  project_name: "cn-north-1"
  access_key: "xdfsdfsdfsdfsdfsdf"
  secret_key: "xsdfsddfsdfsdfsdfsdfsdfsdfsdfsdfsdfsdfdsfsd"
  region: "cn-north-1"

```
or

```
auth:
  auth_url: "https://iam.cn-north-1.myhwclouds.com/v3"
  project_name: "cn-north-1"
  user_name: "username"
  password: "password"
  region: "cn-north-1"
  domain_name: "domain_name"

```

Notes :
* `global.retrieve_offset` (default "0") is an offset (example "-5m") to ask Cloudeye for some older metrics.
* `global.cloudeye_timestamp` (default false) allows Cloudeye Exporter to send metrics with their Cloudeye timestamp
* `global.ignore_empty_datapoints` (default false), when set, will ignore empty datapoints (no warnings)

## Prometheus Configuration
The huaweicloud exporter needs to be passed the address as a parameter, this can be done with relabelling.

Example config:

```
global:
  scrape_interval: 1m # Set the scrape interval to every 1 minute seconds. Default is every 1 minute.
  scrape_timeout: 1m
scrape_configs:
  - job_name: 'huaweicloud'
    static_configs:
    - targets: ['10.0.0.10:8087']
    params:
      services: ['SYS.VPC,SYS.ELB']
```
