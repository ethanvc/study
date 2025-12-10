-- ============================================
-- mysql-3 初始化脚本
-- ============================================

-- 显示当前实例信息
SELECT 'Initializing mysql-3 (server-id=3)' AS Info;

-- 执行公共初始化脚本
SOURCE /replica_scripts/init.sql;

-- mysql-3 特定的初始化逻辑（如果有）
-- 例如：创建特定的数据库或表

SELECT 'mysql-3 initialization completed' AS Info;

