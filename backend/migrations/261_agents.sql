-- 代理分润：在邀请返利账本上加"代理"身份。
-- 代理按名下用户的消费（真充值的部分）分润，不再拿这些用户的充值返利。
ALTER TABLE user_affiliates ADD COLUMN IF NOT EXISTS is_agent BOOLEAN NOT NULL DEFAULT false;
ALTER TABLE user_affiliates ADD COLUMN IF NOT EXISTS agent_share_percent DECIMAL(5,2);
ALTER TABLE user_affiliates ADD COLUMN IF NOT EXISTS agent_since TIMESTAMPTZ;

-- action = 'agent_share' 的流水记下它是按多少消费算出来的。
ALTER TABLE user_affiliate_ledger ADD COLUMN IF NOT EXISTS base_amount DECIMAL(20,8);

CREATE INDEX IF NOT EXISTS idx_user_affiliate_ledger_agent_share
    ON user_affiliate_ledger (user_id, created_at)
    WHERE action = 'agent_share';

-- 代理申请提现，管理员线下打款后标记已付。
CREATE TABLE IF NOT EXISTS agent_payout_requests (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    amount DECIMAL(20,8) NOT NULL,
    payee_note TEXT NOT NULL DEFAULT '',
    status VARCHAR(16) NOT NULL DEFAULT 'pending',
    admin_note TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    handled_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_agent_payout_requests_status ON agent_payout_requests (status, created_at);

-- 一个代理同一时间只有一笔待处理的申请。
CREATE UNIQUE INDEX IF NOT EXISTS idx_agent_payout_requests_pending
    ON agent_payout_requests (user_id)
    WHERE status = 'pending';
