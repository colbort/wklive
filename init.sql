INSERT INTO `sys_role` (id, tenant_id, name, code, enabled, remark, create_times, update_times)
VALUES
(1, 0, '超级管理员', 'super_admin', 1, '', UNIX_TIMESTAMP()*1000, UNIX_TIMESTAMP()*1000),
(2, 0, '租户超级管理员', 'tenant_super_admin', 1, '', UNIX_TIMESTAMP() * 1000, UNIX_TIMESTAMP() * 1000);

INSERT INTO `sys_user` (
  id, tenant_id, user_type, is_owner,
  username, password, nickname, avatar, enabled,
  google_secret, google_enabled, perms_ver,
  last_login_ip, last_login_at, create_by, create_times, update_times
)
VALUES
(
  1, 0, 1, 2,
  'admin',
  '$2a$10$KdJbtCoUCeO.jcI9LJb6me4YAnMt8JScsCWyA9FEPfuaz4bRCfMee',
  '超级管理员', '', 1,
  '', 2, 1,
  '', 0, 0, UNIX_TIMESTAMP()*1000, UNIX_TIMESTAMP()*1000
);
INSERT INTO `sys_user_role` (tenant_id, user_id, role_id)
VALUES
(0, 1, 1);


INSERT INTO `sys_role_menu` (`tenant_id`, `role_id`, `menu_id`) VALUES
(0, 1, 10),
(0, 1, 11),
(0, 1, 12),
(0, 1, 13),
(0, 1, 14),
(0, 1, 15),
(0, 1, 16),
(0, 1, 17),
(0, 1, 18),
(0, 1, 19),
(0, 1, 20),
(0, 1, 21),
(0, 1, 22),
(0, 1, 23),
(0, 1, 24),
(0, 1, 40),
(0, 1, 41),
(0, 1, 60),
(0, 1, 61),
(0, 1, 62),
(0, 1, 63),
(0, 1, 64),
(0, 1, 65),
(0, 1, 66),

(0, 1, 100),
(0, 1, 101),
(0, 1, 102),
(0, 1, 103),
(0, 1, 104),
(0, 1, 105),
(0, 1, 106),
(0, 1, 120),
(0, 1, 121),
(0, 1, 122),
(0, 1, 123),
(0, 1, 124),
(0, 1, 140),
(0, 1, 141),
(0, 1, 142),
(0, 1, 143),
(0, 1, 144),
(0, 1, 160),
(0, 1, 161),
(0, 1, 162),
(0, 1, 163),
(0, 1, 164),
(0, 1, 180),
(0, 1, 181),
(0, 1, 182),
(0, 1, 183),
(0, 1, 184),
(0, 1, 200),
(0, 1, 201),
(0, 1, 202),
(0, 1, 203),
(0, 1, 204),
(0, 1, 220),
(0, 1, 221),
(0, 1, 230),
(0, 1, 231),
(0, 1, 232),
(0, 1, 233),
(0, 1, 234),
(0, 1, 250),
(0, 1, 251),
(0, 1, 260),
(0, 1, 261),
(0, 1, 262),
(0, 1, 280),
(0, 1, 281),
(0, 1, 290),
(0, 1, 291),
(0, 1, 292),
(0, 1, 293),
(0, 1, 294),
(0, 1, 300),
(0, 1, 301),
(0, 1, 302),
(0, 1, 303),
(0, 1, 304),
(0, 1, 310),
(0, 1, 311),

(0, 1, 400),
(0, 1, 401),
(0, 1, 402),
(0, 1, 403),
(0, 1, 404),
(0, 1, 405),
(0, 1, 406),
(0, 1, 420),
(0, 1, 421),
(0, 1, 422),
(0, 1, 423),
(0, 1, 424),
(0, 1, 430),
(0, 1, 431),
(0, 1, 432),
(0, 1, 433),
(0, 1, 434),
(0, 1, 450),
(0, 1, 451),
(0, 1, 452),
(0, 1, 453),
(0, 1, 454),
(0, 1, 470),

(0, 1, 500),
(0, 1, 510),
(0, 1, 511),
(0, 1, 512),
(0, 1, 513),
(0, 1, 514),
(0, 1, 520),
(0, 1, 521),
(0, 1, 522),
(0, 1, 523),
(0, 1, 530),
(0, 1, 540),
(0, 1, 541),
(0, 1, 542),
(0, 1, 550),
(0, 1, 551),
(0, 1, 552),

