-- ============================================
-- 公共初始化脚本 - 用于主从复制环境
-- ============================================

-- 创建复制用户
CREATE USER IF NOT EXISTS 'repl_user'@'%' IDENTIFIED WITH mysql_native_password BY 'repl123';

-- 授予复制权限
GRANT REPLICATION SLAVE ON *.* TO 'repl_user'@'%';

-- 刷新权限
FLUSH PRIVILEGES;

-- 显示创建的用户
SELECT CONCAT('Replication user created: ', User, '@', Host) AS Info 
FROM mysql.user 
WHERE User='repl';

