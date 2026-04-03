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
(1, 11006);



INSERT INTO sys_menu (id, parent_id, name, menu_type, icon, sort)
VALUES (10, 0, '用户管理', 1, 'Users', 10);
INSERT INTO sys_menu (id, parent_id, name, menu_type, method, path, perms, component, icon, sort)
VALUES 
(11, 10, '用户列表', 2, 'GET', '/users', 'users:user:list', 'users/user', 'User', 11),
(12, 10, '创建用户', 3, 'POST', '/users', 'users:user:add', '', '', 12),
(13, 10, '获取用户详情', 3, 'GET', '/users/{id}', 'users:user:detail', '', '', 13),
(14, 10, '更新用户基本信息', 3, 'PUT', '/users/{id}', 'users:user:update', '', '', 14),
(15, 10, '更新用户状态', 3, 'PUT', '/users/{id}/status', 'users:user:update:status', '', '', 15),
(16, 10, '更新用户会员等级', 3, 'PUT', '/users/{id}/level', 'users:user:update:level', '', '', 16),
(17, 10, '重置登录密码', 3, 'POST', '/users/{id}/reset-loginpwd', 'users:user:reset:loginpwd', '', '', 17),
(18, 10, '重置支付密码', 3, 'POST', '/users/{id}/reset-paypwd', 'users:user:reset:paypwd', '', '', 18),
(19, 10, '解锁用户', 3, 'POST', '/users/{id}/unlock', 'users:user:unlock', '', '', 19),
(20, 10, '更新用户风险等级', 3, 'PUT', '/users/{id}/risk-level', 'users:user:update:risklevel', '', '', 20),
(21, 10, '删除用户', 3, 'DELETE', '/users/{id}', 'users:user:delete', '', '', 21),
(22, 10, '获取用户安全设置', 3, 'GET', '/users/{id}/security', 'users:user:security:detail', '', '', 22),
(23, 10, '重置用户谷歌2FA', 3, 'POST', '/users/{id}/reset-google2fa', 'users:user:reset:google2fa', '', '', 23),
(24, 10, '实名认证信息列表', 2, 'GET', '/users/{id}/identities', 'users:user:identities:list', 'users/identity', '', 24),
(25, 10, '审核实名认证信息', 3, 'POST', '/users/{id}/identities/review', 'users:user:identities:review', '', '', 25),
(26, 10, '用户银行卡列表', 2, 'GET', '/users/{id}/banks', 'users:user:banks:list', 'users/bank', '', 26),
(27, 10, '获取用户银行卡详情', 3, 'GET', '/users/{id}/banks/{bankId}', 'users:user:bank:detail', '', '', 27),
(28, 10, '添加用户银行卡', 3, 'POST', '/users/{id}/banks', 'users:user:bank:add', '', '', 28),
(29, 10, '更新用户银行卡', 3, 'PUT', '/users/{id}/banks/{bankId}', 'users:user:bank:update', '', '', 29),
(30, 10, '删除用户银行卡', 3, 'DELETE', '/users/{id}/banks/{bankId}', 'users:user:bank:delete', '', '', 30),
(31, 10, '更新用户银行卡状态', 3, 'PUT', '/users/{id}/banks/{bankId}/status', 'users:user:bank:update:status', '', '', 31),
(32, 10, '设置默认用户银行卡', 3, 'POST', '/users/{id}/banks/{bankId}/default', 'users:user:bank:setdefault', '', '', 32);


INSERT INTO sys_menu (id, parent_id, name, menu_type, icon, sort)
VALUES (100, 0, '支付管理', 1, 'Payment', 40);

INSERT INTO sys_menu (id, parent_id, name, menu_type, method, path, perms, component, icon, sort)
VALUES
(101, 100, '平台管理', 2, 'GET', '/payment/platforms', 'payment:platform:list', 'payment/platforms', 'Bank', 41),
(102, 101, '创建支付平台', 3, 'POST', '/payment/platforms', 'payment:platform:add', '', '', 42),
(103, 101, '更新支付平台', 3, 'PUT', '/payment/platforms/{id}', 'payment:platform:update', '', '', 43),
(104, 101, '获取支付平台详情', 3, 'GET', '/payment/platforms/{id}', 'payment:platform:detail', '', '', 44),

