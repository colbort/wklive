ALTER TABLE `t_tenant_pay_account`
  ADD COLUMN `credential_ref` varchar(255) NOT NULL DEFAULT ''
  COMMENT '密钥管理系统引用，生产环境优先使用' AFTER `cert_cipher`;

ALTER TABLE `t_recharge_order`
  ADD COLUMN `credit_status` tinyint NOT NULL DEFAULT 1
  COMMENT '入账状态：1待入账 2入账中 3入账成功 4入账失败' AFTER `notify_data`,
  ADD COLUMN `credited_time` bigint NOT NULL DEFAULT 0 COMMENT '资产入账时间' AFTER `credit_status`,
  ADD COLUMN `credit_retry_count` int NOT NULL DEFAULT 0 COMMENT '入账重试次数' AFTER `credited_time`,
  ADD COLUMN `last_credit_error` varchar(1000) NOT NULL DEFAULT ''
  COMMENT '最近入账错误' AFTER `credit_retry_count`;

ALTER TABLE `t_recharge_notify_log`
  ADD COLUMN `notify_id` varchar(128) NOT NULL DEFAULT ''
  COMMENT '三方通知唯一标识或请求内容哈希' AFTER `notify_body`;

UPDATE `t_recharge_notify_log`
SET `notify_id` = CONCAT('LEGACY-', `id`)
WHERE `notify_id` = '';

ALTER TABLE `t_recharge_notify_log`
  ADD UNIQUE KEY `uk_platform_notify` (`platform_id`, `notify_id`);

CREATE TABLE `t_pay_request_log` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `tenant_id` bigint NOT NULL DEFAULT 0,
  `order_type` tinyint NOT NULL,
  `order_id` bigint NOT NULL,
  `order_no` varchar(64) NOT NULL,
  `platform_id` bigint NOT NULL,
  `account_id` bigint NOT NULL,
  `request_type` tinyint NOT NULL,
  `request_no` varchar(64) NOT NULL,
  `request_data` json DEFAULT NULL,
  `response_data` json DEFAULT NULL,
  `http_status` int NOT NULL DEFAULT 0,
  `third_code` varchar(64) NOT NULL DEFAULT '',
  `third_message` varchar(1000) NOT NULL DEFAULT '',
  `status` tinyint NOT NULL DEFAULT 1,
  `duration_ms` bigint NOT NULL DEFAULT 0,
  `create_times` bigint NOT NULL DEFAULT 0,
  `update_times` bigint NOT NULL DEFAULT 0,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_request_no` (`request_no`),
  KEY `idx_order_no` (`order_no`),
  KEY `idx_order` (`order_type`, `order_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='支付三方请求流水';

CREATE TABLE `t_pay_outbox` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `event_no` varchar(64) NOT NULL,
  `event_type` varchar(64) NOT NULL,
  `aggregate_type` varchar(32) NOT NULL,
  `aggregate_id` bigint NOT NULL,
  `aggregate_no` varchar(64) NOT NULL,
  `payload` json NOT NULL,
  `status` tinyint NOT NULL DEFAULT 1,
  `retry_count` int NOT NULL DEFAULT 0,
  `next_retry_at` bigint NOT NULL DEFAULT 0,
  `last_error_msg` varchar(1000) NOT NULL DEFAULT '',
  `create_times` bigint NOT NULL DEFAULT 0,
  `update_times` bigint NOT NULL DEFAULT 0,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_event_no` (`event_no`),
  KEY `idx_status_retry` (`status`, `next_retry_at`),
  KEY `idx_aggregate` (`aggregate_type`, `aggregate_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='支付可靠事件表';
