-- dbinit:baseline-safe

INSERT INTO `t_itick_authority_registry`
  (`authority`,`provider_code`,`producer_type`,`allowed_kinds`,`status`,`version`,`create_times`,`update_times`)
VALUES
  ('twelvedata-ws','TWELVEDATA','TWELVEDATA_WS',JSON_ARRAY('FINAL_QUOTE'),1,0,0,0),
  ('twelvedata-rest','TWELVEDATA','TWELVEDATA_REST',JSON_ARRAY('FINAL_QUOTE'),1,0,0,0)
ON DUPLICATE KEY UPDATE
  `provider_code`=VALUES(`provider_code`),
  `producer_type`=VALUES(`producer_type`),
  `allowed_kinds`=VALUES(`allowed_kinds`),
  `status`=VALUES(`status`),
  `update_times`=VALUES(`update_times`);
