INSERT INTO `sys_role` (`id`, `name`, `code`, `status`, `remark`)
VALUES
	('1', '超级管理员', 'admin', '1', '');
INSERT INTO `sys_user` (`id`, `username`, `password`, `nickname`, `avatar`, `status`, `google_secret`, `google_enabled`, `perms_ver`, `last_login_ip`, `last_login_at`)
VALUES
	('1', 'admin', '$2a$10$KdJbtCoUCeO.jcI9LJb6me4YAnMt8JScsCWyA9FEPfuaz4bRCfMee', '超级管理员', '', '1', '', '0', '1', NULL, NULL);
INSERT INTO `sys_user_role` (`user_id`, `role_id`)
VALUES
	('1', '1');


INSERT INTO `sys_role_menu` (`role_id`, `menu_id`) VALUES
(1, 10),
(1, 11),
(1, 12),
(1, 13),
(1, 14),
(1, 15),
(1, 16),
(1, 17),
(1, 18),
(1, 19),
(1, 20),
(1, 21),
(1, 22),
(1, 23),
(1, 24),
(1, 25),
(1, 40),
(1, 41),
(1, 60),
(1, 61),
(1, 62),
(1, 63),
(1, 64),
(1, 65),
(1, 66),
(1, 80),

(1, 100),
(1, 101),
(1, 102),
(1, 103),
(1, 104),
(1, 105),
(1, 106),
(1, 120),
(1, 121),
(1, 122),
(1, 123),
(1, 124),
(1, 140),
(1, 141),
(1, 142),
(1, 143),
(1, 144),
(1, 160),
(1, 161),
(1, 162),
(1, 163),
(1, 164),
(1, 180),
(1, 181),
(1, 182),
(1, 183),
(1, 184),
(1, 200),
(1, 201),
(1, 202),
(1, 203),
(1, 220),
(1, 221),
(1, 230),
(1, 231),
(1, 232),
(1, 233),
(1, 234),
(1, 250),
(1, 251),
(1, 260),
(1, 261),
(1, 262),
(1, 280),
(1, 281),
(1, 290),

(1, 300),
(1, 301),
(1, 302),
(1, 303),
(1, 304),
(1, 305),
(1, 306),
(1, 320),
(1, 321),
(1, 322),
(1, 323),
(1, 324),
(1, 330),
(1, 331),
(1, 332),
(1, 333),
(1, 334),
(1, 350),
(1, 351),
(1, 352),
(1, 353),
(1, 354),
(1, 370),
(1, 380),

(1, 400),
(1, 401),
(1, 410),
(1, 411),
(1, 412),
(1, 413),
(1, 420),
(1, 430),
(1, 431),
(1, 432),
(1, 440),
(1, 441),
(1, 442),

(1, 600),
(1, 601),
(1, 610),
(1, 611),
(1, 612),
(1, 613),
(1, 620),
(1, 621),
(1, 630),
(1, 640),
(1, 641),
(1, 650),
(1, 651),
(1, 660),
(1, 661),
(1, 670),
(1, 671),
(1, 680),
(1, 681),
(1, 690),
(1, 691),
(1, 700),
(1, 701),

(1, 800),
(1, 801),
(1, 810),
(1, 811),
(1, 812),
(1, 813),
(1, 814),
(1, 820),
(1, 821),
(1, 830),
(1, 831),
(1, 840),
(1, 841),

(1, 1000),
(1, 1001),
(1, 1010),
(1, 1011),
(1, 1012),
(1, 1013),
(1, 1014),
(1, 1015),
(1, 1020),
(1, 1021),
(1, 1030),
(1, 1031),
(1, 1040),
(1, 1041),
(1, 1050),
(1, 1060),
(1, 1070),
(1, 1080),
(1, 1081),
(1, 1090),
(1, 1091),
(1, 1100),
(1, 1101),
(1, 1110),
(1, 1120),
(1, 1121),
(1, 1130),
(1, 1131),
(1, 1132),
(1, 10000),
(1, 10100),
(1, 10101),
(1, 10102),
(1, 10103),
(1, 10104),
(1, 10105),
(1, 10106),
(1, 10107),
(1, 10108),
(1, 10109),
(1, 10110),
(1, 10111),
(1, 10112),
(1, 10113),

(1, 10200),
(1, 10201),
(1, 10202),
(1, 10203),
(1, 10204),
(1, 10205),
(1, 10206),

(1, 10300),
(1, 10301),
(1, 10302),
(1, 10303),
(1, 10304),

(1, 10400),
(1, 10401),
(1, 10402),
(1, 10403),

(1, 10500),
(1, 10600),
(1, 10601),
(1, 10602),
(1, 10603),
(1, 10604),
(1, 10605),
(1, 10606),
(1, 10607),

(1, 10700),
(1, 10800),
(1, 10900),

(1, 11000),
(1, 11001),
(1, 11002),
(1, 11003),
(1, 11004),
(1, 11005);



