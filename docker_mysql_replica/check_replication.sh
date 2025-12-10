#!/bin/bash
# MySQL 主从复制状态检查脚本

echo "=========================================="
echo "  MySQL 主从复制状态检查"
echo "=========================================="
echo ""

# 检查容器状态
echo "📦 容器状态："
docker compose ps
echo ""

# 检查从库自动配置日志
echo "📋 从库自动配置日志："
echo "----------------------------------------"
if docker compose exec -T mysql-slave test -f /var/log/replication-setup.log 2>/dev/null; then
    docker compose exec -T mysql-slave cat /var/log/replication-setup.log
else
    echo "⚠️  日志文件不存在，从库可能还在初始化中..."
fi
echo ""

# 检查主库状态
echo "🔵 主库状态："
echo "----------------------------------------"
docker compose exec -T mysql-master mysql -uroot -proot -e "SHOW MASTER STATUS\G" 2>/dev/null
echo ""

# 检查从库复制状态
echo "🟢 从库复制状态："
echo "----------------------------------------"
SLAVE_STATUS=$(docker compose exec -T mysql-slave mysql -uroot -proot -e "SHOW REPLICA STATUS\G" 2>/dev/null || \
               docker compose exec -T mysql-slave mysql -uroot -proot -e "SHOW SLAVE STATUS\G" 2>/dev/null)

if [ -n "$SLAVE_STATUS" ]; then
    echo "$SLAVE_STATUS"
    echo ""
    
    # 提取关键指标
    echo "📊 关键指标："
    echo "----------------------------------------"
    IO_RUNNING=$(echo "$SLAVE_STATUS" | grep "Slave_IO_Running" | head -1 | awk '{print $2}')
    SQL_RUNNING=$(echo "$SLAVE_STATUS" | grep "Slave_SQL_Running" | head -1 | awk '{print $2}')
    SECONDS_BEHIND=$(echo "$SLAVE_STATUS" | grep "Seconds_Behind_Master" | head -1 | awk '{print $2}')
    LAST_ERROR=$(echo "$SLAVE_STATUS" | grep "Last_Error:" | head -1 | cut -d: -f2-)
    
    echo -n "IO 线程运行: "
    if [ "$IO_RUNNING" = "Yes" ]; then
        echo "✅ Yes"
    else
        echo "❌ $IO_RUNNING"
    fi
    
    echo -n "SQL 线程运行: "
    if [ "$SQL_RUNNING" = "Yes" ]; then
        echo "✅ Yes"
    else
        echo "❌ $SQL_RUNNING"
    fi
    
    echo -n "复制延迟: "
    if [ "$SECONDS_BEHIND" = "0" ] || [ "$SECONDS_BEHIND" = "NULL" ]; then
        echo "✅ $SECONDS_BEHIND 秒"
    else
        echo "⚠️  $SECONDS_BEHIND 秒"
    fi
    
    if [ -n "$LAST_ERROR" ] && [ "$LAST_ERROR" != " " ]; then
        echo "❌ 最后错误: $LAST_ERROR"
    else
        echo "✅ 无错误"
    fi
    echo ""
    
    # 总结
    if [ "$IO_RUNNING" = "Yes" ] && [ "$SQL_RUNNING" = "Yes" ] && ([ "$SECONDS_BEHIND" = "0" ] || [ "$SECONDS_BEHIND" = "NULL" ]); then
        echo "🎉 主从复制运行正常！"
    else
        echo "⚠️  主从复制可能存在问题，请检查上述状态"
    fi
else
    echo "❌ 无法获取从库状态，可能还未配置复制"
    echo ""
    echo "💡 提示："
    echo "   1. 等待 30-60 秒让自动配置完成"
    echo "   2. 查看配置日志: docker compose exec mysql-slave cat /var/log/replication-setup.log"
    echo "   3. 手动配置: docker compose exec mysql-slave bash /docker-entrypoint-initdb.d/auto-setup-replication.sh"
fi

echo ""
echo "=========================================="

