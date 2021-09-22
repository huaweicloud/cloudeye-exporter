
# 华为云 Exporter

[华为云](https://www.huaweicloud.com/)云监控的 Prometheus Exporter.


## 介绍
Prometheus是用于展示大型测量数据的开源可视化工具，在工业监控、气象监控、家居自动化和过程管理等领域也有着较广泛的用户基础。将华为云Cloudeye服务接入 prometheus后，您可以利用 prometheus更好地监控和分析来自 Cloudeye服务的数据。

## 环境准备
以Ubuntu 18.04系统和Prometheus 2.14.0版本为例
| Prometheus | prometheus-2.14.0.linux-amd64 |
| ------------ | ------------ |
| ECS | Ubuntu 18.04 |
| Ubuntu private ip | 192.168.0.xx |

## 安装配置cloudeye-exporter
1. 在ubuntu vm上安装cloudeye-exporter
```
git clone https://github.com/huaweicloud/cloudeye-exporter
cd cloudeye-exporter
```
2. 编辑clouds.yml文件配置公有云信息
```
global:
  port: ":8087"
auth:
auth_url: "https://iam.cn-north-1.myhwclouds.com/v3"
project_name: "cn-north-1"
access_key: ""
secret_key: ""
region: "cn-north-1"
```
注：默认的监控端口为8087.

3. 启动cloudeye-exporter
```
./cloudeye-exporter
```

## 安装配置prometheus接入cloudeye
1. 下载Prometheus (https://prometheus.io/download/)
```
$ wget https://github.com/prometheus/prometheus/releases/download/v2.14.0/prometheus-2.14.0.linux-amd64.tar.gz 
$ tar xzf prometheus-2.14.0.linux-amd64.tar.gz
$ cd prometheus-2.14.0.linux-amd64
```
2. 配置接入cloudeye exporter结点

   修改prometheus中的prometheus.yml文件配置。如下配置所示在scrape_configs下新增job_name名为’huaweicloud’的结点。其中targets中配置的是访问cloudeye-exporter服务的ip地址和端口号，services配置的是你想要监控的服务，比如SYS.VPC,SYS.RDS。
	```
	$ vi prometheus.yml
	global:
	  scrape_interval: 1m # Set the scrape interval to every 1 minute seconds. Default is every 1 minute.
	  scrape_timeout: 1m
	scrape_configs:
	  - job_name: 'huaweicloud'
		static_configs:
		- targets: ['192.168.0.xx:8087']
		params:
		  services: ['SYS.VPC,SYS.RDS']
	```
3. 启动prometheus监控华为云服务
```
./prometheus
```
 * 登录http://127.0.0.1:9090/graph
 * 查看指定指标的监控结果