(0, 1, 600),
(0, 1, 610),
(0, 1, 611),
(0, 1, 612),
(0, 1, 613),
(0, 1, 620),
(0, 1, 621),
(0, 1, 630),
(0, 1, 640),
(0, 1, 641),
(0, 1, 650),
(0, 1, 651),
(0, 1, 660),
(0, 1, 661),
(0, 1, 670),
(0, 1, 671),
(0, 1, 680),
(0, 1, 681),
(0, 1, 690),
(0, 1, 691),
(0, 1, 700),
(0, 1, 701),

(0, 1, 800),
(0, 1, 810),
(0, 1, 811),
(0, 1, 812),
(0, 1, 813),
(0, 1, 814),
(0, 1, 820),
(0, 1, 821),
(0, 1, 830),
(0, 1, 831),
(0, 1, 840),
(0, 1, 841),

(0, 1, 1000),
(0, 1, 1010),
(0, 1, 1011),
(0, 1, 1012),
(0, 1, 1013),
(0, 1, 1014),
(0, 1, 1015),
(0, 1, 1016),
(0, 1, 1017),
(0, 1, 1018),
(0, 1, 1020),
(0, 1, 1021),
(0, 1, 1030),
(0, 1, 1031),
(0, 1, 1040),
(0, 1, 1041),
(0, 1, 1050),
(0, 1, 1060),
(0, 1, 1070),
(0, 1, 1080),
(0, 1, 1081),
(0, 1, 1090),
(0, 1, 1091),
(0, 1, 1100),
(0, 1, 1101),
(0, 1, 1110),
(0, 1, 1120),
(0, 1, 1121),
(0, 1, 1130),
(0, 1, 1131),
(0, 1, 1132),
(0, 1, 10000),
(0, 1, 10100),
(0, 1, 10101),
(0, 1, 10102),
(0, 1, 10103),
(0, 1, 10104),
(0, 1, 10105),
(0, 1, 10106),
(0, 1, 10107),
(0, 1, 10108),
(0, 1, 10109),
(0, 1, 10110),
(0, 1, 10111),
(0, 1, 10112),
(0, 1, 10113),

(0, 1, 10200),
(0, 1, 10201),
(0, 1, 10202),
(0, 1, 10203),
(0, 1, 10204),
(0, 1, 10205),
(0, 1, 10206),

(0, 1, 10300),
(0, 1, 10301),
(0, 1, 10302),
(0, 1, 10303),
(0, 1, 10304),

(0, 1, 10400),
(0, 1, 10401),
(0, 1, 10402),
(0, 1, 10403),

(0, 1, 10500),
(0, 1, 10600),
(0, 1, 10601),
(0, 1, 10602),
(0, 1, 10603),
(0, 1, 10604),
(0, 1, 10605),
(0, 1, 10606),
(0, 1, 10607),

(0, 1, 10700),
(0, 1, 10800),
(0, 1, 10900),

(0, 1, 11000),
(0, 1, 11001),
(0, 1, 11002),
(0, 1, 11003),
(0, 1, 11004),
(0, 1, 11100),
(0, 1, 11101),
(0, 1, 11102),
(0, 1, 11103),
(0, 1, 11104),
(0, 1, 11200),
(0, 1, 11201),
(0, 1, 11202);

INSERT INTO sys_role_menu (tenant_id, role_id, menu_id) VALUES
-- 用户管理
(0, 2, 10),
(0, 2, 11),
(0, 2, 12),
(0, 2, 13),
(0, 2, 14),
(0, 2, 15),
(0, 2, 16),
(0, 2, 17),
(0, 2, 18),
(0, 2, 19),
(0, 2, 20),
(0, 2, 21),
(0, 2, 22),
(0, 2, 23),
(0, 2, 24),
(0, 2, 40),
(0, 2, 41),
(0, 2, 60),
(0, 2, 61),
(0, 2, 62),
(0, 2, 63),
(0, 2, 64),
(0, 2, 65),
(0, 2, 66),

