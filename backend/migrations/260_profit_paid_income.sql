-- 利润：把"赠送的余额"和"真充值的余额"分开。
-- paid_cost 是 actual_cost 里按用户"真充值占比"折算出来的部分；NULL 表示这个小时是在有这一列之前汇总的，
-- 报表里按全部算收入。
ALTER TABLE profit_hourly ADD COLUMN IF NOT EXISTS paid_cost DECIMAL(20,10);
