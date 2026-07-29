-- 首次 position_mode 回填若最新 History 来自资金费、交割或强平，
-- 该记录可能没有 ref_order_id。改为选择最近一条可关联订单的仓位历史，
-- 确保历史双向仓位不会因非订单 History 排在最后而保留单向默认值。
UPDATE `t_contract_position` AS p
JOIN (
  SELECT h.position_id, MAX(h.id) AS history_id
  FROM `t_contract_position_history` AS h
  JOIN `t_trade_order` AS o
    ON o.tenant_id = h.tenant_id
   AND o.id = h.ref_order_id
  GROUP BY h.position_id
) AS latest_order_history
  ON latest_order_history.position_id = p.id
JOIN `t_contract_position_history` AS h
  ON h.id = latest_order_history.history_id
JOIN `t_trade_order` AS o
  ON o.tenant_id = h.tenant_id
 AND o.id = h.ref_order_id
SET p.position_mode = CASE WHEN o.position_side = 1 THEN 1 ELSE 2 END;