INSERT INTO sys_menu (id, parent_id, name, menu_type, icon, sort)
VALUES (10, 0, '用户管理', 1, 'Users', 10);
INSERT INTO sys_menu (id, parent_id, name, menu_type, method, path, perms, component, icon, sort)
VALUES 
(11, 10, '用户列表', 2, 'GET', '/member/users', 'users:user:list', 'users/user', 'User', 11),
(12, 11, '创建用户', 3, 'POST', '/member/users', 'users:user:add', '', '', 12),
(13, 11, '获取用户详情', 3, 'GET', '/member/users/{id}', 'users:user:detail', '', '', 13),
(14, 11, '更新用户基本信息', 3, 'PUT', '/member/users/{id}/base', 'users:user:update', '', '', 14),
(15, 11, '更新用户状态', 3, 'PUT', '/member/users/{id}/status', 'users:user:update:status', '', '', 15),
(16, 11, '更新用户会员等级', 3, 'PUT', '/member/users/{id}/level', 'users:user:update:level', '', '', 16),
(17, 11, '重置登录密码', 3, 'PUT', '/member/users/{id}/reset-login-password', 'users:user:reset:loginpwd', '', '', 17),
(18, 11, '重置支付密码', 3, 'PUT', '/member/users/{id}/reset-pay-password', 'users:user:reset:paypwd', '', '', 18),
(19, 11, '解锁用户', 3, 'PUT', '/member/users/{id}/unlock', 'users:user:unlock', '', '', 19),
(20, 11, '更新用户风险等级', 3, 'PUT', '/member/users/{id}/risk-level', 'users:user:update:risklevel', '', '', 20),
(21, 11, '删除用户', 3, 'DELETE', '/member/users/{id}', 'users:user:delete', '', '', 21),
(22, 11, '获取用户安全设置', 3, 'GET', '/member/users/{id}/security', 'users:user:security:detail', '', '', 22),
(23, 11, '重置用户谷歌2FA', 3, 'PUT', '/member/users/{id}/reset2fa', 'users:user:reset:google2fa', '', '', 23),
(24, 11, '获取用户选项', 3, 'GET', '/member/options', 'users:user:options', '', '', 24),
(25, 11, '校验推荐人', 3, 'GET', '/member/users/referrer/check', 'users:user:referrer:check', '', '', 25),

(40, 10, '实名认证信息列表', 2, 'GET', '/member/user-identities', 'users:user:identities:list', 'users/identity', '', 40),
(41, 40, '审核实名认证信息', 3, 'PUT', '/member/user-identities/{id}/review', 'users:user:identities:review', '', '', 41),

(60, 10, '用户银行卡列表', 2, 'GET', '/member/user-banks', 'users:user:banks:list', 'users/bank', '', 60),
(61, 60, '获取用户银行卡详情', 3, 'GET', '/member/user-banks/{id}', 'users:user:bank:detail', '', '', 61),
(62, 60, '添加用户银行卡', 3, 'POST', '/member/user-banks', 'users:user:bank:add', '', '', 62),
(63, 60, '更新用户银行卡', 3, 'PUT', '/member/user-banks/{id}', 'users:user:bank:update', '', '', 63),
(64, 60, '删除用户银行卡', 3, 'DELETE', '/member/user-banks/{id}', 'users:user:bank:delete', '', '', 64),
(65, 60, '更新用户银行卡状态', 3, 'PUT', '/member/user-banks/{id}/status', 'users:user:bank:update:status', '', '', 65),
(66, 60, '设置默认用户银行卡', 3, 'PUT', '/member/user-banks/{id}/default', 'users:user:bank:setdefault', '', '', 66),
(80, 10, '会员服务选项', 3, 'GET', '/member/options', 'users:options', '', '', 80);


INSERT INTO sys_menu (id, parent_id, name, menu_type, icon, sort)
VALUES (100, 0, '支付管理', 1, 'Payment', 100);

INSERT INTO sys_menu (id, parent_id, name, menu_type, method, path, perms, component, icon, sort)
VALUES
(101, 100, '平台管理', 2, 'GET', '/payment/platforms', 'payment:platform:list', 'payment/platforms', 'Bank', 101),
(102, 101, '创建支付平台', 3, 'POST', '/payment/platform', 'payment:platform:add', '', '', 102),
(103, 101, '更新支付平台', 3, 'PUT', '/payment/platform', 'payment:platform:update', '', '', 103),
(104, 101, '获取支付平台详情', 3, 'GET', '/payment/platform', 'payment:platform:detail', '', '', 104),
(105, 101, '删除支付平台', 3, 'DELETE', '/payment/platform/{id}', 'payment:platform:delete', '', '', 105),
(106, 101, '平台选项列表', 3, 'GET', '/payment/platforms/options', 'payment:platform:options', '', '', 106),

