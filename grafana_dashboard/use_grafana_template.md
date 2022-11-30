# Grafana监控面板使用
1. 下载Grafana (https://grafana.com/grafana/download)
   ```
   wget https://dl.grafana.com/enterprise/release/grafana-enterprise-9.0.5-1.x86_64.rpm
   sudo yum install grafana-enterprise-9.0.5-1.x86_64.rpm
   service grafana-server start
   ```
2. Grafana接入Prometheus数据源
   >(1) 登录Grafana
   >> 浏览器访问http://127.0.0.1:3000，登录
   >> ![load.png](pic/login.jpg)
   
   >(2) 配置Prometheus数据源
   >> Configuration—》Data source—》Add data source —》Prometheus —》填写Prometheus地址 —》保存&测试
   >> ![config_prometheus.gif](pic/config_prometheus.gif)
3. 配置相关云服务监控视图

   如果需要直接使用以下模板，需要修改prometheus配置，增加获取企业项目信息的任务，配置参考如下：
   ```
   $ vi prometheus.yml
   global:
     scrape_interval: 1m # Set the scrape interval to every 1 minute seconds. Default is every 1 minute.
     scrape_timeout: 1m
   scrape_configs:
     # 如果开启了企业项目，则配置该任务获取企业项目信息，用于模板
     - job_name: 'huaweicloud-eps'
       metrics_path: "/eps-info"
       static_configs:
       - targets: ['192.168.0.xx:8087']
   ```
   ><font size=6>+</font> —》Import —》输入json模板文件—》load
   >> ![import.png](pic/import.jpg)
   >> ![img.png](pic/load.jpg)
   
   模板文件获取地址: 
   + [云搜索服务 CSS](templates/css(es)_dashboard_template.json)
   + [云专线 DCAAS](templates/dcaas_dashboard_template.json)
   + [分布式缓存服务 DCS](templates/dcs_dashboard_template.json)
   + [弹性云服务器 ECS](templates/ecs_dashboard_template.json)
   + [弹性负载均衡 ELB](templates/elb_dashboard_template.json)
   + [关系型数据库 RDS](templates/rds_dashboard_template.json)
   + [Web应用防火墙 WAF](templates/waf_dashboard_template.json)
   + [弹性公网IP和带宽 VPC](templates/vpc_dashboard_template.json)
4. 效果展示：
   >ECS:
   > ![img.png](pic/ecs.jpg)
   