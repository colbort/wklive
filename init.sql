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
(1, 26),
(1, 27),
(1, 28),
(1, 29),
(1, 30),
(1, 31),
(1, 32),

(1, 100),
(1, 101),
(1, 102),
(1, 103),
(1, 104),
(1, 105),
(1, 106),
(1, 107),
(1, 108),
(1, 109),
(1, 110),
(1, 111),
(1, 112),
(1, 113),
(1, 114),
(1, 115),
(1, 116),
(1, 117),
(1, 118),
(1, 119),
(1, 120),
(1, 121),
(1, 122),
(1, 123),
(1, 124),
(1, 125),
(1, 126),
(1, 127),
(1, 128),
(1, 129),
(1, 130),
(1, 131),
(1, 132),
(1, 133),
(1, 134),
(1, 135),
(1, 136),
(1, 137),
(1, 138),
(1, 139),
(1, 140),
(1, 141),
(1, 142),

(1, 300),
(1, 301),
(1, 302),
(1, 303),
(1, 304),
(1, 305),
(1, 306),
(1, 307),
(1, 308),
(1, 309),
(1, 310),
(1, 311),
(1, 312),
(1, 313),
(1, 314),
(1, 315),
(1, 316),
(1, 317),
(1, 318),
(1, 319),
(1, 320),
(1, 321),
(1, 322),

(1, 400),
(1, 401),
(1, 402),
(1, 403),
(1, 404),
(1, 405),
(1, 406),
(1, 407),
(1, 408),
(1, 409),
(1, 410),
(1, 411),

(1, 500),
(1, 501),
(1, 502),
(1, 503),
(1, 504),
(1, 505),
(1, 506),
(1, 507),
(1, 508),
(1, 509),
(1, 510),
(1, 511),
(1, 512),
(1, 513),
(1, 514),
(1, 515),
(1, 516),
(1, 517),
(1, 518),
(1, 519),
(1, 520),
(1, 521),

(1, 600),
(1, 601),
(1, 602),
(1, 603),
(1, 604),
(1, 605),
(1, 606),
(1, 607),
(1, 608),
(1, 609),
(1, 610),
(1, 611),

(1, 700),
(1, 701),
(1, 702),
(1, 703),
(1, 704),
(1, 705),
(1, 706),
(1, 707),
(1, 708),
(1, 709),
(1, 710),
(1, 711),
(1, 712),
(1, 713),
(1, 714),
(1, 715),
(1, 716),
(1, 717),
(1, 718),
(1, 719),
(1, 720),
(1, 721),
(1, 722),
(1, 723),
(1, 724),
(1, 725),
(1, 726),
(1, 727),

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

(1, 10200),
(1, 10201),
(1, 10202),
(1, 10203),
(1, 10204),

(1, 10300),
(1, 10301),
(1, 10302),
(1, 10303),

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
(1, 11005),
(1, 11006),
(1, 11007);