(120, 100, '产品管理', 2, 'GET', '/payment/products', 'payment:product:list', 'payment/products', 'Gold', 120),
(121, 120, '创建支付产品', 3, 'POST', '/payment/product', 'payment:product:add', '', '', 121),
(122, 120, '更新支付产品', 3, 'PUT', '/payment/product', 'payment:product:update', '', '', 122),
(123, 120, '获取支付产品详情', 3, 'GET', '/payment/product', 'payment:product:detail', '', '', 123),
(124, 120, '删除支付产品', 3, 'DELETE', '/payment/product/{id}', 'payment:product:delete', '', '', 124),

(140, 100, '租户平台', 2, 'GET', '/payment/tenant-platforms', 'payment:tenantplatform:list', 'payment/tenant-platforms', 'Gold', 140),
(141, 140, '开通租户平台', 3, 'POST', '/payment/tenant-platform', 'payment:tenantplatform:add', '', '', 141),
(142, 140, '更新租户平台', 3, 'PUT', '/payment/tenant-platform', 'payment:tenantplatform:update', '', '', 142),
(143, 140, '获取租户平台详情', 3, 'GET', '/payment/tenant-platform', 'payment:tenantplatform:detail', '', '', 143),
(144, 140, '删除租户平台', 3, 'DELETE', '/payment/tenant-platform/{id}', 'payment:tenantplatform:delete', '', '', 144),

(160, 100, '租户支付账号', 2, 'GET', '/payment/tenant-accounts', 'payment:tenantaccount:list', 'payment/tenant-accounts', 'Gold', 160),
(161, 160, '创建租户支付账号', 3, 'POST', '/payment/tenant-account', 'payment:tenantaccount:add', '', '', 161),
(162, 160, '更新租户支付账号', 3, 'PUT', '/payment/tenant-account', 'payment:tenantaccount:update', '', '', 162),
(163, 160, '获取租户支付账号详情', 3, 'GET', '/payment/tenant-account', 'payment:tenantaccount:detail', '', '', 163),
(164, 160, '删除租户支付账号', 3, 'DELETE', '/payment/tenant-account/{id}', 'payment:tenantaccount:delete', '', '', 164),

(180, 100, '租户支付通道', 2, 'GET', '/payment/tenant-channels', 'payment:tenantchannel:list', 'payment/tenant-channels', 'Gold', 180),
(181, 180, '创建租户支付通道', 3, 'POST', '/payment/tenant-channel', 'payment:tenantchannel:add', '', '', 181),
(182, 180, '更新租户支付通道', 3, 'PUT', '/payment/tenant-channel', 'payment:tenantchannel:update', '', '', 182),
(183, 180, '获取租户支付通道详情', 3, 'GET', '/payment/tenant-channel', 'payment:tenantchannel:detail', '', '', 183),
(184, 180, '删除租户支付通道', 3, 'DELETE', '/payment/tenant-channel/{id}', 'payment:tenantchannel:delete', '', '', 184),

(200, 100, '租户支付通道规则', 2, 'GET', '/payment/tenant-channel-rules', 'payment:channelrule:list', 'payment/tenant-channel-rules', 'Gold', 200),
(201, 200, '创建支付通道规则', 3, 'POST', '/payment/tenant-channel-rule', 'payment:channelrule:add', '', '', 201),
(202, 200, '更新支付通道规则', 3, 'PUT', '/payment/tenant-channel-rule', 'payment:channelrule:update', '', '', 202),
(203, 200, '删除支付通道规则', 3, 'DELETE', '/payment/tenant-channel-rule/{id}', 'payment:channelrule:delete', '', '', 203),

(220, 100, '用户充值统计', 2, 'GET', '/payment/user-recharge-stats', 'payment:userrechargestat:list', 'payment/user-recharge-stats', 'Gold', 220),
(221, 220, '用户充值统计详情', 3, 'GET', '/payment/user-recharge-stat', 'payment:userrechargestat:detail', '', '', 221),

(230, 100, '充值管理', 2, 'GET', '/payment/recharge-orders', 'payment:recharge-order:list', 'payment/recharge-orders', 'Gold', 230),
(231, 230, '获取充值订单详情', 3, 'GET', '/payment/recharge-order/{orderNo}', 'payment:recharge-order:detail', '', '', 231),
(232, 230, '关闭充值订单', 3, 'POST', '/payment/recharge-order/{orderNo}/close', 'payment:recharge-order:close', '', '', 232),
(233, 230, '人工标记充值订单支付成功', 3, 'POST', '/payment/recharge-order/{orderNo}/manual-success', 'payment:recharge-order:manualsuccess', '', '', 233),
(234, 230, '重试充值回调', 3, 'POST', '/payment/recharge-order/{orderNo}/retry-notify', 'payment:recharge-order:retrynotify', '', '', 234),

(250, 100, '充值回调日志', 2, 'GET', '/payment/recharge-notify-logs', 'payment:recharge-notifylog:list', 'payment/recharge-notify-logs', 'Gold', 250),
(251, 250, '获取充值回调日志详情', 3, 'GET', '/payment/recharge-notify-log/{id}', 'payment:recharge-notifylog:detail', '', '', 251),

