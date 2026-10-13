> CES Exporter支持导出的“关系型数据库”指标如下表所示

|维度|指标名|指标描述|指标单位|
|:--|:--|:--|:--|
|rds_cluster_id<br>（MySQL实例）|rds001_cpu_util|CPU使用率|%|
|rds_cluster_id<br>（MySQL实例）|rds002_mem_util|内存使用率|%|
|rds_cluster_id<br>（MySQL实例）|rds003_iops|IOPS|次/秒|
|rds_cluster_id<br>（MySQL实例）|rds004_bytes_in|网络输入吞吐量|byte/s|
|rds_cluster_id<br>（MySQL实例）|rds005_bytes_out|网络输出吞吐量|byte/s|
|rds_cluster_id<br>（MySQL实例）|rds006_conn_count|数据库总连接数|个|
|rds_cluster_id<br>（MySQL实例）|rds007_conn_active_count|当前活跃连接数|个|
|rds_cluster_id<br>（MySQL实例）|rds008_qps|QPS|次/秒|
|rds_cluster_id<br>（MySQL实例）|rds009_tps|TPS|次/秒|
|rds_cluster_id<br>（MySQL实例）|rds010_innodb_buf_usage|缓冲池利用率|比率|
|rds_cluster_id<br>（MySQL实例）|rds011_innodb_buf_hit|缓冲池命中率|比率|
|rds_cluster_id<br>（MySQL实例）|rds012_innodb_buf_dirty|缓冲池脏块率|比率|
|rds_cluster_id<br>（MySQL实例）|rds013_innodb_reads|InnoDB读取吞吐量|byte/s|
|rds_cluster_id<br>（MySQL实例）|rds014_innodb_writes|InnoDB写入吞吐量|byte/s|
|rds_cluster_id<br>（MySQL实例）|rds015_innodb_read_count|InnoDB文件读取频率|次/秒|
|rds_cluster_id<br>（MySQL实例）|rds016_innodb_write_count|InnoDB文件写入频率|次/秒|
|rds_cluster_id<br>（MySQL实例）|rds017_innodb_log_write_req_count|InnoDB日志写请求频率|次/秒|
|rds_cluster_id<br>（MySQL实例）|rds018_innodb_log_write_count|InnoDB日志物理写频率|次/秒|
|rds_cluster_id<br>（MySQL实例）|rds019_innodb_log_fsync_count|InnoDB日志fsync()写频率|次/秒|
|rds_cluster_id<br>（MySQL实例）|rds021_myisam_buf_usage|Key Buffer利用率|比率|
|rds_cluster_id<br>（MySQL实例）|rds022_myisam_buf_write_hit|Key Buffer写命中率|比率|
|rds_cluster_id<br>（MySQL实例）|rds023_myisam_buf_read_hit|Key Buffer读命中率|比率|
|rds_cluster_id<br>（MySQL实例）|rds024_myisam_disk_write_count|MyISAM硬盘写入频率|次/秒|
|rds_cluster_id<br>（MySQL实例）|rds025_myisam_disk_read_count|MyISAM硬盘读取频率|次/秒|
|rds_cluster_id<br>（MySQL实例）|rds026_myisam_buf_write_count|MyISAM缓冲池写入频率|次/秒|
|rds_cluster_id<br>（MySQL实例）|rds027_myisam_buf_read_count|MyISAM缓冲池读取频率|次/秒|
|rds_cluster_id<br>（MySQL实例）|rds028_comdml_del_count|Delete语句执行频率|次/秒|
|rds_cluster_id<br>（MySQL实例）|rds029_comdml_ins_count|Insert语句执行频率|次/秒|
|rds_cluster_id<br>（MySQL实例）|rds030_comdml_ins_sel_count|Insert_Select语句执行频率|次/秒|
|rds_cluster_id<br>（MySQL实例）|rds031_comdml_rep_count|Replace语句执行频率|次/秒|
|rds_cluster_id<br>（MySQL实例）|rds032_comdml_rep_sel_count|Replace_Selection语句执行频率|次/秒|
|rds_cluster_id<br>（MySQL实例）|rds033_comdml_sel_count|Select语句执行频率|次/秒|
|rds_cluster_id<br>（MySQL实例）|rds034_comdml_upd_count|Update语句执行频率|次/秒|
|rds_cluster_id<br>（MySQL实例）|rds035_innodb_del_row_count|行删除速率|行/秒|
|rds_cluster_id<br>（MySQL实例）|rds036_innodb_ins_row_count|行插入速率|行/秒|
|rds_cluster_id<br>（MySQL实例）|rds037_innodb_read_row_count|行读取速率|行/秒|
|rds_cluster_id<br>（MySQL实例）|rds038_innodb_upd_row_count|行更新速率|行/秒|
|rds_cluster_id<br>（MySQL实例）|rds039_disk_util|磁盘利用率|%|
|rds_cluster_id<br>（MySQL实例）|rds047_disk_total_size|磁盘总大小|GB|
|rds_cluster_id<br>（MySQL实例）|rds048_disk_used_size|磁盘使用量|GB|
|rds_cluster_id<br>（MySQL实例）|rds049_disk_read_throughput|硬盘读吞吐量|byte/s|
|rds_cluster_id<br>（MySQL实例）|rds050_disk_write_throughput|硬盘写吞吐量|byte/s|
|rds_cluster_id<br>（MySQL实例）|rds072_conn_usage|连接数使用率|%|
|rds_cluster_id<br>（MySQL实例）|rds073_replication_delay|实时复制时延|秒|
|rds_cluster_id<br>（MySQL实例）|rds074_slow_queries|慢日志个数统计|个/分钟|
|rds_cluster_id<br>（MySQL实例）|rds075_avg_disk_ms_per_read|硬盘读耗时|毫秒|
|rds_cluster_id<br>（MySQL实例）|rds076_avg_disk_ms_per_write|硬盘写耗时|毫秒|
|rds_cluster_id<br>（MySQL实例）|rds020_temp_tbl_rate|临时表创建速率|个/秒|
|rds_cluster_id<br>（MySQL实例）|sys_swap_usage|swap利用率|%|
|rds_cluster_id<br>（MySQL实例）|rds_innodb_lock_waits|等待行锁事务数|个数|
|rds_cluster_id<br>（MySQL实例）|rds_bytes_recv_rate|数据库每秒接收字节|byte/s|
|rds_cluster_id<br>（MySQL实例）|rds_bytes_sent_rate|数据库每秒发送字节|byte/s|
|rds_cluster_id<br>（MySQL实例）|rds_innodb_pages_read_rate|innodb平均每秒读取的数据量|页数/秒|
|rds_cluster_id<br>（MySQL实例）|rds_innodb_pages_written_rate|innodb平均每秒写入的数据量|页数/秒|
|rds_cluster_id<br>（MySQL实例）|rds_innodb_os_log_written_rate|平均每秒写入redo log的大小|byte/s|
|rds_cluster_id<br>（MySQL实例）|rds_innodb_buffer_pool_read_requests_rate|innodb_buffer_pool每秒读请求次数|次数/秒|
|rds_cluster_id<br>（MySQL实例）|rds_innodb_buffer_pool_write_requests_rate|innodb_buffer_pool每秒写请求次数|次数/秒|
|rds_cluster_id<br>（MySQL实例）|rds_innodb_buffer_pool_pages_flushed_rate|innodb_buffer_pool每秒页面刷新数|次数/秒|
|rds_cluster_id<br>（MySQL实例）|rds_innodb_log_waits_rate|因log buffer不足导致等待flush到磁盘次数|次数/秒|
|rds_cluster_id<br>（MySQL实例）|rds_created_tmp_tables_rate|每秒创建临时表数|个数/秒|
|rds_cluster_id<br>（MySQL实例）|rds_wait_thread_count|等待线程数|个数|
|rds_cluster_id<br>（MySQL实例）|rds_innodb_row_lock_time_avg|历史行锁平均等待时间|ms|
|rds_cluster_id<br>（MySQL实例）|rds_innodb_row_lock_current_waits|当前行锁等待数|个数|
|rds_cluster_id<br>（MySQL实例）|rds077_vma|VMA数量|个|
|rds_cluster_id<br>（MySQL实例）|rds078_threads|进程中线程数量|个|
|rds_cluster_id<br>（MySQL实例）|rds079_vm_hwm|进程的物理内存占用峰值|KB|
|rds_cluster_id<br>（MySQL实例）|rds080_vm_peak|进程的虚拟内存占用峰值|KB|
|rds_cluster_id<br>（MySQL实例）|rds081_vm_ioutils|磁盘I/O处于非空闲状态的时间百分比|%|
|rds_cluster_id<br>（MySQL实例）|rds082_semi_sync_tx_avg_wait_time|事务平均等待时间|微秒|
|rds_cluster_id<br>（MySQL实例）|rds173_replication_delay_avg|平均复制时延|s|
|rds_cluster_id<br>（MySQL实例）|rds_innodb_log_waits_count|日志等待次数|count|
|rds_cluster_id<br>（MySQL实例）|rds_buffer_pool_wait_free|缓冲池空闲页等待次数|count|
|rds_cluster_id<br>（MySQL实例）|rds_conn_active_usage|活跃连接数使用率|%|
|rds_cluster_id<br>（MySQL实例）|rds_long_transaction|长事务指标|s|
|rds_cluster_id<br>（MySQL实例）|rds_mdl_lock_count|MDL锁数量|count|
|rds_cluster_id<br>（MySQL实例）|iops_usage|IOPS 使用率|%|
|rds_cluster_id<br>（MySQL实例）|rds_threadpool_waiting_threads|线程池中等待线程数|count|
|postgresql_cluster_id<br>（PostgreSQL实例）|rds001_cpu_util|CPU使用率|%|
|postgresql_cluster_id<br>（PostgreSQL实例）|rds002_mem_util|内存使用率|%|
|postgresql_cluster_id<br>（PostgreSQL实例）|rds003_iops|IOPS|次/秒|
|postgresql_cluster_id<br>（PostgreSQL实例）|rds004_bytes_in|网络输入吞吐量|byte/s|
|postgresql_cluster_id<br>（PostgreSQL实例）|rds005_bytes_out|网络输出吞吐量|byte/s|
|postgresql_cluster_id<br>（PostgreSQL实例）|rds039_disk_util|磁盘利用率|%|
|postgresql_cluster_id<br>（PostgreSQL实例）|rds040_transaction_logs_usage|事务日志使用量|MB|
|postgresql_cluster_id<br>（PostgreSQL实例）|rds041_replication_slot_usage|复制插槽使用量|MB|
|postgresql_cluster_id<br>（PostgreSQL实例）|rds042_database_connections|数据库连接数|个|
|postgresql_cluster_id<br>（PostgreSQL实例）|rds043_maximum_used_transaction_ids|事务最大已使用ID数|个|
|postgresql_cluster_id<br>（PostgreSQL实例）|rds044_transaction_logs_generations|事务日志生成速率|MB/s|
|postgresql_cluster_id<br>（PostgreSQL实例）|rds045_oldest_replication_slot_lag|最滞后副本滞后量|MB|
|postgresql_cluster_id<br>（PostgreSQL实例）|rds046_replication_lag|复制时延|毫秒|
|postgresql_cluster_id<br>（PostgreSQL实例）|rds047_disk_total_size|磁盘总大小|GB|
|postgresql_cluster_id<br>（PostgreSQL实例）|rds048_disk_used_size|磁盘使用量|GB|
|postgresql_cluster_id<br>（PostgreSQL实例）|rds049_disk_read_throughput|硬盘读吞吐量|byte/s|
|postgresql_cluster_id<br>（PostgreSQL实例）|rds050_disk_write_throughput|硬盘写吞吐量|byte/s|
|postgresql_cluster_id<br>（PostgreSQL实例）|rds082_tps|TPS|次/秒|
|postgresql_cluster_id<br>（PostgreSQL实例）|rds083_conn_usage|连接数使用率|比率|
|postgresql_cluster_id<br>（PostgreSQL实例）|row_per_second|操作行数|Row/s|
|postgresql_cluster_id<br>（PostgreSQL实例）|active_connections|活跃连接数|个|
|postgresql_cluster_id<br>（PostgreSQL实例）|idle_transaction_connections|事务空闲连接数|个|
|postgresql_cluster_id<br>（PostgreSQL实例）|oldest_transaction_duration|最长事务存活时长|毫秒|
|postgresql_cluster_id<br>（PostgreSQL实例）|oldest_transaction_duration_2pc|最长未决事务存活时长|毫秒|
|postgresql_cluster_id<br>（PostgreSQL实例）|disk_io_usage|磁盘IO使用率|%|
|postgresql_cluster_id<br>（PostgreSQL实例）|lock_waiting_sessions|等待锁的会话数|个|
|postgresql_cluster_id<br>（PostgreSQL实例）|swap_in_rate|swap in速率|KB/s|
|postgresql_cluster_id<br>（PostgreSQL实例）|swap_out_rate|swap out速率|KB/s|
|postgresql_cluster_id<br>（PostgreSQL实例）|swap_total_size|交换区总容量大小|MB|
|postgresql_cluster_id<br>（PostgreSQL实例）|swap_usage|交换区容量使用率|%|
|postgresql_cluster_id<br>（PostgreSQL实例）|db_max_age|最大数据库年龄|个|
|postgresql_cluster_id<br>（PostgreSQL实例）|cpu_user_usage|用户态CPU时间占比|%|
|postgresql_cluster_id<br>（PostgreSQL实例）|cpu_sys_usage|内核态CPU时间占比|%|
|postgresql_cluster_id<br>（PostgreSQL实例）|cpu_wait_usage|硬盘IO等待时间占比|%|
|postgresql_cluster_id<br>（PostgreSQL实例）|io_read_delay|IO读响应延迟|毫秒|
|postgresql_cluster_id<br>（PostgreSQL实例）|io_write_delay|IO写响应延迟|毫秒|
|postgresql_cluster_id<br>（PostgreSQL实例）|read_count_per_second|读IOPS|次|
|postgresql_cluster_id<br>（PostgreSQL实例）|write_count_per_second|写IOPS|次|
|postgresql_cluster_id<br>（PostgreSQL实例）|slow_sql_one_second|已执行1s的SQL数|个|
|postgresql_cluster_id<br>（PostgreSQL实例）|slow_sql_three_second|已执行3s的SQL数|个|
|postgresql_cluster_id<br>（PostgreSQL实例）|slow_sql_five_second|已执行5s的SQL数|个|
|postgresql_cluster_id<br>（PostgreSQL实例）|slow_sql_log_min_duration_statement|已执行log_min_duration_statement时长的SQL数|个|
|postgresql_cluster_id<br>（PostgreSQL实例）|pg_dr_repl_stat|灾备节点复制状态||
|postgresql_cluster_id<br>（PostgreSQL实例）|pg_dr_wal_delay|主机与灾备机间lsn延迟|byte|
|postgresql_cluster_id<br>（PostgreSQL实例）|round_trip_time|主机与灾备机间网络延迟|毫秒|
|postgresql_cluster_id<br>（PostgreSQL实例）|packet_loss_rate|主机与灾备机间丢包率|%|
|postgresql_cluster_id<br>（PostgreSQL实例）|inactive_logical_replication_slot|非活跃逻辑复制槽数量|个|
|postgresql_cluster_id<br>（PostgreSQL实例）|write_lsn_replication_latency_size|主备间wal日志写入延迟|Byte|
|postgresql_cluster_id<br>（PostgreSQL实例）|wal_size|WAL日志占用空间大小|GB|
|postgresql_cluster_id<br>（PostgreSQL实例）|user_current_connections|用户使用连接数|Counts|
|postgresql_cluster_id<br>（PostgreSQL实例）|user_active_connections|用户活跃连接数|Counts|
|postgresql_cluster_id<br>（PostgreSQL实例）|temporary_files_generation_size|每分钟临时文件生成大小|Byte/min|
|postgresql_cluster_id<br>（PostgreSQL实例）|temporary_files_generation_num|每分钟临时文件生成数量|Count/min|
|postgresql_cluster_id<br>（PostgreSQL实例）|sys_memory_hit_rate|内存命中率|%|
|postgresql_cluster_id<br>（PostgreSQL实例）|synchronous_replication_blocking_time|同步复制阻塞时间|s|
|postgresql_cluster_id<br>（PostgreSQL实例）|slave_replication_status|备机（只读）流复制状态|Count|
|postgresql_cluster_id<br>（PostgreSQL实例）|sent_lsn_replication_latency_size|主备间wal日志发送延迟|Byte|
|postgresql_cluster_id<br>（PostgreSQL实例）|replay_lsn_replication_latency_size|主备间wal日志回放延迟|Byte|
|postgresql_cluster_id<br>（PostgreSQL实例）|pgaudit_log_size|审计日志大小|GB|
|postgresql_cluster_id<br>（PostgreSQL实例）|flush_lsn_replication_latency_size|主备间wal日志落盘延迟|Byte|
|postgresql_cluster_id<br>（PostgreSQL实例）|dbuser_passwd_deadline|数据库用户最快过期时间|s|
|postgresql_cluster_id<br>（PostgreSQL实例）|data_disk_inode_used|inode数|Counts|
|rds_cluster_sqlserver_id<br>（Microsoft SQL Server实例）|rds001_cpu_util|CPU使用率|%|
|rds_cluster_sqlserver_id<br>（Microsoft SQL Server实例）|rds002_mem_util|内存使用率|%|
|rds_cluster_sqlserver_id<br>（Microsoft SQL Server实例）|rds003_iops|IOPS|次/秒|
|rds_cluster_sqlserver_id<br>（Microsoft SQL Server实例）|rds004_bytes_in|网络输入吞吐量|byte/s|
|rds_cluster_sqlserver_id<br>（Microsoft SQL Server实例）|rds005_bytes_out|网络输出吞吐量|byte/s|
|rds_cluster_sqlserver_id<br>（Microsoft SQL Server实例）|rds039_disk_util|磁盘利用率|%|
|rds_cluster_sqlserver_id<br>（Microsoft SQL Server实例）|rds047_disk_total_size|磁盘总大小|GB|
|rds_cluster_sqlserver_id<br>（Microsoft SQL Server实例）|rds048_disk_used_size|磁盘使用量|GB|
|rds_cluster_sqlserver_id<br>（Microsoft SQL Server实例）|rds049_disk_read_throughput|硬盘读吞吐量|byte/s|
|rds_cluster_sqlserver_id<br>（Microsoft SQL Server实例）|rds050_disk_write_throughput|硬盘写吞吐量|byte/s|
|rds_cluster_sqlserver_id<br>（Microsoft SQL Server实例）|rds053_avg_disk_queue_length|磁盘平均队列长度||
|rds_cluster_sqlserver_id<br>（Microsoft SQL Server实例）|rds054_db_connections_in_use|使用中的数据库连接数|个|
|rds_cluster_sqlserver_id<br>（Microsoft SQL Server实例）|rds055_transactions_per_sec|平均每秒事务数|个/秒|
|rds_cluster_sqlserver_id<br>（Microsoft SQL Server实例）|rds056_batch_per_sec|平均每秒batch数|个/秒|
|rds_cluster_sqlserver_id<br>（Microsoft SQL Server实例）|rds057_logins_per_sec|每秒登录次数|个/秒|
|rds_cluster_sqlserver_id<br>（Microsoft SQL Server实例）|rds058_logouts_per_sec|每秒登出次数|个/秒|
|rds_cluster_sqlserver_id<br>（Microsoft SQL Server实例）|rds059_cache_hit_ratio|缓存命中率|%|
|rds_cluster_sqlserver_id<br>（Microsoft SQL Server实例）|rds060_sql_compilations_per_sec|平均每秒SQL编译数|个/秒|
|rds_cluster_sqlserver_id<br>（Microsoft SQL Server实例）|rds061_sql_recompilations_per_sec|平均每秒SQL重编译数|个/秒|
|rds_cluster_sqlserver_id<br>（Microsoft SQL Server实例）|rds062_full_scans_per_sec|每秒全表扫描数|个/秒|
|rds_cluster_sqlserver_id<br>（Microsoft SQL Server实例）|rds063_errors_per_sec|每秒用户错误数|个/秒|
|rds_cluster_sqlserver_id<br>（Microsoft SQL Server实例）|rds064_latch_waits_per_sec|每秒闩锁等待数|个/秒|
|rds_cluster_sqlserver_id<br>（Microsoft SQL Server实例）|rds065_lock_waits_per_sec|每秒锁等待次数|个/秒|
|rds_cluster_sqlserver_id<br>（Microsoft SQL Server实例）|rds066_lock_requests_per_sec|每秒锁请求次数|个/秒|
|rds_cluster_sqlserver_id<br>（Microsoft SQL Server实例）|rds067_timeouts_per_sec|每秒锁超时次数|个/秒|
|rds_cluster_sqlserver_id<br>（Microsoft SQL Server实例）|rds068_avg_lock_wait_time|平均锁等待延迟|毫秒|
|rds_cluster_sqlserver_id<br>（Microsoft SQL Server实例）|rds069_deadlocks_per_sec|每秒死锁次数|个/秒|
|rds_cluster_sqlserver_id<br>（Microsoft SQL Server实例）|rds070_checkpoint_pages_per_sec|每秒检查点写入Page数|个/秒|
|rds_cluster_sqlserver_id<br>（Microsoft SQL Server实例）|rds077_replication_delay|数据同步延迟|s|
|rds_cluster_sqlserver_id<br>（Microsoft SQL Server实例）|mssql_mem_grant_pending|待内存授权进程数|个|
|rds_cluster_sqlserver_id<br>（Microsoft SQL Server实例）|mssql_lazy_write_per_sec|每秒惰性写入缓存数|个/秒|
|rds_cluster_sqlserver_id<br>（Microsoft SQL Server实例）|mssql_page_life_expectancy|无引用页缓冲池停留时间|秒|
|rds_cluster_sqlserver_id<br>（Microsoft SQL Server实例）|mssql_page_reads_per_sec|每秒页读取次数|个|
|rds_cluster_sqlserver_id<br>（Microsoft SQL Server实例）|mssql_worker_threads_usage_rate|工作线程使用率|%|
|rds_cluster_sqlserver_id<br>（Microsoft SQL Server实例）|mssql_tempdb_disk_size|临时表空间大小|MB(IEC)|