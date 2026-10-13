> CES Exporter支持导出的“数据仓库服务”指标如下表所示

|维度|指标名|指标描述|指标单位|
|:--|:--|:--|:--|
|datastore_id<br>（数据仓库服务）|dws001_shared_buffer_hit_ratio|缓存命中率|%|
|datastore_id<br>（数据仓库服务）|dws002_in_memory_sort_ratio|内存中排序比率|%|
|datastore_id<br>（数据仓库服务）|dws003_physical_reads|文件读取次数|次数|
|datastore_id<br>（数据仓库服务）|dws004_physical_writes|文件写入次数|次数|
|datastore_id<br>（数据仓库服务）|dws005_physical_reads_per_second|每秒文件读取次数|次数/秒|
|datastore_id<br>（数据仓库服务）|dws006_physical_writes_per_second|每秒文件写入次数|次数/秒|
|datastore_id<br>（数据仓库服务）|dws007_db_size|数据量大小|MB|
|datastore_id<br>（数据仓库服务）|dws008_active_sql_count|活跃SQL数|个数|
|datastore_id<br>（数据仓库服务）|dws009_session_count|会话数|个数|
|dws_instance_id<br>（数据仓库节点）|dws010_cpu_usage|CPU使用率|%|
|dws_instance_id<br>（数据仓库节点）|dws011_mem_usage|内存使用率|%|
|dws_instance_id<br>（数据仓库节点）|dws012_iops|IOPS|个数/秒|
|dws_instance_id<br>（数据仓库节点）|dws013_bytes_in|网络输入吞吐量|byte/s|
|dws_instance_id<br>（数据仓库节点）|dws014_bytes_out|网络输出吞吐量|byte/s|
|dws_instance_id<br>（数据仓库节点）|dws015_disk_usage|磁盘利用率|%|
|dws_instance_id<br>（数据仓库节点）|dws016_disk_total_size|磁盘总大小|GB|
|dws_instance_id<br>（数据仓库节点）|dws017_disk_used_size|磁盘使用量|GB|
|dws_instance_id<br>（数据仓库节点）|dws018_disk_read_throughput|硬盘读吞吐量|byte/s|
|dws_instance_id<br>（数据仓库节点）|dws019_disk_write_throughput|硬盘写吞吐量|byte/s|
|dws_instance_id<br>（数据仓库节点）|dws020_avg_disk_sec_per_read|硬盘读耗时|秒|
|dws_instance_id<br>（数据仓库节点）|dws021_avg_disk_sec_per_write|硬盘写耗时|秒|
|dws_instance_id<br>（数据仓库节点）|dws022_avg_disk_queue_length|磁盘平均队列长度||