(260, 100, '提现管理', 2, 'GET', '/payment/withdraw-orders', 'payment:withdraw-order:list', 'payment/withdraw-orders', 'Gold', 260),
(261, 260, '获取提现订单详情', 3, 'GET', '/payment/withdraw-order/{id}', 'payment:withdraw-order:detail', '', '', 261),
(262, 260, '审核提现订单', 3, 'POST', '/payment/withdraw-order/{orderNo}/audit', 'payment:withdraw-order:audit', '', '', 262),

(280, 100, '提现回调日志', 2, 'GET', '/payment/withdraw-notify-logs', 'payment:withdraw-notifylog:list', 'payment/withdraw-notify-logs', 'Gold', 280),
(281, 280, '获取提现回调日志详情', 3, 'GET', '/payment/withdraw-notify-log/{id}', 'payment:withdraw-notifylog:detail', '', '', 281),
(290, 100, '支付服务选项', 3, 'GET', '/payment/options', 'payment:options', '', '', 290);



INSERT INTO sys_menu (id, parent_id, name, menu_type, icon, sort)
VALUES (300, 0, 'ITICK数据管理', 1, 'Goods', 300);

INSERT INTO sys_menu (id, parent_id, name, menu_type, method, path, perms, component, icon, sort)
VALUES
-- 产品类型管理
(301, 300, '产品类型列表', 2, 'GET', '/itick/categories', 'itick:category:list', 'itick/categories', 'Menu', 301),
(302, 301, '创建产品类型', 3, 'POST', '/itick/categories', 'itick:category:add', '', '', 302),
(303, 301, '更新产品类型', 3, 'PUT', '/itick/categories', 'itick:category:update', '', '', 303),
(304, 301, '获取产品类型详情', 3, 'GET', '/itick/categories/{id}', 'itick:category:detail', '', '', 304),
(305, 301, '同步类型下的产品', 3, 'POST', '/itick/categories/sync-products', 'itick:category:syncProducts', '', '', 305),
(306, 301, '同步任务状态', 3, 'GET', '/itick/sync-tasks/{taskNo}/status', 'itick:sync-task:status', '', '', 306),

-- 产品管理 320
(320, 300, '产品列表', 2, 'GET', '/itick/products', 'itick:product:list', 'itick/products', 'Goods', 320),
(321, 320, '创建产品', 3, 'POST', '/itick/products', 'itick:product:add', '', '', 321),
(322, 320, '更新产品', 3, 'PUT', '/itick/products', 'itick:product:update', '', '', 322),
(323, 320, '获取产品详情', 3, 'GET', '/itick/products/{id}', 'itick:product:detail', '', '', 323),
(324, 320, 'K线查看', 3, 'GET', '/itick/kline', 'itick:kline:view', '', '', 324),

-- 租户产品类型管理 330
(330, 300, '租户产品类型列表', 2, 'GET', '/itick/tenant-categories', 'itick:tenant-category:list', 'itick/tenant-categories', 'OfficeBuilding', 330),
(331, 330, '创建租户产品类型', 3, 'POST', '/itick/tenant-categories', 'itick:tenant-category:add', '', '', 331),
(332, 330, '更新租户产品类型', 3, 'PUT', '/itick/tenant-categories', 'itick:tenant-category:update', '', '', 332),
(333, 330, '批量更新租户产品类型', 3, 'POST', '/itick/tenant-categories/batch', 'itick:tenant-category:batchUpsert', '', '', 333),
(334, 330, '获取租户产品类型详情', 3, 'GET', '/itick/tenant-categories/{id}', 'itick:tenant-category:detail', '', '', 334),

-- 租户产品管理 350
(350, 300, '租户产品列表', 2, 'GET', '/itick/tenant-products', 'itick:tenant-itick:list', 'itick/tenant-products', 'Grid', 350),
(351, 350, '创建租户产品', 3, 'POST', '/itick/tenant-products', 'itick:tenant-itick:add', '', '', 351),
(352, 350, '更新租户产品', 3, 'PUT', '/itick/tenant-products', 'itick:tenant-itick:update', '', '', 352),
(353, 350, '批量更新租户产品', 3, 'POST', '/itick/tenant-products/batch', 'itick:tenant-itick:batchUpsert', '', '', 353),
(354, 350, '获取租户产品详情', 3, 'GET', '/itick/tenant-products/{id}', 'itick:tenant-itick:detail', '', '', 354),

-- 初始化租户展示配置
(370, 300, '初始化租户展示配置', 2, 'POST', '/itick/tenant-display/init', 'itick:tenant-display:init', 'itick/tenant-display-init', 'Setting', 370),
(380, 300, 'ITICK服务选项', 3, 'GET', '/itick/options', 'itick:options', '', '', 371);


-- 资产（asset）
INSERT INTO sys_menu (id, parent_id, name, menu_type, icon, sort)
VALUES (400, 0, '资产管理', 1, 'Wallet', 400);