-- 支付管理：只保留租户相关
(0, 2, 100),
(0, 2, 140),
(0, 2, 141),
(0, 2, 142),
(0, 2, 143),
(0, 2, 144),
(0, 2, 160),
(0, 2, 161),
(0, 2, 162),
(0, 2, 163),
(0, 2, 164),
(0, 2, 180),
(0, 2, 181),
(0, 2, 182),
(0, 2, 183),
(0, 2, 184),
(0, 2, 200),
(0, 2, 201),
(0, 2, 202),
(0, 2, 203),
(0, 2, 204),
(0, 2, 220),
(0, 2, 221),
(0, 2, 230),
(0, 2, 231),
(0, 2, 232),
(0, 2, 233),
(0, 2, 234),
(0, 2, 250),
(0, 2, 251),
(0, 2, 260),
(0, 2, 261),
(0, 2, 262),
(0, 2, 280),
(0, 2, 281),
(0, 2, 290),
(0, 2, 291),
(0, 2, 292),
(0, 2, 293),
(0, 2, 294),
(0, 2, 300),
(0, 2, 301),
(0, 2, 302),
(0, 2, 303),
(0, 2, 304),
(0, 2, 310),
(0, 2, 311),

-- ITICK数据管理：排除平台产品类型、平台产品，只保留租户相关
(0, 2, 400),
(0, 2, 401),
(0, 2, 404),
(0, 2, 430),
(0, 2, 431),
(0, 2, 432),
(0, 2, 433),
(0, 2, 434),
(0, 2, 450),
(0, 2, 451),
(0, 2, 452),
(0, 2, 453),
(0, 2, 454),
(0, 2, 470),

-- 资产管理
(0, 2, 500),
(0, 2, 510),
(0, 2, 511),
(0, 2, 512),
(0, 2, 513),
(0, 2, 514),
(0, 2, 520),
(0, 2, 521),
(0, 2, 522),
(0, 2, 523),
(0, 2, 530),
(0, 2, 540),
(0, 2, 541),
(0, 2, 542),
(0, 2, 550),
(0, 2, 551),
(0, 2, 552),

-- 期权管理
(0, 2, 600),
(0, 2, 610),
(0, 2, 611),
(0, 2, 612),
(0, 2, 613),
(0, 2, 620),
(0, 2, 621),
(0, 2, 630),
(0, 2, 640),
(0, 2, 641),
(0, 2, 650),
(0, 2, 651),
(0, 2, 660),
(0, 2, 661),
(0, 2, 670),
(0, 2, 671),
(0, 2, 680),
(0, 2, 681),
(0, 2, 690),
(0, 2, 691),
(0, 2, 700),
(0, 2, 701),

-- 质押管理
(0, 2, 800),
(0, 2, 810),
(0, 2, 811),
(0, 2, 812),
(0, 2, 813),
(0, 2, 814),
(0, 2, 820),
(0, 2, 821),
(0, 2, 830),
(0, 2, 831),
(0, 2, 840),
(0, 2, 841),

-- 交易管理
(0, 2, 1000),
(0, 2, 1010),
(0, 2, 1011),
(0, 2, 1012),
(0, 2, 1013),
(0, 2, 1014),
(0, 2, 1015),
(0, 2, 1016),
(0, 2, 1017),
(0, 2, 1018),
(0, 2, 1020),
(0, 2, 1021),
(0, 2, 1030),
(0, 2, 1031),
(0, 2, 1040),
(0, 2, 1041),
(0, 2, 1050),
(0, 2, 1060),
(0, 2, 1070),
(0, 2, 1080),
(0, 2, 1081),
(0, 2, 1090),
(0, 2, 1091),
(0, 2, 1100),
(0, 2, 1101),
(0, 2, 1110),
(0, 2, 1120),
(0, 2, 1121),
(0, 2, 1130),
(0, 2, 1131),
(0, 2, 1132),

-- 系统管理
(0, 2, 10000),
(0, 2, 10100),
(0, 2, 10101),
(0, 2, 10102),
(0, 2, 10103),
(0, 2, 10104),
(0, 2, 10105),
(0, 2, 10106),
(0, 2, 10107),
(0, 2, 10108),
(0, 2, 10109),
(0, 2, 10110),
(0, 2, 10111),
(0, 2, 10112),
(0, 2, 10113),
(0, 2, 10200),
(0, 2, 10201),
(0, 2, 10202),
(0, 2, 10203),
(0, 2, 10204),
(0, 2, 10205),
(0, 2, 10206),
(0, 2, 10300),
(0, 2, 10304),
(0, 2, 11100),
(0, 2, 11101),
(0, 2, 11102);



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
(24, 11, '校验推荐人', 3, 'GET', '/member/users/referrer/check', 'users:user:referrer:check', '', '', 24),

