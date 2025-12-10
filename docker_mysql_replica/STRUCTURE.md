# 项目结构说明

```
docker_mysql_replica/
├── docker-compose.yml              # Docker Compose 配置文件
├── readme.md                       # 完整文档
├── QUICKSTART.md                   # 5 分钟快速开始指南
│
├── master/                         # 主库配置
│   ├── my.cnf                      # 主库 MySQL 配置
│   └── init.sql                    # 主库初始化脚本（创建复制用户）
│
├── slave/                          # 从库配置
│   ├── my.cnf                      # 从库 MySQL 配置
│   ├── init.sql                    # 从库初始化脚本（启动自动配置）
│   └── auto-setup-replication.sh   # 自动配置复制的核心脚本
│
├── check_replication.sh            # 检查复制状态脚本
└── test_replication.sh             # 端到端测试脚本
```

## 核心文件说明

### 配置文件

**docker-compose.yml**
- 定义 master 和 slave 两个服务
- 配置健康检查确保服务就绪
- 映射端口：master(3306), slave(3307)
- 挂载配置和初始化脚本

**master/my.cnf**
- 主库 MySQL 配置
- 启用 binlog（二进制日志）
- 配置 GTID 模式
- 设置 server-id=1

**slave/my.cnf**
- 从库 MySQL 配置
- 启用 relay-log（中继日志）
- 配置只读模式
- 设置 server-id=2

### 初始化脚本

**master/init.sql**
```sql
-- 创建复制用户
CREATE USER 'repl'@'%' IDENTIFIED BY 'repl123';
GRANT REPLICATION SLAVE ON *.* TO 'repl'@'%';
```

**slave/init.sql**
```sql
-- 在后台启动自动配置脚本
\! nohup bash /docker-entrypoint-initdb.d/auto-setup-replication.sh > /dev/null 2>&1 &
```

**slave/auto-setup-replication.sh**
- 等待本地 MySQL 就绪
- 等待主库 MySQL 就绪
- 执行 `CHANGE REPLICATION SOURCE TO`
- 启动复制线程
- 记录日志到 `/var/log/replication-setup.log`

### 便捷脚本

**check_replication.sh**
- 显示容器状态
- 显示自动配置日志
- 显示主库状态
- 显示从库复制状态
- 提取并高亮关键指标

**test_replication.sh**
- 测试 1: 检查容器运行状态
- 测试 2: 检查主库状态
- 测试 3: 检查从库复制状态
- 测试 4: 数据同步测试（创建测试数据，验证同步）
- 测试 5: GTID 模式验证
- 测试 6: 从库只读模式验证

## 工作流程

```
启动阶段
┌─────────────────────────────────────────────────┐
│ 1. docker compose up -d                         │
│    ↓                                            │
│ 2. 启动 master 容器                             │
│    ├── 加载 master/my.cnf                       │
│    ├── 等待健康检查通过                         │
│    └── 执行 master/init.sql (创建复制用户)      │
│    ↓                                            │
│ 3. 启动 slave 容器 (depends_on master healthy) │
│    ├── 加载 slave/my.cnf                        │
│    ├── 执行 slave/init.sql                      │
│    └── 启动后台脚本 auto-setup-replication.sh  │
└─────────────────────────────────────────────────┘

自动配置阶段 (后台执行)
┌─────────────────────────────────────────────────┐
│ auto-setup-replication.sh:                      │
│    ↓                                            │
│ 1. 循环检测本地 MySQL (最多 30 次 * 2秒)       │
│    ↓                                            │
│ 2. 循环检测主库 MySQL (最多 30 次 * 2秒)       │
│    ↓                                            │
│ 3. 额外等待 5 秒 (确保主库完全初始化)          │
│    ↓                                            │
│ 4. 执行 CHANGE REPLICATION SOURCE TO           │
│    ↓                                            │
│ 5. 执行 START REPLICA                           │
│    ↓                                            │
│ 6. 记录日志到 /var/log/replication-setup.log   │
└─────────────────────────────────────────────────┘

运行状态
┌─────────────────────────────────────────────────┐
│ Master (3306)                                   │
│   ├── binlog: mysql-bin.000001                  │
│   ├── GTID: uuid:1-N                            │
│   └── 接受写入请求                              │
│       ↓ 复制                                    │
│ Slave (3307)                                    │
│   ├── relay-log: mysql-relay-bin.000001         │
│   ├── IO线程: 从主库拉取 binlog                 │
│   ├── SQL线程: 执行 relay-log                   │
│   └── 只读模式 (read_only=ON)                   │
└─────────────────────────────────────────────────┘
```

## 关键技术点

### 1. 后台脚本启动
使用 SQL 的 `\!` 命令执行 shell 命令：
```sql
\! nohup bash script.sh > /dev/null 2>&1 &
```

### 2. 健康检查
```yaml
healthcheck:
  test: ["CMD", "mysqladmin", "ping", "-h", "localhost", "-uroot", "-proot"]
  interval: 10s
  timeout: 5s
  retries: 5
```

### 3. 依赖关系
```yaml
depends_on:
  mysql-master:
    condition: service_healthy
```

### 4. 主动检测
```bash
for i in {1..30}; do
    if mysqladmin ping -h mysql-master -uroot -p"${MYSQL_ROOT_PASSWORD}" --silent; then
        break
    fi
    sleep 2
done
```

### 5. 命令兼容
```bash
# 尝试新命令
CHANGE REPLICATION SOURCE TO ...
if [ $? -ne 0 ]; then
    # 降级到旧命令
    CHANGE MASTER TO ...
fi
```

## 日志位置

- **自动配置日志**: `/var/log/replication-setup.log` (容器内)
- **MySQL 错误日志**: `/var/lib/mysql/mysql-error.log` (容器内)
- **容器日志**: `docker logs mysql-slave`

## 数据持久化

使用 Docker 卷存储数据：
- `mysql-master-data`: 主库数据
- `mysql-slave-data`: 从库数据

## 环境变量

在 docker-compose.yml 中定义：
- `MYSQL_ROOT_PASSWORD`: root 密码
- `MYSQL_DATABASE`: 默认数据库
- `MYSQL_USER`: 应用用户
- `MYSQL_PASSWORD`: 应用用户密码

