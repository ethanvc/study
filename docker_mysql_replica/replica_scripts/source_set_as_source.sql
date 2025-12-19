-- ============================================
-- 将当前节点设置为主库（Source）
-- 用法: mysql -uroot -p < /replica_scripts/source_set_as_source.sql
-- ============================================

-- 启用半同步主库
SET GLOBAL rpl_semi_sync_source_enabled = ON;

-- 至少等待多少个从库确认（生产建议至少1个）
SET GLOBAL rpl_semi_sync_source_wait_for_replica_count = 1;

-- 等待从库确认的超时时间（毫秒），超时后降级为异步复制
-- 86400000 = 24小时，表示宁可阻塞也不丢数据
SET GLOBAL rpl_semi_sync_source_timeout = 86400000;

-- 等待从库将事务写入 relay log 后确认（AFTER_SYNC 更安全）
SET GLOBAL rpl_semi_sync_source_wait_point = 'AFTER_SYNC';

SET GLOBAL read_only = OFF;
SET GLOBAL super_read_only = OFF;



-- 显示当前半同步主库状态
SHOW VARIABLES LIKE 'rpl_semi_sync_source%';