INSERT INTO sys_menu (id, parent_id, name, menu_type, icon, sort)
VALUES (10, 0, '用户管理', 1, 'Users', 10);
INSERT INTO sys_menu (id, parent_id, name, menu_type, method, path, perms, component, icon, sort)
VALUES 
(11, 10, '用户列表', 2, 'GET', '/member/users', 'users:user:list', 'users/user', 'User', 11),
(12, 10, '创建用户', 3, 'POST', '/member/users', 'users:user:add', '', '', 12),
(13, 10, '获取用户详情', 3, 'GET', '/member/users/{id}', 'users:user:detail', '', '', 13),
(14, 10, '更新用户基本信息', 3, 'PUT', '/member/users/{id}/base', 'users:user:update', '', '', 14),
(15, 10, '更新用户状态', 3, 'PUT', '/member/users/{id}/status', 'users:user:update:status', '', '', 15),
(16, 10, '更新用户会员等级', 3, 'PUT', '/member/users/{id}/level', 'users:user:update:level', '', '', 16),
(17, 10, '重置登录密码', 3, 'PUT', '/member/users/{id}/reset-login-password', 'users:user:reset:loginpwd', '', '', 17),
(18, 10, '重置支付密码', 3, 'PUT', '/member/users/{id}/reset-pay-password', 'users:user:reset:paypwd', '', '', 18),
(19, 10, '解锁用户', 3, 'PUT', '/member/users/{id}/unlock', 'users:user:unlock', '', '', 19),
(20, 10, '更新用户风险等级', 3, 'PUT', '/member/users/{id}/risk-level', 'users:user:update:risklevel', '', '', 20),
(21, 10, '删除用户', 3, 'DELETE', '/member/users/{id}', 'users:user:delete', '', '', 21),
(22, 10, '获取用户安全设置', 3, 'GET', '/member/users/{id}/security', 'users:user:security:detail', '', '', 22),
(23, 10, '重置用户谷歌2FA', 3, 'PUT', '/member/users/{id}/reset2fa', 'users:user:reset:google2fa', '', '', 23),
(24, 10, '实名认证信息列表', 2, 'GET', '/member/user-identities', 'users:user:identities:list', 'users/identity', '', 24),
(25, 10, '审核实名认证信息', 3, 'PUT', '/member/user-identities/{id}/review', 'users:user:identities:review', '', '', 25),
(26, 10, '用户银行卡列表', 2, 'GET', '/member/user-banks', 'users:user:banks:list', 'users/bank', '', 26),
(27, 10, '获取用户银行卡详情', 3, 'GET', '/member/user-banks/{id}', 'users:user:bank:detail', '', '', 27),
(28, 10, '添加用户银行卡', 3, 'POST', '/member/user-banks', 'users:user:bank:add', '', '', 28),
(29, 10, '更新用户银行卡', 3, 'PUT', '/member/user-banks/{id}', 'users:user:bank:update', '', '', 29),
(30, 10, '删除用户银行卡', 3, 'DELETE', '/member/user-banks/{id}', 'users:user:bank:delete', '', '', 30),
(31, 10, '更新用户银行卡状态', 3, 'PUT', '/member/user-banks/{id}/status', 'users:user:bank:update:status', '', '', 31),
(32, 10, '设置默认用户银行卡', 3, 'PUT', '/member/user-banks/{id}/default', 'users:user:bank:setdefault', '', '', 32);


INSERT INTO sys_menu (id, parent_id, name, menu_type, icon, sort)
VALUES (100, 0, '支付管理', 1, 'Payment', 100);

INSERT INTO sys_menu (id, parent_id, name, menu_type, method, path, perms, component, icon, sort)
VALUES
(101, 100, '平台管理', 2, 'GET', '/payment/platforms', 'payment:platform:list', 'payment/platforms', 'Bank', 101),
(102, 101, '创建支付平台', 3, 'POST', '/payment/platform', 'payment:platform:add', '', '', 102),
(103, 101, '更新支付平台', 3, 'PUT', '/payment/platform', 'payment:platform:update', '', '', 103),
(104, 101, '获取支付平台详情', 3, 'GET', '/payment/platform', 'payment:platform:detail', '', '', 104),
(105, 101, '删除支付平台', 3, 'DELETE', '/payment/platform/{id}', 'payment:platform:delete', '', '', 105),

(106, 100, '产品管理', 2, 'GET', '/payment/products', 'payment:itick:list', 'payment/products', 'Gold', 106),
(107, 106, '创建支付产品', 3, 'POST', '/payment/product', 'payment:itick:add', '', '', 107),
(108, 106, '更新支付产品', 3, 'PUT', '/payment/product', 'payment:itick:update', '', '', 108),
(109, 106, '获取支付产品详情', 3, 'GET', '/payment/product', 'payment:itick:detail', '', '', 109),
(110, 106, '删除支付产品', 3, 'DELETE', '/payment/product/{id}', 'payment:itick:delete', '', '', 110),

