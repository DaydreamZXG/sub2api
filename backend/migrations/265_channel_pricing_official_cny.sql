-- Official prices some vendors list in yuan. Only the price list shows them and compares with them;
-- billing never reads them. NULL means restate the official USD price at the daily rate.
ALTER TABLE channel_model_pricing ADD COLUMN IF NOT EXISTS official_cny_input_price NUMERIC(20,12);
ALTER TABLE channel_model_pricing ADD COLUMN IF NOT EXISTS official_cny_output_price NUMERIC(20,12);
ALTER TABLE channel_model_pricing ADD COLUMN IF NOT EXISTS official_cny_cache_read_price NUMERIC(20,12);
