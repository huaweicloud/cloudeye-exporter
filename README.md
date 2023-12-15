# cloudeye-exporter

Prometheus cloudeye exporter for [Huaweicloud](https://www.huaweicloud.com/).

Note: The plug-in is applicable only to the Huaweicloud regions.

[中文](./README_cn.md)

## Download
```
$ git clone https://github.com/huaweicloud/cloudeye-exporter
```

## (Option) Building The Discovery with Exact steps on clean Ubuntu 16.04 
```
$ wget https://dl.google.com/go/go1.17.6.linux-amd64.tar.gz
$ sudo tar -C /usr/local -xzf go1.17.6.linux-amd64.tar.gz
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


To avoid ak&sk leaks, you can input your ak&sk through the command line with '-s' param like this
```shell
./cloudeye-exporter -s true
```
To startup the cloudeye-exporter using script, here is an example
```shell
#!/bin/bash
# Don't write plain text ak&sk in script.
# Encrypt the ak&sk instead and use your own decrypt function to assign the return value.
huaweiCloud_AK=your_decrypt_function("your encrypted ak")
huaweiCloud_SK=your_decrypt_function("your encrypted sk")
$(./cloudeye-exporter -s true<<EOF
$huaweiCloud_AK $huaweiCloud_SK
EOF)
```

## Help
```
Usage of ./cloudeye-exporter:
  -config string
        Path to the cloud configuration file (default "./clouds.yml") 
```

## Example of config file(clouds.yml)
The "URL" value can be get from [Identity and Access Management (IAM) endpoint list (Internal)](https://developer.huaweicloud.com/intl/en-us/endpoint?IAM) and [Identity and Access Management (IAM) endpoint list (China)](https://developer.huaweicloud.com/endpoint?IAM).
```
global:
  prefix: "huaweicloud"
  port: ":8087"
  metric_path: "/metrics"
  scrape_batch_size: 300

auth:
  auth_url: "https://iam.{region_id}.myhuaweicloud.com/v3"
  project_name: "{project_name}"
  access_key: "{access_key}"
  secret_key: "{secret_key}"
  region: "{region}"

```

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