INSERT INTO sys_menu (id, parent_id, name, menu_type, method, path, perms, component, icon, sort)
VALUES
(401, 400, '资产服务选项', 3, 'GET', '/asset/options', 'asset:options', '', '', 401),
(410, 400, '用户资产列表', 2, 'GET', '/asset/user-assets', 'asset:user-asset:list', 'asset/user-assets', 'Wallet', 410),
(411, 410, '获取用户资产详情', 3, 'GET', '/asset/user-assets/detail', 'asset:user-asset:detail', '', '', 411),
(412, 410, '管理员加资产', 3, 'POST', '/asset/add', 'asset:user-asset:add', '', '', 412),
(413, 410, '管理员减资产', 3, 'POST', '/asset/sub', 'asset:user-asset:sub', '', '', 413),
(420, 400, '资产流水列表', 2, 'GET', '/asset/flows', 'asset:flow:list', 'asset/flows', 'Tickets', 420),
(430, 400, '资产冻结列表', 2, 'GET', '/asset/freezes', 'asset:freeze:list', 'asset/freezes', 'Lock', 430),
(431, 430, '管理员冻结资产', 3, 'POST', '/asset/freeze', 'asset:freeze:add', '', '', 431),
(432, 430, '管理员解冻资产', 3, 'POST', '/asset/unfreeze', 'asset:freeze:unfreeze', '', '', 432),
(440, 400, '资产锁定列表', 2, 'GET', '/asset/locks', 'asset:lock:list', 'asset/locks', 'Lock', 440),
(441, 440, '管理员锁定资产', 3, 'POST', '/asset/lock', 'asset:lock:add', '', '', 441),
(442, 440, '管理员解锁资产', 3, 'POST', '/asset/unlock', 'asset:lock:unlock', '', '', 442);

-- 期权（option）
INSERT INTO sys_menu (id, parent_id, name, menu_type, icon, sort)
VALUES (600, 0, '期权管理', 1, 'TrendCharts', 600);

INSERT INTO sys_menu (id, parent_id, name, menu_type, method, path, perms, component, icon, sort)
VALUES
(601, 600, '期权服务选项', 3, 'GET', '/option/options', 'option:options', '', '', 601),
(610, 600, '合约列表', 2, 'GET', '/option/contracts', 'option:contract:list', 'option/contracts', 'Tickets', 610),
(611, 610, '创建合约', 3, 'POST', '/option/contracts', 'option:contract:add', '', '', 611),
(612, 610, '更新合约', 3, 'POST', '/option/contracts/update', 'option:contract:update', '', '', 612),
(613, 610, '获取合约详情', 3, 'GET', '/option/contracts/detail', 'option:contract:detail', '', '', 613),
(620, 600, '行情详情', 2, 'GET', '/option/market/detail', 'option:market:detail', 'option/market-detail', 'TrendCharts', 620),
(621, 620, '更新行情', 3, 'POST', '/option/market/update', 'option:market:update', '', '', 621),
(630, 600, '行情快照列表', 2, 'GET', '/option/market/snapshots', 'option:market-snapshot:list', 'option/market-snapshots', 'Histogram', 630),
(640, 600, '订单列表', 2, 'GET', '/option/orders', 'option:order:list', 'option/orders', 'List', 640),
(641, 640, '获取订单详情', 3, 'GET', '/option/orders/detail', 'option:order:detail', '', '', 641),
(650, 600, '成交列表', 2, 'GET', '/option/trades', 'option:trade:list', 'option/trades', 'DataLine', 650),
(651, 650, '获取成交详情', 3, 'GET', '/option/trades/detail', 'option:trade:detail', '', '', 651),
(660, 600, '持仓列表', 2, 'GET', '/option/positions', 'option:position:list', 'option/positions', 'PieChart', 660),
(661, 660, '获取持仓详情', 3, 'GET', '/option/positions/detail', 'option:position:detail', '', '', 661),
(670, 600, '行权列表', 2, 'GET', '/option/exercises', 'option:exercise:list', 'option/exercises', 'Operation', 670),
(671, 670, '获取行权详情', 3, 'GET', '/option/exercises/detail', 'option:exercise:detail', '', '', 671),
(680, 600, '结算列表', 2, 'GET', '/option/settlements', 'option:settlement:list', 'option/settlements', 'Checked', 680),
(681, 680, '获取结算详情', 3, 'GET', '/option/settlements/detail', 'option:settlement:detail', '', '', 681),
(690, 600, '账户列表', 2, 'GET', '/option/accounts', 'option:account:list', 'option/accounts', 'Avatar', 690),
(691, 690, '获取账户详情', 3, 'GET', '/option/accounts/detail', 'option:account:detail', '', '', 691),
(700, 600, '账单列表', 2, 'GET', '/option/bills', 'option:bill:list', 'option/bills', 'Document', 700),
(701, 700, '获取账单详情', 3, 'GET', '/option/bills/detail', 'option:bill:detail', '', '', 701);

