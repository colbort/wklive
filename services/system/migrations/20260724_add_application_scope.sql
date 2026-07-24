-- 一个系统账号只属于一个后台；用户、角色、菜单必须具有相同 app_scope。
ALTER TABLE sys_user
  ADD COLUMN app_scope TINYINT NOT NULL DEFAULT 1 COMMENT '1综合管理后台 2做市管理后台' AFTER tenant_id,
  DROP INDEX uk_tenant_username,
  ADD UNIQUE KEY uk_tenant_scope_username(tenant_id, app_scope, username),
  ADD INDEX idx_app_scope(app_scope);

ALTER TABLE sys_role
  ADD COLUMN app_scope TINYINT NOT NULL DEFAULT 1 COMMENT '1综合管理后台 2做市管理后台' AFTER tenant_id,
  DROP INDEX uk_tenant_role_name,
  DROP INDEX uk_tenant_role_code,
  ADD UNIQUE KEY uk_tenant_scope_role_name(tenant_id, app_scope, name),
  ADD UNIQUE KEY uk_tenant_scope_role_code(tenant_id, app_scope, code),
  ADD INDEX idx_app_scope(app_scope);

ALTER TABLE sys_menu
  ADD COLUMN app_scope TINYINT NOT NULL DEFAULT 1 COMMENT '1综合管理后台 2做市管理后台' AFTER parent_id,
  ADD INDEX idx_app_scope(app_scope);

-- 如果之前部署过 sys_user_app，优先保留做市后台专用账号的归属。
CREATE TABLE IF NOT EXISTS sys_user_app (
  id BIGINT AUTO_INCREMENT PRIMARY KEY,
  user_id BIGINT NOT NULL DEFAULT 0,
  app_scope TINYINT NOT NULL DEFAULT 0,
  enabled TINYINT NOT NULL DEFAULT 1
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

UPDATE sys_user u
JOIN sys_user_app a ON a.user_id = u.id AND a.app_scope = 2 AND a.enabled = 1
SET u.app_scope = 2;

DROP TABLE IF EXISTS sys_user_app;
