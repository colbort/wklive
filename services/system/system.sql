SET NAMES utf8mb4;
SET FOREIGN_KEY_CHECKS = 0;

-- =============================
-- 管理员用户
-- =============================
DROP TABLE IF EXISTS sys_user;
CREATE TABLE sys_user (
  id BIGINT AUTO_INCREMENT COMMENT '用户ID',

  username VARCHAR(64) NOT NULL UNIQUE COMMENT '登录账号',
  password VARCHAR(255) NOT NULL COMMENT 'bcrypt密码',

  nickname VARCHAR(64) DEFAULT '' COMMENT '昵称',
  avatar VARCHAR(255) DEFAULT '' COMMENT '头像',

  status TINYINT DEFAULT 1 COMMENT '状态 1正常 2禁用',

  -- google 2fa
  google_secret VARCHAR(255) DEFAULT '' COMMENT '2FA secret(加密存储)',
  google_enabled TINYINT DEFAULT 0 COMMENT '是否开启2FA',

  perms_ver INT DEFAULT 1 COMMENT '权限版本(角色变化强制token失效)',

  last_login_ip VARCHAR(64),
  last_login_at DATETIME,

  created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

  PRIMARY KEY (id),
  INDEX idx_status(status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='系统管理员';


-- =============================
-- 角色
-- =============================
DROP TABLE IF EXISTS sys_role;
CREATE TABLE sys_role (
  id BIGINT AUTO_INCREMENT,

  name VARCHAR(64) NOT NULL UNIQUE COMMENT '角色名称',
  code VARCHAR(64) NOT NULL UNIQUE COMMENT '角色标识(如admin)',

  status TINYINT DEFAULT 1 COMMENT '1启用 2禁用',

  remark VARCHAR(255) DEFAULT '',

  created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='角色表';


-- =============================
-- 用户-角色
-- =============================
DROP TABLE IF EXISTS sys_user_role;
CREATE TABLE sys_user_role (
  id BIGINT AUTO_INCREMENT,
  user_id BIGINT NOT NULL,
  role_id BIGINT NOT NULL,
  PRIMARY KEY (id),
  UNIQUE KEY uk_user_role(user_id, role_id),
  INDEX idx_role(role_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='用户角色关联';


-- =============================
-- 菜单/按钮（核心RBAC）
-- =============================
DROP TABLE IF EXISTS sys_menu;
CREATE TABLE sys_menu (
  id BIGINT AUTO_INCREMENT,

  parent_id BIGINT DEFAULT 0 COMMENT '父级ID',

  name VARCHAR(64) NOT NULL COMMENT '名称',

  menu_type TINYINT NOT NULL COMMENT '1目录 2菜单 3按钮',

  method VARCHAR(16) DEFAULT '' COMMENT '请求方法 GET POST PUT DELETE',
  path VARCHAR(255) DEFAULT '' COMMENT '路由路径',
  component VARCHAR(255) DEFAULT '' COMMENT '前端组件',

  perms VARCHAR(128) DEFAULT '' COMMENT '按钮权限标识 sys:user:add',

  icon VARCHAR(64) DEFAULT '',
  sort INT DEFAULT 0,

  visible TINYINT DEFAULT 1 COMMENT '1显示 2隐藏',
  status TINYINT DEFAULT 1 COMMENT '1启用 2禁用',

  created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

  PRIMARY KEY (id),
  INDEX idx_parent(parent_id),
  INDEX idx_perms(perms)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='菜单权限';


-- =============================
-- 角色-菜单权限
-- =============================
DROP TABLE IF EXISTS sys_role_menu;
CREATE TABLE sys_role_menu (
  id BIGINT AUTO_INCREMENT,
  role_id BIGINT NOT NULL,
  menu_id BIGINT NOT NULL,
  PRIMARY KEY (id),
  UNIQUE KEY uk_role_menu(role_id, menu_id),
  INDEX idx_menu(menu_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='角色菜单权限';


-- =============================
-- 登录日志
-- =============================
DROP TABLE IF EXISTS sys_login_log;
CREATE TABLE sys_login_log (
  id BIGINT AUTO_INCREMENT,

  user_id BIGINT,
  username VARCHAR(64),

  ip VARCHAR(64),
  ua VARCHAR(255),

  success TINYINT COMMENT '1成功 0失败',
  msg VARCHAR(255),

  login_at DATETIME DEFAULT CURRENT_TIMESTAMP,
  
  PRIMARY KEY (id),
  INDEX idx_user(user_id),
  INDEX idx_time(login_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='登录日志';


-- =============================
-- 操作日志
-- =============================
DROP TABLE IF EXISTS sys_op_log;
CREATE TABLE sys_op_log (
  id BIGINT AUTO_INCREMENT,

  user_id BIGINT,
  username VARCHAR(64),

  method VARCHAR(16),
  path VARCHAR(255),

  req TEXT,
  resp TEXT,

  ip VARCHAR(64),

  cost_ms INT COMMENT '耗时',

  created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

  PRIMARY KEY (id),
  INDEX idx_user(user_id),
  INDEX idx_time(created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='操作日志';


-- =============================
-- 系统配置（可选）
-- =============================
DROP TABLE IF EXISTS sys_config;
CREATE TABLE sys_config (
  id BIGINT AUTO_INCREMENT,

  config_key VARCHAR(64) UNIQUE,
  config_value JSON,
  remark VARCHAR(255),

  created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='系统配置';


CREATE TABLE `sys_job` (
  `id` bigint NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `job_name` varchar(100) NOT NULL COMMENT '任务名称',
  `job_group` varchar(50) DEFAULT 'DEFAULT' COMMENT '任务分组',
  `invoke_target` varchar(500) NOT NULL COMMENT '调用目标',
  `cron_expression` varchar(100) NOT NULL COMMENT 'cron表达式',
  `status` tinyint NOT NULL DEFAULT 1 COMMENT '状态：0停用 1启用',
  `remark` varchar(500) DEFAULT NULL COMMENT '备注',
  `create_by` varchar(64) DEFAULT NULL COMMENT '创建人',
  `create_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `update_by` varchar(64) DEFAULT NULL COMMENT '更新人',
  `update_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  KEY `idx_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='定时任务表';


CREATE TABLE `sys_job_log` (
  `id` bigint NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `job_id` bigint NOT NULL COMMENT '任务ID',
  `job_name` varchar(100) NOT NULL COMMENT '任务名称',
  `invoke_target` varchar(500) NOT NULL COMMENT '调用目标',
  `cron_expression` varchar(100) DEFAULT NULL COMMENT 'cron表达式',
  `status` tinyint NOT NULL COMMENT '执行状态：0失败 1成功',
  `message` varchar(2000) DEFAULT NULL COMMENT '执行信息',
  `exception_info` text COMMENT '异常信息',
  `start_time` datetime DEFAULT NULL COMMENT '开始时间',
  `end_time` datetime DEFAULT NULL COMMENT '结束时间',
  `create_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  PRIMARY KEY (`id`),
  KEY `idx_job_id` (`job_id`),
  KEY `idx_create_time` (`create_time`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='定时任务日志表';


CREATE TABLE `sys_tenant` (
  `id` bigint NOT NULL AUTO_INCREMENT COMMENT '租户ID',
  `tenant_code` varchar(64) NOT NULL COMMENT '租户编码',
  `tenant_username` varchar(64) NOT NULL COMMENT '租户管理员账号',
  `tenant_password` varchar(255) NOT NULL COMMENT '租户管理员密码（bcrypt加密）',
  `tenant_name` varchar(128) NOT NULL COMMENT '租户名称',
  `status` tinyint NOT NULL DEFAULT 1 COMMENT '状态：1正常 2禁用',
  `expire_time` datetime DEFAULT NULL COMMENT '到期时间',
  `contact_name` varchar(64) DEFAULT NULL COMMENT '联系人',
  `contact_phone` varchar(32) DEFAULT NULL COMMENT '联系电话',
  `login_ip` varchar(64) DEFAULT NULL COMMENT '最后登录IP',
  `login_time` datetime DEFAULT NULL COMMENT '最后登录时间',
  `login_count` int DEFAULT 0 COMMENT '登录次数',
  `remark` varchar(255) DEFAULT NULL COMMENT '备注',
  `create_by` varchar(64) DEFAULT NULL COMMENT '创建人',
  `create_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `update_by` varchar(64) DEFAULT NULL COMMENT '更新人',
  `update_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_tenant_code` (`tenant_code`),
  KEY `idx_status` (`status`),
  KEY `idx_expire_time` (`expire_time`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='租户表';


SET FOREIGN_KEY_CHECKS = 1;

INSERT INTO `sys_role` (`id`, `name`, `code`, `status`, `remark`)
VALUES
	('1', '超级管理员', 'admin', '1', '');
INSERT INTO `sys_user` (`id`, `username`, `password`, `nickname`, `avatar`, `status`, `google_secret`, `google_enabled`, `perms_ver`, `last_login_ip`, `last_login_at`)
VALUES
	('1', 'admin', '$2a$10$KdJbtCoUCeO.jcI9LJb6me4YAnMt8JScsCWyA9FEPfuaz4bRCfMee', '超级管理员', '', '1', '', '0', '1', NULL, NULL);
INSERT INTO `sys_user_role` (`user_id`, `role_id`)
VALUES
	('1', '1');

INSERT INTO `sys_role_menu` (`role_id`, `menu_id`)
VALUES
	('1', '90'),
	('1', '100'),
	('1', '200'),
	('1', '300'),
	('1', '400'),
	('1', '500'),
  ('1', '600'),
  ('1', '700'),
  ('1', '800'),
  ('1', '900'),
  ('1', '1000'),
	('1', '101'),
	('1', '102'),
	('1', '103'),
	('1', '104'),
	('1', '105'),
	('1', '106'),
	('1', '107'),
	('1', '108'),
	('1', '109'),
	('1', '110'),
  ('1', '111'),
	('1', '201'),
	('1', '202'),
	('1', '203'),
	('1', '204'),
	('1', '301'),
	('1', '302'),
	('1', '303'),
  ('1', '401'),
  ('1', '402'),
  ('1', '403'),
  ('1', '601'),
  ('1', '602'),
  ('1', '603'),
  ('1', '604'),
  ('1', '605'),
  ('1', '606'),
  ('1', '607'),
  ('1', '1001'),
  ('1', '1002'),
  ('1', '1003'),
  ('1', '1004'),
  ('1', '1005'),
  ('1', '1006');


INSERT INTO sys_menu (id, parent_id, name, menu_type, icon, sort)
VALUES (90, 0, '系统管理', 1, 'Setting', 90);

INSERT INTO sys_menu (id, parent_id, name, menu_type, method, path, perms, component, icon, sort)
VALUES (100, 90, '用户管理', 2, 'GET', '/users', 'sys:user:list', 'system/users', 'User', 100);

INSERT INTO sys_menu (parent_id, name, menu_type, method, path, perms, sort)
VALUES
(100, '新增用户', 3, 'POST', '/sys/users', 'sys:user:add', 101),
(100, '编辑用户', 3, 'PUT', '/sys/users', 'sys:user:update', 102),
(100, '删除用户', 3, 'DELETE', '/sys/users', 'sys:user:delete', 103),
(100, '重置密码', 3, 'POST', '/sys/users/resetPwd', 'sys:user:resetpwd', 104),
(100, '分配角色', 3, 'POST', '/sys/users/assignRoles', 'sys:user:assignrole', 105),
(100, 'Google2FA管理', 3, 'GET', '/sys/users/google2fa', 'sys:user:google2fa', 106),
(100, '2FA初始化', 3, 'POST', '/sys/users/google2fa/init', 'sys:user:2fa:init', 107),
(100, '2FA绑定', 3, 'POST', '/sys/users/google2fa/bind', 'sys:user:2fa:bind', 108),
(100, '2FA启用', 3, 'POST', '/sys/users/google2fa/enable', 'sys:user:2fa:enable', 109),
(100, '2FA禁用', 3, 'POST', '/sys/users/google2fa/disable', 'sys:user:2fa:disable', 110),
(100, '2FA重置', 3, 'POST', '/sys/users/google2fa/reset', 'sys:user:2fa:reset', 111);


INSERT INTO sys_menu (id, parent_id, name, menu_type, method, path, perms, component, icon, sort)
VALUES (200, 90, '角色管理', 2, 'GET', '/roles', 'sys:role:list', 'system/roles', 'Guide', 200);

INSERT INTO sys_menu (parent_id, name, menu_type, method, path, perms, sort)
VALUES
(200, '新增角色', 3, 'POST', '/roles', 'sys:role:add', 201),
(200, '编辑角色', 3, 'PUT', '/roles', 'sys:role:update', 202),
(200, '删除角色', 3, 'DELETE', '/roles', 'sys:role:delete', 203),
(200, '菜单授权', 3, 'POST', '/roles/grant', 'sys:role:grant', 204);

INSERT INTO sys_menu (id, parent_id, name, menu_type, method, path, perms, component, icon, sort)
VALUES (300, 90, '菜单管理', 2, 'GET', '/menus', 'sys:menu:list', 'system/menus', 'Menu', 300);

INSERT INTO sys_menu (parent_id, name, menu_type, method, path, perms, sort)
VALUES
(300, '新增菜单', 3, 'POST', '/menus', 'sys:menu:add', 301),
(300, '编辑菜单', 3, 'PUT', '/menus', 'sys:menu:update', 302),
(300, '删除菜单', 3, 'DELETE', '/menus', 'sys:menu:delete', 303);

INSERT INTO sys_menu (id, parent_id, name, menu_type, method, path, perms, component, icon, sort)
VALUES (400, 90, '系统配置', 2, 'GET', '/configs', 'sys:config:list', 'system/configs', 'Cpu', 400);
INSERT INTO sys_menu (parent_id, name, menu_type, method, path, perms, sort)
VALUES
(400, '新增配置', 3, 'POST', '/configs', 'sys:config:add', 401),
(400, '编辑配置', 3, 'PUT', '/configs', 'sys:config:update', 402),
(400, '删除配置', 3, 'DELETE', '/configs', 'sys:config:delete', 403);

INSERT INTO sys_menu (id, parent_id, name, menu_type, path, component, icon, sort)
VALUES (500, 90, '定时任务', 1, '', '', 'AlarmClock', 500);
INSERT INTO sys_menu (id, parent_id, name, menu_type, method, path, perms, component, icon, sort)
VALUES (600, 500, '定时任务列表', 2, 'GET', '/cronjobs', 'sys:job:list', 'system/cronjobs', 'Clock', 600);
INSERT INTO sys_menu (parent_id, name, menu_type, method, path, perms, sort)
VALUES
(600, '新增任务', 3, 'POST', '/jobs', 'sys:job:add', 601),
(600, '编辑任务', 3, 'PUT', '/jobs', 'sys:job:update', 602),
(600, '删除任务', 3, 'DELETE', '/jobs', 'sys:job:delete', 603),
(600, '运行任务', 3, 'POST', '/jobs/run', 'sys:job:run', 604),
(600, '启动任务', 3, 'POST', '/jobs/start', 'sys:job:start', 605),
(600, '停止任务', 3, 'POST', '/jobs/stop', 'sys:job:stop', 606),
(600, '任务处理器', 3, 'GET', '/jobs/handlers', 'sys:job:handlers', 607);
INSERT INTO sys_menu (id, parent_id, name, menu_type, method, path, perms, component, icon, sort)
VALUES (700, 500, '定时任务日志', 2, 'GET', '/cronjobs-log', 'sys:job:log:list', 'system/cronjobs-log', 'Paperclip', 700);

INSERT INTO sys_menu (id, parent_id, name, menu_type, method, path, perms, component, icon, sort)
VALUES (800, 90, '登录日志', 2, 'GET', '/logs/login', 'sys:log:login:list', 'system/login-log', 'Reading', 800);

INSERT INTO sys_menu (id, parent_id, name, menu_type, method, path, perms, component, icon, sort)
VALUES (900, 90, '操作日志', 2, 'GET', '/logs/op', 'sys:log:op:list', 'system/op-log', 'Document', 900);

INSERT INTO sys_menu (id, parent_id, name, menu_type, method, path, perms, component, icon, sort)
VALUES (1000, 90, '租户管理', 2, 'GET', '/tenants', 'sys:tenant:list', 'system/tenants', 'Team', 1000);
INSERT INTO sys_menu (parent_id, name, menu_type, method, path, perms, sort)
VALUES
(1000, '新增租户', 3, 'POST', '/tenants', 'sys:tenant:add', 1001),
(1000, '编辑租户', 3, 'PUT', '/tenants', 'sys:tenant:update', 1002),
(1000, '删除租户', 3, 'DELETE', '/tenants', 'sys:tenant:delete', 1003),
(1000, '重置密码', 3, 'POST', '/tenants/resetPwd', 'sys:tenant:resetpwd', 1004),
(1000, '禁用租户', 3, 'POST', '/tenants/disable', 'sys:tenant:disable', 1005),
(1000, '启用租户', 3, 'POST', '/tenants/enable', 'sys:tenant:enable', 1006);


INSERT INTO sys_menu (id, parent_id, name, menu_type, icon, sort)
VALUES (10, 0, '用户管理', 1, 'Users', 10);
INSERT INTO sys_menu (id, parent_id, name, menu_type, method, path, perms, component, icon, sort)
VALUES 
(11, 10, '用户列表', 2, 'GET', '/users', 'app:user:list', 'system/users', 'User', 11),
(12, 10, '创建用户', 3, 'POST', '/users', 'app:user:add', '', '', 12),
(13, 10, '获取用户详情', 3, 'GET', '/users/{id}', 'app:user:detail', '', '', 13),
(14, 10, '更新用户基本信息', 3, 'PUT', '/users/{id}', 'app:user:update', '', '', 14),
(15, 10, '更新用户状态', 3, 'PUT', '/users/{id}/status', 'app:user:update:status', '', '', 15),
(16, 10, '更新用户会员等级', 3, 'PUT', '/users/{id}/level', 'app:user:update:level', '', '', 16),
(17, 10, '重置登录密码', 3, 'POST', '/users/{id}/reset-loginpwd', 'app:user:reset:loginpwd', '', '', 17),
(18, 10, '重置支付密码', 3, 'POST', '/users/{id}/reset-paypwd', 'app:user:reset:paypwd', '', '', 18),
(19, 10, '解锁用户', 3, 'POST', '/users/{id}/unlock', 'app:user:unlock', '', '', 19),
(20, 10, '更新用户风险等级', 3, 'PUT', '/users/{id}/risk-level', 'app:user:update:risklevel', '', '', 20),
(21, 10, '删除用户', 3, 'DELETE', '/users/{id}', 'app:user:delete', '', '', 21),
(22, 10, '获取用户安全设置', 3, 'GET', '/users/{id}/security', 'app:user:security:detail', '', '', 22),
(23, 10, '重置用户谷歌2FA', 3, 'POST', '/users/{id}/reset-google2fa', 'app:user:reset:google2fa', '', '', 23),
(24, 10, '实名认证信息列表', 2, 'GET', '/users/{id}/identities', 'app:user:identities:list', '', '', 24),
(25, 10, '审核实名认证信息', 3, 'POST', '/users/{id}/identities/review', 'app:user:identities:review', '', '', 25),
(26, 10, '用户银行卡列表', 2, 'GET', '/users/{id}/banks', 'app:user:banks:list', '', '', 26),
(27, 10, '获取用户银行卡详情', 3, 'GET', '/users/{id}/banks/{bankId}', 'app:user:bank:detail', '', '', 27),
(28, 10, '添加用户银行卡', 3, 'POST', '/users/{id}/banks', 'app:user:bank:add', '', '', 28),
(29, 10, '更新用户银行卡', 3, 'PUT', '/users/{id}/banks/{bankId}', 'app:user:bank:update', '', '', 29),
(30, 10, '删除用户银行卡', 3, 'DELETE', '/users/{id}/banks/{bankId}', 'app:user:bank:delete', '', '', 30),
(31, 10, '更新用户银行卡状态', 3, 'PUT', '/users/{id}/banks/{bankId}/status', 'app:user:bank:update:status', '', '', 31),
(32, 10, '设置默认用户银行卡', 3, 'POST', '/users/{id}/banks/{bankId}/default', 'app:user:bank:setdefault', '', '', 32);


INSERT INTO sys_menu (id, parent_id, name, menu_type, icon, sort)
VALUES (40, 0, '支付管理', 1, 'Payment', 40);
INSERT INTO sys_menu (id, parent_id, name, menu_type, method, path, perms, component, icon, sort)
VALUES 
(41, 40, '平台管理', 2, 'GET', '/pay/platforms', 'sys:pay:platform:list', 'system/pay-platforms', 'Bank', 41),
(42, 40, '创建支付平台', 3, 'POST', '/pay/platforms', 'sys:pay:platform:add', '', '', 42),
(43, 40, '更新支付平台', 3, 'PUT', '/pay/platforms/{id}', 'sys:pay:platform:update', '', '', 43),
(44, 40, '获取支付平台详情', 3, 'GET', '/pay/platforms/{id}', 'sys:pay:platform:detail', '', '', 44),
(45, 40, '产品管理', 2, 'GET', '/pay/products', 'sys:pay:product:list', 'system/pay-products', 'Gold', 45),
(46, 45, '创建支付产品', 3, 'POST', '/pay/products', 'sys:pay:product:add', '', '', 46),
(47, 45, '更新支付产品', 3, 'PUT', '/pay/products/{id}', 'sys:pay:product:update', '', '', 47),
(48, 45, '获取支付产品详情', 3, 'GET', '/pay/products/{id}', 'sys:pay:product:detail', '', '', 48),
(49, 40, '租户开通平台列表', 2, 'GET', '/pay/tenant-platforms', 'sys:pay:tenantplatform:list', 'system/pay-tenant-platforms', 'Gold', 49),
(50, 49, '开通租户平台', 3, 'POST', '/pay/tenant-platforms', 'sys:pay:tenantplatform:add', '', '', 50),
(51, 49, '更新租户平台', 3, 'PUT', '/pay/tenant-platforms/{id}', 'sys:pay:tenantplatform:update', '', '', 51),
(52, 49, '获取租户平台详情', 3, 'GET', '/pay/tenant-platforms/{id}', 'sys:pay:tenantplatform:detail', '', '', 52),
(53, 40, '租户支付账号列表', 2, 'GET', '/pay/tenant-accounts', 'sys:pay:tenantaccount:list', 'system/pay-tenant-accounts', 'Gold', 53),
(54, 53, '创建租户支付账号', 3, 'POST', '/pay/tenant-accounts', 'sys:pay:tenantaccount:add', '', '', 54),
(55, 53, '更新租户支付账号', 3, 'PUT', '/pay/tenant-accounts/{id}', 'sys:pay:tenantaccount:update', '', '', 55),
(56, 53, '获取租户支付账号详情', 3, 'GET', '/pay/tenant-accounts/{id}', 'sys:pay:tenantaccount:detail', '', '', 56),
(57, 40, '租户支付通道列表', 2, 'GET', '/pay/tenant-channels', 'sys:pay:tenantchannel:list', 'system/pay-tenant-channels', 'Gold', 57),
(58, 57, '创建租户支付通道', 3, 'POST', '/pay/tenant-channels', 'sys:pay:tenantchannel:add', '', '', 58),
(59, 57, '更新租户支付通道', 3, 'PUT', '/pay/tenant-channels/{id}', 'sys:pay:tenantchannel:update', '', '', 59),
(60, 57, '获取租户支付通道详情', 3, 'GET', '/pay/tenant-channels/{id}', 'sys:pay:tenantchannel:detail', '', '', 60),
(61, 40, '通道规则列表', 2, 'GET', '/pay/channel-rules', 'sys:pay:channelrule:list', 'system/pay-channel-rules', 'Gold', 61),
(62, 61, '创建通道规则', 3, 'POST', '/pay/channel-rules', 'sys:pay:channelrule:add', '', '', 62),
(63, 61, '更新通道规则', 3, 'PUT', '/pay/channel-rules/{id}', 'sys:pay:channelrule:update', '', '', 63),
(65, 40, '用户充值统计', 2, 'GET', '/pay/user-recharge-stats', 'sys:pay:userrechargestat:list', 'system/pay-user-recharge-stats', 'Gold', 65),
(66, 40, '订单列表', 2, 'GET', '/pay/orders', 'sys:pay:order:list', 'system/pay-orders', 'Gold', 66),
(67, 66, '获取订单详情', 3, 'GET', '/pay/orders/{id}', 'sys:pay:order:detail', '', '', 67),
(68, 66, '关闭订单', 3, 'POST', '/pay/orders/{id}/close', 'sys:pay:order:close', '', '', 68),
(69, 66, '人工标记订单支付成功', 3, 'POST', '/pay/orders/{id}/manual-success', 'sys:pay:order:manualsuccess', '', '', 69),
(70, 66, '重试回调', 3, 'POST', '/pay/orders/{id}/retry-notify', 'sys:pay:order:retrynotify', '', '', 70),
(71, 40, '回调日志列表', 2, 'GET', '/pay/notify-logs', 'sys:pay:notifylog:list', 'system/pay-notify-logs', 'Gold', 71),
(72, 71, '获取回调日志详情', 3, 'GET', '/pay/notify-logs/{id}', 'sys:pay:notifylog:detail', '', '', 72);