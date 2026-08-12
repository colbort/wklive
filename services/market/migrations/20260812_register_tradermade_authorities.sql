-- Register TraderMade's independent WebSocket and REST producers under one
-- provider identity. Both are allowed to publish exact FINAL_QUOTE snapshots.
-- dbinit:baseline-safe
INSERT INTO `t_itick_authority_registry`
(`authority`,`provider_code`,`producer_type`,`allowed_kinds`,`status`,`version`,`create_times`,`update_times`)
VALUES
  ('tradermade-ws','TRADERMADE','TRADERMADE_WS',JSON_ARRAY('FINAL_QUOTE'),1,0,0,0),
  ('tradermade-rest','TRADERMADE','TRADERMADE_REST',JSON_ARRAY('FINAL_QUOTE'),1,0,0,0)
ON DUPLICATE KEY UPDATE
  `provider_code`=VALUES(`provider_code`),
  `producer_type`=VALUES(`producer_type`),
  `allowed_kinds`=VALUES(`allowed_kinds`);
