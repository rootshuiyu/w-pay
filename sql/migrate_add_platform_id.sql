-- 迁移脚本：为 pay_channel 表添加 platform_id 列
-- 执行方式：sudo -u postgres psql -d wpay -f sql/migrate_add_platform_id.sql

-- 添加 platform_id 列到 pay_channel 表
ALTER TABLE pay_channel ADD COLUMN IF NOT EXISTS platform_id BIGINT DEFAULT 0;

-- 添加索引
CREATE INDEX IF NOT EXISTS idx_channel_platform ON pay_channel (platform_id);

-- 添加注释
COMMENT ON COLUMN pay_channel.platform_id IS '所属代收平台，0=未分配';
