select @@GLOBAL.read_only;
select @@GLOBAL.super_read_only;
select @@GLOBAL.offline_mode;

-- 显示当前半同步主库状态
SHOW VARIABLES LIKE 'rpl_semi_sync_source%';

-- 显示当前半同步从库状态
SHOW VARIABLES LIKE 'rpl_semi_sync_replica%';

SHOW STATUS LIKE 'Rpl_semi_sync_%';


-- 显示当前复制状态
SHOW REPLICA STATUS\G

-- 显示当前复制状态
SHOW MASTER STATUS\G

-- 显示当前复制状态
SHOW SLAVE STATUS\G