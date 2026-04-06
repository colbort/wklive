CREATE TABLE `t_itick_category` (
  `id` bigint NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `category_type` tinyint NOT NULL DEFAULT '0' COMMENT '产品类型: 1-forex 2-crypto 3-stock 4-future 5-indices 6-fund',
  `category_name` varchar(64) NOT NULL DEFAULT '' COMMENT '产品类型名称',
  `category_code` varchar(64) NOT NULL DEFAULT '' COMMENT '产品类型标识, 如 forex/crypto/stock/future/indices/fund',
  `enabled` tinyint NOT NULL DEFAULT '1' COMMENT '是否启用: 0-否 1-是',
  `app_visible` tinyint NOT NULL DEFAULT '1' COMMENT 'APP是否可见: 0-否 1-是',
  `sort` int NOT NULL DEFAULT '0' COMMENT '排序值,越小越靠前',
  `icon` varchar(255) NOT NULL DEFAULT '' COMMENT '图标',
  `remark` varchar(500) NOT NULL DEFAULT '' COMMENT '备注',
  `create_times` bigint NOT NULL DEFAULT '0' COMMENT '创建时间(毫秒时间戳)',
  `update_times` bigint NOT NULL DEFAULT '0' COMMENT '更新时间(毫秒时间戳)',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_category_type` (`category_type`),
  KEY `idx_enabled_visible_sort` (`enabled`, `app_visible`, `sort`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='itick产品类型表';

INSERT INTO `t_itick_category` (`id`, `category_type`, `category_type_name`, `category_code`, `enabled`, `app_visible`, `sort`, `icon`, `remark`, `create_times`, `update_times`) VALUES
(1, 1, '外汇', 'forex', 1, 1, 1, '', '', 0, 0),
(2, 2, '加密货币', 'crypto', 1, 1, 2, '', '', 0, 0),
(3, 3, '股票', 'stock', 1, 1, 3, '', '', 0, 0),
(4, 4, '期货', 'future', 1, 1, 4, '', '', 0, 0),
(5, 5, '指数', 'indices', 1, 1, 5, '', '', 0, 0),
(6, 6, '基金', 'fund', 1, 1, 6, '', '', 0, 0);


CREATE TABLE `t_itick_product` (
  `id` bigint NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `category_type` tinyint NOT NULL DEFAULT '0' COMMENT '产品类型: 1-forex 2-crypto 3-stock 4-future 5-indices 6-fund',
  `category_name` varchar(64) NOT NULL DEFAULT '' COMMENT '产品类型名称',
  `category_code` varchar(64) NOT NULL DEFAULT '' COMMENT '产品类型标识, 如 forex/crypto/stock/future/indices/fund',
  `market` varchar(64) NOT NULL DEFAULT '' COMMENT '市场/来源, 如 binance/hk/us/forex',
  `symbol` varchar(64) NOT NULL DEFAULT '' COMMENT '产品标识, 如 BTCUSDT/AAPL/EURUSD',
  `code` varchar(128) NOT NULL DEFAULT '' COMMENT '第三方原始code',
  `name` varchar(128) NOT NULL DEFAULT '' COMMENT '产品名称',
  `display_name` varchar(128) NOT NULL DEFAULT '' COMMENT '前端展示名称',
  `exchange` varchar(64) NOT NULL DEFAULT '' COMMENT '交易所, 如 binance/forex/hk/us',
  `sector` varchar(64) NOT NULL DEFAULT '' COMMENT '行业/领域, 如 technology/forex',
  `lug` varchar(64) NOT NULL DEFAULT '' COMMENT 'slug, URL友好标识',
  `base_coin` varchar(64) NOT NULL DEFAULT '' COMMENT '基础币种, 如 BTC',
  `quote_coin` varchar(64) NOT NULL DEFAULT '' COMMENT '计价币种, 如 USDT',
  `enabled` tinyint NOT NULL DEFAULT '1' COMMENT '是否启用: 0-否 1-是',
  `app_visible` tinyint NOT NULL DEFAULT '1' COMMENT 'APP是否可见: 0-否 1-是',
  `sort` int NOT NULL DEFAULT '0' COMMENT '排序值,越小越靠前',
  `icon` varchar(255) NOT NULL DEFAULT '' COMMENT '图标',
  `remark` varchar(500) NOT NULL DEFAULT '' COMMENT '备注',
  `create_times` bigint NOT NULL DEFAULT '0' COMMENT '创建时间(毫秒时间戳)',
  `update_times` bigint NOT NULL DEFAULT '0' COMMENT '更新时间(毫秒时间戳)',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_type_region_symbol` (`category_type`, `market`, `symbol`),
  KEY `idx_category_type` (`category_type`),
  KEY `idx_region` (`market`),
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
  `create_times` bigint NOT NULL DEFAULT '0' COMMENT '创建时间(毫秒时间戳)',
  `update_times` bigint NOT NULL DEFAULT '0' COMMENT '更新时间(毫秒时间戳)',
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
  `create_times` bigint NOT NULL DEFAULT '0' COMMENT '创建时间(毫秒时间戳)',
  `update_times` bigint NOT NULL DEFAULT '0' COMMENT '更新时间(毫秒时间戳)',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_tenant_product` (`tenant_id`, `product_id`),
  KEY `idx_tenant_visible_sort` (`tenant_id`, `enabled`, `app_visible`, `sort`),
  KEY `idx_product_id` (`product_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='租户产品可见配置表';


CREATE TABLE `t_itick_sync_task` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `task_no` varchar(64) NOT NULL DEFAULT '' COMMENT '任务号',
  `task_type` varchar(64) NOT NULL DEFAULT '' COMMENT '任务类型',
  `biz_id` bigint NOT NULL DEFAULT 0 COMMENT '业务id，比如category_id',
  `status` tinyint NOT NULL DEFAULT 0 COMMENT '0待执行 1执行中 2成功 3失败',
  `message` varchar(500) NOT NULL DEFAULT '' COMMENT '结果描述',
  `create_times` bigint NOT NULL DEFAULT 0,
  `update_times` bigint NOT NULL DEFAULT 0,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_task_no` (`task_no`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;


CREATE TABLE `t_itick_quote` (
  `id` bigint NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `category_code` varchar(64) NOT NULL DEFAULT '' COMMENT '产品类型标识, 如 forex/crypto/stock/future/indices/fund',
  `market` varchar(32) NOT NULL DEFAULT '' COMMENT '市场/地区，如 GB',
  `symbol` varchar(64) NOT NULL DEFAULT '' COMMENT '代码，如 EURUSD',

  `last_price` decimal(20,8) NOT NULL DEFAULT '0.00000000' COMMENT '最新价，对应 ld',
  `open_price` decimal(20,8) NOT NULL DEFAULT '0.00000000' COMMENT '开盘价，对应 o',
  `high_price` decimal(20,8) NOT NULL DEFAULT '0.00000000' COMMENT '最高价，对应 h',
  `low_price` decimal(20,8) NOT NULL DEFAULT '0.00000000' COMMENT '最低价，对应 l',
  `prev_close_price` decimal(20,8) NOT NULL DEFAULT '0.00000000' COMMENT '昨收价，按 ld - ch 计算',

  `change_value` decimal(20,8) NOT NULL DEFAULT '0.00000000' COMMENT '涨跌额，对应 ch',
  `change_rate` decimal(10,4) NOT NULL DEFAULT '0.0000' COMMENT '涨跌幅(%)，对应 chp',

  `volume` decimal(20,4) NOT NULL DEFAULT '0.0000' COMMENT '成交量，对应 v',
  `turnover` decimal(20,8) NOT NULL DEFAULT '0.00000000' COMMENT '成交额，对应 tu',

  `quote_ts` bigint NOT NULL DEFAULT 0 COMMENT '行情时间戳(毫秒)，对应 t',
  `trade_status` tinyint NOT NULL DEFAULT 0 COMMENT '交易状态，对应 ts',

  `create_times` bigint NOT NULL DEFAULT 0 COMMENT '创建时间(毫秒)',
  `update_times` bigint NOT NULL DEFAULT 0 COMMENT '更新时间(毫秒)',

  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_region_symbol` (`market`, `symbol`),
  KEY `idx_symbol` (`symbol`),
  KEY `idx_quote_ts` (`quote_ts`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='iTick实时报价表';