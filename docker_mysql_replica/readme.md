# MySQL 三节点环境

## 环境说明

- **MySQL 版本**: Ubuntu MySQL 8.0
- **节点数量**: 3 个独立的 MySQL 实例
- **端口映射**: 
  - mysql-1: 3306
  - mysql-2: 3307
  - mysql-3: 3308

## 文件结构

```
docker_mysql_replica/
├── docker-compose.yml          # Docker Compose 配置
├── replica_scripts/            # 公共脚本目录
│   └── init.sql                # 公共初始化脚本（创建复制用户等）
├── mysql-1/
│   ├── my.cnf                  # 配置文件（server-id=1）
│   └── init.sql                # 初始化脚本（引用公共脚本）
├── mysql-2/
│   ├── my.cnf                  # 配置文件（server-id=2）
│   └── init.sql                # 初始化脚本（引用公共脚本）
└── mysql-3/
    ├── my.cnf                  # 配置文件（server-id=3）
    └── init.sql                # 初始化脚本（引用公共脚本）
```

## 初始化脚本设计

### 1. 公共脚本（replica_scripts/init.sql）

包含所有实例共享的初始化逻辑：
- 创建复制用户
- 授予复制权限
- 其他公共配置

```sql
-- 创建复制用户
CREATE USER IF NOT EXISTS 'repl'@'%' IDENTIFIED WITH mysql_native_password BY 'repl123';

-- 授予复制权限
GRANT REPLICATION SLAVE ON *.* TO 'repl'@'%';

-- 刷新权限
FLUSH PRIVILEGES;
```

### 2. 容器级别脚本（mysql-X/init.sql）

每个容器有自己的 init.sql，通过 `SOURCE` 命令引用公共脚本：

```sql
-- 显示当前实例信息
SELECT 'Initializing mysql-1 (server-id=1)' AS Info;

-- 执行公共初始化脚本
SOURCE /replica_scripts/init.sql;

-- 容器特定的初始化逻辑
-- ...

SELECT 'mysql-1 initialization completed' AS Info;
```

### 3. 挂载配置

在 docker-compose.yml 中：

```yaml
services:
  mysql-1:
    volumes:
      - ./mysql-1/my.cnf:/etc/mysql/conf.d/my.cnf
      - ./mysql-1/init.sql:/docker-entrypoint-initdb.d/init.sql
      - ./replica_scripts:/replica_scripts:ro  # 只读挂载公共脚本目录
```

说明：
- `./mysql-1/init.sql` → 容器特定的初始化脚本
- `./replica_scripts:/replica_scripts:ro` → 公共脚本目录（只读模式）
- `:ro` 表示 read-only，防止容器修改公共脚本

## 快速启动

```bash
# 启动所有容器
docker compose up -d

# 查看容器状态
docker compose ps

# 查看初始化日志
docker logs mysql-1
docker logs mysql-2
docker logs mysql-3

# 连接到各个 MySQL
docker exec -it mysql-1 mysql -uroot -proot
docker exec -it mysql-2 mysql -uroot -proot
docker exec -it mysql-3 mysql -uroot -proot
```

## 验证初始化

### 1. 检查复制用户

```bash
docker exec -it mysql-1 mysql -uroot -proot -e "SELECT User, Host FROM mysql.user WHERE User='repl';"
```

应该看到：
```
+------+------+
| User | Host |
+------+------+
| repl | %    |
+------+------+
```

### 2. 测试复制用户登录

```bash
docker exec -it mysql-1 mysql -urepl -prepl123 -e "SHOW GRANTS;"
```

### 3. 检查 server-id

```bash
docker exec -it mysql-1 mysql -uroot -proot -e "SHOW VARIABLES LIKE 'server_id';"
docker exec -it mysql-2 mysql -uroot -proot -e "SHOW VARIABLES LIKE 'server_id';"
docker exec -it mysql-3 mysql -uroot -proot -e "SHOW VARIABLES LIKE 'server_id';"
```

应该分别显示 1、2、3。

## 配置主从复制

### 将 mysql-1 作为主库，mysql-2 和 mysql-3 作为从库

**在 mysql-2 上执行**：

```bash
docker exec -it mysql-2 mysql -uroot -proot
```

