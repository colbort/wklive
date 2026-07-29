-- 注册真实公开交易所行情生产方。只配置公开市场数据连接，不代表数据许可审批通过。

INSERT INTO `t_market_authority_registry`
(`authority`,`provider_code`,`producer_type`,`allowed_kinds`,`status`,`version`,`create_times`,`update_times`)
VALUES
  ('binance-public','BINANCE','PUBLIC_REST',JSON_ARRAY('FINAL_QUOTE'),1,0,0,0),
  ('okx-public','OKX','PUBLIC_REST',JSON_ARRAY('FINAL_QUOTE'),1,0,0,0),
  ('bybit-public','BYBIT','PUBLIC_REST',JSON_ARRAY('FINAL_QUOTE'),1,0,0,0),
  ('binance-futures-public','BINANCE','PUBLIC_REST',JSON_ARRAY('FINAL_QUOTE'),1,0,0,0)
ON DUPLICATE KEY UPDATE
  `provider_code`=VALUES(`provider_code`),
  `producer_type`=VALUES(`producer_type`),
  `allowed_kinds`=VALUES(`allowed_kinds`);
