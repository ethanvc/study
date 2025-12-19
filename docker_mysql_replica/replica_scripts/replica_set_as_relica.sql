-- ============================================
-- 将当前节点设置为从库（Replica）
-- 用法: mysql -uroot -p < /replica_scripts/replica_set_as_relica.sql
-- 注意: 需要先修改 SOURCE_HOST 为实际的主库地址
-- ============================================

-- 启用半同步从库
SET GLOBAL rpl_semi_sync_replica_enabled = ON;

-- 设置只读模式
SET GLOBAL read_only = ON;
SET GLOBAL super_read_only = ON;

-- 配置复制源（主库）
-- 请根据实际情况修改 SOURCE_HOST
CHANGE REPLICATION SOURCE TO
    SOURCE_HOST = 'mysql-1',
    SOURCE_PORT = 3306,
    SOURCE_USER = 'repl_user',
    SOURCE_PASSWORD = '123456',
    SOURCE_AUTO_POSITION = 1,
    GET_SOURCE_PUBLIC_KEY = 1;

-- 启动复制
START REPLICA;

-- 显示复制状态
SHOW REPLICA STATUS\G
