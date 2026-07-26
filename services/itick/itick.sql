DROP TABLE IF EXISTS `t_itick_category`;
CREATE TABLE `t_itick_category` (
  `id` bigint NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `category_type` tinyint NOT NULL DEFAULT '0' COMMENT '产品类型: 1-forex 2-crypto 3-stock 4-future 5-indices 6-fund',
  `category_name` varchar(64) NOT NULL DEFAULT '' COMMENT '产品类型名称',
  `category_code` varchar(64) NOT NULL DEFAULT '' COMMENT '产品类型标识, 如 forex/crypto/stock/future/indices/fund',
  `enabled` tinyint NOT NULL DEFAULT '1' COMMENT '启用状态: 1-启用 2-禁用',
  `app_visible` tinyint NOT NULL DEFAULT '1' COMMENT 'APP可见开关: 1-显示 2-隐藏',
  `sync_priority` tinyint NOT NULL DEFAULT '2' COMMENT 'K线同步优先级: 1-高 2-普通 3-低',
  `sort` int NOT NULL DEFAULT '0' COMMENT '排序值,越小越靠前',
  `icon` varchar(255) NOT NULL DEFAULT '' COMMENT '图标',
  `remark` varchar(500) NOT NULL DEFAULT '' COMMENT '备注',
  `create_times` bigint NOT NULL DEFAULT '0' COMMENT '创建时间(毫秒时间戳)',
  `update_times` bigint NOT NULL DEFAULT '0' COMMENT '更新时间(毫秒时间戳)',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_category_type` (`category_type`),
  KEY `idx_enabled_visible_sort` (`enabled`, `app_visible`, `sort`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='itick产品类型表';

INSERT INTO `t_itick_category` (`id`, `category_type`, `category_name`, `category_code`, `enabled`, `app_visible`, `sort`, `icon`, `remark`, `create_times`, `update_times`) VALUES
(1, 1, '外汇',     'forex',   1, 1, 1, '', '', UNIX_TIMESTAMP() * 1000, UNIX_TIMESTAMP() * 1000),
(2, 2, '加密货币', 'crypto',  1, 1, 2, '', '', UNIX_TIMESTAMP() * 1000, UNIX_TIMESTAMP() * 1000),
(3, 3, '股票',     'stock',   1, 1, 3, '', '', UNIX_TIMESTAMP() * 1000, UNIX_TIMESTAMP() * 1000),
(4, 4, '期货',     'future',  1, 1, 4, '', '', UNIX_TIMESTAMP() * 1000, UNIX_TIMESTAMP() * 1000),
(5, 5, '指数',     'indices', 1, 1, 5, '', '', UNIX_TIMESTAMP() * 1000, UNIX_TIMESTAMP() * 1000),
(6, 6, '基金',     'fund',    1, 1, 6, '', '', UNIX_TIMESTAMP() * 1000, UNIX_TIMESTAMP() * 1000);


DROP TABLE IF EXISTS `t_itick_product`;
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
  `enabled` tinyint NOT NULL DEFAULT '1' COMMENT '启用状态: 1-启用 2-禁用',
  `app_visible` tinyint NOT NULL DEFAULT '1' COMMENT 'APP可见开关: 1-显示 2-隐藏',
  `sync_priority` tinyint NOT NULL DEFAULT '2' COMMENT 'K线同步优先级: 1-高 2-普通 3-低',
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


DROP TABLE IF EXISTS `t_itick_tenant_category`;
CREATE TABLE `t_itick_tenant_category` (
  `id` bigint NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `tenant_id` bigint NOT NULL DEFAULT '0' COMMENT '租户ID',
  `category_id` bigint NOT NULL DEFAULT '0' COMMENT '产品类型ID, 对应 itick_category.id',
  `enabled` tinyint NOT NULL DEFAULT '1' COMMENT '启用状态: 1-启用 2-禁用',
  `app_visible` tinyint NOT NULL DEFAULT '1' COMMENT 'APP可见开关: 1-显示 2-隐藏',
  `sort` int NOT NULL DEFAULT '0' COMMENT '租户排序, 越小越靠前',
  `remark` varchar(500) NOT NULL DEFAULT '' COMMENT '备注',
  `create_times` bigint NOT NULL DEFAULT '0' COMMENT '创建时间(毫秒时间戳)',
  `update_times` bigint NOT NULL DEFAULT '0' COMMENT '更新时间(毫秒时间戳)',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_tenant_category` (`tenant_id`, `category_id`),
  KEY `idx_tenant_visible_sort` (`tenant_id`, `enabled`, `app_visible`, `sort`),
  KEY `idx_category_id` (`category_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='租户产品类型可见配置表';


DROP TABLE IF EXISTS `t_itick_tenant_product`;
CREATE TABLE `t_itick_tenant_product` (
  `id` bigint NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `tenant_id` bigint NOT NULL DEFAULT '0' COMMENT '租户ID',
  `product_id` bigint NOT NULL DEFAULT '0' COMMENT '产品ID, 对应 itick_product.id',
  `enabled` tinyint NOT NULL DEFAULT '1' COMMENT '启用状态: 1-启用 2-禁用',
  `app_visible` tinyint NOT NULL DEFAULT '1' COMMENT 'APP可见开关: 1-显示 2-隐藏',
  `display_name` varchar(128) NOT NULL DEFAULT '' COMMENT '租户自定义展示名称，为空时使用产品展示名称',
  `sort` int NOT NULL DEFAULT '0' COMMENT '租户排序, 越小越靠前',
  `remark` varchar(500) NOT NULL DEFAULT '' COMMENT '备注',
  `create_times` bigint NOT NULL DEFAULT '0' COMMENT '创建时间(毫秒时间戳)',
  `update_times` bigint NOT NULL DEFAULT '0' COMMENT '更新时间(毫秒时间戳)',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_tenant_product` (`tenant_id`, `product_id`),
  KEY `idx_tenant_visible_sort` (`tenant_id`, `enabled`, `app_visible`, `sort`),
  KEY `idx_product_id` (`product_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='租户产品可见配置表';


DROP TABLE IF EXISTS `t_itick_sync_task`;
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


DROP TABLE IF EXISTS `t_itick_quote`;
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

DROP TABLE IF EXISTS `t_itick_authority_registry`;
CREATE TABLE `t_itick_authority_registry` (
  `id` BIGINT NOT NULL AUTO_INCREMENT,
  `authority` VARCHAR(32) NOT NULL,
  `producer_type` VARCHAR(32) NOT NULL COMMENT 'ITICK_WS/ITICK_REST/PRICE_ENGINE',
  `allowed_kinds` JSON NOT NULL COMMENT '允许发布的快照类型',
  `status` TINYINT NOT NULL DEFAULT 1 COMMENT '1启用 2禁用',
  `version` BIGINT NOT NULL DEFAULT 0,
  `create_times` BIGINT NOT NULL,
  `update_times` BIGINT NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_authority` (`authority`),
  CONSTRAINT `chk_authority_registry` CHECK (`status` IN (1,2) AND `version` >= 0)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='权威行情生产方注册表';

INSERT INTO `t_itick_authority_registry`
(`authority`,`producer_type`,`allowed_kinds`,`status`,`version`,`create_times`,`update_times`)
VALUES ('itick-ws','ITICK_WS',JSON_ARRAY('FINAL_QUOTE'),1,0,0,0)
ON DUPLICATE KEY UPDATE `producer_type`=VALUES(`producer_type`),`allowed_kinds`=VALUES(`allowed_kinds`);

INSERT INTO `t_itick_authority_registry`
(`authority`,`producer_type`,`allowed_kinds`,`status`,`version`,`create_times`,`update_times`)
VALUES ('itick-rest','ITICK_REST',JSON_ARRAY('FINAL_QUOTE'),1,0,0,0)
ON DUPLICATE KEY UPDATE `producer_type`=VALUES(`producer_type`),`allowed_kinds`=VALUES(`allowed_kinds`);

DROP TABLE IF EXISTS `t_itick_authoritative_snapshot`;
CREATE TABLE `t_itick_authoritative_snapshot` (
  `id` BIGINT NOT NULL AUTO_INCREMENT,
  `snapshot_id` VARCHAR(64) NOT NULL COMMENT '内容哈希ID',
  `authority` VARCHAR(32) NOT NULL COMMENT '权威生产方',
  `snapshot_kind` VARCHAR(32) NOT NULL COMMENT 'FINAL_QUOTE/MARK/INDEX/FUNDING/DELIVERY',
  `category_code` VARCHAR(64) NOT NULL DEFAULT '',
  `market` VARCHAR(32) NOT NULL DEFAULT '',
  `symbol` VARCHAR(64) NOT NULL,
  `price` DECIMAL(65,30) NOT NULL,
  `source_timestamp` BIGINT NOT NULL,
  `snapshot_timestamp` BIGINT NOT NULL,
  `revision` BIGINT NOT NULL,
  `formula_version` VARCHAR(64) NOT NULL DEFAULT '',
  `raw_payload` JSON NOT NULL,
  `create_times` BIGINT NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_snapshot_id` (`snapshot_id`),
  KEY `idx_product_time` (`authority`,`snapshot_kind`,`category_code`,`market`,`symbol`,`source_timestamp`,`revision`),
  CONSTRAINT `chk_authoritative_snapshot` CHECK ((`snapshot_kind`='FUNDING' OR `price` > 0) AND `source_timestamp` > 0 AND `snapshot_timestamp` > 0 AND `revision` > 0)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='iTick/Price Engine权威行情永久档案';

DROP TABLE IF EXISTS `t_itick_snapshot_outbox`;
CREATE TABLE `t_itick_snapshot_outbox` (
  `id` BIGINT NOT NULL AUTO_INCREMENT, `snapshot_id` VARCHAR(64) NOT NULL,
  `payload` JSON NOT NULL, `status` TINYINT NOT NULL DEFAULT 1 COMMENT '1 pending 2 processing 3 success 4 failed 5 manual',
  `retry_count` INT NOT NULL DEFAULT 0, `next_retry_at` BIGINT NOT NULL DEFAULT 0,
  `redis_published_at` BIGINT NOT NULL DEFAULT 0 COMMENT 'Redis权威快照发布完成时间',
  `event_published_at` BIGINT NOT NULL DEFAULT 0 COMMENT '权威行情Kafka事件发布完成时间；无需发布时同样置完成',
  `last_error_msg` VARCHAR(500) NOT NULL DEFAULT '', `create_times` BIGINT NOT NULL, `update_times` BIGINT NOT NULL,
  PRIMARY KEY (`id`), UNIQUE KEY `uk_snapshot_outbox` (`snapshot_id`),
  KEY `idx_snapshot_outbox_retry` (`status`,`next_retry_at`,`id`),
  KEY `idx_snapshot_outbox_cleanup` (`status`,`update_times`,`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='权威行情异步发布与Redis修复任务';

DROP TABLE IF EXISTS `t_itick_snapshot_revocation`;
CREATE TABLE `t_itick_snapshot_revocation` (
  `id` BIGINT NOT NULL AUTO_INCREMENT,
  `snapshot_id` VARCHAR(64) NOT NULL,
  `replacement_snapshot_id` VARCHAR(64) NOT NULL DEFAULT '',
  `reason` VARCHAR(512) NOT NULL,
  `create_times` BIGINT NOT NULL,
  PRIMARY KEY (`id`), UNIQUE KEY `uk_snapshot_revocation` (`snapshot_id`),
  KEY `idx_snapshot_revocation_rebuild` (`id`),
  CONSTRAINT `chk_snapshot_revocation_reason` CHECK (CHAR_LENGTH(`reason`) > 0)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='权威快照不可变撤销事实';

DROP TABLE IF EXISTS `t_itick_price_formula`;
CREATE TABLE `t_itick_price_formula` (
  `id` BIGINT NOT NULL AUTO_INCREMENT, `formula_no` VARCHAR(64) NOT NULL,
  `authority` VARCHAR(32) NOT NULL DEFAULT 'price-engine', `snapshot_kind` VARCHAR(32) NOT NULL,
  `category_code` VARCHAR(64) NOT NULL DEFAULT '', `market` VARCHAR(32) NOT NULL DEFAULT '', `symbol` VARCHAR(64) NOT NULL,
  `algorithm` TINYINT NOT NULL COMMENT '1 weighted mean,2 median,3 premium rate', `formula_version` VARCHAR(64) NOT NULL,
  `components` JSON NOT NULL, `max_lookback_ms` BIGINT NOT NULL DEFAULT 30000,
  `max_deviation_bps` INT NOT NULL DEFAULT 0, `interval_ms` BIGINT NOT NULL DEFAULT 1000,
  `last_target_time` BIGINT NOT NULL DEFAULT 0, `status` TINYINT NOT NULL DEFAULT 2 COMMENT '1 active,2 inactive,3 revoked',
  `version` BIGINT NOT NULL DEFAULT 0 COMMENT '配置版本', `run_version` BIGINT NOT NULL DEFAULT 0 COMMENT 'Worker CAS版本', `create_times` BIGINT NOT NULL, `update_times` BIGINT NOT NULL,
  PRIMARY KEY (`id`), UNIQUE KEY `uk_price_formula_no` (`formula_no`),
  UNIQUE KEY `uk_price_formula_output` (`authority`,`snapshot_kind`,`category_code`,`market`,`symbol`,`formula_version`),
  KEY `idx_price_formula_due` (`status`,`last_target_time`),
  CONSTRAINT `chk_price_formula` CHECK (`snapshot_kind` IN ('MARK','INDEX','FUNDING','DELIVERY') AND `algorithm` IN (1,2,3) AND `max_lookback_ms` > 0 AND `interval_ms` > 0 AND `status` IN (1,2,3))
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='版本化权威价格公式';

INSERT INTO `t_itick_authority_registry` (`authority`,`producer_type`,`allowed_kinds`,`status`,`version`,`create_times`,`update_times`)
VALUES ('price-engine','PRICE_ENGINE',JSON_ARRAY('MARK','INDEX','FUNDING','DELIVERY'),1,0,0,0)
ON DUPLICATE KEY UPDATE `producer_type`=VALUES(`producer_type`),`allowed_kinds`=VALUES(`allowed_kinds`);

DROP TABLE IF EXISTS `t_itick_kline_sync_progress`;
CREATE TABLE `t_itick_kline_sync_progress` (
  `id` bigint NOT NULL AUTO_INCREMENT COMMENT '主键ID',

  `category_code` varchar(32) NOT NULL DEFAULT '' COMMENT '品类代码：stock/forex/indices/crypto/future/fund',
  `market` varchar(32) NOT NULL DEFAULT '' COMMENT '市场代码：如 US、HK、CN、BA、GB',
  `symbol` varchar(64) NOT NULL DEFAULT '' COMMENT '产品代码/交易代码，如 AAPL、BTCUSDT',
  `interval` varchar(16) NOT NULL DEFAULT '' COMMENT 'K线周期：1m、5m、15m、30m、1h、1d、1w、1mo',

  `latest_ts` bigint NOT NULL DEFAULT '0' COMMENT '最新已同步K线时间戳（毫秒）。仅表示REST见过的最大时间，不代表中间连续',
  `contiguous_ts` bigint NOT NULL DEFAULT '0' COMMENT '最后连续完整已确认K线时间戳（毫秒）',
  `recent_check_ts` bigint NOT NULL DEFAULT '0' COMMENT '最近一次REST校准时间（毫秒）',
  `oldest_ts` bigint NOT NULL DEFAULT '0' COMMENT '最早已同步K线时间戳（毫秒）。当 full_synced=0 时，从该值继续向前回补历史',

  `full_synced` tinyint NOT NULL DEFAULT '0' COMMENT '历史是否补齐：0=未补齐，1=已补齐',
  `sync_status` tinyint NOT NULL DEFAULT '0' COMMENT '同步状态：0=未开始，1=同步中，2=成功，3=失败',

  `last_sync_mode` varchar(16) NOT NULL DEFAULT '' COMMENT '最近一次同步模式：history=历史回补，incremental=增量同步',
  `last_sync_message` varchar(255) NOT NULL DEFAULT '' COMMENT '最近一次同步结果或错误信息',

  `last_success_time` bigint NOT NULL DEFAULT '0' COMMENT '最近一次同步成功时间（毫秒时间戳）',
  `last_fail_time` bigint NOT NULL DEFAULT '0' COMMENT '最近一次同步失败时间（毫秒时间戳）',

  `create_times` bigint NOT NULL DEFAULT '0' COMMENT '创建时间（毫秒时间戳）',
  `update_times` bigint NOT NULL DEFAULT '0' COMMENT '更新时间（毫秒时间戳）',

  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_category_market_symbol_interval` (`category_code`, `market`, `symbol`, `interval`),
  KEY `idx_sync_status` (`sync_status`),
  KEY `idx_full_synced` (`full_synced`),
  KEY `idx_update_times` (`update_times`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='iTick K线同步进度表';

DROP TABLE IF EXISTS `t_itick_market_calendar`;
CREATE TABLE `t_itick_market_calendar` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `category_code` varchar(32) NOT NULL DEFAULT '',
  `market` varchar(32) NOT NULL DEFAULT '',
  `exchange` varchar(64) NOT NULL DEFAULT '',
  `timezone` varchar(64) NOT NULL DEFAULT 'UTC' COMMENT 'IANA时区',
  `trading_day_offset` int NOT NULL DEFAULT 0 COMMENT '夜盘归属交易日偏移',
  `week_start` tinyint NOT NULL DEFAULT 1 COMMENT '1=周一,7=周日',
  `enabled` tinyint NOT NULL DEFAULT 1,
  `remark` varchar(255) NOT NULL DEFAULT '',
  `create_times` bigint NOT NULL DEFAULT 0,
  `update_times` bigint NOT NULL DEFAULT 0,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_category_market_exchange` (`category_code`,`market`,`exchange`),
  KEY `idx_enabled` (`enabled`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='iTick市场交易日历';

DROP TABLE IF EXISTS `t_itick_market_session`;
CREATE TABLE `t_itick_market_session` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `calendar_id` bigint NOT NULL,
  `session_type` varchar(32) NOT NULL DEFAULT 'regular',
  `start_time` varchar(8) NOT NULL DEFAULT '',
  `end_time` varchar(8) NOT NULL DEFAULT '',
  `cross_day` tinyint NOT NULL DEFAULT 0,
  `sort` int NOT NULL DEFAULT 0,
  PRIMARY KEY (`id`),
  KEY `idx_calendar_sort` (`calendar_id`,`sort`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='iTick市场交易时段';

DROP TABLE IF EXISTS `t_itick_market_holiday`;
CREATE TABLE `t_itick_market_holiday` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `calendar_id` bigint NOT NULL,
  `trade_date` date NOT NULL,
  `day_type` varchar(32) NOT NULL DEFAULT 'closed',
  `open_time` varchar(8) NOT NULL DEFAULT '',
  `close_time` varchar(8) NOT NULL DEFAULT '',
  `remark` varchar(255) NOT NULL DEFAULT '',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_calendar_date` (`calendar_id`,`trade_date`),
  KEY `idx_trade_date` (`trade_date`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='iTick市场节假日及特殊交易日';
