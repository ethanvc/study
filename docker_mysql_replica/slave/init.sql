-- 等待主库启动完成
SELECT SLEEP(10);

-- 配置主从复制（使用 GTID 自动定位）
CHANGE MASTER TO
  MASTER_HOST='mysql-master',
  MASTER_PORT=3306,
  MASTER_USER='repl',
  MASTER_PASSWORD='repl123',
  MASTER_AUTO_POSITION=1;

-- 启动从库复制
START SLAVE;

-- 显示从库状态
SHOW SLAVE STATUS\G