(40, 10, '实名认证信息列表', 2, 'GET', '/member/user-identities', 'users:user:identities:list', 'users/identity', '', 40),
(41, 40, '审核实名认证信息', 3, 'PUT', '/member/user-identities/{id}/review', 'users:user:identities:review', '', '', 41),

(60, 10, '用户银行卡列表', 2, 'GET', '/member/user-banks', 'users:user:banks:list', 'users/bank', '', 60),
(61, 60, '获取用户银行卡详情', 3, 'GET', '/member/user-banks/{id}', 'users:user:bank:detail', '', '', 61),
(62, 60, '添加用户银行卡', 3, 'POST', '/member/user-banks', 'users:user:bank:add', '', '', 62),
(63, 60, '更新用户银行卡', 3, 'PUT', '/member/user-banks/{id}', 'users:user:bank:update', '', '', 63),
(64, 60, '删除用户银行卡', 3, 'DELETE', '/member/user-banks/{id}', 'users:user:bank:delete', '', '', 64),
(65, 60, '更新用户银行卡状态', 3, 'PUT', '/member/user-banks/{id}/status', 'users:user:bank:update:status', '', '', 65),
(66, 60, '设置默认用户银行卡', 3, 'PUT', '/member/user-banks/{id}/default', 'users:user:bank:default', '', '', 66);


INSERT INTO sys_menu (id, parent_id, name, menu_type, icon, sort)
VALUES (100, 0, '支付管理', 1, 'Payment', 100);

INSERT INTO sys_menu (id, parent_id, name, menu_type, method, path, perms, component, icon, sort)
VALUES
(101, 100, '平台管理', 2, 'GET', '/payment/platforms', 'payment:platform:list', 'payment/platforms', 'Bank', 101),
(102, 101, '创建支付平台', 3, 'POST', '/payment/platform', 'payment:platform:add', '', '', 102),
(103, 101, '更新支付平台', 3, 'PUT', '/payment/platform', 'payment:platform:update', '', '', 103),
(104, 101, '获取支付平台详情', 3, 'GET', '/payment/platform', 'payment:platform:detail', '', '', 104),
(105, 101, '删除支付平台', 3, 'DELETE', '/payment/platform/{id}', 'payment:platform:delete', '', '', 105),
(106, 101, '平台支付通道', 3, 'GET', '/payment/platforms/options', 'payment:platform:options', '', '', 106),

(120, 100, '产品管理', 2, 'GET', '/payment/products', 'payment:product:list', 'payment/products', 'Gold', 120),
(121, 120, '创建支付产品', 3, 'POST', '/payment/product', 'payment:product:add', '', '', 121),
(122, 120, '更新支付产品', 3, 'PUT', '/payment/product', 'payment:product:update', '', '', 122),
(123, 120, '获取支付产品详情', 3, 'GET', '/payment/product', 'payment:product:detail', '', '', 123),
(124, 120, '删除支付产品', 3, 'DELETE', '/payment/product/{id}', 'payment:product:delete', '', '', 124),

(140, 100, '租户平台', 2, 'GET', '/payment/tenant-platforms', 'payment:tenant-platform:list', 'payment/tenant-platforms', 'Gold', 140),
(141, 140, '开通租户平台', 3, 'POST', '/payment/tenant-platform', 'payment:tenant-platform:add', '', '', 141),
(142, 140, '更新租户平台', 3, 'PUT', '/payment/tenant-platform', 'payment:tenant-platform:update', '', '', 142),
(143, 140, '获取租户平台详情', 3, 'GET', '/payment/tenant-platform', 'payment:tenant-platform:detail', '', '', 143),
(144, 140, '删除租户平台', 3, 'DELETE', '/payment/tenant-platform/{id}', 'payment:tenant-platform:delete', '', '', 144),

(160, 100, '租户支付账号', 2, 'GET', '/payment/tenant-accounts', 'payment:tenant-account:list', 'payment/tenant-accounts', 'Gold', 160),
(161, 160, '创建租户支付账号', 3, 'POST', '/payment/tenant-account', 'payment:tenant-account:add', '', '', 161),
(162, 160, '更新租户支付账号', 3, 'PUT', '/payment/tenant-account', 'payment:tenant-account:update', '', '', 162),
(163, 160, '获取租户支付账号详情', 3, 'GET', '/payment/tenant-account', 'payment:tenant-account:detail', '', '', 163),
(164, 160, '删除租户支付账号', 3, 'DELETE', '/payment/tenant-account/{id}', 'payment:tenant-account:delete', '', '', 164),

