-- Closing fills created before the position projector synchronized
-- realized_pnl kept the schema default (zero), while the immutable position
-- history already contained the authoritative realized PnL.
UPDATE t_trade_fill AS f
JOIN (
    SELECT
        tenant_id,
        ref_fill_id,
        SUM(realized_pnl_delta) AS realized_pnl
    FROM t_contract_position_history
    WHERE ref_fill_id > 0
    GROUP BY tenant_id, ref_fill_id
) AS h
    ON h.tenant_id = f.tenant_id
   AND h.ref_fill_id = f.id
SET f.realized_pnl = h.realized_pnl
WHERE f.product_type = 2
  AND f.realized_pnl = 0
  AND h.realized_pnl <> 0;
