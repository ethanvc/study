#!/bin/bash
# MySQL Slave 自动配置脚本
# 此脚本会在后台运行，等待 master 就绪后自动配置主从复制

set -e

echo "[$(date)] Slave replication auto-setup script started" >> /var/log/replication-setup.log

# 等待本地 MySQL 完全启动
echo "[$(date)] Waiting for local MySQL to be ready..." >> /var/log/replication-setup.log
for i in {1..30}; do
    if mysqladmin ping -h localhost -uroot -p"${MYSQL_ROOT_PASSWORD}" --silent 2>/dev/null; then
        echo "[$(date)] Local MySQL is ready" >> /var/log/replication-setup.log
        break
    fi
    echo "[$(date)] Attempt $i: Local MySQL not ready yet, waiting..." >> /var/log/replication-setup.log
    sleep 2
done

# 等待 master 可连接
echo "[$(date)] Waiting for master MySQL to be ready..." >> /var/log/replication-setup.log
for i in {1..30}; do
    if mysqladmin ping -h mysql-master -uroot -p"${MYSQL_ROOT_PASSWORD}" --silent 2>/dev/null; then
        echo "[$(date)] Master MySQL is ready" >> /var/log/replication-setup.log
        break
    fi
    echo "[$(date)] Attempt $i: Master not ready yet, waiting..." >> /var/log/replication-setup.log
    sleep 2
done

# 额外等待，确保 master 的复制用户已创建
echo "[$(date)] Waiting additional 5 seconds for master to complete initialization..." >> /var/log/replication-setup.log
sleep 5

# 配置主从复制
echo "[$(date)] Configuring replication..." >> /var/log/replication-setup.log

# 尝试使用新版本命令
mysql -h localhost -uroot -p"${MYSQL_ROOT_PASSWORD}" <<EOF 2>> /var/log/replication-setup.log
STOP SLAVE;
RESET SLAVE ALL;

CHANGE REPLICATION SOURCE TO
  SOURCE_HOST='mysql-master',
  SOURCE_PORT=3306,
  SOURCE_USER='repl',
  SOURCE_PASSWORD='repl123',
  SOURCE_AUTO_POSITION=1;

START REPLICA;
EOF

# 如果新命令失败，尝试旧命令
if [ $? -ne 0 ]; then
    echo "[$(date)] New syntax failed, trying legacy syntax..." >> /var/log/replication-setup.log
    mysql -h localhost -uroot -p"${MYSQL_ROOT_PASSWORD}" <<EOF 2>> /var/log/replication-setup.log
STOP SLAVE;
RESET SLAVE ALL;

CHANGE MASTER TO
  MASTER_HOST='mysql-master',
  MASTER_PORT=3306,
  MASTER_USER='repl',
  MASTER_PASSWORD='repl123',
  MASTER_AUTO_POSITION=1;

START SLAVE;
EOF
fi

if [ $? -eq 0 ]; then
    echo "[$(date)] ✅ Replication configured successfully!" >> /var/log/replication-setup.log
    
    # 显示复制状态
    echo "[$(date)] Replication status:" >> /var/log/replication-setup.log
    mysql -h localhost -uroot -p"${MYSQL_ROOT_PASSWORD}" -e "SHOW REPLICA STATUS\G" >> /var/log/replication-setup.log 2>&1 || \
    mysql -h localhost -uroot -p"${MYSQL_ROOT_PASSWORD}" -e "SHOW SLAVE STATUS\G" >> /var/log/replication-setup.log 2>&1
else
    echo "[$(date)] ❌ Failed to configure replication" >> /var/log/replication-setup.log
    exit 1
fi

echo "[$(date)] Slave replication auto-setup completed" >> /var/log/replication-setup.log

