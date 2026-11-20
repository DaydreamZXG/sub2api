-- 利润：每个上游账号每小时的用量小结。usage_logs 只保留一段时间，利润要能一直往回看。
CREATE TABLE IF NOT EXISTS profit_hourly (
    bucket_start TIMESTAMPTZ NOT NULL,
    account_id BIGINT NOT NULL,
    -- 没有分组的调用记 0。
    group_id BIGINT NOT NULL DEFAULT 0,
    -- 0 = 扣余额，1 = 扣包月套餐的额度。
    billing_type SMALLINT NOT NULL DEFAULT 0,
    requests BIGINT NOT NULL DEFAULT 0,
    standard_cost DECIMAL(20,10) NOT NULL DEFAULT 0,
    actual_cost DECIMAL(20,10) NOT NULL DEFAULT 0,
    account_cost DECIMAL(20,10) NOT NULL DEFAULT 0,
    PRIMARY KEY (bucket_start, account_id, group_id, billing_type)
);

CREATE INDEX IF NOT EXISTS idx_profit_hourly_account ON profit_hourly (account_id, bucket_start);

-- 成本包：一笔实际付出去的钱，挂在一个或多个上游账号上。
-- cycle_days = 0 是一次性的（充值包）；> 0 是每隔这么多天付一次 cycle_amount（订阅）。
CREATE TABLE IF NOT EXISTS cost_packs (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    note TEXT NOT NULL DEFAULT '',
    cycle_days INT NOT NULL DEFAULT 0,
    cycle_amount DECIMAL(20,8) NOT NULL DEFAULT 0,
    started_at DATE NOT NULL,
    ended_at DATE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 一次性付的钱：买包那一笔，以及之后往同一批账号上追加的。
CREATE TABLE IF NOT EXISTS cost_pack_payments (
    id BIGSERIAL PRIMARY KEY,
    pack_id BIGINT NOT NULL REFERENCES cost_packs(id) ON DELETE CASCADE,
    amount DECIMAL(20,8) NOT NULL,
    paid_at DATE NOT NULL,
    note TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_cost_pack_payments_pack ON cost_pack_payments (pack_id, paid_at);

-- 一个账号同一时间只属于一个成本包，否则它的收入会被算两遍。
CREATE TABLE IF NOT EXISTS cost_pack_accounts (
    pack_id BIGINT NOT NULL REFERENCES cost_packs(id) ON DELETE CASCADE,
    account_id BIGINT NOT NULL,
    PRIMARY KEY (pack_id, account_id)
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_cost_pack_accounts_account ON cost_pack_accounts (account_id);
