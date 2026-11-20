-- 种菜线路：仅限野菜 app 的分组，以及非 app 请求改用的普通价分组。
ALTER TABLE groups ADD COLUMN IF NOT EXISTS app_only BOOLEAN NOT NULL DEFAULT FALSE;
ALTER TABLE groups ADD COLUMN IF NOT EXISTS app_fallback_group_id BIGINT;
