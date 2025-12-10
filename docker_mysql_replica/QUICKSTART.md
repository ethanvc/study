# 快速开始 - 5 分钟搭建 MySQL 主从复制

## 一键启动

```bash
cd docker_mysql_replica
docker compose up -d
```

就是这么简单！🎉

## 等待自动配置

从库会在后台自动配置复制，大约需要 30-60 秒。你可以：

- 喝口水 ☕
- 或者运行检查脚本看进度：

```bash
./check_replication.sh
```

## 验证复制

### 在主库创建测试数据

```bash
docker compose exec mysql-master mysql -uroot -proot -e "
CREATE DATABASE demo;
USE demo;
CREATE TABLE users (id INT PRIMARY KEY, name VARCHAR(50));
INSERT INTO users VALUES (1, 'Alice'), (2, 'Bob'), (3, 'Charlie');
SELECT * FROM users;
"
```

### 在从库查看数据（应该自动同步）

```bash
docker compose exec mysql-slave mysql -uroot -proot -e "
USE demo;
SELECT * FROM users;
"
```

如果看到相同的数据，恭喜！主从复制工作正常！🎊

## 管理命令

```bash
# 查看容器状态
docker compose ps

# 查看主库状态
docker compose exec mysql-master mysql -uroot -proot -e "SHOW MASTER STATUS\G"

# 查看从库状态
docker compose exec mysql-slave mysql -uroot -proot -e "SHOW REPLICA STATUS\G"

# 查看自动配置日志
docker compose exec mysql-slave cat /var/log/replication-setup.log

# 停止环境
docker compose down

# 完全清理（包括数据）
docker compose down -v
```

## 连接信息

### 主库
- **地址**: localhost:3306
- **用户**: root / testuser
- **密码**: root / testpass
- **数据库**: testdb

### 从库（只读）
- **地址**: localhost:3307
- **用户**: root / testuser
- **密码**: root / testpass
- **数据库**: testdb

## 故障排查

如果复制没有自动配置成功：

```bash
# 1. 查看配置日志
docker compose exec mysql-slave cat /var/log/replication-setup.log

# 2. 手动重新配置
docker compose exec mysql-slave bash /docker-entrypoint-initdb.d/auto-setup-replication.sh

# 3. 重建从库（会触发自动配置）
docker compose down mysql-slave
docker volume rm docker_mysql_replica_mysql-slave-data
docker compose up -d mysql-slave
```

## 特性

✅ **完全自动化** - 无需手动配置，启动即用  
✅ **健康检查** - 确保服务就绪  
✅ **GTID 模式** - 简化故障恢复  
✅ **详细日志** - 方便排查问题  
✅ **MySQL 8.0** - 使用最新版本  

更多详细信息请查看 [README.md](readme.md)

