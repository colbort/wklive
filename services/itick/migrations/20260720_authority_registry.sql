CREATE TABLE IF NOT EXISTS `t_itick_authority_registry` (
  `id` BIGINT NOT NULL AUTO_INCREMENT,
  `authority` VARCHAR(32) NOT NULL,
  `producer_type` VARCHAR(32) NOT NULL,
  `allowed_kinds` JSON NOT NULL,
  `status` TINYINT NOT NULL DEFAULT 1,
  `version` BIGINT NOT NULL DEFAULT 0,
  `create_times` BIGINT NOT NULL,
  `update_times` BIGINT NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_authority` (`authority`),
  CONSTRAINT `chk_authority_registry` CHECK (`status` IN (1,2) AND `version` >= 0)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

INSERT INTO `t_itick_authority_registry`
(`authority`,`producer_type`,`allowed_kinds`,`status`,`version`,`create_times`,`update_times`)
VALUES ('itick-ws','ITICK_WS',JSON_ARRAY('FINAL_QUOTE'),1,0,0,0)
ON DUPLICATE KEY UPDATE `producer_type`=VALUES(`producer_type`),`allowed_kinds`=VALUES(`allowed_kinds`);
