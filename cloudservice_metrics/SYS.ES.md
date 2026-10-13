> CES Exporter支持导出的“云搜索服务”指标如下表所示

|维度|指标名|指标描述|指标单位|
|:--|:--|:--|:--|
|cluster_id<br>（CSS集群）|disk_util|最大磁盘使用率|%|
|cluster_id<br>（CSS集群）|status|集群健康状态||
|cluster_id<br>（CSS集群）|max_jvm_heap_usage|最大JVM堆使用率|%|
|cluster_id<br>（CSS集群）|max_jvm_young_gc_time|最大JVM Young GC耗时|ms|
|cluster_id<br>（CSS集群）|max_jvm_young_gc_count|最大JVM Young GC次数||
|cluster_id<br>（CSS集群）|max_jvm_old_gc_time|最大JVM Old GC耗时|ms|
|cluster_id<br>（CSS集群）|max_jvm_old_gc_count|最大JVM Old GC次数||
|cluster_id<br>（CSS集群）|total_fs_size|文件系统总大小|byte|
|cluster_id<br>（CSS集群）|free_fs_size|文件系统可用大小|byte|
|cluster_id<br>（CSS集群）|max_cpu_usage|最大CPU利用率|%|
|cluster_id<br>（CSS集群）|max_cpu_time_of_jvm_process|最大JVM进程使用的CPU时间|ms|
|cluster_id<br>（CSS集群）|max_virtual_memory_size_of_jvm_process|最大JVM进程使用的虚拟内存大小|byte|
|cluster_id<br>（CSS集群）|max_current_opened_http_count|最大当前打开的Http连接数||
|cluster_id<br>（CSS集群）|max_total_opened_http_count|最大全部打开的Http连接数||
|cluster_id<br>（CSS集群）|indices_count|索引数量||
|cluster_id<br>（CSS集群）|total_shards_count|分片数量||
|cluster_id<br>（CSS集群）|primary_shards_count|主分片数量||
|cluster_id<br>（CSS集群）|docs_count|文档数量||
|cluster_id<br>（CSS集群）|docs_deleted_count|被删除的文档数量||
|cluster_id<br>（CSS集群）|nodes_count|节点数量||
|cluster_id<br>（CSS集群）|data_nodes_count|数据节点数量||
|cluster_id<br>（CSS集群）|coordinating_nodes_count|协调节点数量||
|cluster_id<br>（CSS集群）|master_nodes_count|Master节点数量||
|cluster_id<br>（CSS集群）|ingest_nodes_count|Client节点数量||
|cluster_id<br>（CSS集群）|max_load_average|最大节点Load值||
|cluster_id<br>（CSS集群）|avg_cpu_usage|平均CPU使用率|%|
|cluster_id<br>（CSS集群）|avg_load_average|平均节点Load值||
|cluster_id<br>（CSS集群）|avg_jvm_heap_usage|平均JVM堆使用率|%|
|cluster_id<br>（CSS集群）|max_open_file_descriptors|已打开的最大文件描述符数||
|cluster_id<br>（CSS集群）|avg_open_file_descriptors|已打开的平均文件描述符数||
|cluster_id<br>（CSS集群）|sum_max_file_descriptors|最大允许的文件描述符数||
|cluster_id<br>（CSS集群）|sum_open_file_descriptors|已打开的文件描述符数||
|cluster_id<br>（CSS集群）|sum_thread_pool_bulk_queue|Bulk队列中总排队任务数||
|cluster_id<br>（CSS集群）|sum_thread_pool_write_queue|Write队列中总排队任务数||
|cluster_id<br>（CSS集群）|sum_thread_pool_search_queue|Search队列中总排队任务数||
|cluster_id<br>（CSS集群）|sum_thread_pool_index_queue|Index队列中总排队任务数||
|cluster_id<br>（CSS集群）|sum_thread_pool_force_merge_queue|ForceMerge队列中总排队任务数||
|cluster_id<br>（CSS集群）|sum_thread_pool_bulk_rejected|Bulk队列中总的已拒绝任务数||
|cluster_id<br>（CSS集群）|sum_thread_pool_write_rejected|Write队列中总的已拒绝任务数||
|cluster_id<br>（CSS集群）|sum_thread_pool_search_rejected|Search队列中总的已拒绝任务数||
|cluster_id<br>（CSS集群）|sum_thread_pool_index_rejected|Index队列中总的已拒绝任务数||
|cluster_id<br>（CSS集群）|sum_thread_pool_force_merge_rejected|Forcemerge队列中总的已拒绝任务数||
|cluster_id<br>（CSS集群）|max_thread_pool_bulk_queue|Bulk队列中最大排队任务数||
|cluster_id<br>（CSS集群）|max_thread_pool_search_queue|Search队列中最大排队任务数||
|cluster_id<br>（CSS集群）|max_thread_pool_index_queue|Index队列中最大排队任务数||
|cluster_id<br>（CSS集群）|max_thread_pool_force_merge_queue|ForceMerge队列中最大排队任务数||
|cluster_id<br>（CSS集群）|sum_thread_pool_bulk_threads|Bulk线程池总大小||
|cluster_id<br>（CSS集群）|sum_thread_pool_write_threads|Write线程池总大小||
|cluster_id<br>（CSS集群）|sum_thread_pool_search_threads|Search线程池总大小||
|cluster_id<br>（CSS集群）|sum_thread_pool_index_threads|Index线程池总大小||
|cluster_id<br>（CSS集群）|sum_thread_pool_force_merge_threads|ForceMerge线程池总大小||
|cluster_id<br>（CSS集群）|avg_thread_pool_bulk_queue|Bulk队列中平均排队任务数||
|cluster_id<br>（CSS集群）|avg_thread_pool_write_queue|Write队列中平均排队任务数||
|cluster_id<br>（CSS集群）|avg_thread_pool_search_queue|Search队列中平均排队任务数||
|cluster_id<br>（CSS集群）|avg_thread_pool_index_queue|Index队列中平均排队任务数||
|cluster_id<br>（CSS集群）|avg_thread_pool_force_merge_queue|ForceMerge队列中平均排队任务数||
|cluster_id<br>（CSS集群）|avg_thread_pool_bulk_threads|Bulk线程池平均大小||
|cluster_id<br>（CSS集群）|avg_thread_pool_search_threads|Search线程池平均大小||
|cluster_id<br>（CSS集群）|avg_thread_pool_write_threads|Write线程池平均大小||
|cluster_id<br>（CSS集群）|avg_thread_pool_index_threads|Index线程池平均大小||
|cluster_id<br>（CSS集群）|avg_thread_pool_force_merge_threads|ForceMerge线程池平均大小||
|cluster_id<br>（CSS集群）|avg_thread_pool_write_rejected|Write队列中平均已拒绝任务数||
|cluster_id<br>（CSS集群）|min_free_fs_size|最小可用存储空间|byte|
|cluster_id<br>（CSS集群）|avg_jvm_old_gc_count|JVM老年代平均GC次数||
|cluster_id<br>（CSS集群）|avg_jvm_old_gc_time|JVM老年代平均GC时间|ms|
|cluster_id<br>（CSS集群）|avg_jvm_young_gc_count|JVM年轻代平均GC次数||
|cluster_id<br>（CSS集群）|avg_jvm_young_gc_time|JVM年轻代平均GC时间|ms|
|cluster_id<br>（CSS集群）|avg_max_file_descriptors|最大允许的文件描述符数-平均值||
|cluster_id<br>（CSS集群）|avg_mem_free_in_bytes|平均可用内存空间|byte|
|cluster_id<br>（CSS集群）|avg_mem_free_percent|平均可用内存比例|%|
|cluster_id<br>（CSS集群）|avg_mem_used_in_bytes|平均已用内存空间|byte|
|cluster_id<br>（CSS集群）|avg_mem_used_percent|平均已用内存比例|%|
|cluster_id<br>（CSS集群）|max_mem_free_in_bytes|最大可用内存空间|byte|
|cluster_id<br>（CSS集群）|max_mem_free_percent|最大可用内存比例|%|
|cluster_id<br>（CSS集群）|max_mem_used_in_bytes|最大已用内存空间|byte|
|cluster_id<br>（CSS集群）|max_mem_used_percent|最大已用内存比例|%|
|cluster_id<br>（CSS集群）|sum_jvm_old_gc_count|JVM老年代总GC次数||
|cluster_id<br>（CSS集群）|sum_jvm_old_gc_time|JVM老年代总GC时间|ms|
|cluster_id<br>（CSS集群）|sum_jvm_young_gc_count|JVM年轻代总GC次数||
|cluster_id<br>（CSS集群）|sum_jvm_young_gc_time|JVM年轻代总GC时间|ms|
|cluster_id<br>（CSS集群）|sum_current_opened_http_count|当前已打开http连接数||
|cluster_id<br>（CSS集群）|sum_total_opened_http_count|历史已打开http连接数||
|cluster_id<br>（CSS集群）|IndexingLatency|平均索引延迟|ms|
|cluster_id<br>（CSS集群）|IndexingRate|平均索引速率||
|cluster_id<br>（CSS集群）|SearchLatency|平均查询延迟|ms|
|cluster_id<br>（CSS集群）|SearchRate|平均查询速率||
|cluster_id<br>（CSS集群）|sum_thread_pool_flush_queue|Flush队列中总排队任务数|count|
|cluster_id<br>（CSS集群）|sum_thread_pool_flush_rejected|Flush队列中总的已拒绝任务数|count|
|cluster_id<br>（CSS集群）|max_thread_pool_flush_queue|Flush队列中最大排队任务数|count|
|cluster_id<br>（CSS集群）|sum_thread_pool_flush_threads|Flush线程池总大小|count|
|cluster_id<br>（CSS集群）|avg_thread_pool_flush_queue|Flush队列中平均排队任务数|count|
|cluster_id<br>（CSS集群）|avg_thread_pool_flush_threads|Flush线程池平均大小|count|
|cluster_id<br>（CSS集群）|sum_thread_pool_generic_queue|Generic队列中总排队任务数|count|
|cluster_id<br>（CSS集群）|sum_thread_pool_generic_rejected|Generic队列中总的已拒绝任务数|count|
|cluster_id<br>（CSS集群）|max_thread_pool_generic_queue|Generic队列中最大排队任务数|count|
|cluster_id<br>（CSS集群）|sum_thread_pool_generic_threads|Generic线程池总大小|count|
|cluster_id<br>（CSS集群）|avg_thread_pool_generic_queue|Generic队列中平均排队任务数|count|
|cluster_id<br>（CSS集群）|avg_thread_pool_generic_threads|Generic线程池平均大小|count|
|cluster_id<br>（CSS集群）|sum_thread_pool_management_queue|Management队列中总排队任务数|count|
|cluster_id<br>（CSS集群）|sum_thread_pool_management_rejected|Management队列中总的已拒绝任务数|count|
|cluster_id<br>（CSS集群）|max_thread_pool_management_queue|Management队列中最大排队任务数|count|
|cluster_id<br>（CSS集群）|sum_thread_pool_management_threads|Management线程池总大小|count|
|cluster_id<br>（CSS集群）|avg_thread_pool_management_queue|Management队列中平均排队任务数|count|
|cluster_id<br>（CSS集群）|avg_thread_pool_management_threads|Management线程池平均大小|count|
|cluster_id<br>（CSS集群）|sum_thread_pool_refresh_queue|Refresh队列中总排队任务数|count|
|cluster_id<br>（CSS集群）|sum_thread_pool_refresh_rejected|Refresh队列中总的已拒绝任务数|count|
|cluster_id<br>（CSS集群）|max_thread_pool_refresh_queue|Refresh队列中最大排队任务数|count|
|cluster_id<br>（CSS集群）|sum_thread_pool_refresh_threads|Refresh线程池总大小|count|
|cluster_id<br>（CSS集群）|avg_thread_pool_refresh_queue|Refresh队列中平均排队任务数|count|
|cluster_id<br>（CSS集群）|avg_thread_pool_refresh_threads|Refresh线程池平均大小|count|
|cluster_id<br>（CSS集群）|sum_thread_pool_obs_searcher_queue|OBS Searcher队列中总排队任务数|count|
|cluster_id<br>（CSS集群）|sum_thread_pool_obs_searcher_rejected|OBS Searcher队列中总的已拒绝任务数|count|
|cluster_id<br>（CSS集群）|max_thread_pool_obs_searcher_queue|OBS Searcher队列中最大排队任务数|count|
|cluster_id<br>（CSS集群）|sum_thread_pool_obs_searcher_threads|OBS Searcher线程池总大小|count|
|cluster_id<br>（CSS集群）|avg_thread_pool_obs_searcher_queue|OBS Searcher队列中平均排队任务数|count|
|cluster_id<br>（CSS集群）|avg_thread_pool_obs_searcher_threads|OBS Searcher线程池平均大小|count|
|cluster_id<br>（CSS集群）|sum_thread_pool_obs_queue|OBS队列中总排队任务数|count|
|cluster_id<br>（CSS集群）|sum_thread_pool_obs_rejected|OBS队列中总的已拒绝任务数|count|
|cluster_id<br>（CSS集群）|max_thread_pool_obs_queue|OBS队列中最大排队任务数|count|
|cluster_id<br>（CSS集群）|sum_thread_pool_obs_threads|OBS线程池总大小|count|
|cluster_id<br>（CSS集群）|avg_thread_pool_obs_queue|OBS队列中平均排队任务数|count|
|cluster_id<br>（CSS集群）|avg_thread_pool_obs_threads|OBS线程池平均大小|count|
|cluster_id<br>（CSS集群）|sum_thread_pool_obs_upload_queue|OBS Upload队列中总排队任务数|count|
|cluster_id<br>（CSS集群）|sum_thread_pool_obs_upload_rejected|OBS Upload队列中总的已拒绝任务数|count|
|cluster_id<br>（CSS集群）|max_thread_pool_obs_upload_queue|OBS Upload队列中最大排队任务数|count|
|cluster_id<br>（CSS集群）|sum_thread_pool_obs_upload_threads|OBS Upload线程池总大小|count|
|cluster_id<br>（CSS集群）|avg_thread_pool_obs_upload_queue|OBS Upload队列中平均排队任务数|count|
|cluster_id<br>（CSS集群）|avg_thread_pool_obs_upload_threads|OBS Upload线程池平均大小|count|
|cluster_id<br>（CSS集群）|sum_thread_pool_obs_download_queue|OBS Download队列中总排队任务数|count|
|cluster_id<br>（CSS集群）|sum_thread_pool_obs_download_rejected|OBS Download队列中总的已拒绝任务数|count|
|cluster_id<br>（CSS集群）|max_thread_pool_obs_download_queue|OBS Download队列中最大排队任务数|count|
|cluster_id<br>（CSS集群）|sum_thread_pool_obs_download_threads|OBS Download线程池总大小|count|
|cluster_id<br>（CSS集群）|avg_thread_pool_obs_download_queue|OBS Download队列中平均排队任务数|count|
|cluster_id<br>（CSS集群）|avg_thread_pool_obs_download_threads|OBS Download线程池平均大小|count|
|cluster_id<br>（CSS集群）|task_max_running_time|最大Task运行时长|ms|
|cluster_id<br>（CSS集群）|number_of_pending_tasks|Pending Task排队任务数|count|
|cluster_id<br>（CSS集群）|sum_disk_read_requests_rate|磁盘读总IOPS|request/s|
|cluster_id<br>（CSS集群）|sum_disk_write_requests_rate|磁盘写总IOPS|request/s|
|cluster_id<br>（CSS集群）|sum_disk_read_bytes_rate|磁盘读总带宽|Byte/s(IEC)|
|cluster_id<br>（CSS集群）|sum_disk_write_bytes_rate|磁盘写总带宽|Byte/s(IEC)|
|cluster_id<br>（CSS集群）|avg_vector_index_off_heap_used_in_bytes|平均向量索引堆外内存使用量|byte(IEC)|
|cluster_id<br>（CSS集群）|avg_vector_index_off_heap_usage|平均向量索引堆外内存使用率|%|
|cluster_id<br>（CSS集群）|max_vector_index_off_heap_used_in_bytes|最大向量索引堆外内存使用量|byte(IEC)|
|cluster_id<br>（CSS集群）|max_vector_index_off_heap_usage|最大向量索引堆外内存使用率|%|
|cluster_id<br>（CSS集群）|vector_index_circuit_breaker_status|向量索引熔断状态||
|cluster_id<br>（CSS集群）|request_count|请求总次数|count|
|cluster_id<br>（CSS集群）|successfully_request_count|请求成功次数|count|
|cluster_id<br>（CSS集群）|failed_request_count|请求失败次数|count|
|cluster_id<br>（CSS集群）|limited_request_count|请求限流次数|count|
|cluster_id<br>（CSS集群）|cold_data_storage|冷数据存储量|byte(IEC)|
|cluster_id<br>（CSS集群）|number_of_index_creation_failures|索引创建失败次数|count|
|cluster_id<br>（CSS集群）|shard_doc_exceed_threshold_count|文档数超过阈值的分片数量|count|
|cluster_id<br>（CSS集群）|sum_events_in|集群下所有节点经过input插件的数据总数|count|
|cluster_id<br>（CSS集群）|sum_events_filtered|集群下所有节点经过filtered插件的数据总数|count|
|cluster_id<br>（CSS集群）|sum_events_out|集群下所有节点经过out插件的数据总数|count|
|cluster_id,instance_id<br>（CSS集群,云服务节点）|load_average|节点Load值||
|cluster_id,instance_id<br>（CSS集群,云服务节点）|cpu_usage|CPU利用率|%|
|cluster_id,instance_id<br>（CSS集群,云服务节点）|disk_usage|磁盘利用率|%|
|cluster_id,instance_id<br>（CSS集群,云服务节点）|jvm_heap_usage|JVM堆使用率|%|
|cluster_id,instance_id<br>（CSS集群,云服务节点）|open_file_descriptors|已打开的文件描述符数||
|cluster_id,instance_id<br>（CSS集群,云服务节点）|max_file_descriptors|最大允许的文件描述符数||
|cluster_id,instance_id<br>（CSS集群,云服务节点）|thread_pool_bulk_queue|Bulk队列中总排队任务数||
|cluster_id,instance_id<br>（CSS集群,云服务节点）|thread_pool_write_queue|Write队列中总排队任务数||
|cluster_id,instance_id<br>（CSS集群,云服务节点）|thread_pool_search_queue|Search队列中总排队任务数||
|cluster_id,instance_id<br>（CSS集群,云服务节点）|thread_pool_index_queue|Index队列中总排队任务数||
|cluster_id,instance_id<br>（CSS集群,云服务节点）|thread_pool_force_merge_queue|ForceMerge队列中总排队任务数||
|cluster_id,instance_id<br>（CSS集群,云服务节点）|thread_pool_bulk_rejected|Bulk队列中总的已拒绝任务数||
|cluster_id,instance_id<br>（CSS集群,云服务节点）|thread_pool_write_rejected|Write队列中总的已拒绝任务数||
|cluster_id,instance_id<br>（CSS集群,云服务节点）|thread_pool_search_rejected|Search队列中总的已拒绝任务数||
|cluster_id,instance_id<br>（CSS集群,云服务节点）|thread_pool_index_rejected|Index队列中总的已拒绝任务数||
|cluster_id,instance_id<br>（CSS集群,云服务节点）|thread_pool_force_merge_rejected|Forcemerge队列中总的已拒绝任务数||
|cluster_id,instance_id<br>（CSS集群,云服务节点）|thread_pool_bulk_threads|Bulk线程池总大小||
|cluster_id,instance_id<br>（CSS集群,云服务节点）|thread_pool_write_threads|Write线程池总大小||
|cluster_id,instance_id<br>（CSS集群,云服务节点）|thread_pool_search_threads|Search线程池总大小||
|cluster_id,instance_id<br>（CSS集群,云服务节点）|thread_pool_index_threads|Index线程池总大小||
|cluster_id,instance_id<br>（CSS集群,云服务节点）|thread_pool_force_merge_threads|ForceMerge线程池总大小||
|cluster_id,instance_id<br>（CSS集群,云服务节点）|free_fs_size|文件系统可用大小|byte|
|cluster_id,instance_id<br>（CSS集群,云服务节点）|total_fs_size|文件系统总大小|byte|
|cluster_id,instance_id<br>（CSS集群,云服务节点）|jvm_old_gc_count|JVM老年代总GC次数||
|cluster_id,instance_id<br>（CSS集群,云服务节点）|jvm_old_gc_time|JVM老年代总GC时间|ms|
|cluster_id,instance_id<br>（CSS集群,云服务节点）|jvm_young_gc_count|JVM年轻代总GC次数||
|cluster_id,instance_id<br>（CSS集群,云服务节点）|jvm_young_gc_time|JVM年轻代GC时间|ms|
|cluster_id,instance_id<br>（CSS集群,云服务节点）|mem_free_in_bytes|可用内存空间|byte|
|cluster_id,instance_id<br>（CSS集群,云服务节点）|mem_free_percent|可用内存比例||
|cluster_id,instance_id<br>（CSS集群,云服务节点）|mem_used_in_bytes|已用内存空间|byte|
|cluster_id,instance_id<br>（CSS集群,云服务节点）|current_opened_http_count|当前已打开http连接数||
|cluster_id,instance_id<br>（CSS集群,云服务节点）|total_opened_http_count|全部打开的http连接数||
|cluster_id,instance_id<br>（CSS集群,云服务节点）|disk_util|最大磁盘使用率|%|
|cluster_id,instance_id<br>（CSS集群,云服务节点）|thread_pool_flush_queue|Flush队列中总排队任务数|count|
|cluster_id,instance_id<br>（CSS集群,云服务节点）|thread_pool_flush_rejected|Flush队列中总的已拒绝任务数|count|
|cluster_id,instance_id<br>（CSS集群,云服务节点）|thread_pool_flush_threads|Flush线程池总大小|count|
|cluster_id,instance_id<br>（CSS集群,云服务节点）|thread_pool_generic_queue|Generic队列中总排队任务数|count|
|cluster_id,instance_id<br>（CSS集群,云服务节点）|thread_pool_generic_rejected|Generic队列中总的已拒绝任务数|count|
|cluster_id,instance_id<br>（CSS集群,云服务节点）|thread_pool_generic_threads|Generic线程池总大小|count|
|cluster_id,instance_id<br>（CSS集群,云服务节点）|thread_pool_management_queue|Management队列中总排队任务数|count|
|cluster_id,instance_id<br>（CSS集群,云服务节点）|thread_pool_management_rejected|Management队列中总的已拒绝任务数|count|
|cluster_id,instance_id<br>（CSS集群,云服务节点）|thread_pool_management_threads|Management线程池总大小|count|
|cluster_id,instance_id<br>（CSS集群,云服务节点）|thread_pool_refresh_queue|Refresh队列中总排队任务数|count|
|cluster_id,instance_id<br>（CSS集群,云服务节点）|thread_pool_refresh_rejected|Refresh队列中总的已拒绝任务数|count|
|cluster_id,instance_id<br>（CSS集群,云服务节点）|thread_pool_refresh_threads|Rfresh线程池总大小|count|
|cluster_id,instance_id<br>（CSS集群,云服务节点）|thread_pool_obs_searcher_queue|OBS Searcher队列中总排队任务数|count|
|cluster_id,instance_id<br>（CSS集群,云服务节点）|thread_pool_obs_searcher_rejected|OBS Searcher队列中总的已拒绝任务数|count|
|cluster_id,instance_id<br>（CSS集群,云服务节点）|thread_pool_obs_searcher_threads|OBS Searcher线程池总大小|count|
|cluster_id,instance_id<br>（CSS集群,云服务节点）|thread_pool_obs_queue|OBS队列中总排队任务数|count|
|cluster_id,instance_id<br>（CSS集群,云服务节点）|thread_pool_obs_rejected|OBS队列中总的已拒绝任务数|count|
|cluster_id,instance_id<br>（CSS集群,云服务节点）|thread_pool_obs_threads|OBS线程池总大小|count|
|cluster_id,instance_id<br>（CSS集群,云服务节点）|thread_pool_obs_upload_queue|OBS Upload队列中总排队任务数|count|
|cluster_id,instance_id<br>（CSS集群,云服务节点）|thread_pool_obs_upload_rejected|OBS Upload队列中总的已拒绝任务数|count|
|cluster_id,instance_id<br>（CSS集群,云服务节点）|thread_pool_obs_upload_threads|OBS Upload线程池总大小|count|
|cluster_id,instance_id<br>（CSS集群,云服务节点）|thread_pool_obs_download_queue|OBS Download队列中总排队任务数|count|
|cluster_id,instance_id<br>（CSS集群,云服务节点）|thread_pool_obs_download_rejected|OBS Download队列中总的已拒绝任务数|count|
|cluster_id,instance_id<br>（CSS集群,云服务节点）|thread_pool_obs_download_threads|OBS Download线程池总大小|count|
|cluster_id,instance_id<br>（CSS集群,云服务节点）|disk_read_requests_rate|磁盘读IOPS|request/s|
|cluster_id,instance_id<br>（CSS集群,云服务节点）|disk_write_requests_rate|磁盘写IOPS|request/s|
|cluster_id,instance_id<br>（CSS集群,云服务节点）|disk_read_bytes_rate|磁盘读带宽|Byte/s(IEC)|
|cluster_id,instance_id<br>（CSS集群,云服务节点）|disk_write_bytes_rate|磁盘写带宽|Byte/s(IEC)|
|cluster_id,instance_id<br>（CSS集群,云服务节点）|shards_count|分片数量|count|
|cluster_id,instance_id<br>（CSS集群,云服务节点）|vector_index_off_heap_used_in_bytes|向量索引堆外内存使用量|byte(IEC)|
|cluster_id,instance_id<br>（CSS集群,云服务节点）|vector_index_off_heap_usage|向量索引对外内存使用率|%|
|cluster_id,instance_id<br>（CSS集群,云服务节点）|events_in|当前节点经过input插件的数据数|count|
|cluster_id,instance_id<br>（CSS集群,云服务节点）|events_filtered|当前节点经过filtered插件的数据数|count|
|cluster_id,instance_id<br>（CSS集群,云服务节点）|events_out|当前节点经过out插件的数据数|count|