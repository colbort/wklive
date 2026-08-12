ALTER TABLE `t_itick_market_session`
  ADD COLUMN `weekday_mask` TINYINT UNSIGNED NOT NULL DEFAULT 62
  COMMENT 'bit0=Sunday ... bit6=Saturday' AFTER `cross_day`;

UPDATE `t_itick_market_session` s
JOIN `t_itick_market_calendar` c ON c.id=s.calendar_id
SET s.weekday_mask=127
WHERE c.category_code='crypto';

CREATE TABLE `t_itick_product_calendar` (
  `id` BIGINT NOT NULL AUTO_INCREMENT,
  `category_code` VARCHAR(32) NOT NULL DEFAULT '',
  `market` VARCHAR(32) NOT NULL DEFAULT '',
  `symbol` VARCHAR(64) NOT NULL DEFAULT '',
  `calendar_id` BIGINT NOT NULL,
  `source` VARCHAR(64) NOT NULL DEFAULT 'itick-product-list',
  `create_times` BIGINT NOT NULL DEFAULT 0,
  `update_times` BIGINT NOT NULL DEFAULT 0,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_product_calendar_identity` (`category_code`,`market`,`symbol`),
  KEY `idx_product_calendar_calendar` (`calendar_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='iTick产品交易日历映射';
