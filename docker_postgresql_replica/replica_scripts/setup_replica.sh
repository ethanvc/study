#!/bin/bash
# ============================================
# 初始化从库脚本
# 用法: ./setup_replica.sh <replica_name> <slot_name>
# 示例: ./setup_replica.sh postgres-2 replica_slot_2
# ============================================

REPLICA_NAME=${1:-postgres-2}
SLOT_NAME=${2:-replica_slot_2}
PRIMARY_HOST="postgres-1"
REPL_USER="repl_user"
REPL_PASSWORD="123456"

echo "=== 初始化从库 $REPLICA_NAME ==="

# 停止 PostgreSQL（如果正在运行）
pg_ctl stop -D /var/lib/postgresql/data -m fast 2>/dev/null || true

# 清空数据目录
rm -rf /var/lib/postgresql/data/*

# 使用 pg_basebackup 从主库复制数据
echo "正在从主库复制数据..."
PGPASSWORD=$REPL_PASSWORD pg_basebackup \
    -h $PRIMARY_HOST \
    -p 5432 \
    -U $REPL_USER \
    -D /var/lib/postgresql/data \
    -Fp -Xs -P -R \
    -S $SLOT_NAME

# 创建 standby.signal 文件（PostgreSQL 12+ 表示这是从库）
touch /var/lib/postgresql/data/standby.signal

echo "=== 从库 $REPLICA_NAME 初始化完成 ==="
echo "请重启容器以应用配置"
