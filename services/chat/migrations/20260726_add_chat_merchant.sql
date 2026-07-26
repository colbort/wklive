CREATE TABLE IF NOT EXISTS `t_chat_merchant` (
  `id` bigint NOT NULL AUTO_INCREMENT COMMENT '客服商户ID',
  `merchant_code` varchar(64) NOT NULL DEFAULT '' COMMENT '商户编码/主账号用户名',
  `merchant_name` varchar(128) NOT NULL DEFAULT '' COMMENT '商户名称',
  `enabled` tinyint NOT NULL DEFAULT '1' COMMENT '启用状态:1启用 2禁用',
  `expire_time` bigint NOT NULL DEFAULT '0' COMMENT '到期时间戳(毫秒)',
  `contact_name` varchar(64) NOT NULL DEFAULT '' COMMENT '联系人',
  `contact_phone` varchar(32) NOT NULL DEFAULT '' COMMENT '联系电话',
  `contact_email` varchar(128) NOT NULL DEFAULT '' COMMENT '联系邮箱',
  `remark` varchar(255) NOT NULL DEFAULT '' COMMENT '备注',
  `create_by` varchar(64) NOT NULL DEFAULT '' COMMENT '创建人',
  `create_times` bigint NOT NULL DEFAULT '0' COMMENT '创建时间戳(毫秒)',
  `update_by` varchar(64) NOT NULL DEFAULT '' COMMENT '更新人',
  `update_times` bigint NOT NULL DEFAULT '0' COMMENT '更新时间戳(毫秒)',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_chat_merchant_code` (`merchant_code`),
  KEY `idx_chat_merchant_enabled` (`enabled`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='客服商户主数据';
