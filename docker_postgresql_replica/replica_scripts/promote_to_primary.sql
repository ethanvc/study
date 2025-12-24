-- ============================================
-- 提升从库为主库
-- 在要提升的从库执行
-- ============================================

-- 方式1: 使用 SQL 函数（PostgreSQL 12+）
SELECT pg_promote();

-- 方式2: 使用命令行
-- pg_ctl promote -D /var/lib/postgresql/data

-- 提升后，需要更新配置：
-- 1. 修改 postgresql.conf，添加 synchronous_standby_names
-- 2. 其他从库需要修改 primary_conninfo 指向新主库
-- 3. 重新加载配置: SELECT pg_reload_conf();