(105, 100, '产品管理', 2, 'GET', '/payment/products', 'payment:itick:list', 'payment/products', 'Gold', 45),
(106, 105, '创建支付产品', 3, 'POST', '/payment/products', 'payment:itick:add', '', '', 46),
(107, 105, '更新支付产品', 3, 'PUT', '/payment/products/{id}', 'payment:itick:update', '', '', 47),
(108, 105, '获取支付产品详情', 3, 'GET', '/payment/products/{id}', 'payment:itick:detail', '', '', 48),

(109, 100, '租户开通平台列表', 2, 'GET', '/payment/tenant-platforms', 'payment:tenantplatform:list', 'payment/tenant-platforms', 'Gold', 49),
(110, 109, '开通租户平台', 3, 'POST', '/payment/tenant-platforms', 'payment:tenantplatform:add', '', '', 50),
(111, 109, '更新租户平台', 3, 'PUT', '/payment/tenant-platforms/{id}', 'payment:tenantplatform:update', '', '', 51),
(112, 109, '获取租户平台详情', 3, 'GET', '/payment/tenant-platforms/{id}', 'payment:tenantplatform:detail', '', '', 52),

(113, 100, '租户支付账号列表', 2, 'GET', '/payment/tenant-accounts', 'payment:tenantaccount:list', 'payment/tenant-accounts', 'Gold', 53),
(114, 113, '创建租户支付账号', 3, 'POST', '/payment/tenant-accounts', 'payment:tenantaccount:add', '', '', 54),
(115, 113, '更新租户支付账号', 3, 'PUT', '/payment/tenant-accounts/{id}', 'payment:tenantaccount:update', '', '', 55),
(116, 113, '获取租户支付账号详情', 3, 'GET', '/payment/tenant-accounts/{id}', 'payment:tenantaccount:detail', '', '', 56),

(117, 100, '租户支付通道列表', 2, 'GET', '/payment/tenant-channels', 'payment:tenantchannel:list', 'payment/tenant-channels', 'Gold', 57),
(118, 117, '创建租户支付通道', 3, 'POST', '/payment/tenant-channels', 'payment:tenantchannel:add', '', '', 58),
(119, 117, '更新租户支付通道', 3, 'PUT', '/payment/tenant-channels/{id}', 'payment:tenantchannel:update', '', '', 59),
(120, 117, '获取租户支付通道详情', 3, 'GET', '/payment/tenant-channels/{id}', 'payment:tenantchannel:detail', '', '', 60),

(121, 100, '通道规则列表', 2, 'GET', '/payment/channel-rules', 'payment:channelrule:list', 'payment/channel-rules', 'Gold', 61),
(122, 121, '创建通道规则', 3, 'POST', '/payment/channel-rules', 'payment:channelrule:add', '', '', 62),
(123, 121, '更新通道规则', 3, 'PUT', '/payment/channel-rules/{id}', 'payment:channelrule:update', '', '', 63),

(124, 100, '用户充值统计', 2, 'GET', '/payment/user-recharge-stats', 'payment:userrechargestat:list', 'payment/user-recharge-stats', 'Gold', 65),

(125, 100, '充值管理', 2, 'GET', '/payment/recharge-orders', 'payment:recharge-order:list', 'payment/recharge-orders', 'Gold', 66),
(126, 125, '获取充值订单详情', 3, 'GET', '/payment/recharge-orders/{id}', 'payment:recharge-order:detail', '', '', 67),
(127, 125, '关闭充值订单', 3, 'POST', '/payment/recharge-orders/{id}/close', 'payment:recharge-order:close', '', '', 68),
(128, 125, '人工标记充值订单支付成功', 3, 'POST', '/payment/recharge-orders/{id}/manual-success', 'payment:recharge-order:manualsuccess', '', '', 69),
(129, 125, '重试充值回调', 3, 'POST', '/payment/recharge-orders/{id}/retry-notify', 'payment:recharge-order:retrynotify', '', '', 70),

(130, 100, '充值回调日志列表', 2, 'GET', '/payment/recharge-notify-logs', 'payment:recharge-notifylog:list', 'payment/recharge-notify-logs', 'Gold', 71),
(131, 130, '获取充值回调日志详情', 3, 'GET', '/payment/recharge-notify-logs/{id}', 'payment:recharge-notifylog:detail', '', '', 72),

