-- Crops the site gives away: one to the inviter when an invited user pays for the first time, and ones an admin hands out.
-- They are not taken out of anyone's balance; ref_user_id / ref_order_id say which invitee and order earned an invite crop.
ALTER TABLE garden_crops ADD COLUMN IF NOT EXISTS ref_user_id BIGINT;
ALTER TABLE garden_crops ADD COLUMN IF NOT EXISTS ref_order_id BIGINT;

-- An invitee earns their inviter one crop, ever.
CREATE UNIQUE INDEX IF NOT EXISTS idx_garden_crops_invitee
    ON garden_crops (ref_user_id) WHERE source = 'invite';

CREATE INDEX IF NOT EXISTS idx_garden_crops_ref_order
    ON garden_crops (ref_order_id) WHERE ref_order_id IS NOT NULL;