```sql
-- 配置主库信息
CHANGE MASTER TO
  MASTER_HOST='mysql-1',        -- 主库容器名
  MASTER_PORT=3306,             -- 主库端口
  MASTER_USER='repl',           -- 复制用户名
  MASTER_PASSWORD='repl123',    -- 复制密码
  MASTER_AUTO_POSITION=1;       -- 使用 GTID 自动定位

-- 启动复制
START SLAVE;

-- 查看复制状态
SHOW SLAVE STATUS\G
```

**在 mysql-3 上执行同样的步骤**。

检查关键字段：
- `Slave_IO_Running: Yes` - IO 线程运行正常
- `Slave_SQL_Running: Yes` - SQL 线程运行正常
- `Seconds_Behind_Master: 0` - 没有延迟

### 测试主从复制

**在主库 (mysql-1) 创建数据**：

```bash
docker exec -it mysql-1 mysql -uroot -proot -e "
CREATE DATABASE repl_test;
USE repl_test;
CREATE TABLE users (id INT PRIMARY KEY, name VARCHAR(50));
INSERT INTO users VALUES (1, 'Alice'), (2, 'Bob'), (3, 'Charlie');
SELECT * FROM users;
"
```

**在从库验证数据已同步**：

```bash
# mysql-2
docker exec -it mysql-2 mysql -uroot -proot -e "USE repl_test; SELECT * FROM users;"

# mysql-3
docker exec -it mysql-3 mysql -uroot -proot -e "USE repl_test; SELECT * FROM users;"
```

## 修改公共脚本

如果需要修改公共初始化逻辑，只需编辑 `replica_scripts/init.sql` 文件即可。

**注意**：
- 修改后需要重新初始化容器才会生效
- 初始化脚本只在容器**首次创建**时执行

```bash
# 删除容器和数据
docker compose down

# 重新启动，执行更新后的脚本
docker compose up -d
```

## 添加更多公共脚本

可以在 `replica_scripts/` 目录下添加更多脚本：

```
replica_scripts/
├── init.sql              # 主初始化脚本
├── create_databases.sql  # 创建数据库
└── grant_permissions.sql # 权限配置
```

然后在容器级别的 `init.sql` 中按需引用：

```sql
SOURCE /replica_scripts/init.sql;
SOURCE /replica_scripts/create_databases.sql;
SOURCE /replica_scripts/grant_permissions.sql;
```

## 优点

这种设计的优势：

1. **集中管理**：公共逻辑在一个地方维护
2. **代码复用**：避免重复代码
3. **易于维护**：修改一次，所有容器生效
4. **灵活扩展**：容器可以添加特定的初始化逻辑
5. **只读保护**：`:ro` 防止容器意外修改公共脚本

## 常用命令

```bash
# 查看容器日志
docker logs mysql-1
docker logs mysql-2
docker logs mysql-3

# 查看主库状态
docker exec -it mysql-1 mysql -uroot -proot -e "SHOW MASTER STATUS\G"

# 查看从库状态
docker exec -it mysql-2 mysql -uroot -proot -e "SHOW SLAVE STATUS\G"
docker exec -it mysql-3 mysql -uroot -proot -e "SHOW SLAVE STATUS\G"

# 停止所有容器
docker compose down

# 查看挂载的脚本
docker exec -it mysql-1 ls -la /replica_scripts/
docker exec -it mysql-1 ls -la /docker-entrypoint-initdb.d/

# 手动执行公共脚本（容器运行时）
docker exec -it mysql-1 mysql -uroot -proot < replica_scripts/init.sql
```

## 注意事项

### 1. 初始化时机

- 初始化脚本只在容器**第一次创建**时执行
- 容器重启不会重新执行
- 如需重新初始化，必须删除容器：`docker compose down`

### 2. SQL SOURCE 命令

`SOURCE` 命令需要使用容器内的绝对路径：
```sql
SOURCE /replica_scripts/init.sql;  -- ✅ 正确
SOURCE ./replica_scripts/init.sql; -- ❌ 错误
```

### 3. 只读挂载

公共脚本目录使用只读挂载（`:ro`），防止容器修改：
```yaml
- ./replica_scripts:/replica_scripts:ro
```

### 4. 数据持久化

当前配置**不使用持久化存储**，数据存储在容器内部：
- ✅ 容器重启时数据保留
- ❌ 容器删除时数据丢失
- ⚠️ 适合测试环境
