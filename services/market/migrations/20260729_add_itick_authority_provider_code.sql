-- dbinit:baseline-safe
-- 明确数据供应商身份，防止同一供应商的 WS/REST 通道被当作独立价格源。

SET @dbinit_exists := (
  SELECT COUNT(*)
  FROM information_schema.columns
  WHERE table_schema = DATABASE()
    AND table_name = 't_itick_authority_registry'
    AND column_name = 'provider_code'
);
SET @dbinit_sql := IF(
  @dbinit_exists = 0,
  'ALTER TABLE `t_itick_authority_registry` ADD COLUMN `provider_code` VARCHAR(32) NOT NULL DEFAULT '''' COMMENT ''独立数据供应商标识；同一供应商不同传输通道必须相同'' AFTER `authority`',
  'SELECT 1'
);
PREPARE dbinit_stmt FROM @dbinit_sql;
EXECUTE dbinit_stmt;
DEALLOCATE PREPARE dbinit_stmt;

UPDATE `t_itick_authority_registry`
SET `provider_code` = CASE
  WHEN `authority` IN ('itick-ws','itick-rest') THEN 'ITICK'
  WHEN `authority` = 'price-engine' THEN 'PRICE_ENGINE'
  ELSE UPPER(REPLACE(`authority`, '-', '_'))
END
WHERE `provider_code` = '';

SET @dbinit_exists := (
  SELECT COUNT(*)
  FROM information_schema.table_constraints
  WHERE constraint_schema = DATABASE()
    AND table_name = 't_itick_authority_registry'
    AND constraint_name = 'chk_authority_registry'
    AND constraint_type = 'CHECK'
);
SET @dbinit_sql := IF(
  @dbinit_exists > 0,
  'ALTER TABLE `t_itick_authority_registry` DROP CHECK `chk_authority_registry`',
  'SELECT 1'
);
PREPARE dbinit_stmt FROM @dbinit_sql;
EXECUTE dbinit_stmt;
DEALLOCATE PREPARE dbinit_stmt;

ALTER TABLE `t_itick_authority_registry`
  ADD CONSTRAINT `chk_authority_registry`
  CHECK (
    CHAR_LENGTH(`provider_code`) > 0
    AND `status` IN (1,2)
    AND `version` >= 0
  );
