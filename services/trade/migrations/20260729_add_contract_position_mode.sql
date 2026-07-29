-- 仓位实际方向（多/空）不能表达订单采用的是单向还是双向模式。
-- 持久化 position_mode，避免把单向净持仓投影出的 LONG/SHORT 误判为双向持仓。
ALTER TABLE `t_contract_position`
  ADD COLUMN `position_mode` TINYINT NOT NULL DEFAULT 1
    COMMENT '持仓模式：1单向 2双向' AFTER `position_side`;

-- 用最近一次仓位变更对应订单恢复已有仓位模式；无历史的旧仓位保守保持单向默认值。
UPDATE `t_contract_position` AS p
JOIN (
  SELECT h.position_id, MAX(h.id) AS history_id
  FROM `t_contract_position_history` AS h
  GROUP BY h.position_id
) AS latest
  ON latest.position_id = p.id
JOIN `t_contract_position_history` AS h
  ON h.id = latest.history_id
JOIN `t_trade_order` AS o
  ON o.tenant_id = h.tenant_id
 AND o.id = h.ref_order_id
SET p.position_mode = CASE WHEN o.position_side = 1 THEN 1 ELSE 2 END;

ALTER TABLE `t_contract_position`
  DROP CHECK `chk_position_dimensions`,
  ADD CONSTRAINT `chk_position_dimensions`
    CHECK (`contract_type` IN (1, 2) AND
           `contract_value_type` IN (1, 2) AND
           `position_side` IN (1, 2, 3) AND
           `position_mode` IN (1, 2) AND
           `margin_mode` IN (1, 2) AND
           `status` BETWEEN 1 AND 6 AND
           `leverage` > 0);
