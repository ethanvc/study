-- 从库初始化脚本
-- 启动后台脚本自动配置主从复制
-- 使用 nohup 确保脚本在后台运行，不阻塞初始化过程

\! nohup bash /docker-entrypoint-initdb.d/auto-setup-replication.sh > /dev/null 2>&1 &

-- 提示信息
SELECT 'Replication auto-setup script started in background. Check /var/log/replication-setup.log for progress.' AS info;