(180, 100, '租户支付通道', 2, 'GET', '/payment/tenant-channels', 'payment:tenant-channel:list', 'payment/tenant-channels', 'Gold', 180),
(181, 180, '创建租户支付通道', 3, 'POST', '/payment/tenant-channel', 'payment:tenant-channel:add', '', '', 181),
(182, 180, '更新租户支付通道', 3, 'PUT', '/payment/tenant-channel', 'payment:tenant-channel:update', '', '', 182),
(183, 180, '获取租户支付通道详情', 3, 'GET', '/payment/tenant-channel', 'payment:tenant-channel:detail', '', '', 183),
(184, 180, '删除租户支付通道', 3, 'DELETE', '/payment/tenant-channel/{id}', 'payment:tenant-channel:delete', '', '', 184),

(200, 100, '租户支付通道规则', 2, 'GET', '/payment/tenant-channel-rules', 'payment:tenant-channel-rule:list', 'payment/tenant-channel-rules', 'Gold', 200),
(201, 200, '创建支付通道规则', 3, 'POST', '/payment/tenant-channel-rule', 'payment:tenant-channel-rule:add', '', '', 201),
(202, 200, '更新支付通道规则', 3, 'PUT', '/payment/tenant-channel-rule', 'payment:tenant-channel-rule:update', '', '', 202),
(203, 200, '删除支付通道规则', 3, 'DELETE', '/payment/tenant-channel-rule/{id}', 'payment:tenant-channel-rule:delete', '', '', 203),
(204, 200, '获取支付通道规则详情', 3, 'GET', '/payment/tenant-channel-rule', 'payment:tenant-channel-rule:detail', '', '', 204),

(220, 100, '用户充值统计', 2, 'GET', '/payment/user-recharge-stats', 'payment:user-recharge-stat:list', 'payment/user-recharge-stats', 'Gold', 220),
(221, 220, '用户充值统计详情', 3, 'GET', '/payment/user-recharge-stat', 'payment:user-recharge-stat:detail', '', '', 221),

(230, 100, '充值管理', 2, 'GET', '/payment/recharge-orders', 'payment:recharge-order:list', 'payment/recharge-orders', 'Gold', 230),
(231, 230, '获取充值订单详情', 3, 'GET', '/payment/recharge-order/{orderNo}', 'payment:recharge-order:detail', '', '', 231),
(232, 230, '关闭充值订单', 3, 'POST', '/payment/recharge-order/{orderNo}/close', 'payment:recharge-order:close', '', '', 232),
(233, 230, '人工标记充值订单支付成功', 3, 'POST', '/payment/recharge-order/{orderNo}/manual-success', 'payment:recharge-order:manual-success', '', '', 233),
(234, 230, '重试充值回调', 3, 'POST', '/payment/recharge-order/{orderNo}/retry-notify', 'payment:recharge-order:retry-notify', '', '', 234),

(250, 100, '充值回调日志', 2, 'GET', '/payment/recharge-notify-logs', 'payment:recharge-notify-log:list', 'payment/recharge-notify-logs', 'Gold', 250),
(251, 250, '获取充值回调日志详情', 3, 'GET', '/payment/recharge-notify-log/{id}', 'payment:recharge-notify-log:detail', '', '', 251),

(260, 100, '提现管理', 2, 'GET', '/payment/withdraw-orders', 'payment:withdraw-order:list', 'payment/withdraw-orders', 'Gold', 260),
(261, 260, '获取提现订单详情', 3, 'GET', '/payment/withdraw-order/{id}', 'payment:withdraw-order:detail', '', '', 261),
(262, 260, '审核提现订单', 3, 'POST', '/payment/withdraw-order/{orderNo}/audit', 'payment:withdraw-order:audit', '', '', 262),

(280, 100, '提现回调日志', 2, 'GET', '/payment/withdraw-notify-logs', 'payment:withdraw-notify-log:list', 'payment/withdraw-notify-logs', 'Gold', 280),
(281, 280, '获取提现回调日志详情', 3, 'GET', '/payment/withdraw-notify-log/{id}', 'payment:withdraw-notify-log:detail', '', '', 281),

