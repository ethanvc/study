-- ============================================
-- PostgreSQL 主库初始化脚本 (postgres-1)
-- ============================================

-- 创建复制用户
CREATE USER repl_user WITH REPLICATION ENCRYPTED PASSWORD '123456';

-- 创建复制槽（可选，用于确保从库不会丢失 WAL）
SELECT pg_create_physical_replication_slot('replica_slot_2', true);
SELECT pg_create_physical_replication_slot('replica_slot_3', true);

-- 显示创建的用户
SELECT usename, userepl FROM pg_user WHERE usename = 'repl_user';

-- 显示复制槽
SELECT slot_name, slot_type, active FROM pg_replication_slots;
