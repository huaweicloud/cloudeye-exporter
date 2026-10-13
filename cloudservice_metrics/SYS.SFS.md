> CES Exporter支持导出的“弹性文件服务”指标如下表所示

|维度|指标名|指标描述|指标单位|
|:--|:--|:--|:--|
|share_id<br>（SFS容量型）|read_bandwidth|读带宽|byte/s|
|share_id<br>（SFS容量型）|write_bandwidth|写带宽|byte/s|
|share_id<br>（SFS容量型）|rw_bandwidth|读写带宽|byte/s|
|share_id<br>（SFS容量型）|read_ops|读OPS|次/秒|
|share_id<br>（SFS容量型）|write_ops|写OPS|次/秒|
|share_id<br>（SFS容量型）|rw_ops|读写OPS|次/秒|
|share_id<br>（SFS容量型）|used_capicity|已用容量|byte|
|bucket_name<br>（通用文件系统）|read_bytes_intranet|文件系统读带宽|byte/s|
|bucket_name<br>（通用文件系统）|write_bytes_intranet|文件系统写带宽|byte/s|
|bucket_name<br>（通用文件系统）|read_tps|文件系统读TPS|个|
|bucket_name<br>（通用文件系统）|write_tps|文件系统写TPS|个|
|bucket_name<br>（通用文件系统）|read_latency|读时延|ms|
|bucket_name<br>（通用文件系统）|write_latency|写时延|ms|
|bucket_name<br>（通用文件系统）|capacity_standard|容量型存储用量|Byte|
|bucket_name<br>（通用文件系统）|capacity_infrequentaccess|低频型存储用量|Byte|
|bucket_name<br>（通用文件系统）|capacity_performance|性能型存储用量|Byte|