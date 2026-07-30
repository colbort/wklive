CREATE TABLE IF NOT EXISTS `t_option_client_order_key` (
  `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `tenant_id` BIGINT NOT NULL DEFAULT 0 COMMENT '租户ID',
  `user_id` BIGINT NOT NULL DEFAULT 0 COMMENT '用户ID',
  `client_order_id` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '客户端幂等订单号，禁止为空',
  `order_id` BIGINT NOT NULL DEFAULT 0 COMMENT 'Option订单ID',
  `order_no` VARCHAR(64) NOT NULL DEFAULT '' COMMENT 'Option订单号',
  `create_times` BIGINT NOT NULL DEFAULT 0 COMMENT '创建时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_tenant_user_client_order` (`tenant_id`, `user_id`, `client_order_id`),
  KEY `idx_tenant_order_id` (`tenant_id`, `order_id`),
  CONSTRAINT `chk_option_client_order_key` CHECK (`client_order_id` <> '' AND `order_id` > 0)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='期权客户端订单幂等键';

INSERT INTO `t_option_client_order_key`
  (`tenant_id`, `user_id`, `client_order_id`, `order_id`, `order_no`, `create_times`)
SELECT `tenant_id`, `user_id`, `client_order_id`, `id`, `order_no`, `create_times`
FROM `t_option_order`
WHERE `client_order_id` <> ''
ON DUPLICATE KEY UPDATE
  `order_id` = VALUES(`order_id`),
  `order_no` = VALUES(`order_no`);

SET @option_client_order_index_sql = (
  SELECT IF(
    EXISTS (
      SELECT 1
      FROM information_schema.statistics
      WHERE table_schema = DATABASE()
        AND table_name = 't_option_order'
        AND index_name = 'uk_tenant_uid_client_order_id'
    ),
    'ALTER TABLE `t_option_order`
       DROP INDEX `uk_tenant_uid_client_order_id`,
       ADD KEY `idx_tenant_uid_client_order_id` (`tenant_id`, `user_id`, `client_order_id`)',
    'SELECT 1'
  )
);
PREPARE option_client_order_index_stmt FROM @option_client_order_index_sql;
EXECUTE option_client_order_index_stmt;
DEALLOCATE PREPARE option_client_order_index_stmt;
