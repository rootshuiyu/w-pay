-- 清空业务测试数据（保留 admin 账号、IP 白名单配置）
-- 用法: psql -U wpay -d wpay -f sql/clear_business_data.sql

BEGIN;

TRUNCATE TABLE
    orders,
    pay_channel_history,
    pay_channel,
    pay_platform,
    sensitive_logs,
    store
RESTART IDENTITY CASCADE;

COMMIT;
