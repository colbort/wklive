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

-- 旧的多应用关联表不再使用；一个账号只归属一个应用范围。
DROP TABLE IF EXISTS sys_user_app;
