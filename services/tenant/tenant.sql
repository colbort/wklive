-- =============================
-- 租户
-- =============================
DROP TABLE IF EXISTS tenant_user;
CREATE TABLE tenant_user (
  id BIGINT AUTO_INCREMENT COMMENT '租户ID',

  tenant_code VARCHAR(64) NOT NULL UNIQUE COMMENT '所属租户code',
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
  last_login_at bigint NOT NULL DEFAULT 0,

  create_times bigint NOT NULL DEFAULT 0,
  update_times bigint NOT NULL DEFAULT 0,

  PRIMARY KEY (id),
  INDEX idx_status(status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='租户';


-- =============================
-- 角色
-- =============================
DROP TABLE IF EXISTS tenant_role;
CREATE TABLE tenant_role (
  id BIGINT AUTO_INCREMENT,

  name VARCHAR(64) NOT NULL UNIQUE COMMENT '角色名称',
  code VARCHAR(64) NOT NULL UNIQUE COMMENT '角色标识(如admin)',

  status TINYINT DEFAULT 1 COMMENT '1启用 2禁用',

  remark VARCHAR(255) DEFAULT '',

  create_times bigint NOT NULL DEFAULT 0,
  update_times bigint NOT NULL DEFAULT 0,
  PRIMARY KEY (id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='角色表';


-- =============================
-- 租户-角色
-- =============================
DROP TABLE IF EXISTS tenant_user_role;
CREATE TABLE tenant_user_role (
  id BIGINT AUTO_INCREMENT,
  user_id BIGINT NOT NULL,
  role_id BIGINT NOT NULL,
  PRIMARY KEY (id),
  UNIQUE KEY uk_user_role(user_id, role_id),
  INDEX idx_role(role_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='租户角色关联';


-- =============================
-- 菜单/按钮（核心RBAC）
-- =============================
DROP TABLE IF EXISTS tenant_menu;
CREATE TABLE tenant_menu (
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

  create_times bigint NOT NULL DEFAULT 0,
  update_times bigint NOT NULL DEFAULT 0,

  PRIMARY KEY (id),
  INDEX idx_parent(parent_id),
  INDEX idx_perms(perms)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='菜单权限';


-- =============================
-- 角色-菜单权限
-- =============================
DROP TABLE IF EXISTS tenant_role_menu;
CREATE TABLE tenant_role_menu (
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
DROP TABLE IF EXISTS tenant_login_log;
CREATE TABLE tenant_login_log (
  id BIGINT AUTO_INCREMENT,

  user_id BIGINT,
  username VARCHAR(64),

  ip VARCHAR(64),
  ua VARCHAR(255),

  success TINYINT COMMENT '1成功 0失败',
  msg VARCHAR(255),

  login_at bigint NOT NULL DEFAULT 0,
  
  PRIMARY KEY (id),
  INDEX idx_user(user_id),
  INDEX idx_time(login_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='登录日志';


-- =============================
-- 操作日志
-- =============================
DROP TABLE IF EXISTS tenant_op_log;
CREATE TABLE tenant_op_log (
  id BIGINT AUTO_INCREMENT,

  user_id BIGINT,
  username VARCHAR(64),

  method VARCHAR(16),
  path VARCHAR(255),

  req TEXT,
  resp TEXT,

  ip VARCHAR(64),

  cost_ms INT COMMENT '耗时',

  create_times bigint NOT NULL DEFAULT 0,
  update_times bigint NOT NULL DEFAULT 0,

  PRIMARY KEY (id),
  INDEX idx_user(user_id),
  INDEX idx_time(create_times)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='操作日志';


-- =============================
-- 租户配置（可选）
-- =============================
DROP TABLE IF EXISTS tenant_config;
CREATE TABLE tenant_config (
  id BIGINT AUTO_INCREMENT,

  config_key VARCHAR(64) UNIQUE,
  config_value JSON,
  remark VARCHAR(255),

  create_times bigint NOT NULL DEFAULT 0,
  update_times bigint NOT NULL DEFAULT 0,
  PRIMARY KEY (id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='租户配置';
