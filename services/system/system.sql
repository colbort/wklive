-- =============================
-- 管理员用户（统一用户表）
-- 说明：
-- 1. tenant_id = 0       -> 系统侧用户
-- 2. tenant_id > 0       -> 租户侧用户
-- 3. user_type = 1       -> 系统管理员
-- 4. user_type = 2       -> 租户主账号
-- 5. user_type = 3       -> 租户管理员
-- =============================
DROP TABLE IF EXISTS sys_user;
CREATE TABLE sys_user (
  id BIGINT AUTO_INCREMENT COMMENT '用户ID',

  tenant_id BIGINT NOT NULL DEFAULT 0 COMMENT '所属租户ID：0=系统侧，>0=租户ID',
  user_type TINYINT NOT NULL DEFAULT 1 COMMENT '用户类型：1系统管理员 2租户主账号 3租户管理员',
  is_owner TINYINT NOT NULL DEFAULT 1 COMMENT '是否租户主账号：1否 2是',

  username VARCHAR(64) NOT NULL DEFAULT '' COMMENT '登录账号',
  password VARCHAR(255) NOT NULL DEFAULT '' COMMENT 'bcrypt密码',

  nickname VARCHAR(64) DEFAULT '' COMMENT '昵称',
  avatar VARCHAR(255) DEFAULT '' COMMENT '头像',

  enabled TINYINT DEFAULT 1 COMMENT '启用开关：1启用 2禁用',

  -- google 2fa
  google_secret VARCHAR(255) DEFAULT '' COMMENT '2FA secret(加密存储)',
  google_enabled TINYINT DEFAULT 2 COMMENT 'Google2FA开关：1启用 2禁用',

  perms_ver INT DEFAULT 1 COMMENT '权限版本(角色变化强制token失效)',

  last_login_ip VARCHAR(64) DEFAULT '' COMMENT '最后登录IP',
  last_login_at BIGINT NOT NULL DEFAULT 0 COMMENT '最后登录时间',

  create_by BIGINT NOT NULL DEFAULT 0 COMMENT '创建人ID',
  create_times BIGINT NOT NULL DEFAULT 0,
  update_times BIGINT NOT NULL DEFAULT 0,

  PRIMARY KEY (id),
  UNIQUE KEY uk_tenant_username (tenant_id, username),
  INDEX idx_tenant_id(tenant_id),
  INDEX idx_user_type(user_type),
  INDEX idx_owner(is_owner),
  INDEX idx_enabled(enabled)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='统一用户表';


-- =============================
-- 角色
-- 说明：
-- 1. tenant_id = 0  -> 系统角色
-- 2. tenant_id > 0  -> 某个租户自己的角色
-- =============================
DROP TABLE IF EXISTS sys_role;
CREATE TABLE sys_role (
  id BIGINT AUTO_INCREMENT COMMENT '角色ID',

  tenant_id BIGINT NOT NULL DEFAULT 0 COMMENT '所属租户ID：0=系统角色，>0=租户角色',

  name VARCHAR(64) NOT NULL DEFAULT '' COMMENT '角色名称',
  code VARCHAR(64) NOT NULL DEFAULT '' COMMENT '角色标识(如admin)',

  enabled TINYINT DEFAULT 1 COMMENT '启用开关：1启用 2禁用',

  remark VARCHAR(255) DEFAULT '',

  create_times BIGINT NOT NULL DEFAULT 0,
  update_times BIGINT NOT NULL DEFAULT 0,

  PRIMARY KEY (id),
  UNIQUE KEY uk_tenant_role_name(tenant_id, name),
  UNIQUE KEY uk_tenant_role_code(tenant_id, code),
  INDEX idx_tenant_id(tenant_id),
  INDEX idx_enabled(enabled)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='角色表';


-- =============================
-- 用户-角色
-- =============================
DROP TABLE IF EXISTS sys_user_role;
CREATE TABLE sys_user_role (
  id BIGINT AUTO_INCREMENT COMMENT '主键ID',

  tenant_id BIGINT NOT NULL DEFAULT 0 COMMENT '所属租户ID：0=系统侧，>0=租户ID',
  user_id BIGINT NOT NULL DEFAULT 0 COMMENT '用户ID',
  role_id BIGINT NOT NULL DEFAULT 0 COMMENT '角色ID',

  PRIMARY KEY (id),
  UNIQUE KEY uk_user_role(user_id, role_id),
  INDEX idx_tenant_id(tenant_id),
  INDEX idx_user(user_id),
  INDEX idx_role(role_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='用户角色关联';


-- =============================
-- 菜单/按钮（核心RBAC）
-- 说明：
-- 1. scope = 1 -> 系统端菜单
-- 2. scope = 2 -> 租户端菜单
-- =============================
DROP TABLE IF EXISTS sys_menu;
CREATE TABLE sys_menu (
  id BIGINT AUTO_INCREMENT,

  parent_id BIGINT DEFAULT 0 COMMENT '父级ID',

  name VARCHAR(64) NOT NULL DEFAULT '' COMMENT '名称',

  menu_type TINYINT NOT NULL DEFAULT 0 COMMENT '菜单类型：0未知 1目录 2菜单 3按钮',

  method VARCHAR(16) DEFAULT '' COMMENT '请求方法 GET POST PUT DELETE',
  path VARCHAR(255) DEFAULT '' COMMENT '路由路径',
  component VARCHAR(255) DEFAULT '' COMMENT '前端组件',

  perms VARCHAR(128) DEFAULT '' COMMENT '按钮权限标识 sys:user:add',

  icon VARCHAR(64) DEFAULT '',
  sort INT DEFAULT 0,

  visible TINYINT DEFAULT 1 COMMENT '显示开关：1显示 2隐藏',
  enabled TINYINT DEFAULT 1 COMMENT '启用开关：1启用 2禁用',

  create_times BIGINT NOT NULL DEFAULT 0,
  update_times BIGINT NOT NULL DEFAULT 0,

  PRIMARY KEY (id),
  INDEX idx_parent(parent_id),
  INDEX idx_perms(perms)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='菜单权限';


-- =============================
-- 角色-菜单权限
-- =============================
DROP TABLE IF EXISTS sys_role_menu;
CREATE TABLE sys_role_menu (
  id BIGINT AUTO_INCREMENT COMMENT '主键ID',

  tenant_id BIGINT NOT NULL DEFAULT 0 COMMENT '所属租户ID：0=系统侧，>0=租户ID',
  role_id BIGINT NOT NULL DEFAULT 0 COMMENT '角色ID',
  menu_id BIGINT NOT NULL DEFAULT 0 COMMENT '菜单ID',

  PRIMARY KEY (id),
  UNIQUE KEY uk_role_menu(role_id, menu_id),
  INDEX idx_tenant_id(tenant_id),
  INDEX idx_role(role_id),
  INDEX idx_menu(menu_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='角色菜单权限';


-- =============================
-- 登录日志
-- 说明：
-- 增加 tenant_id，方便系统侧/租户侧隔离查询
-- =============================
DROP TABLE IF EXISTS sys_login_log;
CREATE TABLE sys_login_log (
  id BIGINT AUTO_INCREMENT,

  tenant_id BIGINT NOT NULL DEFAULT 0 COMMENT '所属租户ID：0=系统侧，>0=租户ID',

  user_id BIGINT,
  username VARCHAR(64),

  ip VARCHAR(64),
  ua VARCHAR(255),

  success TINYINT COMMENT '1成功 0失败',
  msg VARCHAR(255),

  login_at BIGINT NOT NULL DEFAULT 0,
  
  PRIMARY KEY (id),
  INDEX idx_tenant_id(tenant_id),
  INDEX idx_user(user_id),
  INDEX idx_time(login_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='登录日志';


-- =============================
-- 操作日志
-- 说明：
-- 增加 tenant_id，方便系统侧/租户侧隔离查询
-- =============================
DROP TABLE IF EXISTS sys_op_log;

CREATE TABLE sys_op_log (
  id BIGINT AUTO_INCREMENT,

  tenant_id BIGINT NOT NULL DEFAULT 0 COMMENT '所属租户ID：0=系统侧，>0=租户ID',

  user_id BIGINT DEFAULT 0 COMMENT '操作人ID',
  username VARCHAR(64) DEFAULT '' COMMENT '操作人账号',

  module VARCHAR(64) DEFAULT '' COMMENT '模块',
  action VARCHAR(64) DEFAULT '' COMMENT '操作',

  method VARCHAR(16) DEFAULT '' COMMENT '请求方法',
  path VARCHAR(255) DEFAULT '' COMMENT '请求路径',

  req TEXT COMMENT '请求参数',
  resp TEXT COMMENT '响应内容',

  ip VARCHAR(64) DEFAULT '' COMMENT 'IP',

  cost_ms INT DEFAULT 0 COMMENT '耗时',

  create_times BIGINT NOT NULL DEFAULT 0,
  update_times BIGINT NOT NULL DEFAULT 0,

  PRIMARY KEY (id),
  INDEX idx_tenant_id(tenant_id),
  INDEX idx_user(user_id),
  INDEX idx_time(create_times),
  INDEX idx_path(path)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='操作日志';


-- =============================
-- 系统配置（系统级）
-- 说明：
-- 这张表保留为系统配置，不做租户隔离
-- 如需租户配置，建议单独增加 tenant_config
-- =============================
DROP TABLE IF EXISTS sys_config;
CREATE TABLE sys_config (
  id BIGINT AUTO_INCREMENT,
  tenant_id BIGINT NOT NULL DEFAULT 0 COMMENT '所属租户ID：0=系统侧，>0=租户ID',
  config_key VARCHAR(64),
  config_value JSON,
  remark VARCHAR(255),

  create_times BIGINT NOT NULL DEFAULT 0,
  update_times BIGINT NOT NULL DEFAULT 0,

  PRIMARY KEY (id),
  UNIQUE KEY uk_config_key(tenant_id, config_key)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='系统配置';

-- =============================
-- 验证码发送记录
-- =============================
DROP TABLE IF EXISTS sys_verification_code_record;
CREATE TABLE sys_verification_code_record (
  id BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  tenant_id BIGINT NOT NULL DEFAULT 0 COMMENT '所属租户ID：0=系统侧，>0=租户ID',
  channel TINYINT NOT NULL DEFAULT 0 COMMENT '发送渠道：0未知 1邮箱 2手机短信',
  target VARCHAR(128) NOT NULL DEFAULT '' COMMENT '发送目标：邮箱或手机号',
  scene SMALLINT NOT NULL DEFAULT 0 COMMENT '业务场景：0未知 1注册 2登录 3重置密码 4绑定邮箱 5绑定手机 6修改密码 7提现 100测试',
  code VARCHAR(16) NOT NULL DEFAULT '' COMMENT '验证码',
  status TINYINT NOT NULL DEFAULT 0 COMMENT '发送状态：0未知 1成功 2失败',
  provider VARCHAR(64) DEFAULT NULL COMMENT '服务商',
  error_message VARCHAR(512) DEFAULT NULL COMMENT '失败原因',
  create_times BIGINT NOT NULL DEFAULT 0,
  update_times BIGINT NOT NULL DEFAULT 0,
  PRIMARY KEY (id),
  KEY idx_tenant_channel_target (tenant_id, channel, target),
  KEY idx_scene_status (scene, status),
  KEY idx_create_times (create_times)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='验证码发送记录';


-- =============================
-- 定时任务表
-- 说明：
-- 默认为系统级任务，不开放给租户
-- =============================
DROP TABLE IF EXISTS sys_job;
CREATE TABLE sys_job (
  id BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  job_name VARCHAR(100) NOT NULL DEFAULT '' COMMENT '任务名称',
  job_group VARCHAR(50) DEFAULT 'DEFAULT' COMMENT '任务分组',
  invoke_target VARCHAR(500) NOT NULL DEFAULT '' COMMENT '调用目标',
  cron_expression VARCHAR(100) NOT NULL DEFAULT '' COMMENT 'cron表达式',
  status TINYINT NOT NULL DEFAULT 1 COMMENT '任务状态：0停用 1启用',
  remark VARCHAR(500) DEFAULT NULL COMMENT '备注',
  create_by VARCHAR(64) DEFAULT NULL COMMENT '创建人',
  create_times BIGINT NOT NULL DEFAULT 0 COMMENT '创建时间',
  update_by VARCHAR(64) DEFAULT NULL COMMENT '更新人',
  update_times BIGINT NOT NULL DEFAULT 0 COMMENT '更新时间',
  PRIMARY KEY (id),
  KEY idx_status (status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='定时任务表';


-- =============================
-- 定时任务日志表
-- =============================
DROP TABLE IF EXISTS sys_job_log;
CREATE TABLE sys_job_log (
  id BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  job_id BIGINT NOT NULL DEFAULT 0 COMMENT '任务ID',
  job_name VARCHAR(100) NOT NULL DEFAULT '' COMMENT '任务名称',
  invoke_target VARCHAR(500) NOT NULL DEFAULT '' COMMENT '调用目标',
  cron_expression VARCHAR(100) DEFAULT NULL COMMENT 'cron表达式',
  status TINYINT NOT NULL DEFAULT 0 COMMENT '执行状态：0失败 1成功',
  message VARCHAR(2000) DEFAULT NULL COMMENT '执行信息',
  exception_info TEXT COMMENT '异常信息',
  start_time BIGINT DEFAULT 0 COMMENT '开始时间',
  end_time BIGINT DEFAULT 0 COMMENT '结束时间',
  create_times BIGINT NOT NULL DEFAULT 0 COMMENT '创建时间',
  PRIMARY KEY (id),
  KEY idx_job_id (job_id),
  KEY idx_create_times (create_times)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='定时任务日志表';


-- =============================
-- 租户表
-- 说明：
-- 这里只保留租户资料，不再存登录账号密码
-- 租户主账号统一存到 sys_user
-- =============================
DROP TABLE IF EXISTS sys_tenant;
CREATE TABLE sys_tenant (
  id BIGINT NOT NULL AUTO_INCREMENT COMMENT '租户ID',
  tenant_code VARCHAR(64) NOT NULL DEFAULT '' COMMENT '租户编码',
  tenant_name VARCHAR(128) NOT NULL DEFAULT '' COMMENT '租户名称',
  enabled TINYINT NOT NULL DEFAULT 1 COMMENT '启用开关：1启用 2禁用',
  expire_time BIGINT DEFAULT 0 COMMENT '到期时间',
  contact_name VARCHAR(64) DEFAULT NULL COMMENT '联系人',
  contact_phone VARCHAR(32) DEFAULT NULL COMMENT '联系电话',
  login_ip VARCHAR(64) DEFAULT NULL COMMENT '最后登录IP',
  login_time BIGINT DEFAULT 0 COMMENT '最后登录时间',
  login_count INT DEFAULT 0 COMMENT '登录次数',
  remark VARCHAR(255) DEFAULT NULL COMMENT '备注',
  create_by VARCHAR(64) DEFAULT NULL COMMENT '创建人',
  create_times BIGINT NOT NULL DEFAULT 0 COMMENT '创建时间',
  update_by VARCHAR(64) DEFAULT NULL COMMENT '更新人',
  update_times BIGINT NOT NULL DEFAULT 0 COMMENT '更新时间',
  PRIMARY KEY (id),
  UNIQUE KEY uk_tenant_code (tenant_code),
  KEY idx_enabled (enabled),
  KEY idx_expire_time (expire_time)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='租户表';
