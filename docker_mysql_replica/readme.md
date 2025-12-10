# MySQL 主从复制环境

## 环境说明

- **MySQL 版本**: 8.0 (最新版)
- **主库端口**: 3306
- **从库端口**: 3307
- **复制模式**: GTID 模式（自动故障恢复）
- **复制用户**: repl / repl123
- **配置方式**: 🤖 完全自动化，启动即用

## 便捷脚本

| 脚本                                                                                                                                      | 功能                               |
| ----------------------------------------------------------------------------------------------------------------------------------------- | ---------------------------------- |
| [`check_replication.sh`](/Users/hongtao.xu/root/opensource/ethanvc/study/docker_mysql_replica/check_replication.sh)                       | 检查主从复制状态，显示关键指标     |
| [`test_replication.sh`](/Users/hongtao.xu/root/opensource/ethanvc/study/docker_mysql_replica/test_replication.sh)                         | 端到端测试，验证数据同步功能       |
| [`slave/auto-setup-replication.sh`](/Users/hongtao.xu/root/opensource/ethanvc/study/docker_mysql_replica/slave/auto-setup-replication.sh) | 从库自动配置脚本（容器内自动执行） |

另外还有 [`QUICKSTART.md`](/Users/hongtao.xu/root/opensource/ethanvc/study/docker_mysql_replica/QUICKSTART.md) 快速开始指南。

## 快速启动

### 1. 启动服务

```bash
docker compose up -d
```

**从库会自动配置复制！** 内置脚本会：
- ✅ 自动检测主库是否就绪
- ✅ 等待主库完全启动
- ✅ 自动配置主从复制关系
- ✅ 启动复制线程

整个过程大约需要 30-60 秒。

### 2. 检查容器状态

```bash
docker compose ps
```

应该看到两个容器都在运行：
- `mysql-master` - STATUS 显示 healthy
- `mysql-slave` - STATUS 显示 healthy

### 3. 查看自动配置进度

从库容器会在后台自动配置复制，使用便捷脚本查看状态：

```bash
./check_replication.sh
```

或手动查看配置日志：

```bash
docker compose exec mysql-slave cat /var/log/replication-setup.log
```

成功的日志应该包含：
```
✅ Replication configured successfully!
```

### 4. 验证主库状态

```bash
docker compose exec mysql-master mysql -uroot -proot -e "SHOW MASTER STATUS\G"
```

### 5. 验证从库复制状态

```bash
# 新版本命令
docker compose exec mysql-slave mysql -uroot -proot -e "SHOW REPLICA STATUS\G"

# 或使用兼容命令
docker compose exec mysql-slave mysql -uroot -proot -e "SHOW SLAVE STATUS\G"
```

检查以下关键字段：
- ✅ `Slave_IO_Running: Yes` (或 `Replica_IO_Running: Yes`)
- ✅ `Slave_SQL_Running: Yes` (或 `Replica_SQL_Running: Yes`)
- ✅ `Seconds_Behind_Source: 0` (或 `Seconds_Behind_Master: 0`)
- ✅ 如果看到任何 `Last_Error`，说明配置有问题

## 手动重新配置（可选）

如果自动配置失败，可以手动配置：

```bash
# 进入从库容器
docker compose exec mysql-slave bash

# 手动执行配置脚本
bash /docker-entrypoint-initdb.d/auto-setup-replication.sh

# 或直接执行 SQL
mysql -uroot -proot -e "
CHANGE REPLICATION SOURCE TO
  SOURCE_HOST='mysql-master',
  SOURCE_PORT=3306,
  SOURCE_USER='repl',
  SOURCE_PASSWORD='repl123',
  SOURCE_AUTO_POSITION=1;
START REPLICA;
"
```

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

### 查看自动配置日志

如果从库复制没有自动配置成功，首先查看日志：

```bash
docker compose exec mysql-slave cat /var/log/replication-setup.log
```

### 从库复制出错

```bash
# 查看错误信息
docker compose exec mysql-slave mysql -uroot -proot -e "SHOW SLAVE STATUS\G" | grep -E "Last_Error|Last_SQL_Error"

# 手动重新配置
docker compose exec mysql-slave bash /docker-entrypoint-initdb.d/auto-setup-replication.sh

# 跳过错误（仅限开发环境）
docker compose exec mysql-slave mysql -uroot -proot -e "STOP SLAVE; SET GLOBAL SQL_SLAVE_SKIP_COUNTER=1; START SLAVE;"
```

### 重建从库

```bash
# 停止从库
docker compose stop mysql-slave

# 删除从库数据（会触发重新初始化和自动配置）
docker compose down mysql-slave
docker volume rm docker_mysql_replica_mysql-slave-data

# 重新启动（会自动配置复制）
docker compose up -d mysql-slave

# 等待 30-60 秒后查看日志
docker compose exec mysql-slave cat /var/log/replication-setup.log
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

## 自动化配置原理

本环境实现了 **完全自动化的主从复制配置**，无需手动干预。

### 工作流程

1. **容器启动**：`docker compose up -d`
   - master 和 slave 容器同时启动
   - `depends_on` 确保 master 健康检查通过后才启动 slave

2. **初始化阶段**：
   - MySQL 执行 `/docker-entrypoint-initdb.d/` 目录下的脚本
   - 主库：执行 `master/init.sql` 创建复制用户
   - 从库：执行 `slave/01-init.sql` 启动后台配置脚本

3. **后台自动配置** ([`slave/auto-setup-replication.sh`](/Users/hongtao.xu/root/opensource/ethanvc/study/docker_mysql_replica/slave/auto-setup-replication.sh))：
   ```bash
   # 等待本地 MySQL 就绪
   while ! mysqladmin ping -h localhost; do sleep 2; done
   
   # 等待主库 MySQL 就绪
   while ! mysqladmin ping -h mysql-master; do sleep 2; done
   
   # 配置主从复制
   CHANGE REPLICATION SOURCE TO ...
   START REPLICA;
   ```

4. **日志记录**：所有操作记录到 `/var/log/replication-setup.log`

### 关键技术点

**问题**：为什么不能在 `init.sql` 中直接配置复制？
- Docker 的初始化脚本是同步执行的
- 如果在 SQL 中执行 `CHANGE MASTER TO`，会阻塞初始化
- 主库可能还未完全就绪

**解决方案**：后台脚本 + 主动检测
- 使用 `\! nohup bash script.sh &` 在后台启动脚本
- 脚本主动检测 master 和 slave 的就绪状态
- 配置完成后记录日志，不影响容器启动

**优点**：
- ✅ 完全自动化，无需手动操作
- ✅ 健壮性强，自动重试检测
- ✅ 日志完整，方便排查问题
- ✅ 不阻塞容器启动

### 文件说明

- [`slave/init.sql`](/Users/hongtao.xu/root/opensource/ethanvc/study/docker_mysql_replica/slave/init.sql) - 启动后台配置脚本
- [`slave/auto-setup-replication.sh`](/Users/hongtao.xu/root/opensource/ethanvc/study/docker_mysql_replica/slave/auto-setup-replication.sh) - 自动配置复制的核心脚本
- `/var/log/replication-setup.log` - 配置过程日志（容器内）

### MySQL 8.0 命令兼容性

脚本自动处理新旧命令兼容：

| MySQL 版本 | 主从配置命令                                     |
| ---------- | ------------------------------------------------ |
| 8.0.23+    | `CHANGE REPLICATION SOURCE TO` + `START REPLICA` |
| < 8.0.23   | `CHANGE MASTER TO` + `START SLAVE`               |

脚本会先尝试新命令，失败后自动降级到旧命令。


