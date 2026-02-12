# ClickHouse 3 节点集群（Docker Compose）

- **节点**：`clickhouse-1`、`clickhouse-2`、`clickhouse-3`（1 分片 3 副本）
- **Keeper**：跑在 `clickhouse-1` 上，2/3 仅作客户端连接

## 端口

| 节点 | HTTP | Native |
|------|------|--------|
| clickhouse-1 | 8123 | 9000 |
| clickhouse-2 | 8124 | 9001 |
| clickhouse-3 | 8125 | 9002 |

## 启动

```bash
cd clickhouse_docker
docker compose up -d
```

## 验证集群

连任意节点（如 `localhost:8123`）：

```sql
SELECT * FROM system.clusters WHERE cluster = 'cluster_3n';
```

## 目录

- [docker-compose.yml](docker-compose.yml) — 服务与卷
- [config/clickhouse-{1,2,3}/](config/) — 每节点 `remote_servers`、keeper、macros、listen
