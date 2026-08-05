-- Pre-production business baseline.
-- This file is bundled into the immutable db-init release image. It contains
-- configuration only: no users, balances, positions, orders, fake prices or
-- approval evidence. Risk-bearing products remain disabled or pending review.

START TRANSACTION;

SET @preprod_now_ms = CAST(UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000 AS UNSIGNED);
SET @preprod_now_s = UNIX_TIMESTAMP();

INSERT INTO sys_tenant
(id,tenant_code,tenant_name,enabled,expire_time,contact_name,contact_phone,
 login_ip,login_time,login_count,remark,create_by,create_times,update_by,update_times)
VALUES
(@preprod_tenant_id,@preprod_tenant_code,@preprod_tenant_name,1,0,NULL,NULL,
 NULL,0,0,'PREPROD_BASELINE','db-init',@preprod_now_ms,'db-init',@preprod_now_ms)
ON DUPLICATE KEY UPDATE tenant_name=VALUES(tenant_name),update_times=VALUES(update_times);

-- Only the crypto category is exposed to the pre-production tenant.
INSERT INTO t_itick_category
(category_type,category_name,category_code,enabled,app_visible,sync_priority,sort,
 icon,remark,create_times,update_times)
VALUES
(2,'加密货币','crypto',1,1,1,1,'','PREPROD_BASELINE',@preprod_now_ms,@preprod_now_ms)
ON DUPLICATE KEY UPDATE category_name=VALUES(category_name),category_code=VALUES(category_code);

SET @preprod_crypto_category_id = (
  SELECT id FROM t_itick_category WHERE category_type=2 LIMIT 1
);

