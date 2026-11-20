-- 菜园：充值赠送的额度先种成一棵菜，成熟后由用户收进余额。
CREATE TABLE IF NOT EXISTS garden_crops (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    -- 充值赠送种出来的菜记订单号；其他来源（活动赠送等）为空。
    order_id BIGINT,
    source VARCHAR(16) NOT NULL DEFAULT 'order',
    amount DECIMAL(20,8) NOT NULL,
    stolen_amount DECIMAL(20,8) NOT NULL DEFAULT 0,
    status VARCHAR(16) NOT NULL DEFAULT 'growing',
    planted_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    ripe_at TIMESTAMPTZ NOT NULL,
    harvested_at TIMESTAMPTZ,
    harvested_amount DECIMAL(20,8) NOT NULL DEFAULT 0
);

CREATE INDEX IF NOT EXISTS idx_garden_crops_user_status ON garden_crops (user_id, status, ripe_at);

-- 一张充值订单最多种一棵菜；重复履约靠这个唯一约束保持幂等。
CREATE UNIQUE INDEX IF NOT EXISTS idx_garden_crops_order ON garden_crops (order_id) WHERE order_id IS NOT NULL;
