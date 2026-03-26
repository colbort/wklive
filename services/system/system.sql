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

INSERT INTO `sys_role_menu` (`id`, `role_id`, `menu_id`)
VALUES
	('1', '1', '1'),
	('2', '1', '100'),
	('3', '1', '200'),
	('4', '1', '300'),
	('5', '1', '400'),
	('6', '1', '500'),
  ('7', '1', '600'),
  ('8', '1', '700'),
  ('9', '1', '800'),
  ('10', '1', '900'),
	('11', '1', '101'),
	('12', '1', '102'),
	('13', '1', '103'),
	('14', '1', '104'),
	('15', '1', '105'),
	('16', '1', '106'),
	('17', '1', '107'),
	('18', '1', '108'),
	('19', '1', '109'),
	('20', '1', '110'),
  ('21', '1', '111'),
	('22', '1', '201'),
	('23', '1', '202'),
	('24', '1', '203'),
	('25', '1', '204'),
	('26', '1', '301'),
	('27', '1', '302'),
	('28', '1', '303'),
  ('31', '1', '401'),
  ('32', '1', '402'),
  ('33', '1', '403'),
  ('34', '1', '601'),
  ('35', '1', '602'),
  ('36', '1', '603'),
  ('37', '1', '604'),
  ('38', '1', '605'),
  ('39', '1', '606'),
  ('40', '1', '607');


INSERT INTO sys_menu (id, parent_id, name, menu_type, icon, sort)
VALUES (1, 0, '系统管理', 1, 'Setting', 1);

INSERT INTO sys_menu (id, parent_id, name, menu_type, path, component, icon, sort)
VALUES (100, 1, '用户管理', 2, '/users', 'system/users', 'User', 100);

INSERT INTO sys_menu (parent_id, name, menu_type, method, perms, sort)
VALUES
(100, '新增用户', 3, 'POST', 'sys:user:add', 101),
(100, '编辑用户', 3, 'PUT', 'sys:user:update', 102),
(100, '删除用户', 3, 'DELETE', 'sys:user:delete', 103),
(100, '重置密码', 3, 'POST', 'sys:user:resetpwd', 104),
(100, '分配角色', 3, 'POST', 'sys:user:assignrole', 105),
(100, 'Google2FA管理', 3, 'GET', 'sys:user:google2fa', 106),
(100, '2FA初始化', 3, 'POST', 'sys:user:2fa:init', 107),
(100, '2FA绑定', 3, 'POST', 'sys:user:2fa:bind', 108),
(100, '2FA启用', 3, 'POST', 'sys:user:2fa:enable', 109),
(100, '2FA禁用', 3, 'POST', 'sys:user:2fa:disable', 110),
(100, '2FA重置', 3, 'POST', 'sys:user:2fa:reset', 111);


INSERT INTO sys_menu (id, parent_id, name, menu_type, path, component, icon, sort)
VALUES (200, 1, '角色管理', 2, '/roles', 'system/roles', 'Guide', 200);

INSERT INTO sys_menu (parent_id, name, menu_type, method, perms, sort)
VALUES
(200, '新增角色', 3, 'POST', 'sys:role:add', 201),
(200, '编辑角色', 3, 'PUT', 'sys:role:update', 202),
(200, '删除角色', 3, 'DELETE', 'sys:role:delete', 203),
(200, '菜单授权', 3, 'POST', 'sys:role:grant', 204);

INSERT INTO sys_menu (id, parent_id, name, menu_type, path, component, icon, sort)
VALUES (300, 1, '菜单管理', 2, '/menus', 'system/menus', 'Menu', 300);

INSERT INTO sys_menu (parent_id, name, menu_type, method, perms, sort)
VALUES
(300, '新增菜单', 3, 'POST', 'sys:menu:add', 301),
(300, '编辑菜单', 3, 'PUT', 'sys:menu:update', 302),
(300, '删除菜单', 3, 'DELETE', 'sys:menu:delete', 303);

INSERT INTO sys_menu (id, parent_id, name, menu_type, path, component, icon, sort)
VALUES (400, 1, '系统配置', 2, '/configs', 'system/configs', 'Cpu', 400);
INSERT INTO sys_menu (parent_id, name, menu_type, method, perms, sort)
VALUES
(400, '新增配置', 3, 'POST', 'sys:config:add', 401),
(400, '编辑配置', 3, 'PUT', 'sys:config:update', 402),
(400, '删除配置', 3, 'DELETE', 'sys:config:delete', 403);

INSERT INTO sys_menu (id, parent_id, name, menu_type, path, component, icon, sort)
VALUES (500, 1, '定时任务', 1, '', '', 'AlarmClock', 500);
INSERT INTO sys_menu (id, parent_id, name, menu_type, path, component, icon, sort)
VALUES (600, 500, '定时任务列表', 2, '/cronjobs', 'system/cronjobs', 'Clock', 600);
INSERT INTO sys_menu (parent_id, name, menu_type, method, perms, sort)
VALUES
(600, '新增任务', 3, 'POST', 'sys:job:add', 601),
(600, '编辑任务', 3, 'PUT', 'sys:job:update', 602),
(600, '删除任务', 3, 'DELETE', 'sys:job:delete', 603),
(600, '运行任务', 3, 'POST', 'sys:job:run', 604),
(600, '启动任务', 3, 'POST', 'sys:job:start', 605),
(600, '停止任务', 3, 'POST', 'sys:job:stop', 606),
(600, '任务处理器', 3, 'GET', 'sys:job:handlers', 607);
INSERT INTO sys_menu (id, parent_id, name, menu_type, path, component, icon, sort)
VALUES (700, 500, '定时任务日志', 2, '/cronjobs-log', 'system/cronjobs-log', 'Paperclip', 700);

INSERT INTO sys_menu (id, parent_id, name, menu_type, path, component, icon, sort)
VALUES (800, 1, '登录日志', 2, '/logs/login', 'system/login-log', 'Reading', 800);

INSERT INTO sys_menu (id, parent_id, name, menu_type, path, component, icon, sort)
VALUES (900, 1, '操作日志', 2, '/logs/op', 'system/op-log', 'Document', 900);


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