CREATE TEMPORARY TABLE preprod_pairs (
  symbol VARCHAR(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL PRIMARY KEY,
  base_asset VARCHAR(32) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL,
  display_name VARCHAR(128) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL,
  sort_no INT NOT NULL
);
INSERT INTO preprod_pairs(symbol,base_asset,display_name,sort_no) VALUES
('BTCUSDT','BTC','BTC/USDT',1),
('ETHUSDT','ETH','ETH/USDT',2);

INSERT INTO t_itick_product
(category_type,category_name,category_code,market,symbol,code,name,display_name,
 exchange,sector,lug,base_coin,quote_coin,enabled,app_visible,sync_priority,sort,
 icon,remark,create_times,update_times)
SELECT 2,'加密货币','crypto','BA',pair.symbol,pair.symbol,pair.display_name,
       pair.display_name,'BINANCE','crypto',LOWER(pair.symbol),pair.base_asset,'USDT',
       1,1,1,pair.sort_no,'','PREPROD_BASELINE',@preprod_now_ms,@preprod_now_ms
FROM preprod_pairs pair
ON DUPLICATE KEY UPDATE
  code=VALUES(code),name=VALUES(name),display_name=VALUES(display_name),
  base_coin=VALUES(base_coin),quote_coin=VALUES(quote_coin),update_times=VALUES(update_times);

INSERT INTO t_itick_tenant_category
(tenant_id,category_id,enabled,app_visible,sort,remark,create_times,update_times)
VALUES
(@preprod_tenant_id,@preprod_crypto_category_id,1,1,1,'PREPROD_BASELINE',
 @preprod_now_ms,@preprod_now_ms)
ON DUPLICATE KEY UPDATE id=LAST_INSERT_ID(t_itick_tenant_category.id);

INSERT INTO t_itick_tenant_product
(tenant_id,product_id,enabled,app_visible,display_name,sort,remark,create_times,update_times)
SELECT @preprod_tenant_id,product.id,1,1,pair.display_name,pair.sort_no,
       'PREPROD_BASELINE',@preprod_now_ms,@preprod_now_ms
FROM preprod_pairs pair
JOIN t_itick_product product
  ON product.category_type=2 AND product.market='BA' AND product.symbol=pair.symbol
ON DUPLICATE KEY UPDATE id=LAST_INSERT_ID(t_itick_tenant_product.id);

-- Crypto trades continuously. The calendar is operational metadata, not a
-- price or approval assertion.
INSERT INTO t_itick_market_calendar
(category_code,market,exchange,timezone,trading_day_offset,week_start,enabled,
 remark,create_times,update_times)
VALUES
('crypto','BA','BINANCE','UTC',0,1,1,'PREPROD_BASELINE',@preprod_now_ms,@preprod_now_ms)
ON DUPLICATE KEY UPDATE id=LAST_INSERT_ID(t_itick_market_calendar.id);
SET @preprod_market_calendar_id = (
  SELECT id FROM t_itick_market_calendar
  WHERE category_code='crypto' AND market='BA' AND exchange='BINANCE' LIMIT 1
);
INSERT INTO t_itick_market_session
(calendar_id,session_type,start_time,end_time,cross_day,sort)
SELECT @preprod_market_calendar_id,'regular','00:00:00','24:00:00',0,1
WHERE NOT EXISTS (
  SELECT 1 FROM t_itick_market_session
  WHERE calendar_id=@preprod_market_calendar_id AND session_type='regular'
    AND start_time='00:00:00' AND end_time='24:00:00'
);

-- Price-engine definitions are installed inactive. Enabling them still
-- requires source/licensing review and fresh multi-source market evidence.
INSERT INTO t_itick_price_formula
(formula_no,authority,snapshot_kind,category_code,market,symbol,algorithm,
 formula_version,components,max_lookback_ms,max_deviation_bps,min_input_count,
 interval_ms,last_target_time,status,version,run_version,create_times,update_times)
SELECT CONCAT(pair.symbol,'-INDEX-v1'),'price-engine','INDEX','crypto','BA',pair.symbol,2,
       'v1',JSON_ARRAY(
         JSON_OBJECT('authority','binance-public','kind','FINAL_QUOTE','category_code','crypto','market','BINANCE','symbol',pair.symbol,'weight','1'),
         JSON_OBJECT('authority','okx-public','kind','FINAL_QUOTE','category_code','crypto','market','OKX','symbol',pair.symbol,'weight','1'),
         JSON_OBJECT('authority','bybit-public','kind','FINAL_QUOTE','category_code','crypto','market','BYBIT','symbol',pair.symbol,'weight','1')
       ),30000,100,3,1000,0,2,0,0,@preprod_now_ms,@preprod_now_ms
FROM preprod_pairs pair
ON DUPLICATE KEY UPDATE id=LAST_INSERT_ID(t_itick_price_formula.id);

INSERT INTO t_itick_price_formula
(formula_no,authority,snapshot_kind,category_code,market,symbol,algorithm,
 formula_version,components,max_lookback_ms,max_deviation_bps,min_input_count,
 interval_ms,last_target_time,status,version,run_version,create_times,update_times)
SELECT CONCAT(pair.symbol,'-MARK-v1'),'price-engine','MARK','crypto','BA',pair.symbol,4,
       'v1',JSON_ARRAY(
         JSON_OBJECT('authority','price-engine','kind','INDEX','category_code','crypto','market','BA','symbol',pair.symbol,'weight','1'),
         JSON_OBJECT('authority','binance-futures-public','kind','FINAL_QUOTE','category_code','crypto','market','BINANCE_PERP','symbol',pair.symbol,'weight','1')
       ),30000,200,2,1000,0,2,0,0,@preprod_now_ms,@preprod_now_ms
FROM preprod_pairs pair
ON DUPLICATE KEY UPDATE id=LAST_INSERT_ID(t_itick_price_formula.id);

INSERT INTO t_itick_price_formula
(formula_no,authority,snapshot_kind,category_code,market,symbol,algorithm,
 formula_version,components,max_lookback_ms,max_deviation_bps,min_input_count,
 interval_ms,last_target_time,status,version,run_version,create_times,update_times)
SELECT CONCAT(pair.symbol,'-FUNDING-v1'),'price-engine','FUNDING','crypto','BA',pair.symbol,3,
       'v1',JSON_ARRAY(
         JSON_OBJECT('authority','price-engine','kind','MARK','category_code','crypto','market','BA','symbol',pair.symbol,'weight','1'),
         JSON_OBJECT('authority','price-engine','kind','INDEX','category_code','crypto','market','BA','symbol',pair.symbol,'weight','1')
       ),30000,0,2,60000,0,2,0,0,@preprod_now_ms,@preprod_now_ms
FROM preprod_pairs pair
ON DUPLICATE KEY UPDATE id=LAST_INSERT_ID(t_itick_price_formula.id);

INSERT INTO t_itick_price_formula
(formula_no,authority,snapshot_kind,category_code,market,symbol,algorithm,
 formula_version,components,max_lookback_ms,max_deviation_bps,min_input_count,
 interval_ms,last_target_time,status,version,run_version,create_times,update_times)
SELECT CONCAT(pair.symbol,'-DELIVERY-v1'),'price-engine','DELIVERY','crypto','BA',pair.symbol,2,
       'v1',JSON_ARRAY(
         JSON_OBJECT('authority','binance-public','kind','FINAL_QUOTE','category_code','crypto','market','BINANCE','symbol',pair.symbol,'weight','1'),
         JSON_OBJECT('authority','okx-public','kind','FINAL_QUOTE','category_code','crypto','market','OKX','symbol',pair.symbol,'weight','1'),
         JSON_OBJECT('authority','bybit-public','kind','FINAL_QUOTE','category_code','crypto','market','BYBIT','symbol',pair.symbol,'weight','1')
       ),60000,100,3,1000,0,2,0,0,@preprod_now_ms,@preprod_now_ms
FROM preprod_pairs pair
ON DUPLICATE KEY UPDATE id=LAST_INSERT_ID(t_itick_price_formula.id);

-- Trade: spot is available for ordinary pre-production testing. Seconds,
-- perpetual and delivery instruments are present but disabled until their
-- respective release gates are approved.
CREATE TEMPORARY TABLE preprod_trade_instruments (
  symbol VARCHAR(64) NOT NULL,
  base_asset VARCHAR(32) NOT NULL,
  product_type TINYINT NOT NULL,
  contract_type TINYINT NOT NULL,
  contract_value_type TINYINT NOT NULL,
  status TINYINT NOT NULL,
  display_symbol VARCHAR(64) NOT NULL,
  sort_no INT NOT NULL,
  PRIMARY KEY(symbol,product_type,contract_type,contract_value_type)
);
INSERT INTO preprod_trade_instruments VALUES
('BTCUSDT','BTC',1,0,0,1,'BTC/USDT Spot',10),
('BTCUSDT','BTC',2,1,1,2,'BTC/USDT Perpetual',11),
('BTCUSDT','BTC',2,2,1,2,'BTC/USDT Delivery',12),
('BTCUSDT','BTC',3,0,0,2,'BTC/USDT Seconds',13),
('ETHUSDT','ETH',1,0,0,1,'ETH/USDT Spot',20),
('ETHUSDT','ETH',2,1,1,2,'ETH/USDT Perpetual',21),
('ETHUSDT','ETH',2,2,1,2,'ETH/USDT Delivery',22),
('ETHUSDT','ETH',3,0,0,2,'ETH/USDT Seconds',23);

INSERT INTO t_trade_symbol
(tenant_id,symbol,display_symbol,product_type,base_asset,quote_asset,settle_asset,
 margin_asset,contract_type,contract_value_type,status,price_scale,qty_scale,
 min_price,max_price,price_tick,min_qty,max_qty,qty_step,min_notional,max_notional,
 listing_time,trading_start_time,trading_end_time,sort,remark,create_times,update_times)
SELECT @preprod_tenant_id,instrument.symbol,instrument.display_symbol,
       instrument.product_type,instrument.base_asset,'USDT','USDT',
       IF(instrument.product_type=2,'USDT',''),instrument.contract_type,
       instrument.contract_value_type,instrument.status,2,6,
       0.01,0,0.01,0.000001,1000,0.000001,5,0,0,0,0,instrument.sort_no,
       'PREPROD_BASELINE',@preprod_now_ms,@preprod_now_ms
FROM preprod_trade_instruments instrument
ON DUPLICATE KEY UPDATE id=LAST_INSERT_ID(t_trade_symbol.id);

INSERT INTO t_trade_symbol_spot
(tenant_id,symbol_id,maker_fee_rate,taker_fee_rate,buy_enabled,sell_enabled,create_times,update_times)
SELECT @preprod_tenant_id,symbol_row.id,0.001,0.001,1,1,@preprod_now_ms,@preprod_now_ms
FROM preprod_trade_instruments instrument
JOIN t_trade_symbol symbol_row
  ON symbol_row.tenant_id=@preprod_tenant_id
 AND symbol_row.symbol=instrument.symbol
 AND symbol_row.product_type=instrument.product_type
 AND symbol_row.contract_type=instrument.contract_type
 AND symbol_row.contract_value_type=instrument.contract_value_type
WHERE instrument.product_type=1
ON DUPLICATE KEY UPDATE id=LAST_INSERT_ID(t_trade_symbol_spot.id);

INSERT INTO t_trade_symbol_contract
(tenant_id,symbol_id,contract_size,multiplier,maintenance_margin_rate,
 initial_margin_rate,maker_fee_rate,taker_fee_rate,funding_interval_minutes,
 funding_rate_cap,funding_rate_floor,funding_rate_source,index_symbol,
 mark_price_source,settlement_price_source,delivery_time,open_cutoff_time,
 matching_stop_time,settlement_window_seconds,settlement_price_algorithm,
 delivery_fee_rate,liquidation_fee_rate,support_cross,support_isolated,
 open_long_enabled,open_short_enabled,close_long_enabled,close_short_enabled,
 create_times,update_times)
SELECT @preprod_tenant_id,symbol_row.id,1,1,0.05,0.10,0.0002,0.0005,
       IF(instrument.contract_type=1,480,0),0.003,-0.003,
       IF(instrument.contract_type=1,CONCAT(instrument.symbol,'-FUNDING-v1'),''),
       CONCAT(instrument.symbol,'-INDEX-v1'),CONCAT(instrument.symbol,'-MARK-v1'),
       CONCAT(instrument.symbol,'-DELIVERY-v1'),
       IF(instrument.contract_type=2,UNIX_TIMESTAMP('2030-12-31 08:00:00')*1000,0),
       IF(instrument.contract_type=2,UNIX_TIMESTAMP('2030-12-31 07:00:00')*1000,0),
       IF(instrument.contract_type=2,UNIX_TIMESTAMP('2030-12-31 07:55:00')*1000,0),
       IF(instrument.contract_type=2,60,0),
       IF(instrument.contract_type=2,'median-v1',''),0.001,0.005,
       0,1,2,2,1,1,@preprod_now_ms,@preprod_now_ms
FROM preprod_trade_instruments instrument
JOIN t_trade_symbol symbol_row
  ON symbol_row.tenant_id=@preprod_tenant_id
 AND symbol_row.symbol=instrument.symbol
 AND symbol_row.product_type=instrument.product_type
 AND symbol_row.contract_type=instrument.contract_type
 AND symbol_row.contract_value_type=instrument.contract_value_type
WHERE instrument.product_type=2
ON DUPLICATE KEY UPDATE id=LAST_INSERT_ID(t_trade_symbol_contract.id);

INSERT INTO t_trade_symbol_seconds
(tenant_id,symbol_id,duration_seconds,payout_rate,fee_rate,draw_rule,
 start_price_source,settlement_price_source,quote_validity_ms,
 settlement_window_ms,settlement_price_algorithm,draw_tolerance,
 max_exposure_amount,min_stake,max_stake,up_enabled,down_enabled,
 create_times,update_times)
SELECT @preprod_tenant_id,symbol_row.id,60,0.80,0,1,
       CONCAT(instrument.symbol,'-INDEX-v1'),CONCAT(instrument.symbol,'-DELIVERY-v1'),
       3000,1000,'median-v1',0,0,5,1000,2,2,@preprod_now_ms,@preprod_now_ms
FROM preprod_trade_instruments instrument
JOIN t_trade_symbol symbol_row
  ON symbol_row.tenant_id=@preprod_tenant_id
 AND symbol_row.symbol=instrument.symbol
 AND symbol_row.product_type=instrument.product_type
 AND symbol_row.contract_type=instrument.contract_type
 AND symbol_row.contract_value_type=instrument.contract_value_type
WHERE instrument.product_type=3
ON DUPLICATE KEY UPDATE id=LAST_INSERT_ID(t_trade_symbol_seconds.id);

INSERT INTO t_trade_symbol_session
(tenant_id,symbol_id,day_of_week,start_second,end_second,timezone,enabled,create_times,update_times)
SELECT @preprod_tenant_id,symbol_row.id,weekday.day_no,0,86400,'UTC',1,
       @preprod_now_ms,@preprod_now_ms
FROM t_trade_symbol symbol_row
JOIN (SELECT 1 day_no UNION ALL SELECT 2 UNION ALL SELECT 3 UNION ALL SELECT 4
      UNION ALL SELECT 5 UNION ALL SELECT 6 UNION ALL SELECT 7) weekday
WHERE symbol_row.tenant_id=@preprod_tenant_id
  AND symbol_row.symbol IN ('BTCUSDT','ETHUSDT')
  AND NOT EXISTS (
    SELECT 1 FROM t_trade_symbol_session session
    WHERE session.tenant_id=@preprod_tenant_id AND session.symbol_id=symbol_row.id
      AND session.day_of_week=weekday.day_no AND session.start_second=0
      AND session.end_second=86400
  );

INSERT INTO t_trade_symbol_leverage_config
(tenant_id,symbol_id,margin_mode,leverage_values,enabled,sort,remark,create_times,update_times)
SELECT @preprod_tenant_id,symbol_row.id,2,JSON_ARRAY(1,2,3,5,10),1,1,
       'PREPROD_BASELINE',@preprod_now_ms,@preprod_now_ms
FROM t_trade_symbol symbol_row
WHERE symbol_row.tenant_id=@preprod_tenant_id
  AND symbol_row.symbol IN ('BTCUSDT','ETHUSDT') AND symbol_row.product_type=2
ON DUPLICATE KEY UPDATE id=LAST_INSERT_ID(t_trade_symbol_leverage_config.id);

INSERT INTO t_trade_symbol_leverage_default
(tenant_id,symbol_id,margin_mode,leverage,create_times,update_times)
SELECT @preprod_tenant_id,symbol_row.id,2,3,@preprod_now_ms,@preprod_now_ms
FROM t_trade_symbol symbol_row
WHERE symbol_row.tenant_id=@preprod_tenant_id
  AND symbol_row.symbol IN ('BTCUSDT','ETHUSDT') AND symbol_row.product_type=2
ON DUPLICATE KEY UPDATE id=LAST_INSERT_ID(t_trade_symbol_leverage_default.id);

INSERT INTO t_contract_risk_limit_tier
(tenant_id,symbol_id,tier_no,notional_floor,notional_cap,max_leverage,
 initial_margin_rate,maintenance_margin_rate,maintenance_amount,enabled,
 create_times,update_times)
SELECT @preprod_tenant_id,symbol_row.id,1,0,1000000,10,0.10,0.05,0,1,
       @preprod_now_ms,@preprod_now_ms
FROM t_trade_symbol symbol_row
WHERE symbol_row.tenant_id=@preprod_tenant_id
  AND symbol_row.symbol IN ('BTCUSDT','ETHUSDT') AND symbol_row.product_type=2
ON DUPLICATE KEY UPDATE id=LAST_INSERT_ID(t_contract_risk_limit_tier.id);

INSERT INTO t_contract_insurance_fund_account
(tenant_id,symbol_id,settle_asset,adl_enabled,status,version,create_times,update_times)
VALUES
(@preprod_tenant_id,0,'USDT',2,2,0,@preprod_now_ms,@preprod_now_ms)
ON DUPLICATE KEY UPDATE id=LAST_INSERT_ID(t_contract_insurance_fund_account.id);

-- Option: install an auditable draft calendar and four pending cash-settled
-- contracts. Seller trading, auto exercise and insurance consumption remain
-- disabled; no approval evidence or market price is fabricated.
INSERT INTO t_option_trading_calendar
(tenant_id,calendar_code,version,status,timezone,effective_from,effective_until,
 supersedes_id,change_reason,evidence_ref,created_by,reviewed_by,review_reason,
 reviewed_at,create_times,update_times)
SELECT @preprod_tenant_id,'CONTINUOUS_24_7',1,1,'UTC',0,0,0,
       'PREPROD_BASELINE_DRAFT','',1,0,'',0,@preprod_now_s,@preprod_now_s
WHERE NOT EXISTS (
  SELECT 1 FROM t_option_trading_calendar
  WHERE tenant_id=@preprod_tenant_id AND calendar_code='CONTINUOUS_24_7' AND version=1
);
SET @preprod_option_calendar_id = (
  SELECT id FROM t_option_trading_calendar
  WHERE tenant_id=@preprod_tenant_id AND calendar_code='CONTINUOUS_24_7' AND version=1 LIMIT 1
);
INSERT INTO t_option_trading_calendar_session
(tenant_id,calendar_id,weekday,open_second,close_second,create_times)
SELECT @preprod_tenant_id,@preprod_option_calendar_id,weekday.day_no,0,86400,@preprod_now_s
FROM (SELECT 0 day_no UNION ALL SELECT 1 UNION ALL SELECT 2 UNION ALL SELECT 3
      UNION ALL SELECT 4 UNION ALL SELECT 5 UNION ALL SELECT 6) weekday
WHERE EXISTS (
  SELECT 1 FROM t_option_trading_calendar
  WHERE id=@preprod_option_calendar_id AND tenant_id=@preprod_tenant_id AND status=1
)
AND NOT EXISTS (
  SELECT 1 FROM t_option_trading_calendar_session session
  WHERE session.tenant_id=@preprod_tenant_id
    AND session.calendar_id=@preprod_option_calendar_id
    AND session.weekday=weekday.day_no AND session.open_second=0
    AND session.close_second=86400
);

CREATE TEMPORARY TABLE preprod_option_contracts (
  contract_code VARCHAR(64) NOT NULL PRIMARY KEY,
  underlying_symbol VARCHAR(32) NOT NULL,
  underlying_coin VARCHAR(16) NOT NULL,
  option_type TINYINT NOT NULL,
  strike_price DECIMAL(32,16) NOT NULL,
  contract_unit DECIMAL(32,16) NOT NULL,
  sort_no INT NOT NULL
);
INSERT INTO preprod_option_contracts VALUES
('BTC-PREPROD-BASE-C','BTCUSDT','BTC',1,100000,0.001,10),
('BTC-PREPROD-BASE-P','BTCUSDT','BTC',2,100000,0.001,11),
('ETH-PREPROD-BASE-C','ETHUSDT','ETH',1,5000,0.01,20),
('ETH-PREPROD-BASE-P','ETHUSDT','ETH',2,5000,0.01,21);

INSERT INTO t_option_contract
(tenant_id,contract_code,underlying_symbol,underlying_coin,settle_coin,quote_coin,
 option_type,exercise_style,settlement_type,strike_price,contract_unit,min_order_qty,
 max_order_qty,price_tick,qty_step,multiplier,list_time,last_trade_time,expire_time,
 deliver_time,exercise_cutoff_time,auto_exercise_threshold,max_user_long_qty,
 max_user_short_qty,max_open_interest,order_price_band_ratio,circuit_breaker_ratio,
 greeks_max_age_seconds,settlement_price_source,settlement_price_method,
 settlement_window_seconds,settlement_min_samples,is_auto_exercise,maker_fee_rate,
 taker_fee_rate,exercise_fee_rate,fee_user_id,fee_account_id,seller_margin_mode,
 initial_margin_rate,maintenance_margin_rate,min_margin_rate,liquidation_fee_rate,
 insurance_user_id,insurance_account_id,liquidation_deficit_policy,
 physical_delivery_policy,physical_delivery_cure_seconds,trading_calendar_code,
 status,sort,remark,is_deleted,create_times,update_times)
SELECT @preprod_tenant_id,template.contract_code,template.underlying_symbol,
       template.underlying_coin,'USDT','USDT',template.option_type,1,1,
       template.strike_price,template.contract_unit,1,10000,0.1,1,1,
       UNIX_TIMESTAMP('2030-12-01 00:00:00'),
       UNIX_TIMESTAMP('2030-12-31 07:00:00'),
       UNIX_TIMESTAMP('2030-12-31 08:00:00'),
       UNIX_TIMESTAMP('2030-12-31 08:05:00'),
       UNIX_TIMESTAMP('2030-12-31 07:00:00'),
       0,1000,1000,10000,0.20,0.30,30,'price-engine','MEDIAN',60,3,2,
       0.0002,0.0005,0.0005,0,0,1,0,0,0,0,0,0,1,0,0,
       'CONTINUOUS_24_7',1,template.sort_no,'PREPROD_BASELINE_PENDING',2,
       @preprod_now_s,@preprod_now_s
FROM preprod_option_contracts template
WHERE NOT EXISTS (
  SELECT 1 FROM t_option_contract contract
  WHERE contract.tenant_id=@preprod_tenant_id
    AND contract.contract_code=template.contract_code
);

INSERT INTO t_option_portfolio_risk_config
(tenant_id,settle_coin,version,status,model_method,initial_shock_rate,
 maintenance_shock_rate,scenario_shocks,concentration_threshold,
 concentration_addon_rate,liquidity_addon_rate,effective_from,effective_until,
 supersedes_id,source_config_id,change_reason,evidence_ref,created_by,reviewed_by,
 review_reason,reviewed_at,create_times,update_times)
SELECT @preprod_tenant_id,'USDT',1,1,1,0.30,0.20,'-0.30,-0.20,-0.10,0.10,0.20,0.30',
       1000000,0.05,0.02,@preprod_now_s+86400,0,0,0,
       'PREPROD_BASELINE_DRAFT','',1,0,'',0,@preprod_now_s,@preprod_now_s
WHERE NOT EXISTS (
  SELECT 1 FROM t_option_portfolio_risk_config
  WHERE tenant_id=@preprod_tenant_id AND settle_coin='USDT' AND version=1
);

-- Staking uses the underlying BTC and ETH coins. Products start disabled with
-- zero APR/capacity until treasury, legal and risk approvals are recorded.
INSERT INTO t_stake_product
(tenant_id,product_no,product_name,product_type,coin_name,coin_symbol,
 reward_coin_name,reward_coin_symbol,apr,lock_days,min_amount,max_amount,
 step_amount,total_amount,staked_amount,user_limit_amount,interest_mode,
 reward_mode,allow_early_redeem,early_redeem_rate,status,sort,remark,
 create_user_id,update_user_id,create_times,update_times)
SELECT @preprod_tenant_id,CONCAT(pair.base_asset,'-FLEX-PREPROD'),
       CONCAT(pair.base_asset,' Flexible Staking'),1,pair.base_asset,pair.base_asset,
       pair.base_asset,pair.base_asset,0,0,0.0001,0,0.0001,0,0,0,1,1,1,0,1,
       pair.sort_no,'PREPROD_BASELINE_DISABLED',1,1,@preprod_now_ms,@preprod_now_ms
FROM preprod_pairs pair
ON DUPLICATE KEY UPDATE id=LAST_INSERT_ID(t_stake_product.id);

DROP TEMPORARY TABLE preprod_option_contracts;
DROP TEMPORARY TABLE preprod_trade_instruments;
DROP TEMPORARY TABLE preprod_pairs;

COMMIT;
