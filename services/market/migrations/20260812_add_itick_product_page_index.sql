SET @schema_name = DATABASE();
SET @index_exists = (
  SELECT COUNT(1)
  FROM information_schema.statistics
  WHERE table_schema = @schema_name
    AND table_name = 't_itick_product'
    AND index_name = 'idx_product_page_filter'
);
SET @ddl = IF(
  @index_exists = 0,
  'ALTER TABLE `t_itick_product` ADD INDEX `idx_product_page_filter` (`category_type`, `market`, `enabled`, `app_visible`)',
  'SELECT 1'
);
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;
