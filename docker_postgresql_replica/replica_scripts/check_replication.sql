-- ============================================
-- 检查复制状态
-- 在主库执行
-- ============================================

-- 查看复制连接状态
SELECT 
    client_addr,
    usename,
    application_name,
    state,
    sync_state,
    sent_lsn,
    write_lsn,
    flush_lsn,
    replay_lsn,
    pg_wal_lsn_diff(sent_lsn, replay_lsn) AS replication_lag_bytes
FROM pg_stat_replication;

-- 查看复制槽状态
SELECT 
    slot_name,
    slot_type,
    active,
    restart_lsn,
    confirmed_flush_lsn
FROM pg_replication_slots;

-- 查看同步从库名称
SHOW synchronous_standby_names;

-- 查看同步提交模式
SHOW synchronous_commit;
