# 自定义标签与系统内置标签重名问题排查指南

## 问题描述

若您配置的自定义标签与exporter系统内置标签重名，将导致对应资源的指标无法导出至prometheus，并在exporter日志中输出错误信息。

## 日志识别

当发生标签重名问题时，exporter日志会打印以下格式的ERROR信息：

```
[xxx] Duplicate label names were found between custom tags and system tags and the metrics export is skipped for this resource(namespace: {服务名}, dimensions info: {维度信息},  labels: {冲突的标签名}).
```

### 日志字段说明

| 字段 | 含义                                    |
|------|---------------------------------------|
| `namespace` | 发生冲突的云服务命名空间（如 `SYS.ECS`、`SYS.ELB` 等） |
| `dimensions info` | 发生冲突的资源维度信息（如ECS的实例ID，磁盘盘符，ELB实例ID等） |
| `labels` | 发生重名问题的标签名称 |


## 解决方案

### 步骤一

根据日志提示的 `namespace` 、 `dimensions info` 和 `labels`，定位具体是哪个服务、哪个资源实例下的哪个标签存在冲突：

1. **定位服务**：根据 `namespace` 确认是哪个云服务（如 ECS、ELB、RDS 等）
2. **定位资源**：根据 `dimensions info` 中的维度信息，定位到具体的资源
3. **定位标签**：根据 `labels` 字段，确认是哪个自定义标签与内置标签冲突

### 步骤二

登录华为云console，在对应资源的编辑页面，调整自定义标签名称。

## 验证修复

修复后，重新启动 Exporter，检查日志中是否还有类似的 ERROR 信息。如果问题已解决，该指标数据应能正常导出。

## 附录：支持的云服务命名空间


具体支持的服务命名空间及维度描述请参考 `cloudservice_metrics` 目录下的文档。