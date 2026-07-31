-- t_option_account 是 Asset OPTION 钱包的只读镜像。历史按业务 account_id
-- 写入会把同一钱包余额复制多份；合并为 (tenant,user,coin,account_id=0)。
INSERT INTO t_option_account (
  tenant_id,user_id,account_id,margin_coin,balance,available_balance,frozen_balance,
  position_margin,order_margin,unrealized_pnl,realized_pnl,risk_rate,status,
  create_times,update_times
)
SELECT
  source.tenant_id,source.user_id,0,source.margin_coin,
  source.balance,source.available_balance,source.frozen_balance,
  source.position_margin,source.order_margin,source.unrealized_pnl,source.realized_pnl,
  source.risk_rate,source.status,source.create_times,source.update_times
FROM (
  SELECT ranked.*
  FROM (
    SELECT account_rows.*,
      ROW_NUMBER() OVER (
        PARTITION BY tenant_id,user_id,margin_coin
        ORDER BY update_times DESC,id DESC
      ) AS row_rank
    FROM t_option_account account_rows
  ) ranked
  WHERE ranked.row_rank = 1
) source
ON DUPLICATE KEY UPDATE
  balance=VALUES(balance),
  available_balance=VALUES(available_balance),
  frozen_balance=VALUES(frozen_balance),
  position_margin=VALUES(position_margin),
  order_margin=VALUES(order_margin),
  unrealized_pnl=VALUES(unrealized_pnl),
  realized_pnl=VALUES(realized_pnl),
  risk_rate=VALUES(risk_rate),
  status=VALUES(status),
  update_times=VALUES(update_times);

DELETE FROM t_option_account WHERE account_id <> 0;
