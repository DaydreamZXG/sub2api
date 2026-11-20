-- When an admin put the user under an agent. The agent earns on this user's spending from then on;
-- NULL for users who came in through the agent's own invite link.
ALTER TABLE user_affiliates ADD COLUMN IF NOT EXISTS agent_bound_at TIMESTAMPTZ;
