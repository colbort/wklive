-- 固化 Price Engine 最终交割价公式版本，避免配置算法名与实际价格来源不一致。
ALTER TABLE `t_contract_delivery_batch`
  ADD COLUMN `formula_version` VARCHAR(64) NOT NULL DEFAULT ''
    COMMENT 'Price Engine交割公式版本' AFTER `price_algorithm`;
