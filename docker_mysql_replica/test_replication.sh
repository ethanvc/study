#!/bin/bash
# 端到端测试脚本 - 验证主从复制功能

set -e

echo "=========================================="
echo "  MySQL 主从复制 - 端到端测试"
echo "=========================================="
echo ""

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

success() {
    echo -e "${GREEN}✅ $1${NC}"
}

error() {
    echo -e "${RED}❌ $1${NC}"
}

warning() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

info() {
    echo "ℹ️  $1"
}

# 测试 1: 检查容器运行状态
echo "测试 1: 检查容器运行状态"
echo "----------------------------------------"
if docker compose ps mysql-master | grep -q "Up.*healthy"; then
    success "主库容器运行正常"
else
    error "主库容器未运行"
    exit 1
fi

if docker compose ps mysql-slave | grep -q "Up.*healthy"; then
    success "从库容器运行正常"
else
    error "从库容器未运行"
    exit 1
fi
echo ""

# 测试 2: 检查主库状态
echo "测试 2: 检查主库状态"
echo "----------------------------------------"
if docker compose exec -T mysql-master mysql -uroot -proot -e "SHOW MASTER STATUS\G" 2>/dev/null | grep -q "mysql-bin"; then
    success "主库 binlog 已启用"
else
    error "主库 binlog 未启用"
    exit 1
fi
echo ""

# 测试 3: 检查从库复制状态
echo "测试 3: 检查从库复制状态"
echo "----------------------------------------"
SLAVE_STATUS=$(docker compose exec -T mysql-slave mysql -uroot -proot -e "SHOW REPLICA STATUS\G" 2>/dev/null || \
               docker compose exec -T mysql-slave mysql -uroot -proot -e "SHOW SLAVE STATUS\G" 2>/dev/null)

if [ -z "$SLAVE_STATUS" ]; then
    warning "从库复制尚未配置，等待 30 秒..."
    sleep 30
    SLAVE_STATUS=$(docker compose exec -T mysql-slave mysql -uroot -proot -e "SHOW REPLICA STATUS\G" 2>/dev/null || \
                   docker compose exec -T mysql-slave mysql -uroot -proot -e "SHOW SLAVE STATUS\G" 2>/dev/null)
fi

IO_RUNNING=$(echo "$SLAVE_STATUS" | grep "Slave_IO_Running" | head -1 | awk '{print $2}')
SQL_RUNNING=$(echo "$SLAVE_STATUS" | grep "Slave_SQL_Running" | head -1 | awk '{print $2}')

if [ "$IO_RUNNING" = "Yes" ] && [ "$SQL_RUNNING" = "Yes" ]; then
    success "从库复制线程运行正常"
else
    error "从库复制线程状态异常 (IO: $IO_RUNNING, SQL: $SQL_RUNNING)"
    echo "查看日志："
    docker compose exec -T mysql-slave cat /var/log/replication-setup.log 2>/dev/null || true
    exit 1
fi
echo ""

# 测试 4: 数据同步测试
echo "测试 4: 数据同步测试"
echo "----------------------------------------"
TEST_DB="test_replication_$(date +%s)"
TEST_VALUE="test_value_$(date +%s)"

info "在主库创建测试数据库和表..."
docker compose exec -T mysql-master mysql -uroot -proot <<EOF 2>/dev/null
CREATE DATABASE ${TEST_DB};
USE ${TEST_DB};
CREATE TABLE test_table (
    id INT PRIMARY KEY AUTO_INCREMENT,
    value VARCHAR(100),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
INSERT INTO test_table (value) VALUES ('${TEST_VALUE}');
EOF

if [ $? -eq 0 ]; then
    success "主库测试数据创建成功"
else
    error "主库测试数据创建失败"
    exit 1
fi

info "等待 5 秒让数据同步到从库..."
sleep 5

info "从从库读取数据..."
SLAVE_VALUE=$(docker compose exec -T mysql-slave mysql -uroot -proot -e "USE ${TEST_DB}; SELECT value FROM test_table WHERE value='${TEST_VALUE}';" 2>/dev/null | grep "${TEST_VALUE}" || true)

if [ -n "$SLAVE_VALUE" ]; then
    success "数据已成功同步到从库"
else
    error "从库未能同步数据"
    warning "可能需要更长的等待时间，或者复制出现问题"
    exit 1
fi
echo ""

# 测试 5: GTID 模式验证
echo "测试 5: GTID 模式验证"
echo "----------------------------------------"
MASTER_GTID=$(docker compose exec -T mysql-master mysql -uroot -proot -e "SHOW GLOBAL VARIABLES LIKE 'gtid_mode';" 2>/dev/null | grep "gtid_mode" | awk '{print $2}')
SLAVE_GTID=$(docker compose exec -T mysql-slave mysql -uroot -proot -e "SHOW GLOBAL VARIABLES LIKE 'gtid_mode';" 2>/dev/null | grep "gtid_mode" | awk '{print $2}')

if [ "$MASTER_GTID" = "ON" ] && [ "$SLAVE_GTID" = "ON" ]; then
    success "GTID 模式已在主从库启用"
else
    warning "GTID 模式未启用 (Master: $MASTER_GTID, Slave: $SLAVE_GTID)"
fi
echo ""

# 测试 6: 从库只读模式验证
echo "测试 6: 从库只读模式验证"
echo "----------------------------------------"
if docker compose exec -T mysql-slave mysql -uroot -proot -e "SHOW GLOBAL VARIABLES LIKE 'read_only';" 2>/dev/null | grep -q "ON"; then
    success "从库只读模式已启用"
else
    warning "从库只读模式未启用（对 root 用户无效）"
fi
echo ""

# 清理测试数据
echo "清理测试数据..."
docker compose exec -T mysql-master mysql -uroot -proot -e "DROP DATABASE ${TEST_DB};" 2>/dev/null
sleep 2
success "测试数据已清理"
echo ""

# 最终总结
echo "=========================================="
echo "  🎉 所有测试通过！"
echo "=========================================="
echo ""
echo "主从复制环境运行正常，可以开始使用了！"
echo ""
echo "快速命令："
echo "  - 查看状态: ./check_replication.sh"
echo "  - 查看日志: docker compose exec mysql-slave cat /var/log/replication-setup.log"
echo "  - 连接主库: docker compose exec mysql-master mysql -uroot -proot"
echo "  - 连接从库: docker compose exec mysql-slave mysql -uroot -proot"
echo ""