(290, 100, '充值地址表', 2, 'GET', '/payment/crypto-recharge-addresses', 'payment:crypto-recharge-addresses:list', 'payment/crypto-recharge-addresses', 'Gold', 290),
(291, 290, '获取充值地址详情', 3, 'GET', '/payment/crypto-recharge-address/{id}', 'payment:crypto-recharge-address:detail', '', '', 291),
(292, 290, '创建充值地址', 3, 'POST', '/payment/crypto-recharge-address', 'payment:crypto-recharge-address:add', '', '', 292),
(293, 290, '更新充值地址', 3, 'PUT', '/payment/crypto-recharge-address', 'payment:crypto-recharge-address:update', '', '', 293),
(294, 290, '删除充值地址', 3, 'DELETE', '/payment/crypto-recharge-address/{id}', 'payment:crypto-recharge-address:delete', '', '', 294),

(300, 100, '币种钱包账户', 2, 'GET', '/payment/crypto-wallet-accounts', 'payment:crypto-wallet-account:list', 'payment/crypto-wallet-accounts', 'Gold', 300),
(301, 300, '获取钱包账户详情', 3, 'GET', '/payment/crypto-wallet-account/{id}', 'payment:crypto-wallet-account:detail', '', '', 301),
(302, 300, '创建钱包账户', 3, 'POST', '/payment/crypto-wallet-account', 'payment:crypto-wallet-account:add', '', '', 302),
(303, 300, '更新钱包账户', 3, 'PUT', '/payment/crypto-wallet-account', 'payment:crypto-wallet-account:update', '', '', 303),
(304, 300, '删除钱包账户', 3, 'DELETE', '/payment/crypto-wallet-account/{id}', 'payment:crypto-wallet-account:delete', '', '', 304),

(310, 100, '充值交易记录', 2, 'GET', '/payment/crypto-recharge-txs', 'payment:crypto-recharge-txs:list', 'payment/crypto-recharge-txs', 'Gold', 310),
(311, 310, '获取充值交易详情', 3, 'GET', '/payment/crypto-recharge-tx/{id}', 'payment:crypto-recharge-tx:detail', '', '', 311);



INSERT INTO sys_menu (id, parent_id, name, menu_type, icon, sort)
VALUES (400, 0, 'ITICK数据管理', 1, 'Goods', 400);

INSERT INTO sys_menu (id, parent_id, name, menu_type, method, path, perms, component, icon, sort)
VALUES
-- 产品类型管理
(401, 400, '产品类型列表', 2, 'GET', '/itick/categories', 'itick:category:list', 'itick/categories', 'Menu', 401),
(402, 401, '创建产品类型', 3, 'POST', '/itick/categories', 'itick:category:add', '', '', 402),
(403, 401, '更新产品类型', 3, 'PUT', '/itick/categories', 'itick:category:update', '', '', 403),
(404, 401, '获取产品类型详情', 3, 'GET', '/itick/categories/{id}', 'itick:category:detail', '', '', 404),
(405, 401, '同步类型下的产品', 3, 'POST', '/itick/categories/sync-products', 'itick:category:syncProducts', '', '', 405),
(406, 401, '同步任务状态', 3, 'GET', '/itick/sync-tasks/{taskNo}/status', 'itick:sync-task:status', '', '', 406),

-- 产品管理 320
(420, 400, '产品列表', 2, 'GET', '/itick/products', 'itick:product:list', 'itick/products', 'Goods', 420),
(421, 420, '创建产品', 3, 'POST', '/itick/products', 'itick:product:add', '', '', 421),
(422, 420, '更新产品', 3, 'PUT', '/itick/products', 'itick:product:update', '', '', 422),
(423, 420, '获取产品详情', 3, 'GET', '/itick/products/{id}', 'itick:product:detail', '', '', 423),
(424, 420, 'K线查看', 3, 'GET', '/itick/kline', 'itick:kline:view', '', '', 424),

-- 租户产品类型管理 330
(430, 400, '租户产品类型列表', 2, 'GET', '/itick/tenant-categories', 'itick:tenant-category:list', 'itick/tenant-categories', 'OfficeBuilding', 430),
(431, 430, '创建租户产品类型', 3, 'POST', '/itick/tenant-categories', 'itick:tenant-category:add', '', '', 431),
(432, 430, '更新租户产品类型', 3, 'PUT', '/itick/tenant-categories', 'itick:tenant-category:update', '', '', 432),
(433, 430, '批量更新租户产品类型', 3, 'POST', '/itick/tenant-categories/batch', 'itick:tenant-category:batchUpsert', '', '', 433),
(434, 430, '获取租户产品类型详情', 3, 'GET', '/itick/tenant-categories/{id}', 'itick:tenant-category:detail', '', '', 434),