-- 质押（staking）
INSERT INTO sys_menu (id, parent_id, name, menu_type, icon, sort)
VALUES (800, 0, '质押管理', 1, 'Coin', 800);

INSERT INTO sys_menu (id, parent_id, name, menu_type, method, path, perms, component, icon, sort)
VALUES
(801, 800, '质押服务选项', 3, 'GET', '/staking/options', 'staking:options', '', '', 801),
(810, 800, '质押产品列表', 2, 'GET', '/staking/products', 'staking:product:list', 'staking/products', 'Coin', 810),
(811, 810, '获取质押产品详情', 3, 'GET', '/staking/products/detail', 'staking:product:detail', '', '', 811),
(812, 810, '创建质押产品', 3, 'POST', '/staking/products', 'staking:product:add', '', '', 812),
(813, 810, '更新质押产品', 3, 'POST', '/staking/products/update', 'staking:product:update', '', '', 813),
(814, 810, '更新质押产品状态', 3, 'POST', '/staking/products/status', 'staking:product:update:status', '', '', 814),
(820, 800, '质押订单列表', 2, 'GET', '/staking/orders', 'staking:order:list', 'staking/orders', 'List', 820),
(821, 820, '获取质押订单详情', 3, 'GET', '/staking/orders/detail', 'staking:order:detail', '', '', 821),
(830, 800, '奖励记录列表', 2, 'GET', '/staking/reward-logs', 'staking:reward-log:list', 'staking/reward-logs', 'Medal', 830),
(831, 830, '手动发放奖励', 3, 'POST', '/staking/manual-reward', 'staking:reward-log:manual', '', '', 831),
(840, 800, '赎回记录列表', 2, 'GET', '/staking/redeem-logs', 'staking:redeem-log:list', 'staking/redeem-logs', 'RefreshLeft', 840),
(841, 840, '手动赎回', 3, 'POST', '/staking/manual-redeem', 'staking:redeem-log:manual', '', '', 841);

-- 币币交易（trade）
INSERT INTO sys_menu (id, parent_id, name, menu_type, icon, sort)
VALUES (1000, 0, '交易管理', 1, 'DataBoard', 1000);

INSERT INTO sys_menu (id, parent_id, name, menu_type, method, path, perms, component, icon, sort)
VALUES
(1001, 1000, '交易服务选项', 3, 'GET', '/trade/options', 'trade:options', '', '', 1001),
(1010, 1000, '交易对列表', 2, 'GET', '/trade/symbols', 'trade:symbol:list', 'trade/symbols', 'Switch', 1010),
(1011, 1010, '获取交易对详情', 3, 'GET', '/trade/symbols/detail', 'trade:symbol:detail', '', '', 1011),
(1012, 1010, '创建交易对', 3, 'POST', '/trade/symbols', 'trade:symbol:add', '', '', 1012),
(1013, 1010, '更新交易对', 3, 'POST', '/trade/symbols/update', 'trade:symbol:update', '', '', 1013),
(1014, 1010, '设置现货交易对配置', 3, 'POST', '/trade/symbols/spot-config', 'trade:symbol:spot-config', '', '', 1014),
(1015, 1010, '设置合约交易对配置', 3, 'POST', '/trade/symbols/contract-config', 'trade:symbol:contract-config', '', '', 1015),
(1020, 1000, '订单列表', 2, 'GET', '/trade/orders', 'trade:order:list', 'trade/orders', 'List', 1020),
(1021, 1020, '获取订单详情', 3, 'GET', '/trade/orders/detail', 'trade:order:detail', '', '', 1021),
(1030, 1000, '成交明细列表', 2, 'GET', '/trade/fills', 'trade:fill:list', 'trade/fills', 'DataLine', 1030),
(1031, 1030, '获取成交明细详情', 3, 'GET', '/trade/fills/detail', 'trade:fill:detail', '', '', 1031),
(1040, 1000, '持仓列表', 2, 'GET', '/trade/positions', 'trade:position:list', 'trade/positions', 'PieChart', 1040),
(1041, 1040, '获取持仓详情', 3, 'GET', '/trade/positions/detail', 'trade:position:detail', '', '', 1041),
(1050, 1000, '持仓历史列表', 2, 'GET', '/trade/position-histories', 'trade:position-history:list', 'trade/position-histories', 'Histogram', 1050),
(1060, 1000, '保证金账户列表', 2, 'GET', '/trade/margin-accounts', 'trade:margin-account:list', 'trade/margin-accounts', 'Wallet', 1060),
(1070, 1000, '撤单日志列表', 2, 'GET', '/trade/cancel-logs', 'trade:cancel-log:list', 'trade/cancel-logs', 'DocumentDelete', 1070),
(1080, 1000, '用户交易限制', 2, 'GET', '/trade/user-trade-limit', 'trade:user-trade-limit:detail', 'trade/user-trade-limit', 'Warning', 1080),
(1081, 1080, '设置用户交易限制', 3, 'POST', '/trade/user-trade-limit', 'trade:user-trade-limit:update', '', '', 1081),
(1090, 1000, '用户交易对限制', 2, 'GET', '/trade/user-symbol-limit', 'trade:user-symbol-limit:detail', 'trade/user-symbol-limit', 'WarningFilled', 1090),
(1091, 1090, '设置用户交易对限制', 3, 'POST', '/trade/user-symbol-limit', 'trade:user-symbol-limit:update', '', '', 1091),
(1100, 1000, '用户交易配置', 2, 'GET', '/trade/user-trade-config', 'trade:user-trade-config:detail', 'trade/user-trade-config', 'Tools', 1100),
(1101, 1100, '设置用户交易配置', 3, 'POST', '/trade/user-trade-config', 'trade:user-trade-config:update', '', '', 1101),
(1110, 1000, '风控校验日志列表', 2, 'GET', '/trade/risk-order-check-logs', 'trade:risk-order-check-log:list', 'trade/risk-order-check-logs', 'Memo', 1110),
(1120, 1000, '用户杠杆配置', 2, 'GET', '/trade/user-leverage-config', 'trade:user-leverage-config:detail', 'trade/user-leverage-config', 'TrendCharts', 1120),
(1121, 1120, '设置用户杠杆配置', 3, 'POST', '/trade/user-leverage-config', 'trade:user-leverage-config:update', '', '', 1121),
(1130, 1000, '交易事件列表', 2, 'GET', '/trade/events', 'trade:event:list', 'trade/events', 'Bell', 1130),
(1131, 1130, '获取交易事件详情', 3, 'GET', '/trade/events/detail', 'trade:event:detail', '', '', 1131),
(1132, 1130, '重试交易事件', 3, 'POST', '/trade/events/retry', 'trade:event:retry', '', '', 1132);