(111, 100, '租户开通平台列表', 2, 'GET', '/payment/tenant-platforms', 'payment:tenantplatform:list', 'payment/tenant-platforms', 'Gold', 111),
(112, 111, '开通租户平台', 3, 'POST', '/payment/tenant-platform', 'payment:tenantplatform:add', '', '', 112),
(113, 111, '更新租户平台', 3, 'PUT', '/payment/tenant-platform', 'payment:tenantplatform:update', '', '', 113),
(114, 111, '获取租户平台详情', 3, 'GET', '/payment/tenant-platform', 'payment:tenantplatform:detail', '', '', 114),
(115, 111, '删除租户平台', 3, 'DELETE', '/payment/tenant-platform/{id}', 'payment:tenantplatform:delete', '', '', 115),

(116, 100, '租户支付账号列表', 2, 'GET', '/payment/tenant-accounts', 'payment:tenantaccount:list', 'payment/tenant-accounts', 'Gold', 116),
(117, 116, '创建租户支付账号', 3, 'POST', '/payment/tenant-account', 'payment:tenantaccount:add', '', '', 117),
(118, 116, '更新租户支付账号', 3, 'PUT', '/payment/tenant-account', 'payment:tenantaccount:update', '', '', 118),
(119, 116, '获取租户支付账号详情', 3, 'GET', '/payment/tenant-account', 'payment:tenantaccount:detail', '', '', 119),
(120, 116, '删除租户支付账号', 3, 'DELETE', '/payment/tenant-account/{id}', 'payment:tenantaccount:delete', '', '', 120),

(121, 100, '租户支付通道列表', 2, 'GET', '/payment/tenant-channels', 'payment:tenantchannel:list', 'payment/tenant-channels', 'Gold', 121),
(122, 121, '创建租户支付通道', 3, 'POST', '/payment/tenant-channel', 'payment:tenantchannel:add', '', '', 122),
(123, 121, '更新租户支付通道', 3, 'PUT', '/payment/tenant-channel', 'payment:tenantchannel:update', '', '', 123),
(124, 121, '获取租户支付通道详情', 3, 'GET', '/payment/tenant-channel', 'payment:tenantchannel:detail', '', '', 124),
(125, 121, '删除租户支付通道', 3, 'DELETE', '/payment/tenant-channel/{id}', 'payment:tenantchannel:delete', '', '', 125),

(126, 100, '通道规则列表', 2, 'GET', '/payment/tenant-channel-rules', 'payment:channelrule:list', 'payment/channel-rules', 'Gold', 126),
(127, 126, '创建通道规则', 3, 'POST', '/payment/tenant-channel-rule', 'payment:channelrule:add', '', '', 127),
(128, 126, '更新通道规则', 3, 'PUT', '/payment/tenant-channel-rule', 'payment:channelrule:update', '', '', 128),
(129, 126, '删除通道规则', 3, 'DELETE', '/payment/tenant-channel-rule/{id}', 'payment:channelrule:delete', '', '', 129),

(130, 100, '用户充值统计', 2, 'GET', '/payment/user-recharge-stats', 'payment:userrechargestat:list', 'payment/user-recharge-stats', 'Gold', 130),

(131, 100, '充值管理', 2, 'GET', '/payment/recharge-orders', 'payment:recharge-order:list', 'payment/recharge-orders', 'Gold', 131),
(132, 131, '获取充值订单详情', 3, 'GET', '/payment/recharge-order/{orderNo}', 'payment:recharge-order:detail', '', '', 132),
(133, 131, '关闭充值订单', 3, 'POST', '/payment/recharge-order/{orderNo}/close', 'payment:recharge-order:close', '', '', 133),
(134, 131, '人工标记充值订单支付成功', 3, 'POST', '/payment/recharge-order/{orderNo}/manual-success', 'payment:recharge-order:manualsuccess', '', '', 134),
(135, 131, '重试充值回调', 3, 'POST', '/payment/recharge-order/{orderNo}/retry-notify', 'payment:recharge-order:retrynotify', '', '', 135),

(136, 100, '充值回调日志列表', 2, 'GET', '/payment/recharge-notify-logs', 'payment:recharge-notifylog:list', 'payment/recharge-notify-logs', 'Gold', 136),
(137, 136, '获取充值回调日志详情', 3, 'GET', '/payment/recharge-notify-log/{id}', 'payment:recharge-notifylog:detail', '', '', 137),

