-- 连锁自营门店内部聚合订单调度系统 PostgreSQL 建表脚本
-- 合规：仅供实控人自有门店内部使用，不涉及资金清算
-- 使用前先建库（Docker 初始化时已自动建库，可跳过）：
--   CREATE DATABASE wpay;
-- 然后：psql -U postgres -d wpay -f sql/init.sql

-- 3.1 admin 管理员账号表
CREATE TABLE IF NOT EXISTS admin (
    id BIGSERIAL PRIMARY KEY,
    username VARCHAR(64) NOT NULL,
    password_hash VARCHAR(128) NOT NULL,
    role VARCHAR(32) NOT NULL,
    phone VARCHAR(20) DEFAULT NULL,
    status SMALLINT NOT NULL DEFAULT 1,
    real_name VARCHAR(64) DEFAULT NULL,
    created_at TIMESTAMPTZ(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMPTZ(3) DEFAULT NULL
);
CREATE UNIQUE INDEX IF NOT EXISTS idx_admin_username ON admin (username);
CREATE INDEX IF NOT EXISTS idx_admin_role ON admin (role);
CREATE INDEX IF NOT EXISTS idx_admin_deleted_at ON admin (deleted_at);
COMMENT ON TABLE admin IS '管理员账号';
COMMENT ON COLUMN admin.password_hash IS '密码 bcrypt 加密';
COMMENT ON COLUMN admin.role IS 'super_admin/finance/operator';
COMMENT ON COLUMN admin.status IS '1正常 0禁用';

-- 3.2 store 门店信息表（无数量上限，store_id 雪花ID）
CREATE TABLE IF NOT EXISTS store (
    id BIGINT PRIMARY KEY,
    store_code VARCHAR(64) DEFAULT NULL,
    store_name VARCHAR(128) NOT NULL,
    address VARCHAR(512) DEFAULT NULL,
    tax_subject VARCHAR(256) DEFAULT NULL,
    subject_info VARCHAR(256) DEFAULT NULL,
    status SMALLINT NOT NULL DEFAULT 1,
    remark VARCHAR(512) DEFAULT NULL,
    created_at TIMESTAMPTZ(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMPTZ(3) DEFAULT NULL
);
-- 部分唯一索引：空编号不参与唯一约束
CREATE UNIQUE INDEX IF NOT EXISTS idx_store_store_code ON store (store_code) WHERE store_code <> '';
CREATE INDEX IF NOT EXISTS idx_store_name ON store (store_name);
CREATE INDEX IF NOT EXISTS idx_store_status ON store (status);
CREATE INDEX IF NOT EXISTS idx_store_deleted_at ON store (deleted_at);
COMMENT ON TABLE store IS '门店信息';
COMMENT ON COLUMN store.id IS 'store_id 雪花主键';
COMMENT ON COLUMN store.tax_subject IS '个体户主体全称';
COMMENT ON COLUMN store.status IS '1正常 0停用';

-- 3.3 pay_channel 支付渠道（同门店可多码，支持轮询代收池）
CREATE TABLE IF NOT EXISTS pay_channel (
    id BIGSERIAL PRIMARY KEY,
    store_id BIGINT NOT NULL,
    pay_type SMALLINT NOT NULL,
    status SMALLINT NOT NULL DEFAULT 1,
    pool_enabled SMALLINT NOT NULL DEFAULT 0,
    daily_limit_fen BIGINT NOT NULL DEFAULT 0,
    single_limit_fen BIGINT NOT NULL DEFAULT 0,
    daily_used_fen BIGINT NOT NULL DEFAULT 0,
    daily_reset_date VARCHAR(10) DEFAULT NULL,
    rotate_weight INT NOT NULL DEFAULT 1,
    mch_no VARCHAR(64) DEFAULT NULL,
    mch_key VARCHAR(256) DEFAULT NULL,
    app_id VARCHAR(64) DEFAULT NULL,
    serial_no VARCHAR(64) DEFAULT NULL,
    private_key TEXT,
    public_key TEXT,
    notify_url VARCHAR(512) DEFAULT NULL,
    cert_file VARCHAR(512) DEFAULT NULL,
    remark VARCHAR(256) DEFAULT NULL,
    created_at TIMESTAMPTZ(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMPTZ(3) DEFAULT NULL
);
CREATE INDEX IF NOT EXISTS idx_channel_store_id ON pay_channel (store_id);
CREATE INDEX IF NOT EXISTS idx_channel_status ON pay_channel (status);
CREATE INDEX IF NOT EXISTS idx_channel_pool ON pay_channel (pool_enabled, pay_type, status);
CREATE INDEX IF NOT EXISTS idx_channel_deleted_at ON pay_channel (deleted_at);
COMMENT ON TABLE pay_channel IS '支付渠道（多码轮询代收）';
COMMENT ON COLUMN pay_channel.pay_type IS '1微信 2支付宝';
COMMENT ON COLUMN pay_channel.status IS '1启用 0关停';
COMMENT ON COLUMN pay_channel.pool_enabled IS '1参与轮询代收池';
COMMENT ON COLUMN pay_channel.daily_limit_fen IS '日收款上限(分)，0不限';
COMMENT ON COLUMN pay_channel.single_limit_fen IS '单笔上限(分)，0不限';
COMMENT ON COLUMN pay_channel.daily_used_fen IS '当日已收(分)';
COMMENT ON COLUMN pay_channel.mch_key IS '商户密钥/APIv3Key';
COMMENT ON COLUMN pay_channel.serial_no IS '微信证书序列号';

-- 3.4 orders 订单主表（预留分表：store_id + created_at）
CREATE TABLE IF NOT EXISTS orders (
    id BIGSERIAL PRIMARY KEY,
    order_no VARCHAR(32) NOT NULL,
    store_id BIGINT NOT NULL,
    pay_type SMALLINT NOT NULL,
    total_amount BIGINT NOT NULL,
    pay_amount BIGINT NOT NULL DEFAULT 0,
    order_status SMALLINT NOT NULL DEFAULT 0,
    device_sn VARCHAR(64) DEFAULT NULL,
    subject VARCHAR(256) DEFAULT NULL,
    notify_data TEXT,
    pay_time TIMESTAMPTZ(3) DEFAULT NULL,
    transaction_id VARCHAR(64) DEFAULT NULL,
    qr_code_url VARCHAR(512) DEFAULT NULL,
    channel_id BIGINT DEFAULT NULL,
    created_at TIMESTAMPTZ(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMPTZ(3) DEFAULT NULL
);
CREATE UNIQUE INDEX IF NOT EXISTS idx_orders_order_no ON orders (order_no);
CREATE INDEX IF NOT EXISTS idx_orders_store_created ON orders (store_id, created_at);
CREATE INDEX IF NOT EXISTS idx_orders_pay_type ON orders (pay_type);
CREATE INDEX IF NOT EXISTS idx_orders_status ON orders (order_status);
CREATE INDEX IF NOT EXISTS idx_orders_deleted_at ON orders (deleted_at);
COMMENT ON TABLE orders IS '收银订单';
COMMENT ON COLUMN orders.order_no IS 'order_id 全局雪花订单号';
COMMENT ON COLUMN orders.total_amount IS '订单金额(分)';
COMMENT ON COLUMN orders.pay_amount IS '实付金额(分)';
COMMENT ON COLUMN orders.order_status IS '0待支付 1已支付 2已关闭 3退款';
COMMENT ON COLUMN orders.notify_data IS '脱敏回调报文摘要';
COMMENT ON COLUMN orders.channel_id IS '下单时 pay_channel.id';

-- 渠道历史密钥（更换商户时归档，7天内供旧订单回调验签）
CREATE TABLE IF NOT EXISTS pay_channel_history (
    id BIGSERIAL PRIMARY KEY,
    channel_id BIGINT NOT NULL,
    store_id BIGINT NOT NULL,
    pay_type SMALLINT NOT NULL,
    mch_no VARCHAR(64) DEFAULT NULL,
    mch_key VARCHAR(256) DEFAULT NULL,
    app_id VARCHAR(64) DEFAULT NULL,
    serial_no VARCHAR(64) DEFAULT NULL,
    private_key TEXT,
    public_key TEXT,
    expires_at TIMESTAMPTZ(3) NOT NULL,
    created_at TIMESTAMPTZ(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMPTZ(3) DEFAULT NULL
);
CREATE INDEX IF NOT EXISTS idx_history_channel_expires ON pay_channel_history (channel_id, expires_at);
CREATE INDEX IF NOT EXISTS idx_history_store_id ON pay_channel_history (store_id);
COMMENT ON TABLE pay_channel_history IS '渠道历史密钥';

CREATE TABLE IF NOT EXISTS sensitive_logs (
    id BIGSERIAL PRIMARY KEY,
    action VARCHAR(64) NOT NULL,
    content TEXT,
    admin_id BIGINT DEFAULT NULL,
    ip VARCHAR(64) DEFAULT NULL,
    created_at TIMESTAMPTZ(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMPTZ(3) DEFAULT NULL
);
CREATE INDEX IF NOT EXISTS idx_sensitive_created_at ON sensitive_logs (created_at);
CREATE INDEX IF NOT EXISTS idx_sensitive_deleted_at ON sensitive_logs (deleted_at);
COMMENT ON TABLE sensitive_logs IS '敏感日志(6个月清理)';

-- 3.7 IP 白名单策略与条目（后台可视化管理，热更新生效）
CREATE TABLE IF NOT EXISTS ip_whitelist_policy (
    scope VARCHAR(32) PRIMARY KEY,
    enabled SMALLINT NOT NULL DEFAULT 0,
    remark VARCHAR(256) DEFAULT NULL,
    updated_at TIMESTAMPTZ(3) NOT NULL DEFAULT CURRENT_TIMESTAMP
);
COMMENT ON TABLE ip_whitelist_policy IS 'IP白名单开关：admin/pay/callback/trusted_proxy';
COMMENT ON COLUMN ip_whitelist_policy.enabled IS '1启用校验 0不限制';

CREATE TABLE IF NOT EXISTS ip_whitelist_entry (
    id BIGSERIAL PRIMARY KEY,
    scope VARCHAR(32) NOT NULL,
    cidr VARCHAR(64) NOT NULL,
    remark VARCHAR(256) DEFAULT NULL,
    status SMALLINT NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMPTZ(3) DEFAULT NULL
);
CREATE UNIQUE INDEX IF NOT EXISTS idx_scope_cidr ON ip_whitelist_entry (scope, cidr) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_whitelist_scope ON ip_whitelist_entry (scope);
CREATE INDEX IF NOT EXISTS idx_whitelist_status ON ip_whitelist_entry (status);
CREATE INDEX IF NOT EXISTS idx_whitelist_deleted_at ON ip_whitelist_entry (deleted_at);
COMMENT ON TABLE ip_whitelist_entry IS 'IP/CIDR 白名单条目';
COMMENT ON COLUMN ip_whitelist_entry.scope IS 'admin/pay/callback/trusted_proxy';
COMMENT ON COLUMN ip_whitelist_entry.status IS '1启用 0停用';

INSERT INTO ip_whitelist_policy (scope, enabled) VALUES
    ('admin', 0),
    ('callback', 0),
    ('pay', 0),
    ('trusted_proxy', 0)
ON CONFLICT (scope) DO NOTHING;