(132, 100, '提现管理', 2, 'GET', '/payment/withdraw-orders', 'payment:withdraw-order:list', 'payment/withdraw-orders', 'Gold', 73),
(133, 132, '获取提现订单详情', 3, 'GET', '/payment/withdraw-orders/{id}', 'payment:withdraw-order:detail', '', '', 74),
(134, 132, '审核提现订单', 3, 'POST', '/withdraw-orders/:orderNo/audit', 'payment:withdraw-order:audit', '', '', 75),

(135, 100, '提现回调日志列表', 2, 'GET', '/payment/withdraw-notify-logs/list', 'payment:withdraw-notifylog:list', 'payment/withdraw-notify-logs', 'Gold', 76),
(136, 135, '获取提现回调日志详情', 3, 'GET', '/payment/withdraw-notify-logs/{id}', 'payment:withdraw-notifylog:detail', '', '', 77);



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






INSERT INTO sys_menu (id, parent_id, name, menu_type, icon, sort)
VALUES (10000, 0, '系统管理', 1, 'Setting', 10000);

INSERT INTO sys_menu (id, parent_id, name, menu_type, method, path, perms, component, icon, sort)
VALUES (10100, 10000, '用户管理', 2, 'GET', '/users', 'sys:user:list', 'system/users', 'User', 10100);

INSERT INTO sys_menu (id, parent_id, name, menu_type, method, path, perms, sort)
VALUES
(10101, 10100, '新增用户', 3, 'POST', '/sys/users', 'sys:user:add', 10101),
(10102, 10100, '编辑用户', 3, 'PUT', '/sys/users', 'sys:user:update', 10102),
(10103, 10100, '删除用户', 3, 'DELETE', '/sys/users', 'sys:user:delete', 10103),
(10104, 10100, '重置密码', 3, 'POST', '/sys/users/resetPwd', 'sys:user:resetpwd', 10104),
(10105, 10100, '分配角色', 3, 'POST', '/sys/users/assignRoles', 'sys:user:assignrole', 10105),
(10106, 10100, 'Google2FA管理', 3, 'GET', '/sys/users/google2fa', 'sys:user:google2fa', 10106),
(10107, 10100, '2FA初始化', 3, 'POST', '/sys/users/google2fa/init', 'sys:user:2fa:init', 10107),
(10108, 10100, '2FA绑定', 3, 'POST', '/sys/users/google2fa/bind', 'sys:user:2fa:bind', 10108),
(10109, 10100, '2FA启用', 3, 'POST', '/sys/users/google2fa/enable', 'sys:user:2fa:enable', 10109),
(10110, 10100, '2FA禁用', 3, 'POST', '/sys/users/google2fa/disable', 'sys:user:2fa:disable', 10110),
(10111, 10100, '2FA重置', 3, 'POST', '/sys/users/google2fa/reset', 'sys:user:2fa:reset', 10111);


INSERT INTO sys_menu (id, parent_id, name, menu_type, method, path, perms, component, icon, sort)
VALUES (10200, 10000, '角色管理', 2, 'GET', '/roles', 'sys:role:list', 'system/roles', 'Guide', 10200);

INSERT INTO sys_menu (id, parent_id, name, menu_type, method, path, perms, sort)
VALUES
(10201, 10200, '新增角色', 3, 'POST', '/roles', 'sys:role:add', 10201),
(10202, 10200, '编辑角色', 3, 'PUT', '/roles', 'sys:role:update', 10202),
(10203, 10200, '删除角色', 3, 'DELETE', '/roles', 'sys:role:delete', 10203),
(10204, 10200, '菜单授权', 3, 'POST', '/roles/grant', 'sys:role:grant', 10204);

INSERT INTO sys_menu (id, parent_id, name, menu_type, method, path, perms, component, icon, sort)
VALUES (10300, 10000, '菜单管理', 2, 'GET', '/menus', 'sys:menu:list', 'system/menus', 'Menu', 10300);

