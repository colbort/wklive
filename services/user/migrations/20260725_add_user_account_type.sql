ALTER TABLE `t_user`
  ADD COLUMN `account_type` TINYINT NOT NULL DEFAULT 1
  COMMENT '账户类型：1普通用户 2内部做市账户 3平台系统账户'
  AFTER `register_type`,
  ADD KEY `idx_tenant_account_type` (`tenant_id`, `account_type`, `status`),
  ADD CONSTRAINT `chk_user_account_type` CHECK (`account_type` IN (1, 2, 3));
