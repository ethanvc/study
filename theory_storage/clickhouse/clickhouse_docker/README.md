# ClickHouse 3 节点集群（Docker Compose）

- **节点**：`clickhouse-1`、`clickhouse-2`、`clickhouse-3`（1 分片 3 副本）
- **Keeper**：跑在 `clickhouse-1` 上，2/3 仅作客户端连接

## 端口

| 节点         | HTTP | Native |
| ------------ | ---- | ------ |
| clickhouse-1 | 8123 | 9000   |
| clickhouse-2 | 8124 | 9001   |
| clickhouse-3 | 8125 | 9002   |

## 启动

```bash
cd clickhouse_docker
docker compose up -d
```

## 连接方式

| 方式                  | 说明                                                                                                                                                                                                 |
| --------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| **HTTP**              | 浏览器或 `curl "http://localhost:8123/?query=SELECT 1"`                                                                                                                                              |
| **clickhouse-client** | 同机安装 [ClickHouse 客户端](https://clickhouse.com/docs/en/integrations/sql-clients/clickhouse-client-local) 后：`clickhouse-client --host 127.0.0.1 --port 9000`（节点 2 用 9001，节点 3 用 9002） |
| **GUI**               | DBeaver、Tabix、DataGrip 等，连接类型选 ClickHouse，主机 `localhost`，HTTP 端口 8123/8124/8125 或 Native 端口 9000/9001/9002                                                                         |

## 验证集群

在任意客户端执行（HTTP 可在浏览器打开 `http://localhost:8123`，在输入框执行）：

```sql
SELECT * FROM system.clusters WHERE cluster = 'cluster_3n';
```

## 目录

- [docker-compose.yml](docker-compose.yml) — 服务与卷
- [config/clickhouse-{1,2,3}/](config/) — 每节点 `remote_servers`、keeper、macros、listen
