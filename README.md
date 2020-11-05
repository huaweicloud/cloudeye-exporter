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

## Building and Running with Docker

```shell script
cp clouds.yml clouds-custom.yml
# Edit clouds-custom.yml
docker build -t you/cloudeye-exporter .
docker run -it -p 8087:8087 --rm -v `pwd`/clouds-custom.yml:/app/clouds.yml you/cloudeye-exporter
```

## Using the Helm Chart

```shell script
# Fill the blanks
export REPO=you/cloudeye-exporter
export OS_DOMAIN_NAME=
export OS_USERNAME=
export OS_PASSWORD=
export OS_PROJECT_NAME=
export OS_AUTH_URL=
export OS_REGION_NAME=

cat <<EOF | helm install my -f - charts/cloudeye-exporter
cloudConfig:
  auth:
    auth_url: ${OS_AUTH_URL}
    project_name: ${OS_PROJECT_NAME}
    user_name: ${OS_USERNAME}
    password: ${OS_PASSWORD}
    domain-name: ${OS_DOMAIN_NAME}
    region: ${OS_REGION_NAME}

image:
  repository: ${REPO}
  pullPolicy: IfNotPresent
  tag: "1.1.2"
#ingress:
#  enabled: true
#  annotations:
#    kubernetes.io/ingress.class: nginx
#  hosts:
#    - host: your.dns.name.com
#      paths: [ "/" ]
#  tls:
#    - secretName: your-secret
#      hosts:
#        - your.dns.name.com
#EOF
  
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

auth:
  auth_url: "https://iam.xxx.yyy.com/v3"
  project_name: "{project_name}"
  access_key: "{access_key}"
  secret_key: "{secret_key}"
  region: "{region}"

```
or

```
auth:
  auth_url: "https://iam.xxx.yyy.com/v3"
  project_name: "{project_name}"
  user_name: "{username}"
  password: "{password}"
  region: "{region}"
  domain_name: "{domain_name}"

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
