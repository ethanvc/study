# MySQL 主从复制环境

## 环境说明

- **MySQL 版本**: 8.0 (最新版)
- **主库端口**: 3306
- **从库端口**: 3307
- **复制模式**: GTID 模式
- **复制用户**: repl / repl123

## 快速启动

### 1. 启动服务

```bash
docker compose up -d
```

### 2. 检查容器状态

```bash
docker compose ps
```

### 3. 验证主库状态

```bash
docker compose exec mysql-master mysql -uroot -proot -e "SHOW MASTER STATUS\G"
```

### 4. 验证从库状态

```bash
docker compose exec mysql-slave mysql -uroot -proot -e "SHOW SLAVE STATUS\G"
```

检查以下关键字段：
- `Slave_IO_Running: Yes`
- `Slave_SQL_Running: Yes`
- `Seconds_Behind_Master: 0`

## 测试主从复制

### 在主库创建数据

```bash
# 连接主库
docker compose exec mysql-master mysql -uroot -proot testdb

# 创建表并插入数据
CREATE TABLE users (
    id INT PRIMARY KEY AUTO_INCREMENT,
    name VARCHAR(50),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

INSERT INTO users (name) VALUES ('Alice'), ('Bob'), ('Charlie');
```

### 在从库验证数据

```bash
# 连接从库
docker compose exec mysql-slave mysql -uroot -proot testdb

# 查询数据
SELECT * FROM users;
```

## 常用命令

### 查看 binlog 文件

```bash
docker compose exec mysql-master mysql -uroot -proot -e "SHOW BINARY LOGS;"
```

### 查看 binlog 内容

```bash
docker compose exec mysql-master mysqlbinlog /var/lib/mysql/mysql-bin.000001
```

### 重置从库复制

```bash
docker compose exec mysql-slave mysql -uroot -proot -e "STOP SLAVE; RESET SLAVE; START SLAVE;"
```

### 查看复制延迟

```bash
docker compose exec mysql-slave mysql -uroot -proot -e "SHOW SLAVE STATUS\G" | grep Seconds_Behind_Master
```

## 故障排查

### 从库复制出错

```bash
# 查看错误信息
docker compose exec mysql-slave mysql -uroot -proot -e "SHOW SLAVE STATUS\G" | grep -E "Last_Error|Last_SQL_Error"

# 跳过错误（仅限开发环境）
docker compose exec mysql-slave mysql -uroot -proot -e "STOP SLAVE; SET GLOBAL SQL_SLAVE_SKIP_COUNTER=1; START SLAVE;"
```

### 重建从库

```bash
# 停止从库
docker compose stop mysql-slave

# 删除从库数据
docker compose down mysql-slave
docker volume rm docker_mysql_replica_mysql-slave-data

# 重新启动
docker compose up -d mysql-slave
```

## 性能监控

### 查看复制线程状态

```bash
docker compose exec mysql-slave mysql -uroot -proot -e "SHOW PROCESSLIST\G" | grep -A 5 "system user"
```

### 查看 GTID 信息

```bash
# 主库
docker compose exec mysql-master mysql -uroot -proot -e "SHOW GLOBAL VARIABLES LIKE 'gtid_executed';"

# 从库
docker compose exec mysql-slave mysql -uroot -proot -e "SHOW GLOBAL VARIABLES LIKE 'gtid_executed';"
```

## 清理环境

```bash
# 停止并删除容器
docker compose down

# 删除数据卷（危险操作！会删除所有数据）
docker compose down -v
```

## 配置说明

### 主库配置重点 (master/my.cnf)

- `server-id=1`: 服务器唯一标识
- `log-bin=mysql-bin`: 开启二进制日志
- `binlog-format=ROW`: 使用行级复制（最安全）
- `gtid-mode=ON`: 开启 GTID 模式（推荐）

### 从库配置重点 (slave/my.cnf)

- `server-id=2`: 服务器唯一标识（不能与主库相同）
- `relay-log=mysql-relay-bin`: 中继日志配置
- `read_only=ON`: 从库只读
- `gtid-mode=ON`: 开启 GTID 模式

## 注意事项

1. **生产环境建议**：
   - 修改默认密码
   - 配置持久化存储路径
   - 设置合适的资源限制
   - 启用 SSL 加密复制

2. **GTID 模式优势**：
   - 简化故障恢复
   - 自动定位复制位置
   - 支持多源复制

3. **监控指标**：
   - Seconds_Behind_Master (复制延迟)
   - Slave_IO_Running (IO 线程状态)
   - Slave_SQL_Running (SQL 线程状态)