INSERT INTO sys_menu (id, parent_id, name, menu_type, icon, sort)
VALUES (10000, 0, '系统管理', 1, 'Setting', 10000);

INSERT INTO sys_menu (id, parent_id, name, menu_type, method, path, perms, component, icon, sort)
VALUES (10100, 10000, '用户管理', 2, 'GET', '/system/users', 'sys:user:list', 'system/users', 'User', 10100);

INSERT INTO sys_menu (id, parent_id, name, menu_type, method, path, perms, sort)
VALUES
(10101, 10100, '新增用户', 3, 'POST', '/system/users', 'sys:user:add', 10101),
(10102, 10100, '编辑用户', 3, 'PUT', '/system/users', 'sys:user:update', 10102),
(10103, 10100, '删除用户', 3, 'DELETE', '/system/users', 'sys:user:delete', 10103),
(10104, 10100, '重置密码', 3, 'POST', '/system/users/resetPwd', 'sys:user:resetpwd', 10104),
(10105, 10100, '分配角色', 3, 'POST', '/system/users/assignRoles', 'sys:user:assignrole', 10105),
(10106, 10100, '用户详情', 3, 'GET', '/system/users/{id}', 'sys:user:detail', 10106),
(10107, 10100, '修改用户状态', 3, 'POST', '/system/users/status', 'sys:user:status', 10107),
(10108, 10100, 'Google2FA管理', 3, 'GET', '/system/users/google2fa', 'sys:user:google2fa', 10108),
(10109, 10100, '2FA初始化', 3, 'POST', '/system/users/google2fa/init', 'sys:user:2fa:init', 10109),
(10110, 10100, '2FA绑定', 3, 'POST', '/system/users/google2fa/bind', 'sys:user:2fa:bind', 10110),
(10111, 10100, '2FA启用', 3, 'POST', '/system/users/google2fa/enable', 'sys:user:2fa:enable', 10111),
(10112, 10100, '2FA禁用', 3, 'POST', '/system/users/google2fa/disable', 'sys:user:2fa:disable', 10112),
(10113, 10100, '2FA重置', 3, 'POST', '/system/users/google2fa/reset', 'sys:user:2fa:reset', 10113);


INSERT INTO sys_menu (id, parent_id, name, menu_type, method, path, perms, component, icon, sort)
VALUES (10200, 10000, '角色管理', 2, 'GET', '/system/roles', 'sys:role:list', 'system/roles', 'Guide', 10200);

INSERT INTO sys_menu (id, parent_id, name, menu_type, method, path, perms, sort)
VALUES
(10201, 10200, '新增角色', 3, 'POST', '/system/roles', 'sys:role:add', 10201),
(10202, 10200, '编辑角色', 3, 'PUT', '/system/roles', 'sys:role:update', 10202),
(10203, 10200, '删除角色', 3, 'DELETE', '/system/roles', 'sys:role:delete', 10203),
(10204, 10200, '菜单授权', 3, 'POST', '/system/roles/grant', 'sys:role:grant', 10204),
(10205, 10200, '授权详情', 3, 'GET', '/system/roles/{id}/grant', 'sys:role:grant:detail', 10205),
(10206, 10000, '权限列表', 3, 'GET', '/system/perms', 'sys:perm:list', 10206);

