-- 创建复制用户
CREATE USER IF NOT EXISTS 'repl'@'%' IDENTIFIED WITH mysql_native_password BY 'repl123';

-- 授予复制权限
GRANT REPLICATION SLAVE ON *.* TO 'repl'@'%';

-- 刷新权限
FLUSH PRIVILEGES;

-- 显示主库状态
SHOW MASTER STATUS;

