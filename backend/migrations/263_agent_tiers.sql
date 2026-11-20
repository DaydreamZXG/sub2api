-- An agent's share can rise with what their users spend in a calendar month: a JSON list of
-- {"from": spending in balance units, "percent": share from there on}, lowest first. NULL or empty: one flat share.
ALTER TABLE user_affiliates ADD COLUMN IF NOT EXISTS agent_tiers JSONB;