(138, 100, '提现管理', 2, 'GET', '/payment/withdraw-orders', 'payment:withdraw-order:list', 'payment/withdraw-orders', 'Gold', 138),
(139, 138, '获取提现订单详情', 3, 'GET', '/payment/withdraw-order/{orderNo}', 'payment:withdraw-order:detail', '', '', 139),
(140, 138, '审核提现订单', 3, 'POST', '/payment/withdraw-order/{orderNo}/audit', 'payment:withdraw-order:audit', '', '', 140),

(141, 100, '提现回调日志列表', 2, 'GET', '/payment/withdraw-notify-logs', 'payment:withdraw-notifylog:list', 'payment/withdraw-notify-logs', 'Gold', 141),
(142, 141, '获取提现回调日志详情', 3, 'GET', '/payment/withdraw-notify-log/{id}', 'payment:withdraw-notifylog:detail', '', '', 142);



INSERT INTO sys_menu (id, parent_id, name, menu_type, icon, sort)
VALUES (300, 0, 'ITICK数据管理', 1, 'Goods', 300);

INSERT INTO sys_menu (id, parent_id, name, menu_type, method, path, perms, component, icon, sort)
VALUES
-- 产品类型管理
(301, 300, '产品类型列表', 2, 'GET', '/itick/categories', 'itick:category:list', 'itick/categories', 'Menu', 301),
(302, 301, '创建产品类型', 3, 'POST', '/itick/categories', 'itick:category:add', '', '', 302),
(303, 301, '更新产品类型', 3, 'PUT', '/itick/categories', 'itick:category:update', '', '', 303),
(304, 301, '获取产品类型详情', 3, 'GET', '/itick/categories/:id', 'itick:category:detail', '', '', 304),
(305, 301, '同步类型下的产品', 3, 'POST', '/itick/categories/sync-products', 'itick:category:syncProducts', '', '', 305),
(306, 301, '同步任务状态', 3, 'GET', '/itick/sync-tasks/:taskNo/status', 'itick:sync-task:status', '', '', 306),

-- 产品管理
(307, 300, '产品列表', 2, 'GET', '/itick/products', 'itick:itick:list', 'itick/products', 'Goods', 307),
(308, 307, '创建产品', 3, 'POST', '/itick/products', 'itick:itick:add', '', '', 308),
(309, 307, '更新产品', 3, 'PUT', '/itick/products', 'itick:itick:update', '', '', 309),
(310, 307, '获取产品详情', 3, 'GET', '/itick/products/:id', 'itick:itick:detail', '', '', 310),
(311, 307, 'K线查看', 3, 'GET', '/itick/kline', 'itick:kline:view', '', '', 311),

-- 租户产品类型管理
(312, 300, '租户产品类型列表', 2, 'GET', '/itick/tenant-categories', 'itick:tenant-category:list', 'itick/tenant-categories', 'OfficeBuilding', 312),
(313, 312, '创建租户产品类型', 3, 'POST', '/itick/tenant-categories', 'itick:tenant-category:add', '', '', 313),
(314, 312, '更新租户产品类型', 3, 'PUT', '/itick/tenant-categories', 'itick:tenant-category:update', '', '', 314),
(315, 312, '批量更新租户产品类型', 3, 'POST', '/itick/tenant-categories/batch', 'itick:tenant-category:batchUpsert', '', '', 315),
(316, 312, '获取租户产品类型详情', 3, 'GET', '/itick/tenant-categories/:id', 'itick:tenant-category:detail', '', '', 316),

-- 租户产品管理
(317, 300, '租户产品列表', 2, 'GET', '/itick/tenant-products', 'itick:tenant-itick:list', 'itick/tenant-products', 'Grid', 317),
(318, 317, '创建租户产品', 3, 'POST', '/itick/tenant-products', 'itick:tenant-itick:add', '', '', 318),
(319, 317, '更新租户产品', 3, 'PUT', '/itick/tenant-products', 'itick:tenant-itick:update', '', '', 319),
(320, 317, '批量更新租户产品', 3, 'POST', '/itick/tenant-products/batch', 'itick:tenant-itick:batchUpsert', '', '', 320),
(321, 317, '获取租户产品详情', 3, 'GET', '/itick/tenant-products/:id', 'itick:tenant-itick:detail', '', '', 321),

