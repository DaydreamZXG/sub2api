-- Site-wide official model prices for the price list (per million tokens). Display only: billing never reads it.
-- Seeded from the vendors' price pages (DeepSeek off-peak list, Zhipu bigmodel.cn / z.ai) as of 2026-11-24.
INSERT INTO settings (key, value)
VALUES ('official_price_book', '[
  {"models":["deepseek-flash","DeepSeek-V4.1-Flash-1017","deepseek-v4.1-flash","deepseek-v4-flash"],"cny_input":1,"cny_output":4,"cny_cache_read":0.02,"usd_input":0.15,"usd_output":0.6,"usd_cache_read":0.003},
  {"models":["glm-5.3"],"cny_input":8,"cny_output":28,"cny_cache_read":2,"usd_input":1.4,"usd_output":4.4,"usd_cache_read":0.26},
  {"models":["glm-5.3-flash"],"cny_input":0.8,"cny_output":2.8,"cny_cache_read":0.23,"usd_input":0.15,"usd_output":0.5,"usd_cache_read":0.03},
  {"models":["glm-5.3-flashx"],"cny_input":2,"cny_output":7,"cny_cache_read":0.57,"usd_input":0.37,"usd_output":1.25,"usd_cache_read":0.075}
]')
ON CONFLICT (key) DO NOTHING;
