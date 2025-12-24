# PostgreSQL 主从复制环境

## 架构

```
postgres-1 (主库) ──同步复制──> postgres-2 (从库)
                └──同步复制──> postgres-3 (从库)
```

## 配置说明

| 配置项   | 值               | 说明                                |
| -------- | ---------------- | ----------------------------------- |
| 复制模式 | 同步复制         | `synchronous_commit = remote_apply` |
| 同步策略 | FIRST 1          | 至少等待 1 个从库确认               |
| 复制用户 | repl_user        | 密码: 123456                        |
| 复制槽   | replica_slot_2/3 | 确保 WAL 不会被清理                 |

## 使用方法

### 1. 启动主库

```bash
docker compose up -d postgres-1
# 等待初始化完成
docker logs -f postgres-1
```

### 2. 初始化从库

从库需要从主库复制初始数据：

```bash
# 进入 postgres-2 容器
docker exec -it postgres-2 bash

# 执行初始化脚本
/replica_scripts/setup_replica.sh postgres-2 replica_slot_2
exit

# 重启容器
docker restart postgres-2
```

对 postgres-3 执行相同操作：

```bash
docker exec -it postgres-3 bash
/replica_scripts/setup_replica.sh postgres-3 replica_slot_3
exit
docker restart postgres-3
```

### 3. 检查复制状态

```bash
# 在主库查看
docker exec -it postgres-1 psql -U postgres -f /replica_scripts/check_replication.sql
```

## 端口映射

| 容器       | 端口 |
| ---------- | ---- |
| postgres-1 | 5432 |
| postgres-2 | 5433 |
| postgres-3 | 5434 |

## 连接测试

```bash
# 连接主库
psql -h localhost -p 5432 -U postgres -d testdb

# 连接从库
psql -h localhost -p 5433 -U postgres -d testdb
psql -h localhost -p 5434 -U postgres -d testdb
```

## 故障切换

如需将 postgres-2 提升为主库：

```bash
docker exec -it postgres-2 psql -U postgres -c "SELECT pg_promote();"
```

然后更新其他节点的 `primary_conninfo` 指向新主库。