INSERT INTO sys_menu (id, parent_id, name, menu_type, method, path, perms, sort)
VALUES
(10301, 10300, '新增菜单', 3, 'POST', '/menus', 'sys:menu:add', 10301),
(10302, 10300, '编辑菜单', 3, 'PUT', '/menus', 'sys:menu:update', 10302),
(10303, 10300, '删除菜单', 3, 'DELETE', '/menus', 'sys:menu:delete', 10303);

INSERT INTO sys_menu (id, parent_id, name, menu_type, method, path, perms, component, icon, sort)
VALUES (10400, 10000, '系统配置', 2, 'GET', '/configs', 'sys:config:list', 'system/configs', 'Cpu', 10400);
INSERT INTO sys_menu (id, parent_id, name, menu_type, method, path, perms, sort)
VALUES
(10401, 10400, '新增配置', 3, 'POST', '/configs', 'sys:config:add', 10401),
(10402, 10400, '编辑配置', 3, 'PUT', '/configs', 'sys:config:update', 10402),
(10403, 10400, '删除配置', 3, 'DELETE', '/configs', 'sys:config:delete', 10403);

INSERT INTO sys_menu (id, parent_id, name, menu_type, path, component, icon, sort)
VALUES (10500, 10000, '定时任务', 1, '', '', 'AlarmClock', 10500);
INSERT INTO sys_menu (id, parent_id, name, menu_type, method, path, perms, component, icon, sort)
VALUES (10600, 10500, '定时任务列表', 2, 'GET', '/cronjobs', 'sys:job:list', 'system/cronjobs', 'Clock', 10600);
INSERT INTO sys_menu (id, parent_id, name, menu_type, method, path, perms, sort)
VALUES
(10601, 10600, '新增任务', 3, 'POST', '/jobs', 'sys:job:add', 10601),
(10602, 10600, '编辑任务', 3, 'PUT', '/jobs', 'sys:job:update', 10602),
(10603, 10600, '删除任务', 3, 'DELETE', '/jobs', 'sys:job:delete', 10603),
(10604, 10600, '运行任务', 3, 'POST', '/jobs/run', 'sys:job:run', 10604),
(10605, 10600, '启动任务', 3, 'POST', '/jobs/start', 'sys:job:start', 10605),
(10606, 10600, '停止任务', 3, 'POST', '/jobs/stop', 'sys:job:stop', 10606),
(10607, 10600, '任务处理器', 3, 'GET', '/jobs/handlers', 'sys:job:handlers', 10607);
INSERT INTO sys_menu (id, parent_id, name, menu_type, method, path, perms, component, icon, sort)
VALUES (10700, 10500, '定时任务日志', 2, 'GET', '/cronjobs-log', 'sys:job:log:list', 'system/cronjobs-log', 'Paperclip', 10700);

INSERT INTO sys_menu (id, parent_id, name, menu_type, method, path, perms, component, icon, sort)
VALUES (10800, 10000, '登录日志', 2, 'GET', '/logs/login', 'sys:log:login:list', 'system/login-log', 'Reading', 10800);

INSERT INTO sys_menu (id, parent_id, name, menu_type, method, path, perms, component, icon, sort)
VALUES (10900, 10000, '操作日志', 2, 'GET', '/logs/op', 'sys:log:op:list', 'system/op-log', 'Document', 10900);

INSERT INTO sys_menu (id, parent_id, name, menu_type, method, path, perms, component, icon, sort)
VALUES (11000, 10000, '租户管理', 2, 'GET', '/tenants', 'sys:tenant:list', 'system/tenants', 'Team', 11000);
INSERT INTO sys_menu (id, parent_id, name, menu_type, method, path, perms, sort)
VALUES
(11001, 11000, '新增租户', 3, 'POST', '/tenants', 'sys:tenant:add', 11001),
(11002, 11000, '编辑租户', 3, 'PUT', '/tenants', 'sys:tenant:update', 11002),
(11003, 11000, '删除租户', 3, 'DELETE', '/tenants', 'sys:tenant:delete', 11003),
(11004, 11000, '重置密码', 3, 'POST', '/tenants/resetPwd', 'sys:tenant:resetpwd', 11004),
(11005, 11000, '禁用租户', 3, 'POST', '/tenants/disable', 'sys:tenant:disable', 11005),
(11006, 11000, '启用租户', 3, 'POST', '/tenants/enable', 'sys:tenant:enable', 11006);