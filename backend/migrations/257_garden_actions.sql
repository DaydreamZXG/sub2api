-- 菜园互动：邻居浇水、偷菜、被旺财抓住的记录，也是“动态”的来源。
CREATE TABLE IF NOT EXISTS garden_actions (
    id BIGSERIAL PRIMARY KEY,
    crop_id BIGINT NOT NULL REFERENCES garden_crops(id) ON DELETE CASCADE,
    owner_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    actor_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    -- water | steal | caught
    kind VARCHAR(16) NOT NULL,
    amount DECIMAL(20,8) NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 每个邻居对同一棵菜只能下手一次（偷到或被抓都算）。
CREATE UNIQUE INDEX IF NOT EXISTS idx_garden_actions_one_try
    ON garden_actions (crop_id, actor_id) WHERE kind IN ('steal', 'caught');
CREATE INDEX IF NOT EXISTS idx_garden_actions_owner ON garden_actions (owner_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_garden_actions_actor ON garden_actions (actor_id, created_at DESC);

-- 每个菜园的旺财：当天喂过才看门；feed_seen_at 用来算未读动态。
CREATE TABLE IF NOT EXISTS garden_dogs (
    user_id BIGINT PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    fed_on DATE,
    feed_seen_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
