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
  config_value TEXT,
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


INSERT INTO sys_menu (id, parent_id, name, menu_type, icon, sort)
VALUES (1, 0, '系统管理', 1, 'Setting', 1);

INSERT INTO sys_menu (id, parent_id, name, menu_type, path, component, sort)
VALUES (100, 1, '用户管理', 2, '/users', 'system/users', 1);

INSERT INTO sys_menu (parent_id, name, menu_type, perms)
VALUES
(100, '新增用户', 3, 'sys:user:add'),
(100, '编辑用户', 3, 'sys:user:update'),
(100, '删除用户', 3, 'sys:user:delete'),
(100, '重置密码', 3, 'sys:user:resetpwd'),
(100, '分配角色', 3, 'sys:user:assignrole'),
(100, 'Google2FA管理', 3, 'sys:user:google2fa');
INSERT INTO sys_menu (parent_id, name, menu_type, perms)
VALUES
(100, '2FA绑定', 3, 'sys:user:2fa:init'),
(100, '2FA启用', 3, 'sys:user:2fa:enable'),
(100, '2FA禁用', 3, 'sys:user:2fa:disable'),
(100, '2FA重置', 3, 'sys:user:2fa:reset');

INSERT INTO sys_menu (id, parent_id, name, menu_type, path, component, sort)
VALUES (200, 1, '角色管理', 2, '/roles', 'system/roles', 2);

INSERT INTO sys_menu (parent_id, name, menu_type, perms)
VALUES
(200, '新增角色', 3, 'sys:role:add'),
(200, '编辑角色', 3, 'sys:role:update'),
(200, '删除角色', 3, 'sys:role:delete'),
(200, '菜单授权', 3, 'sys:role:grant');

INSERT INTO sys_menu (id, parent_id, name, menu_type, path, component, sort)
VALUES (300, 1, '菜单管理', 2, '/menus', 'system/menus', 3);

INSERT INTO sys_menu (parent_id, name, menu_type, perms)
VALUES
(300, '新增菜单', 3, 'sys:menu:add'),
(300, '编辑菜单', 3, 'sys:menu:update'),
(300, '删除菜单', 3, 'sys:menu:delete');

INSERT INTO sys_menu (id, parent_id, name, menu_type, path, component, sort)
VALUES (400, 1, '登录日志', 2, '/logs/login', 'system/login-log', 4);
INSERT INTO sys_menu (id, parent_id, name, menu_type, path, component, sort)
VALUES (500, 1, '操作日志', 2, '/logs/op', 'system/op-log', 5);
