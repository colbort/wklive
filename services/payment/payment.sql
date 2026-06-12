CREATE TABLE `t_pay_platform` (
  `id` bigint NOT NULL AUTO_INCREMENT COMMENT '平台ID',
  `platform_code` varchar(64) NOT NULL COMMENT '平台编码，如 alipay/wechat/stripe/usdt/bank',
  `platform_name` varchar(128) NOT NULL COMMENT '平台名称',
  `platform_type` tinyint NOT NULL DEFAULT 1 COMMENT '类型：1三方支付 2银行转账 3链上支付 4人工代收',
  `notify_url` varchar(255) DEFAULT NULL COMMENT '统一异步通知地址',
  `return_url` varchar(255) DEFAULT NULL COMMENT '默认同步跳转地址',
  `icon` varchar(255) DEFAULT NULL COMMENT '图标',
  `enabled` tinyint NOT NULL DEFAULT 1 COMMENT '启用状态：1启用 2禁用',
  `remark` varchar(255) DEFAULT NULL COMMENT '备注',
  `create_times` bigint NOT NULL DEFAULT 0 COMMENT '创建时间',
  `update_times` bigint NOT NULL DEFAULT 0 COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_platform_code` (`platform_code`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='支付平台表';

CREATE TABLE `t_pay_product` (
  `id` bigint NOT NULL AUTO_INCREMENT COMMENT '产品ID',
  `platform_id` bigint NOT NULL COMMENT '平台ID',
  `product_code` varchar(64) NOT NULL COMMENT '产品编码',
  `product_name` varchar(128) NOT NULL COMMENT '产品名称',
  `scene_type` tinyint NOT NULL DEFAULT 1 COMMENT '场景：1APP 2H5 3WEB 4收银台 5链上',
  `currency` varchar(16) NOT NULL DEFAULT 'CNY' COMMENT '币种',
  `enabled` tinyint NOT NULL DEFAULT 1 COMMENT '启用状态：1启用 2禁用',
  `remark` varchar(255) DEFAULT NULL COMMENT '备注',
  `create_times` bigint NOT NULL DEFAULT 0 COMMENT '创建时间',
  `update_times` bigint NOT NULL DEFAULT 0 COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_product_code` (`product_code`),
  KEY `idx_platform_id` (`platform_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='支付产品表';

CREATE TABLE `t_tenant_pay_platform` (
  `id` bigint NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `tenant_id` bigint NOT NULL COMMENT '租户ID',
  `platform_id` bigint NOT NULL COMMENT '平台ID',
  `enabled` tinyint NOT NULL DEFAULT 1 COMMENT '启用状态：1启用 2禁用',
  `open_status` tinyint NOT NULL DEFAULT 1 COMMENT '开通状态：1待配置 2已开通 3审核中 4驳回',
  `remark` varchar(255) DEFAULT NULL COMMENT '备注',
  `create_times` bigint NOT NULL DEFAULT 0 COMMENT '创建时间',
  `update_times` bigint NOT NULL DEFAULT 0 COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_tenant_platform` (`tenant_id`, `platform_id`),
  KEY `idx_platform_id` (`platform_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='租户开通支付平台表';

CREATE TABLE `t_tenant_pay_account` (
  `id` bigint NOT NULL AUTO_INCREMENT COMMENT '账号ID',
  `tenant_id` bigint NOT NULL COMMENT '租户ID',
  `tenant_pay_platform_id` bigint NOT NULL COMMENT '租户开通平台ID',
  `platform_id` bigint NOT NULL COMMENT '平台ID',
  `account_code` varchar(64) NOT NULL COMMENT '账号编码',
  `account_name` varchar(128) NOT NULL COMMENT '账号名称',
  `app_id` varchar(128) DEFAULT NULL COMMENT '应用ID',
  `merchant_id` varchar(128) DEFAULT NULL COMMENT '商户号',
  `merchant_name` varchar(128) DEFAULT NULL COMMENT '商户名称',
  `api_key_cipher` text COMMENT 'API Key密文',
  `api_secret_cipher` text COMMENT 'API Secret密文',
  `private_key_cipher` longtext COMMENT '私钥密文',
  `public_key` longtext COMMENT '公钥',
  `cert_cipher` longtext COMMENT '证书密文',
  `ext_config` json DEFAULT NULL COMMENT '扩展配置',
  `enabled` tinyint NOT NULL DEFAULT 1 COMMENT '启用状态：1启用 2禁用',
  `is_default` tinyint NOT NULL DEFAULT 1 COMMENT '是否默认账号：1否 2是',
  `remark` varchar(255) DEFAULT NULL COMMENT '备注',
  `create_times` bigint NOT NULL DEFAULT 0 COMMENT '创建时间',
  `update_times` bigint NOT NULL DEFAULT 0 COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_tenant_account_code` (`tenant_id`, `account_code`),
  KEY `idx_tenant_platform` (`tenant_id`, `platform_id`),
  KEY `idx_tenant_enabled` (`tenant_id`, `enabled`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='租户支付账号表';

CREATE TABLE `t_tenant_pay_channel` (
  `id` bigint NOT NULL AUTO_INCREMENT COMMENT '通道ID',
  `tenant_id` bigint NOT NULL COMMENT '租户ID',
  `platform_id` bigint NOT NULL COMMENT '平台ID',
  `product_id` bigint NOT NULL COMMENT '产品ID',
  `account_id` bigint NOT NULL COMMENT '租户支付账号ID',
  `channel_code` varchar(64) NOT NULL COMMENT '通道编码',
  `channel_name` varchar(128) NOT NULL COMMENT '通道名称',
  `display_name` varchar(128) DEFAULT NULL COMMENT '前端展示名称',
  `icon` varchar(255) DEFAULT NULL COMMENT '图标',
  `currency` varchar(16) NOT NULL DEFAULT 'CNY' COMMENT '币种',
  `sort` int NOT NULL DEFAULT 0 COMMENT '排序',
  `visible` tinyint NOT NULL DEFAULT 1 COMMENT '显示开关：1显示 2隐藏',
  `enabled` tinyint NOT NULL DEFAULT 1 COMMENT '启用状态：1启用 2禁用',

  `single_min_amount` bigint NOT NULL DEFAULT 0 COMMENT '单笔最小金额，单位分',
  `single_max_amount` bigint NOT NULL DEFAULT 0 COMMENT '单笔最大金额，0表示不限制，单位分',
  `daily_max_amount` bigint NOT NULL DEFAULT 0 COMMENT '单日最大金额，0表示不限制，单位分',
  `daily_max_count` int NOT NULL DEFAULT 0 COMMENT '单日最大次数，0表示不限制',

  `fee_type` tinyint NOT NULL DEFAULT 1 COMMENT '手续费类型：1比例 2固定',
  `fee_rate` decimal(10,4) NOT NULL DEFAULT 0.0000 COMMENT '手续费比例',
  `fee_fixed_amount` bigint NOT NULL DEFAULT 0 COMMENT '固定手续费，单位分',

  `ext_config` json DEFAULT NULL COMMENT '扩展配置',
  `remark` varchar(255) DEFAULT NULL COMMENT '备注',
  `create_times` bigint NOT NULL DEFAULT 0 COMMENT '创建时间',
  `update_times` bigint NOT NULL DEFAULT 0 COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_tenant_channel_code` (`tenant_id`, `channel_code`),
  KEY `idx_tenant_visible_enabled` (`tenant_id`, `visible`, `enabled`),
  KEY `idx_account_id` (`account_id`),
  KEY `idx_product_id` (`product_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='租户支付通道表';

CREATE TABLE `t_tenant_pay_channel_rule` (
  `id` bigint NOT NULL AUTO_INCREMENT COMMENT '规则ID',
  `tenant_id` bigint NOT NULL COMMENT '租户ID',
  `channel_id` bigint NOT NULL COMMENT '通道ID',
  `rule_name` varchar(128) NOT NULL COMMENT '规则名称',
  `priority` int NOT NULL DEFAULT 100 COMMENT '优先级，越小越优先',
  `enabled` tinyint NOT NULL DEFAULT 1 COMMENT '启用状态：1启用 2禁用',

  `single_amount_min` bigint NOT NULL DEFAULT 0 COMMENT '单笔充值最小金额，单位分',
  `single_amount_max` bigint NOT NULL DEFAULT 0 COMMENT '单笔充值最大金额，0表示不限制，单位分',

  `user_total_recharge_min` bigint NOT NULL DEFAULT 0 COMMENT '用户累计充值最小金额，单位分',
  `user_total_recharge_max` bigint NOT NULL DEFAULT 0 COMMENT '用户累计充值最大金额，0表示不限制，单位分',

  `member_level_min` int NOT NULL DEFAULT 0 COMMENT '会员等级最小值',
  `member_level_max` int NOT NULL DEFAULT 0 COMMENT '会员等级最大值，0表示不限制',

  `kyc_level_min` tinyint NOT NULL DEFAULT 0 COMMENT 'KYC等级最小值',
  `kyc_level_max` tinyint NOT NULL DEFAULT 0 COMMENT 'KYC等级最大值，0表示不限制',

  `allow_new_user` tinyint NOT NULL DEFAULT 2 COMMENT '是否允许新用户：1否 2是',
  `allow_old_user` tinyint NOT NULL DEFAULT 2 COMMENT '是否允许老用户：1否 2是',

  `allow_tags` json DEFAULT NULL COMMENT '允许的用户标签(JSON数组)',
  `deny_tags` json DEFAULT NULL COMMENT '禁止的用户标签(JSON数组)',

  `remark` varchar(255) DEFAULT NULL COMMENT '备注',
  `create_times` bigint NOT NULL DEFAULT 0 COMMENT '创建时间',
  `update_times` bigint NOT NULL DEFAULT 0 COMMENT '更新时间',
  PRIMARY KEY (`id`),
  KEY `idx_tenant_channel_enabled` (`tenant_id`, `channel_id`, `enabled`),
  KEY `idx_tenant_priority` (`tenant_id`, `priority`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='租户支付通道显示规则表';

CREATE TABLE `t_user_recharge_stat` (
  `id` bigint NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `tenant_id` bigint NOT NULL COMMENT '租户ID',
  `user_id` bigint NOT NULL COMMENT '用户ID',
  `success_order_count` int NOT NULL DEFAULT 0 COMMENT '成功充值笔数',
  `success_total_amount` bigint NOT NULL DEFAULT 0 COMMENT '成功累计充值金额，单位分',
  `today_success_amount` bigint NOT NULL DEFAULT 0 COMMENT '今日成功充值金额，单位分',
  `today_success_count` int NOT NULL DEFAULT 0 COMMENT '今日成功充值次数',
  `first_success_time` bigint DEFAULT NULL COMMENT '首次成功充值时间',
  `last_success_time` bigint DEFAULT NULL COMMENT '最近成功充值时间',
  `create_times` bigint NOT NULL DEFAULT 0 COMMENT '创建时间',
  `update_times` bigint NOT NULL DEFAULT 0 COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_tenant_user` (`tenant_id`, `user_id`),
  KEY `idx_tenant_total_amount` (`tenant_id`, `success_total_amount`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='用户充值统计表';

CREATE TABLE `t_recharge_order` (
  `id` bigint NOT NULL AUTO_INCREMENT COMMENT '充值订单ID',
  `tenant_id` bigint NOT NULL COMMENT '租户ID',
  `user_id` bigint NOT NULL COMMENT '用户ID',
  `order_no` varchar(64) NOT NULL COMMENT '平台订单号',
  `biz_order_no` varchar(64) DEFAULT NULL COMMENT '业务订单号',
  `platform_id` bigint NOT NULL COMMENT '平台ID',
  `product_id` bigint NOT NULL COMMENT '产品ID',
  `account_id` bigint NOT NULL COMMENT '账号ID',
  `channel_id` bigint NOT NULL COMMENT '通道ID',
  `recharge_type` tinyint NOT NULL DEFAULT 0 COMMENT '充值类型：0未知 1虚拟币 2三方充值 3银行卡 4人工充值 5其他',
  `wallet_type` tinyint NOT NULL DEFAULT 1 COMMENT '钱包类型:1现金/现货 2股票/资金 3合约 4理财 5期权',
  `currency` varchar(16) NOT NULL DEFAULT 'CNY' COMMENT '币种',
  `order_amount` bigint NOT NULL COMMENT '订单金额，单位分',
  `pay_amount` bigint NOT NULL DEFAULT 0 COMMENT '实际支付金额，单位分',
  `fee_amount` bigint NOT NULL DEFAULT 0 COMMENT '手续费金额，单位分',
  `subject` varchar(255) DEFAULT NULL COMMENT '标题',
  `body` varchar(255) DEFAULT NULL COMMENT '描述',
  `client_type` tinyint NOT NULL DEFAULT 1 COMMENT '客户端类型：1APP 2H5 3WEB',
  `client_ip` varchar(64) DEFAULT NULL COMMENT '客户端IP',
  `status` tinyint NOT NULL DEFAULT 1 COMMENT '状态：1待支付 2支付中 3成功 4失败 5已关闭 6已退款',
  `third_trade_no` varchar(128) DEFAULT NULL COMMENT '三方交易号',
  `third_order_no` varchar(128) DEFAULT NULL COMMENT '三方订单号',
  `pay_url` text COMMENT '支付链接',
  `qr_content` text COMMENT '二维码内容',
  `voucher_image` varchar(128) NOT NULL DEFAULT '' COMMENT '充值凭证图片',
  `request_data` json DEFAULT NULL COMMENT '请求快照',
  `response_data` json DEFAULT NULL COMMENT '响应快照',
  `notify_data` json DEFAULT NULL COMMENT '回调数据',
  `expire_time` bigint NOT NULL DEFAULT 0 COMMENT '过期时间',
  `paid_time` bigint NOT NULL DEFAULT 0 COMMENT '支付时间',
  `notify_time` bigint NOT NULL DEFAULT 0 COMMENT '回调时间',
  `close_time` bigint NOT NULL DEFAULT 0 COMMENT '关闭时间',
  `remark` varchar(255) DEFAULT NULL COMMENT '备注',
  `create_times` bigint NOT NULL DEFAULT 0 COMMENT '创建时间',
  `update_times` bigint NOT NULL DEFAULT 0 COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_order_no` (`order_no`),
  UNIQUE KEY `uk_tenant_biz_order_no` (`tenant_id`, `biz_order_no`),
  KEY `idx_tenant_user` (`tenant_id`, `user_id`),
  KEY `idx_tenant_status` (`tenant_id`, `status`),
  KEY `idx_tenant_recharge_type` (`tenant_id`, `recharge_type`),
  KEY `idx_third_trade_no` (`third_trade_no`),
  KEY `idx_create_times` (`create_times`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='充值订单表';

CREATE TABLE `t_recharge_notify_log` (
  `id` bigint NOT NULL AUTO_INCREMENT COMMENT '回调日志ID',
  `tenant_id` bigint NOT NULL COMMENT '租户ID',
  `order_id` bigint DEFAULT NULL COMMENT '充值订单ID',
  `order_no` varchar(64) DEFAULT NULL COMMENT '平台订单号',
  `platform_id` bigint NOT NULL COMMENT '平台ID',
  `channel_id` bigint DEFAULT NULL COMMENT '通道ID',
  `notify_status` tinyint NOT NULL DEFAULT 1 COMMENT '处理状态：1待处理 2成功 3失败',
  `notify_body` longtext COMMENT '回调原文',
  `sign_result` tinyint NOT NULL DEFAULT 0 COMMENT '验签结果：1未验 2通过 3失败',
  `process_result` varchar(255) DEFAULT NULL COMMENT '处理结果',
  `error_message` varchar(1000) DEFAULT NULL COMMENT '错误信息',
  `notify_time` bigint NOT NULL DEFAULT 0 COMMENT '回调时间',
  `create_times` bigint NOT NULL DEFAULT 0 COMMENT '创建时间',
  PRIMARY KEY (`id`),
  KEY `idx_tenant_order_no` (`tenant_id`, `order_no`),
  KEY `idx_order_id` (`order_id`),
  KEY `idx_notify_status` (`notify_status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='充值回调日志表';


CREATE TABLE `t_withdraw_order` (
  `id` bigint NOT NULL AUTO_INCREMENT COMMENT '提现订单ID',
  `tenant_id` bigint NOT NULL COMMENT '租户ID',
  `user_id` bigint NOT NULL COMMENT '用户ID',
  `order_no` varchar(64) NOT NULL COMMENT '平台订单号',
  `biz_order_no` varchar(64) DEFAULT NULL COMMENT '业务订单号',
  `platform_id` bigint NOT NULL COMMENT '平台ID',
  `product_id` bigint NOT NULL COMMENT '产品ID',
  `account_id` bigint NOT NULL COMMENT '账号ID',
  `channel_id` bigint NOT NULL COMMENT '通道ID',
  `currency` varchar(16) NOT NULL DEFAULT 'CNY' COMMENT '币种',
  `amount` bigint NOT NULL COMMENT '订单金额，单位分',
  `fee_amount` bigint NOT NULL DEFAULT 0 COMMENT '手续费金额，单位分',
  `actual_amount` bigint NOT NULL DEFAULT 0 COMMENT '实际到账金额，单位分',
  `client_type` tinyint NOT NULL DEFAULT 1 COMMENT '客户端类型：1APP 2H5 3WEB',
  `client_ip` varchar(64) DEFAULT NULL COMMENT '客户端IP',
  `status` tinyint NOT NULL DEFAULT 1 COMMENT '状态：1待处理 2处理中 3成功 4失败 5已关闭',
  `third_trade_no` varchar(128) DEFAULT NULL COMMENT '三方交易号',
  `third_order_no` varchar(128) DEFAULT NULL COMMENT '三方订单号',
  `request_data` json DEFAULT NULL COMMENT '请求快照',
  `response_data` json DEFAULT NULL COMMENT '响应快照',
  `notify_data` json DEFAULT NULL COMMENT '回调数据',
  `process_time` bigint NOT NULL DEFAULT 0 COMMENT '处理时间',
  `notify_time` bigint NOT NULL DEFAULT 0 COMMENT '回调时间',
  `close_time` bigint NOT NULL DEFAULT 0 COMMENT '关闭时间',
  `remark` varchar(255) DEFAULT NULL COMMENT '备注',
  `create_times` bigint NOT NULL DEFAULT 0 COMMENT '创建时间',
  `update_times` bigint NOT NULL DEFAULT 0 COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_order_no` (`order_no`),
  UNIQUE KEY `uk_tenant_biz_order_no` (`tenant_id`, `biz_order_no`),
  KEY `idx_tenant_user` (`tenant_id`, `user_id`),
  KEY `idx_tenant_status` (`tenant_id`, `status`),
  KEY `idx_third_trade_no` (`third_trade_no`),
  KEY `idx_create_times` (`create_times`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='提现订单表';

CREATE TABLE `t_withdraw_notify_log` (
  `id` bigint NOT NULL AUTO_INCREMENT COMMENT '回调日志ID',
  `tenant_id` bigint NOT NULL COMMENT '租户ID',
  `order_id` bigint DEFAULT NULL COMMENT '提现订单ID',
  `order_no` varchar(64) DEFAULT NULL COMMENT '平台订单号',
  `platform_id` bigint NOT NULL COMMENT '平台ID',
  `channel_id` bigint DEFAULT NULL COMMENT '通道ID',
  `notify_status` tinyint NOT NULL DEFAULT 1 COMMENT '处理状态：1待处理 2成功 3失败',
  `notify_body` longtext COMMENT '回调原文',
  `sign_result` tinyint NOT NULL DEFAULT 0 COMMENT '验签结果：1未验 2通过 3失败',
  `process_result` varchar(255) DEFAULT NULL COMMENT '处理结果',
  `error_message` varchar(1000) DEFAULT NULL COMMENT '错误信息',
  `notify_time` bigint NOT NULL DEFAULT 0 COMMENT '回调时间',
  `create_times` bigint NOT NULL DEFAULT 0 COMMENT '创建时间',
  PRIMARY KEY (`id`),
  KEY `idx_tenant_order_no` (`tenant_id`, `order_no`),
  KEY `idx_order_id` (`order_id`),
  KEY `idx_notify_status` (`notify_status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='提现回调日志表'; 

CREATE TABLE `t_crypto_recharge_address` (
  `id` bigint NOT NULL AUTO_INCREMENT COMMENT '主键ID',

  `tenant_id` bigint NOT NULL DEFAULT 0 COMMENT '租户ID',
  `user_id` bigint NOT NULL DEFAULT 0 COMMENT '用户ID',

  `wallet_type` tinyint NOT NULL DEFAULT 1 COMMENT '账户类型:1现金/现货 2股票/资金 3合约 4理财 5期权',

  `coin` varchar(20) NOT NULL DEFAULT '' COMMENT '币种:USDT/BTC/ETH',
  `chain_code` tinyint NOT NULL DEFAULT 0 COMMENT '链类型',

  `address` varchar(255) NOT NULL DEFAULT '' COMMENT '充值地址',
  `memo` varchar(128) NOT NULL DEFAULT '' COMMENT 'memo/tag，如XRP/EOS/TON等',

  `address_source` tinyint NOT NULL DEFAULT 1 COMMENT '地址来源:1系统生成 2第三方分配 3手工导入',
  `address_type` tinyint NOT NULL DEFAULT 1 COMMENT '地址类型:1用户独享 2平台公共地址+memo',

  `status` tinyint NOT NULL DEFAULT 2 COMMENT '地址状态:1禁用 2可用 3冻结',

  `last_used_time` bigint NOT NULL DEFAULT 0 COMMENT '最近使用时间',
  `create_times` bigint NOT NULL DEFAULT 0 COMMENT '创建时间',
  `update_times` bigint NOT NULL DEFAULT 0 COMMENT '更新时间',

  PRIMARY KEY (`id`),

  UNIQUE KEY `uk_tenant_user_coin_chain` (`tenant_id`, `user_id`, `wallet_type`, `coin`, `chain_code`),
  UNIQUE KEY `uk_chain_address_memo` (`chain_code`, `address`, `memo`),

  KEY `idx_tenant_user` (`tenant_id`, `user_id`),
  KEY `idx_tenant_coin_chain` (`tenant_id`, `coin`, `chain_code`),
  KEY `idx_address` (`address`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='用户链上充值地址表';

CREATE TABLE `t_crypto_wallet_account` (
  `id` bigint NOT NULL AUTO_INCREMENT COMMENT '主键ID',

  `tenant_id` bigint NOT NULL DEFAULT 0 COMMENT '租户ID',
  `account_code` varchar(64) NOT NULL DEFAULT '' COMMENT '钱包账号编码',
  `account_name` varchar(128) NOT NULL DEFAULT '' COMMENT '钱包账号名称',

  `provider` varchar(64) NOT NULL DEFAULT '' COMMENT '钱包服务商:self/cobo/bitgo/fireblocks等',
  `api_key_cipher` text COMMENT 'API Key密文',
  `api_secret_cipher` text COMMENT 'API Secret密文',
  `callback_secret_cipher` text COMMENT '回调验签密钥密文',

  `ext_config` json DEFAULT NULL COMMENT '扩展配置',

  `enabled` tinyint NOT NULL DEFAULT 1 COMMENT '启用状态:1启用 2禁用',
  `is_default` tinyint NOT NULL DEFAULT 1 COMMENT '是否默认:1否 2是',

  `create_times` bigint NOT NULL DEFAULT 0,
  `update_times` bigint NOT NULL DEFAULT 0,

  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_tenant_account_code` (`tenant_id`, `account_code`),
  KEY `idx_tenant_enabled` (`tenant_id`, `enabled`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='链上钱包账号配置表';

CREATE TABLE `t_crypto_recharge_tx` (
  `id` bigint NOT NULL AUTO_INCREMENT COMMENT '主键ID',

  `tenant_id` bigint NOT NULL DEFAULT 0 COMMENT '租户ID',
  `user_id` bigint NOT NULL DEFAULT 0 COMMENT '用户ID',

  `order_id` bigint NOT NULL DEFAULT 0 COMMENT '充值订单ID',
  `order_no` varchar(64) NOT NULL DEFAULT '' COMMENT '充值订单号',

  `coin` varchar(20) NOT NULL DEFAULT '' COMMENT '币种',
  `chain_code` tinyint NOT NULL DEFAULT 0 COMMENT '链类型',

  `tx_hash` varchar(128) NOT NULL DEFAULT '' COMMENT '交易哈希',
  `from_address` varchar(255) NOT NULL DEFAULT '' COMMENT '付款地址',
  `to_address` varchar(255) NOT NULL DEFAULT '' COMMENT '收款地址',
  `memo` varchar(128) NOT NULL DEFAULT '' COMMENT 'memo/tag',

  `amount` decimal(36,18) NOT NULL DEFAULT 0 COMMENT '链上到账数量',
  `block_height` bigint NOT NULL DEFAULT 0 COMMENT '区块高度',
  `confirm_count` int NOT NULL DEFAULT 0 COMMENT '当前确认数',
  `required_confirm_count` int NOT NULL DEFAULT 0 COMMENT '要求确认数',

  `status` tinyint NOT NULL DEFAULT 1 COMMENT '状态:1待确认 2确认中 3已确认 4失败 5已入账',

  `raw_data` json DEFAULT NULL COMMENT '链上原始数据',

  `create_times` bigint NOT NULL DEFAULT 0,
  `update_times` bigint NOT NULL DEFAULT 0,

  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_chain_tx` (`chain_code`, `tx_hash`),
  KEY `idx_tenant_user` (`tenant_id`, `user_id`),
  KEY `idx_order_no` (`order_no`),
  KEY `idx_to_address` (`to_address`),
  KEY `idx_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='链上充值交易记录表';