-- 初始化租户展示配置
(322, 300, '初始化租户展示配置', 2, 'POST', '/itick/tenant-display/init', 'itick:tenant-display:init', 'itick/tenant-display-init', 'Setting', 322);


-- 资产（asset）
INSERT INTO sys_menu (id, parent_id, name, menu_type, icon, sort)
VALUES (400, 0, '资产管理', 1, 'Wallet', 400);

INSERT INTO sys_menu (id, parent_id, name, menu_type, method, path, perms, component, icon, sort)
VALUES
(401, 400, '用户资产列表', 2, 'GET', '/asset/user-assets', 'asset:user-asset:list', 'asset/user-assets', 'Wallet', 401),
(403, 400, '资产流水列表', 2, 'GET', '/asset/flows', 'asset:flow:list', 'asset/flows', 'Tickets', 403),
(404, 400, '资产冻结列表', 2, 'GET', '/asset/freezes', 'asset:freeze:list', 'asset/freezes', 'Lock', 404),
(405, 400, '资产锁定列表', 2, 'GET', '/asset/locks', 'asset:lock:list', 'asset/locks', 'Lock', 405);

INSERT INTO sys_menu (id, parent_id, name, menu_type, method, path, perms, sort)
VALUES
(402, 401, '获取用户资产详情', 3, 'GET', '/asset/user-assets/detail', 'asset:user-asset:detail', 402),
(406, 401, '管理员加资产', 3, 'POST', '/asset/add', 'asset:user-asset:add', 406),
(407, 401, '管理员减资产', 3, 'POST', '/asset/sub', 'asset:user-asset:sub', 407),
(408, 404, '管理员冻结资产', 3, 'POST', '/asset/freeze', 'asset:freeze:add', 408),
(409, 404, '管理员解冻资产', 3, 'POST', '/asset/unfreeze', 'asset:freeze:unfreeze', 409),
(410, 405, '管理员锁定资产', 3, 'POST', '/asset/lock', 'asset:lock:add', 410),
(411, 405, '管理员解锁资产', 3, 'POST', '/asset/unlock', 'asset:lock:unlock', 411);

-- 期权（option）
INSERT INTO sys_menu (id, parent_id, name, menu_type, icon, sort)
VALUES (500, 0, '期权管理', 1, 'TrendCharts', 500);

INSERT INTO sys_menu (id, parent_id, name, menu_type, method, path, perms, component, icon, sort)
VALUES
(501, 500, '合约列表', 2, 'GET', '/option/contracts', 'option:contract:list', 'option/contracts', 'Tickets', 501),
(505, 500, '行情详情', 2, 'GET', '/option/market/detail', 'option:market:detail', 'option/market-detail', 'TrendCharts', 505),
(507, 500, '行情快照列表', 2, 'GET', '/option/market/snapshots', 'option:market-snapshot:list', 'option/market-snapshots', 'Histogram', 507),
(508, 500, '订单列表', 2, 'GET', '/option/orders', 'option:order:list', 'option/orders', 'List', 508),
(510, 500, '成交列表', 2, 'GET', '/option/trades', 'option:trade:list', 'option/trades', 'DataLine', 510),
(512, 500, '持仓列表', 2, 'GET', '/option/positions', 'option:position:list', 'option/positions', 'PieChart', 512),
(514, 500, '行权列表', 2, 'GET', '/option/exercises', 'option:exercise:list', 'option/exercises', 'Operation', 514),
(516, 500, '结算列表', 2, 'GET', '/option/settlements', 'option:settlement:list', 'option/settlements', 'Checked', 516),
(518, 500, '账户列表', 2, 'GET', '/option/accounts', 'option:account:list', 'option/accounts', 'Avatar', 518),
(520, 500, '账单列表', 2, 'GET', '/option/bills', 'option:bill:list', 'option/bills', 'Document', 520);