-- 租户产品管理 350
(450, 400, '租户产品列表', 2, 'GET', '/itick/tenant-products', 'itick:tenant-itick:list', 'itick/tenant-products', 'Grid', 450),
(451, 450, '创建租户产品', 3, 'POST', '/itick/tenant-products', 'itick:tenant-itick:add', '', '', 451),
(452, 450, '更新租户产品', 3, 'PUT', '/itick/tenant-products', 'itick:tenant-itick:update', '', '', 452),
(453, 450, '批量更新租户产品', 3, 'POST', '/itick/tenant-products/batch', 'itick:tenant-itick:batchUpsert', '', '', 453),
(454, 450, '获取租户产品详情', 3, 'GET', '/itick/tenant-products/{id}', 'itick:tenant-itick:detail', '', '', 454),

-- 初始化租户展示配置
(470, 400, '初始化租户展示配置', 2, 'POST', '/itick/tenant-display/init', 'itick:tenant-display:init', 'itick/tenant-display-init', 'Setting', 470);


-- 资产（asset）
INSERT INTO sys_menu (id, parent_id, name, menu_type, icon, sort)
VALUES (500, 0, '资产管理', 1, 'Wallet', 500);

INSERT INTO sys_menu (id, parent_id, name, menu_type, method, path, perms, component, icon, sort)
VALUES
(510, 500, '资产配置', 2, 'GET', '/asset/coin-configs', 'asset:config:list', 'asset/configs', 'Lock', 510),
(511, 510, '添加资产配置', 3, 'POST', '/asset/coin-configs', 'asset:config:add', '', 'Lock', 511),
(512, 510, '删除资产配置', 3, 'DELETE', '/asset/coin-configs/{id}', 'asset:config:delete', '', 'Lock', 512),
(513, 510, '更新资产配置', 3, 'PUT', '/asset/coin-configs/{id}', 'asset:config:update', '', 'Lock', 513),
(514, 510, '获取资产配置详情', 3, 'GET', '/asset/coin-configs/{id}', 'asset:config:detail', '', 'Lock', 514),
(520, 500, '用户资产列表', 2, 'GET', '/asset/user-assets', 'asset:user-asset:list', 'asset/user-assets', 'Wallet', 520),
(521, 520, '获取用户资产详情', 3, 'GET', '/asset/user-assets/detail', 'asset:user-asset:detail', '', '', 521),
(522, 520, '管理员加资产', 3, 'POST', '/asset/add', 'asset:user-asset:add', '', '', 522),
(523, 520, '管理员减资产', 3, 'POST', '/asset/sub', 'asset:user-asset:sub', '', '', 523),
(530, 500, '资产流水列表', 2, 'GET', '/asset/flows', 'asset:flow:list', 'asset/flows', 'Tickets', 530),
(540, 500, '资产冻结列表', 2, 'GET', '/asset/freezes', 'asset:freeze:list', 'asset/freezes', 'Lock', 540),
(541, 540, '管理员冻结资产', 3, 'POST', '/asset/freeze', 'asset:freeze:add', '', '', 541),
(542, 540, '管理员解冻资产', 3, 'POST', '/asset/unfreeze', 'asset:freeze:unfreeze', '', '', 542),
(550, 500, '资产锁定列表', 2, 'GET', '/asset/locks', 'asset:lock:list', 'asset/locks', 'Lock', 550),
(551, 550, '管理员锁定资产', 3, 'POST', '/asset/lock', 'asset:lock:add', '', '', 551),
(552, 550, '管理员解锁资产', 3, 'POST', '/asset/unlock', 'asset:lock:unlock', '', '', 552);

-- 期权（option）
INSERT INTO sys_menu (id, parent_id, name, menu_type, icon, sort)
VALUES (600, 0, '期权管理', 1, 'TrendCharts', 600);

