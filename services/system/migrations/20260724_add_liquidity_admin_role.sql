-- 做市管理后台默认角色。
INSERT INTO sys_role
  (tenant_id, app_scope, name, code, enabled, remark, create_times, update_times)
SELECT
  0,
  2,
  '做市管理员',
  'liquidity_admin',
  1,
  '做市及外部流动性管理后台角色',
  UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000,
  UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000
WHERE NOT EXISTS (
  SELECT 1
  FROM sys_role
  WHERE tenant_id = 0 AND app_scope = 2 AND code = 'liquidity_admin'
);

-- 兼容按文件名顺序执行迁移：权限资源可能先于默认角色创建。
INSERT IGNORE INTO sys_role_menu (tenant_id, role_id, menu_id)
SELECT r.tenant_id, r.id, m.id
FROM sys_role r
JOIN sys_menu m ON m.app_scope = 2
WHERE r.code = 'liquidity_admin'
  AND r.app_scope = 2;