INSERT INTO sys_menu (id, parent_id, name, menu_type, method, path, perms, sort)
VALUES
(502, 501, '创建合约', 3, 'POST', '/option/contracts', 'option:contract:add', 502),
(503, 501, '更新合约', 3, 'POST', '/option/contracts/update', 'option:contract:update', 503),
(504, 501, '获取合约详情', 3, 'GET', '/option/contracts/detail', 'option:contract:detail', 504),
(506, 505, '更新行情', 3, 'POST', '/option/market/update', 'option:market:update', 506),
(509, 508, '获取订单详情', 3, 'GET', '/option/orders/detail', 'option:order:detail', 509),
(511, 510, '获取成交详情', 3, 'GET', '/option/trades/detail', 'option:trade:detail', 511),
(513, 512, '获取持仓详情', 3, 'GET', '/option/positions/detail', 'option:position:detail', 513),
(515, 514, '获取行权详情', 3, 'GET', '/option/exercises/detail', 'option:exercise:detail', 515),
(517, 516, '获取结算详情', 3, 'GET', '/option/settlements/detail', 'option:settlement:detail', 517),
(519, 518, '获取账户详情', 3, 'GET', '/option/accounts/detail', 'option:account:detail', 519),
(521, 520, '获取账单详情', 3, 'GET', '/option/bills/detail', 'option:bill:detail', 521);

-- 质押（staking）
INSERT INTO sys_menu (id, parent_id, name, menu_type, icon, sort)
VALUES (600, 0, '质押管理', 1, 'Coin', 600);

INSERT INTO sys_menu (id, parent_id, name, menu_type, method, path, perms, component, icon, sort)
VALUES
(601, 600, '质押产品列表', 2, 'GET', '/staking/products', 'staking:product:list', 'staking/products', 'Coin', 601),
(606, 600, '质押订单列表', 2, 'GET', '/staking/orders', 'staking:order:list', 'staking/orders', 'List', 606),
(608, 600, '奖励记录列表', 2, 'GET', '/staking/reward-logs', 'staking:reward-log:list', 'staking/reward-logs', 'Medal', 608),
(609, 600, '赎回记录列表', 2, 'GET', '/staking/redeem-logs', 'staking:redeem-log:list', 'staking/redeem-logs', 'RefreshLeft', 609);

INSERT INTO sys_menu (id, parent_id, name, menu_type, method, path, perms, sort)
VALUES
(602, 601, '获取质押产品详情', 3, 'GET', '/staking/products/detail', 'staking:product:detail', 602),
(603, 601, '创建质押产品', 3, 'POST', '/staking/products', 'staking:product:add', 603),
(604, 601, '更新质押产品', 3, 'POST', '/staking/products/update', 'staking:product:update', 604),
(605, 601, '更新质押产品状态', 3, 'POST', '/staking/products/status', 'staking:product:update:status', 605),
(607, 606, '获取质押订单详情', 3, 'GET', '/staking/orders/detail', 'staking:order:detail', 607),
(610, 608, '手动发放奖励', 3, 'POST', '/staking/manual-reward', 'staking:reward-log:manual', 610),
(611, 609, '手动赎回', 3, 'POST', '/staking/manual-redeem', 'staking:redeem-log:manual', 611);

-- 币币交易（trade）
INSERT INTO sys_menu (id, parent_id, name, menu_type, icon, sort)
VALUES (700, 0, '交易管理', 1, 'DataBoard', 700);

