CREATE TABLE IF NOT EXISTS `t_option_exercise_assignment` (
  `id` BIGINT NOT NULL AUTO_INCREMENT,
  `tenant_id` BIGINT NOT NULL DEFAULT 0,
  `exercise_id` BIGINT NOT NULL DEFAULT 0,
  `exercise_no` VARCHAR(64) NOT NULL DEFAULT '',
  `long_position_id` BIGINT NOT NULL DEFAULT 0,
  `short_position_id` BIGINT NOT NULL DEFAULT 0,
  `short_user_id` BIGINT NOT NULL DEFAULT 0,
  `short_account_id` BIGINT NOT NULL DEFAULT 0,
  `quantity` DECIMAL(32,16) NOT NULL DEFAULT 0,
  `payoff` DECIMAL(32,16) NOT NULL DEFAULT 0,
  `status` TINYINT NOT NULL DEFAULT 1,
  `instruction_no` VARCHAR(96) NOT NULL DEFAULT '',
  `create_times` BIGINT NOT NULL DEFAULT 0,
  `update_times` BIGINT NOT NULL DEFAULT 0,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_exercise_short_position` (`tenant_id`, `exercise_id`, `short_position_id`),
  KEY `idx_assignment_exercise` (`tenant_id`, `exercise_id`, `status`, `id`),
  CONSTRAINT `chk_option_exercise_assignment` CHECK (
    `quantity` > 0 AND `payoff` > 0 AND `status` IN (1,2,3,4)
  )
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='期权主动行权空头指派';
