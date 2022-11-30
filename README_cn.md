
# 华为云 Exporter

[华为云](https://www.huaweicloud.com/)云监控的 Prometheus Exporter.


## 介绍
Prometheus是用于展示大型测量数据的开源可视化工具，在工业监控、气象监控、家居自动化和过程管理等领域也有着较广泛的用户基础。将华为云Cloudeye服务接入 prometheus后，您可以利用 prometheus更好地监控和分析来自 Cloudeye服务的数据。

## 拓展标签支持情况
该插件对于已对接云监控的云服务均支持指标数据的导出。为提高云服务资源的识别度、可读性，插件对于以下服务支持导出资源属性label，如ECS实例会增加hostname、ip等label，同时支持将华为云标签转化为label，满足对资源自定义label的诉求，具体如下：
|云服务|命名空间|支持通过实例TAG增加标签|标签数据来源|
| :--- | :--- | :--: | :--: |
|弹性云服务器|SYS.ECS/AGT.ECS|√|RMS/云服务|
|云硬盘|SYS.EVS|√|RMS/云服务|
|分布式缓存服务|SYS.DCS|√|RMS|
|云专线|SYS.DCAAS|√|RMS|
|弹性公网IP和带宽|SYS.VPC|√|RMS|
|云搜索服务|SYS.ES|√|RMS|
|关系型数据库|SYS.RDS|√|RMS|
|弹性负载均衡|SYS.ELB|√|云服务|
|云数据库 GaussDB(for MySQL)|SYS.GAUSSDB|√|RMS|
|云数据库 GaussDB(for openGauss)|SYS.GAUSSDBV5|√|云服务|
|NAT网关|SYS.NAT|√|RMS|
|弹性伸缩|SYS.AS|√|RMS|
|函数工作流|SYS.FunctionGraph|√|RMS|
|数据复制服务|SYS.DRS|√|RMS|
|Web应用防火墙|SYS.WAF|√|RMS|
|文档数据库服务|SYS.DDS|√|云服务|
|API网关|SYS.APIG|×|云服务|
|云备份|SYS.CBR|√|RMS/云服务|
|数据湖探索|SYS.DLI|√|RMS&云服务|
|弹性文件服务|SYS.SFS|×|云服务|
|弹性文件服务 SFS Turbo|SYS.EFS|√|RMS|
|虚拟专用网络|SYS.VPN|√|RMS|
|云数据迁移|SYS.CDM|×|云服务|
|数据仓库服务|SYS.DWS|√|云服务|
|内容审核Moderation|SYS.MODERATION|×|-|
|Anti-DDoS流量清洗|SYS.DDOS|√|RMS|
|云数据库GaussDB(for Nosql)|SYS.NoSQL|×|云服务|
|分布式消息服务|SYS.DMS|√|RMS|
|分布式数据库中间件|SYS.DDMS|×|RMS&云服务|
|API专享版网关|SYS.APIC|×|云服务|
|裸金属服务器|SYS.BMS/SERVICE.BMS|√|RMS|
|ModelArts|SYS.ModelArts|√|RMS|
|VPC终端节点|SYS.VPCEP |√|RMS|
|图引擎服务GES|SYS.GES|√|RMS|
|数据库安全服务DBSS|SYS.DBSS |√|RMS|

注：自定义标签时，key只能包含大写字母、小写字母以及中划线

## 环境准备
以Ubuntu 18.04系统和Prometheus 2.14.0版本为例
| Prometheus | prometheus-2.14.0.linux-amd64 |
| ------------ | ------------ |
| ECS | Ubuntu 18.04 |
| Ubuntu private ip | 192.168.0.xx |

## 安装配置cloudeye-exporter
1. 在ubuntu vm上安装cloudeye-exporter
   
   登录vm机器，查看插件Releases版本 (https://github.com/huaweicloud/cloudeye-exporter/releases) ，获取插件下载地址，下载解压安装。
```
# 参考命令：
mkdir cloudeye-exporter
cd cloudeye-exporter
wget https://github.com/huaweicloud/cloudeye-exporter/releases/download/v2.0.2/cloudeye-exporter.v2.0.2.tar.gz
tar -xzvf cloudeye-exporter.v2.0.2.tar.gz
```
2. 编辑clouds.yml文件配置公有云信息
```
global:
  port: ":8087"
  scrape_batch_size: 300
auth:
auth_url: "https://iam.cn-north-1.myhuaweicloud.com/v3"
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

## Grafana监控面板使用
Grafana是一个开源的可视化和分析平台，支持多种数据源，提供多种面板、插件来快速将复杂的数据转换为漂亮的图形和可视化的工具。将华为云Cloudeye服务接入 prometheus后，您可以利用Grafana更好地分析和展示来自Cloudeye服务的数据。
目前提供了ECS等服务的监控面板，具体使用方法见：[Grafana监控面板使用](./grafana_dashboard/use_grafana_template.md)