INSERT INTO sys_menu (id, parent_id, name, menu_type, method, path, perms, component, icon, sort)
VALUES
(701, 700, '交易对列表', 2, 'GET', '/trade/symbols', 'trade:symbol:list', 'trade/symbols', 'Switch', 701),
(707, 700, '订单列表', 2, 'GET', '/trade/orders', 'trade:order:list', 'trade/orders', 'List', 707),
(709, 700, '成交明细列表', 2, 'GET', '/trade/fills', 'trade:fill:list', 'trade/fills', 'DataLine', 709),
(711, 700, '持仓列表', 2, 'GET', '/trade/positions', 'trade:position:list', 'trade/positions', 'PieChart', 711),
(713, 700, '持仓历史列表', 2, 'GET', '/trade/position-histories', 'trade:position-history:list', 'trade/position-histories', 'Histogram', 713),
(714, 700, '保证金账户列表', 2, 'GET', '/trade/margin-accounts', 'trade:margin-account:list', 'trade/margin-accounts', 'Wallet', 714),
(715, 700, '撤单日志列表', 2, 'GET', '/trade/cancel-logs', 'trade:cancel-log:list', 'trade/cancel-logs', 'DocumentDelete', 715),
(716, 700, '用户交易限制', 2, 'GET', '/trade/user-trade-limit', 'trade:user-trade-limit:detail', 'trade/user-trade-limit', 'Warning', 716),
(718, 700, '用户交易对限制', 2, 'GET', '/trade/user-symbol-limit', 'trade:user-symbol-limit:detail', 'trade/user-symbol-limit', 'WarningFilled', 718),
(720, 700, '用户交易配置', 2, 'GET', '/trade/user-trade-config', 'trade:user-trade-config:detail', 'trade/user-trade-config', 'Tools', 720),
(722, 700, '风控校验日志列表', 2, 'GET', '/trade/risk-order-check-logs', 'trade:risk-order-check-log:list', 'trade/risk-order-check-logs', 'Memo', 722),
(723, 700, '用户杠杆配置', 2, 'GET', '/trade/user-leverage-config', 'trade:user-leverage-config:detail', 'trade/user-leverage-config', 'TrendCharts', 723),
(725, 700, '交易事件列表', 2, 'GET', '/trade/events', 'trade:event:list', 'trade/events', 'Bell', 725);

INSERT INTO sys_menu (id, parent_id, name, menu_type, method, path, perms, sort)
VALUES
(702, 701, '获取交易对详情', 3, 'GET', '/trade/symbols/detail', 'trade:symbol:detail', 702),
(703, 701, '创建交易对', 3, 'POST', '/trade/symbols', 'trade:symbol:add', 703),
(704, 701, '更新交易对', 3, 'POST', '/trade/symbols/update', 'trade:symbol:update', 704),
(705, 701, '设置现货交易对配置', 3, 'POST', '/trade/symbols/spot-config', 'trade:symbol:spot-config', 705),
(706, 701, '设置合约交易对配置', 3, 'POST', '/trade/symbols/contract-config', 'trade:symbol:contract-config', 706),
(708, 707, '获取订单详情', 3, 'GET', '/trade/orders/detail', 'trade:order:detail', 708),
(710, 709, '获取成交明细详情', 3, 'GET', '/trade/fills/detail', 'trade:fill:detail', 710),
(712, 711, '获取持仓详情', 3, 'GET', '/trade/positions/detail', 'trade:position:detail', 712),
(717, 716, '设置用户交易限制', 3, 'POST', '/trade/user-trade-limit', 'trade:user-trade-limit:update', 717),
(719, 718, '设置用户交易对限制', 3, 'POST', '/trade/user-symbol-limit', 'trade:user-symbol-limit:update', 719),
(721, 720, '设置用户交易配置', 3, 'POST', '/trade/user-trade-config', 'trade:user-trade-config:update', 721),
(724, 723, '设置用户杠杆配置', 3, 'POST', '/trade/user-leverage-config', 'trade:user-leverage-config:update', 724),
(726, 725, '获取交易事件详情', 3, 'GET', '/trade/events/detail', 'trade:event:detail', 726),
(727, 725, '重试交易事件', 3, 'POST', '/trade/events/retry', 'trade:event:retry', 727);



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
(10106, 10100, 'Google2FA管理', 3, 'GET', '/system/users/google2fa', 'sys:user:google2fa', 10106),
(10107, 10100, '2FA初始化', 3, 'POST', '/system/users/google2fa/init', 'sys:user:2fa:init', 10107),
(10108, 10100, '2FA绑定', 3, 'POST', '/system/users/google2fa/bind', 'sys:user:2fa:bind', 10108),
(10109, 10100, '2FA启用', 3, 'POST', '/system/users/google2fa/enable', 'sys:user:2fa:enable', 10109),
(10110, 10100, '2FA禁用', 3, 'POST', '/system/users/google2fa/disable', 'sys:user:2fa:disable', 10110),
(10111, 10100, '2FA重置', 3, 'POST', '/system/users/google2fa/reset', 'sys:user:2fa:reset', 10111);


