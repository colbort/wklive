-- Staking 业务代码统一使用毫秒时间戳。MySQL INT 无法保存当前毫秒值，
-- 因此所有业务时间字段必须使用 BIGINT。

ALTER TABLE t_stake_order
  MODIFY COLUMN `start_times` BIGINT NOT NULL DEFAULT 0 COMMENT '起息时间戳（毫秒）',
  MODIFY COLUMN `end_times` BIGINT NOT NULL DEFAULT 0 COMMENT '到期时间戳（毫秒），活期可为0',
  MODIFY COLUMN `last_reward_times` BIGINT NOT NULL DEFAULT 0 COMMENT '最后一次收益发放时间戳（毫秒）',
  MODIFY COLUMN `next_reward_times` BIGINT NOT NULL DEFAULT 0 COMMENT '下一次收益发放时间戳（毫秒）',
  MODIFY COLUMN `redeem_apply_times` BIGINT NOT NULL DEFAULT 0 COMMENT '申请赎回时间戳（毫秒）',
  MODIFY COLUMN `redeem_times` BIGINT NOT NULL DEFAULT 0 COMMENT '实际赎回时间戳（毫秒）';

ALTER TABLE t_stake_reward_log
  MODIFY COLUMN `reward_times` BIGINT NOT NULL DEFAULT 0 COMMENT '收益发放时间戳（毫秒）';

ALTER TABLE t_stake_redeem_log
  MODIFY COLUMN `redeem_times` BIGINT NOT NULL DEFAULT 0 COMMENT '赎回时间戳（毫秒）';