INSERT INTO sys_menu (id, parent_id, name, menu_type, method, path, perms, component, icon, sort)
VALUES (10300, 10000, '菜单管理', 2, 'GET', '/system/menus', 'sys:menu:list', 'system/menus', 'Menu', 10300);

INSERT INTO sys_menu (id, parent_id, name, menu_type, method, path, perms, sort)
VALUES
(10301, 10300, '新增菜单', 3, 'POST', '/system/menus', 'sys:menu:add', 10301),
(10302, 10300, '编辑菜单', 3, 'PUT', '/system/menus', 'sys:menu:update', 10302),
(10303, 10300, '删除菜单', 3, 'DELETE', '/system/menus', 'sys:menu:delete', 10303),
(10304, 10300, '菜单树', 3, 'GET', '/system/menus/tree', 'sys:menu:tree', 10304);

INSERT INTO sys_menu (id, parent_id, name, menu_type, method, path, perms, component, icon, sort)
VALUES (10400, 10000, '系统配置', 2, 'GET', '/system/configs', 'sys:config:list', 'system/configs', 'Cpu', 10400);
INSERT INTO sys_menu (id, parent_id, name, menu_type, method, path, perms, sort)
VALUES
(10401, 10400, '新增配置', 3, 'POST', '/system/configs', 'sys:config:add', 10401),
(10402, 10400, '编辑配置', 3, 'PUT', '/system/configs', 'sys:config:update', 10402),
(10403, 10400, '删除配置', 3, 'DELETE', '/system/configs', 'sys:config:delete', 10403);

INSERT INTO sys_menu (id, parent_id, name, menu_type, path, component, icon, sort)
VALUES (10500, 10000, '定时任务', 1, '', '', 'AlarmClock', 10500);
INSERT INTO sys_menu (id, parent_id, name, menu_type, method, path, perms, component, icon, sort)
VALUES (10600, 10500, '定时任务列表', 2, 'GET', '/system/jobs', 'sys:job:list', 'system/cronjobs', 'Clock', 10600);
INSERT INTO sys_menu (id, parent_id, name, menu_type, method, path, perms, sort)
VALUES
(10601, 10600, '新增任务', 3, 'POST', '/system/jobs', 'sys:job:add', 10601),
(10602, 10600, '编辑任务', 3, 'PUT', '/system/jobs', 'sys:job:update', 10602),
(10603, 10600, '删除任务', 3, 'DELETE', '/system/jobs/{id}', 'sys:job:delete', 10603),
(10604, 10600, '运行任务', 3, 'POST', '/system/jobs/{id}/run', 'sys:job:run', 10604),
(10605, 10600, '启动任务', 3, 'POST', '/system/jobs/{id}/start', 'sys:job:start', 10605),
(10606, 10600, '停止任务', 3, 'POST', '/system/jobs/{id}/stop', 'sys:job:stop', 10606),
(10607, 10600, '任务处理器', 3, 'GET', '/system/jobs/handlers', 'sys:job:handlers', 10607);
INSERT INTO sys_menu (id, parent_id, name, menu_type, method, path, perms, component, icon, sort)
VALUES (10700, 10500, '定时任务日志', 2, 'GET', '/system/logs/job', 'sys:job:log:list', 'system/cronjobs-log', 'Paperclip', 10700);

INSERT INTO sys_menu (id, parent_id, name, menu_type, method, path, perms, component, icon, sort)
VALUES (10800, 10000, '登录日志', 2, 'GET', '/system/logs/login', 'sys:log:login:list', 'system/login-log', 'Reading', 10800);

INSERT INTO sys_menu (id, parent_id, name, menu_type, method, path, perms, component, icon, sort)
VALUES (10900, 10000, '操作日志', 2, 'GET', '/system/logs/op', 'sys:log:op:list', 'system/op-log', 'Document', 10900);

INSERT INTO sys_menu (id, parent_id, name, menu_type, method, path, perms, component, icon, sort)
VALUES (11000, 10000, '租户管理', 2, 'GET', '/system/tenants', 'sys:tenant:list', 'system/tenants', 'Team', 11000);
INSERT INTO sys_menu (id, parent_id, name, menu_type, method, path, perms, sort)
VALUES
(11001, 11000, '新增租户', 3, 'POST', '/system/tenants', 'sys:tenant:add', 11001),
(11002, 11000, '编辑租户', 3, 'PUT', '/system/tenants', 'sys:tenant:update', 11002),
(11003, 11000, '删除租户', 3, 'DELETE', '/system/tenants', 'sys:tenant:delete', 11003),
(11004, 11000, '获取租户详情', 3, 'GET', '/system/tenant/detail', 'sys:tenant:detail', 11004),
(11005, 10000, '系统服务选项', 3, 'GET', '/system/options', 'sys:options', 11005);