INSERT INTO sys_menu (id, parent_id, name, menu_type, method, path, perms, component, icon, sort)
VALUES (10200, 10000, '角色管理', 2, 'GET', '/system/roles', 'sys:role:list', 'system/roles', 'Guide', 10200);

INSERT INTO sys_menu (id, parent_id, name, menu_type, method, path, perms, sort)
VALUES
(10201, 10200, '新增角色', 3, 'POST', '/system/roles', 'sys:role:add', 10201),
(10202, 10200, '编辑角色', 3, 'PUT', '/system/roles', 'sys:role:update', 10202),
(10203, 10200, '删除角色', 3, 'DELETE', '/system/roles', 'sys:role:delete', 10203),
(10204, 10200, '菜单授权', 3, 'POST', '/system/roles/grant', 'sys:role:grant', 10204);

INSERT INTO sys_menu (id, parent_id, name, menu_type, method, path, perms, component, icon, sort)
VALUES (10300, 10000, '菜单管理', 2, 'GET', '/system/menus', 'sys:menu:list', 'system/menus', 'Menu', 10300);

INSERT INTO sys_menu (id, parent_id, name, menu_type, method, path, perms, sort)
VALUES
(10301, 10300, '新增菜单', 3, 'POST', '/system/menus', 'sys:menu:add', 10301),
(10302, 10300, '编辑菜单', 3, 'PUT', '/system/menus', 'sys:menu:update', 10302),
(10303, 10300, '删除菜单', 3, 'DELETE', '/system/menus', 'sys:menu:delete', 10303);

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
VALUES (10600, 10500, '定时任务列表', 2, 'GET', '/system/cronjobs', 'sys:job:list', 'system/cronjobs', 'Clock', 10600);
INSERT INTO sys_menu (id, parent_id, name, menu_type, method, path, perms, sort)
VALUES
(10601, 10600, '新增任务', 3, 'POST', '/system/jobs', 'sys:job:add', 10601),
(10602, 10600, '编辑任务', 3, 'PUT', '/system/jobs', 'sys:job:update', 10602),
(10603, 10600, '删除任务', 3, 'DELETE', '/system/jobs', 'sys:job:delete', 10603),
(10604, 10600, '运行任务', 3, 'POST', '/system/jobs/run', 'sys:job:run', 10604),
(10605, 10600, '启动任务', 3, 'POST', '/system/jobs/start', 'sys:job:start', 10605),
(10606, 10600, '停止任务', 3, 'POST', '/system/jobs/stop', 'sys:job:stop', 10606),
(10607, 10600, '任务处理器', 3, 'GET', '/system/jobs/handlers', 'sys:job:handlers', 10607);
INSERT INTO sys_menu (id, parent_id, name, menu_type, method, path, perms, component, icon, sort)
VALUES (10700, 10500, '定时任务日志', 2, 'GET', '/system/cronjobs-log', 'sys:job:log:list', 'system/cronjobs-log', 'Paperclip', 10700);

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
(11004, 11000, '重置密码', 3, 'POST', '/system/tenants/resetPwd', 'sys:tenant:resetpwd', 11004),
(11005, 11000, '禁用租户', 3, 'POST', '/system/tenants/disable', 'sys:tenant:disable', 11005),
(11006, 11000, '启用租户', 3, 'POST', '/system/tenants/enable', 'sys:tenant:enable', 11006),
(11007, 11000, '获取租户详情', 3, 'GET', '/system/tenant/detail', 'sys:tenant:detail', 11007);
