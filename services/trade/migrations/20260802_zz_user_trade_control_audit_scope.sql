ALTER TABLE `t_trade_user_control_audit`
  ADD COLUMN `scope_type` TINYINT NOT NULL DEFAULT 1 COMMENT '控制级别：1产品级 2交易对级' AFTER `control_id`,
  DROP INDEX `idx_control_id`,
  ADD KEY `idx_control_id` (`scope_type`, `control_id`, `id`),
  ADD CONSTRAINT `chk_user_control_audit_scope` CHECK (`scope_type` IN (1, 2));

-- 早期版本没有 scope_type，根据 JSON 快照中是否存在 symbol_id 完成回填。
UPDATE `t_trade_user_control_audit`
SET `scope_type` = 2
WHERE JSON_EXTRACT(COALESCE(`after_json`, `before_json`), '$.SymbolId') IS NOT NULL
   OR JSON_EXTRACT(COALESCE(`after_json`, `before_json`), '$.symbol_id') IS NOT NULL;
