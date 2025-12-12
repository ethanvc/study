-- ============================================
-- mysql-1 初始化脚本
-- ============================================

-- 显示当前实例信息
SELECT 'Initializing mysql-1 (server-id=1)' AS Info;

-- 执行公共初始化脚本
SOURCE /replica_scripts/init.sql;

-- mysql-1 特定的初始化逻辑（如果有）
-- 例如：创建特定的数据库或表

SELECT 'mysql-1 initialization completed' AS Info;

