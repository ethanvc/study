-- ============================================
-- mysql 初始化
-- ============================================

-- 创建复制用户
CREATE USER IF NOT EXISTS 'repl_user'@'%' IDENTIFIED WITH mysql_native_password BY '123456';

-- 授予复制权限
GRANT REPLICATION SLAVE,REPLICATION_SLAVE_ADMIN,REPLICATION CLIENT ON *.* TO 'repl_user'@'%';
-- 刷新权限
FLUSH PRIVILEGES;

-- 显示创建的用户
SELECT CONCAT('Replication user created: ', User, '@', Host) AS Info 
FROM mysql.user 
WHERE User='repl_user';


SET PERSIST super_read_only = OFF;
SET PERSIST read_only = OFF;
SET PERSIST offline_mode = OFF;