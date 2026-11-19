-- 桌面助手的设备凭证：app 通过网站“连接到 app”登录后签发，网关用它识别来自 app 的请求。
CREATE TABLE IF NOT EXISTS app_device_tokens (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    device_id VARCHAR(128) NOT NULL,
    platform VARCHAR(16) NOT NULL DEFAULT 'unknown',
    client_version VARCHAR(32) NOT NULL DEFAULT '',
    token_hash VARCHAR(64) NOT NULL UNIQUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    revoked_at TIMESTAMPTZ
);

-- 一台设备同时只有一个有效凭证。
CREATE UNIQUE INDEX IF NOT EXISTS idx_app_device_tokens_active_device
    ON app_device_tokens (user_id, device_id) WHERE revoked_at IS NULL;
