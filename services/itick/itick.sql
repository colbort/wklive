CREATE TABLE `t_itick_category` (
  `id` bigint NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `category_type` tinyint NOT NULL DEFAULT '0' COMMENT '产品类型: 1-forex 2-crypto 3-stock 4-future 5-indices 6-fund',
  `category_type_name` varchar(64) NOT NULL DEFAULT '' COMMENT '产品类型名称',
  `category_code` varchar(64) NOT NULL DEFAULT '' COMMENT '产品类型标识, 如 forex/crypto/stock/future/indices/fund',
  `enabled` tinyint NOT NULL DEFAULT '1' COMMENT '是否启用: 0-否 1-是',
  `app_visible` tinyint NOT NULL DEFAULT '1' COMMENT 'APP是否可见: 0-否 1-是',
  `sort` int NOT NULL DEFAULT '0' COMMENT '排序值,越小越靠前',
  `icon` varchar(255) NOT NULL DEFAULT '' COMMENT '图标',
  `remark` varchar(500) NOT NULL DEFAULT '' COMMENT '备注',
  `create_time` bigint NOT NULL DEFAULT '0' COMMENT '创建时间(毫秒时间戳)',
  `update_time` bigint NOT NULL DEFAULT '0' COMMENT '更新时间(毫秒时间戳)',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_category_type` (`category_type`),
  KEY `idx_enabled_visible_sort` (`enabled`, `app_visible`, `sort`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='itick产品类型表';

INSERT INTO `t_itick_category` (`id`, `category_type`, `category_type_name`, `category_code`, `enabled`, `app_visible`, `sort`, `icon`, `remark`, `create_time`, `update_time`) VALUES
(1, 1, '外汇', 'forex', 1, 1, 1, '', '', 0, 0),
(2, 2, '加密货币', 'crypto', 1, 1, 2, '', '', 0, 0),
(3, 3, '股票', 'stock', 1, 1, 3, '', '', 0, 0),
(4, 4, '期货', 'future', 1, 1, 4, '', '', 0, 0),
(5, 5, '指数', 'indices', 1, 1, 5, '', '', 0, 0),
(6, 6, '基金', 'fund', 1, 1, 6, '', '', 0, 0);


CREATE TABLE `t_itick_product` (
  `id` bigint NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `category_type` tinyint NOT NULL DEFAULT '0' COMMENT '产品类型: 1-forex 2-crypto 3-stock 4-future 5-indices 6-fund',
  `market` varchar(64) NOT NULL DEFAULT '' COMMENT '市场/来源, 如 binance/hk/us/forex',
  `symbol` varchar(64) NOT NULL DEFAULT '' COMMENT '产品标识, 如 BTCUSDT/AAPL/EURUSD',
  `code` varchar(128) NOT NULL DEFAULT '' COMMENT '第三方原始code',
  `name` varchar(128) NOT NULL DEFAULT '' COMMENT '产品名称',
  `display_name` varchar(128) NOT NULL DEFAULT '' COMMENT '前端展示名称',
  `base_coin` varchar(64) NOT NULL DEFAULT '' COMMENT '基础币种, 如 BTC',
  `quote_coin` varchar(64) NOT NULL DEFAULT '' COMMENT '计价币种, 如 USDT',
  `enabled` tinyint NOT NULL DEFAULT '1' COMMENT '是否启用: 0-否 1-是',
  `app_visible` tinyint NOT NULL DEFAULT '1' COMMENT 'APP是否可见: 0-否 1-是',
  `sort` int NOT NULL DEFAULT '0' COMMENT '排序值,越小越靠前',
  `icon` varchar(255) NOT NULL DEFAULT '' COMMENT '图标',
  `remark` varchar(500) NOT NULL DEFAULT '' COMMENT '备注',
  `create_time` bigint NOT NULL DEFAULT '0' COMMENT '创建时间(毫秒时间戳)',
  `update_time` bigint NOT NULL DEFAULT '0' COMMENT '更新时间(毫秒时间戳)',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_type_market_symbol` (`category_type`, `market`, `symbol`),
  KEY `idx_category_type` (`category_type`),
  KEY `idx_market` (`market`),
  KEY `idx_enabled_visible_sort` (`enabled`, `app_visible`, `sort`),
  KEY `idx_keyword_query` (`category_type`, `market`, `name`, `display_name`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='itick产品表';


CREATE TABLE `t_itick_tenant_category` (
  `id` bigint NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `tenant_id` bigint NOT NULL DEFAULT '0' COMMENT '租户ID',
  `category_id` bigint NOT NULL DEFAULT '0' COMMENT '产品类型ID, 对应 itick_category.id',
  `enabled` tinyint NOT NULL DEFAULT '1' COMMENT '是否启用: 0-否 1-是',
  `app_visible` tinyint NOT NULL DEFAULT '1' COMMENT 'APP是否可见: 0-否 1-是',
  `sort` int NOT NULL DEFAULT '0' COMMENT '租户排序, 越小越靠前',
  `remark` varchar(500) NOT NULL DEFAULT '' COMMENT '备注',
  `create_time` bigint NOT NULL DEFAULT '0' COMMENT '创建时间(毫秒时间戳)',
  `update_time` bigint NOT NULL DEFAULT '0' COMMENT '更新时间(毫秒时间戳)',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_tenant_category` (`tenant_id`, `category_id`),
  KEY `idx_tenant_visible_sort` (`tenant_id`, `enabled`, `app_visible`, `sort`),
  KEY `idx_category_id` (`category_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='租户产品类型可见配置表';


CREATE TABLE `t_itick_tenant_product` (
  `id` bigint NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `tenant_id` bigint NOT NULL DEFAULT '0' COMMENT '租户ID',
  `product_id` bigint NOT NULL DEFAULT '0' COMMENT '产品ID, 对应 itick_product.id',
  `enabled` tinyint NOT NULL DEFAULT '1' COMMENT '是否启用: 0-否 1-是',
  `app_visible` tinyint NOT NULL DEFAULT '1' COMMENT 'APP是否可见: 0-否 1-是',
  `sort` int NOT NULL DEFAULT '0' COMMENT '租户排序, 越小越靠前',
  `remark` varchar(500) NOT NULL DEFAULT '' COMMENT '备注',
  `create_time` bigint NOT NULL DEFAULT '0' COMMENT '创建时间(毫秒时间戳)',
  `update_time` bigint NOT NULL DEFAULT '0' COMMENT '更新时间(毫秒时间戳)',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_tenant_product` (`tenant_id`, `product_id`),
  KEY `idx_tenant_visible_sort` (`tenant_id`, `enabled`, `app_visible`, `sort`),
  KEY `idx_product_id` (`product_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='租户产品可见配置表';