INSERT INTO sys_menu (id, parent_id, name, menu_type, method, path, perms, component, icon, sort)
VALUES
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
(1010, 1000, '交易对列表', 2, 'GET', '/trade/symbols', 'trade:symbol:list', 'trade/symbols', 'Switch', 1010),
(1011, 1010, '获取交易对详情', 3, 'GET', '/trade/symbols/detail', 'trade:symbol:detail', '', '', 1011),
(1012, 1010, '创建交易对', 3, 'POST', '/trade/symbols', 'trade:symbol:add', '', '', 1012),
(1013, 1010, '更新交易对', 3, 'POST', '/trade/symbols/update', 'trade:symbol:update', '', '', 1013),
(1014, 1010, '设置现货交易对配置', 3, 'POST', '/trade/symbols/spot-config', 'trade:symbol:spot-config', '', '', 1014),
(1015, 1010, '设置合约交易对配置', 3, 'POST', '/trade/symbols/contract-config', 'trade:symbol:contract-config', '', '', 1015),
(1016, 1010, '保存交易对杠杆配置', 3, 'POST', '/trade/symbols/leverage-config', 'trade:symbol:leverage-config:update', '', '', 1016),
(1017, 1010, '获取交易对杠杆配置', 3, 'GET', '/trade/symbols/leverage-config', 'trade:symbol:leverage-config:detail', '', '', 1017),
(1018, 1010, '获取交易对杠杆配置列表', 3, 'GET', '/trade/symbols/leverage-configs', 'trade:symbol:leverage-config:list', '', '', 1018),
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
(10103, 10100, '删除用户', 3, 'DELETE', '/system/users/{id}', 'sys:user:delete', 10103),
(10104, 10100, '重置密码', 3, 'POST', '/system/users/resetPwd', 'sys:user:resetpwd', 10104),
(10105, 10100, '分配角色', 3, 'POST', '/system/users/assignRoles', 'sys:user:assignrole', 10105),
(10106, 10100, '用户详情', 3, 'GET', '/system/users/detail', 'sys:user:detail', 10106),
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
(10206, 10200, '权限列表', 3, 'GET', '/system/perms', 'sys:perm:list', 10206);

INSERT INTO sys_menu (id, parent_id, name, menu_type, method, path, perms, component, icon, sort)
VALUES (10300, 10000, '菜单管理', 2, 'GET', '/system/menus', 'sys:menu:list', 'system/menus', 'Menu', 10300);

INSERT INTO sys_menu (id, parent_id, name, menu_type, method, path, perms, sort)
VALUES
(10301, 10300, '新增菜单', 3, 'POST', '/system/menus', 'sys:menu:add', 10301),
(10302, 10300, '编辑菜单', 3, 'PUT', '/system/menus', 'sys:menu:update', 10302),
(10303, 10300, '删除菜单', 3, 'DELETE', '/system/menus', 'sys:menu:delete', 10303),
(10304, 10300, '菜单树', 3, 'GET', '/system/menus/tree/{tenantId}', 'sys:menu:tree', 10304);

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
(11003, 11000, '删除租户', 3, 'DELETE', '/system/tenants/{id}', 'sys:tenant:delete', 11003),
(11004, 11000, '获取租户详情', 3, 'GET', '/system/tenant/detail', 'sys:tenant:detail', 11004);

INSERT INTO sys_menu (id, parent_id, name, menu_type, method, path, perms, component, icon, sort)
VALUES (11100, 10000, '客服商户管理', 2, 'GET', '/system/chat_merchants', 'sys:chat_merchant:list', 'system/chat_merchants', 'Team', 11100);
INSERT INTO sys_menu (id, parent_id, name, menu_type, method, path, perms, sort)
VALUES
(11101, 11100, '新增客服商户', 3, 'POST', '/system/chat_merchant', 'sys:chat_merchant:add', 11101),
(11102, 11100, '编辑客服商户', 3, 'PUT', '/system/chat_merchant', 'sys:chat_merchant:update', 11102),
(11103, 11100, '删除客服商户', 3, 'DELETE', '/system/chat_merchant/{id}', 'sys:chat_merchant:delete', 11103),
(11104, 11100, '获取客服商户详情', 3, 'GET', '/system/chat_merchant/detail', 'sys:chat_merchant:detail', 11104);

INSERT INTO sys_menu (id, parent_id, name, menu_type, method, path, perms, component, icon, sort)
VALUES (11200, 10000, '验证码记录', 2, 'GET', '/system/verification-codes', 'sys:verification-code:list', 'system/verification-codes', 'Message', 11200);
INSERT INTO sys_menu (id, parent_id, name, menu_type, method, path, perms, sort)
VALUES
(11201, 11200, '验证码详情', 3, 'GET', '/system/verification-codes/{id}', 'sys:verification-code:detail', 11201),
(11202, 11200, '测试发送验证码', 3, 'POST', '/system/verification-codes/test', 'sys:verification-code:test', 11202);
