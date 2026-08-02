-- Option 平台兜底资金政策：查询、创建与独立复核权限。
-- 创建和复核分属不同角色，避免单一账号同时持有两项写权限。

INSERT INTO sys_menu
(id, parent_id, app_scope, name, menu_type, method, path, perms, component, icon, sort)
SELECT 751, 500, 1, '平台兜底资金政策', 2, 'GET',
       '/asset/platform-backstop-policies', 'asset:platform-backstop-policy:list',
       'asset/platform-backstop-policies', 'Wallet', 751
WHERE NOT EXISTS (
  SELECT 1 FROM sys_menu
  WHERE id = 751 OR perms = 'asset:platform-backstop-policy:list'
);

INSERT INTO sys_menu
(id, parent_id, app_scope, name, menu_type, method, path, perms, component, icon, sort)
SELECT 752, 751, 1, '创建平台兜底资金政策', 3, 'POST',
       '/asset/platform-backstop-policies', 'asset:platform-backstop-policy:create', '', '', 752
WHERE NOT EXISTS (
  SELECT 1 FROM sys_menu
  WHERE id = 752 OR perms = 'asset:platform-backstop-policy:create'
);

INSERT INTO sys_menu
(id, parent_id, app_scope, name, menu_type, method, path, perms, component, icon, sort)
SELECT 753, 751, 1, '复核平台兜底资金政策', 3, 'POST',
       '/asset/platform-backstop-policies/{policyId}/review',
       'asset:platform-backstop-policy:review', '', '', 753
WHERE NOT EXISTS (
  SELECT 1 FROM sys_menu
  WHERE id = 753 OR perms = 'asset:platform-backstop-policy:review'
);

INSERT INTO sys_menu
(id, parent_id, app_scope, name, menu_type, method, path, perms, component, icon, sort)
SELECT 754, 751, 1, '查看平台兜底资金政策详情', 3, 'GET',
       '/asset/platform-backstop-policies/{policyId}',
       'asset:platform-backstop-policy:detail', '', '', 754
WHERE NOT EXISTS (
  SELECT 1 FROM sys_menu
  WHERE id = 754 OR perms = 'asset:platform-backstop-policy:detail'
);

INSERT INTO sys_role
  (tenant_id, app_scope, name, code, enabled, remark, create_times, update_times)
VALUES
  (0, 1, '平台兜底政策申请员', 'platform_backstop_policy_creator', 1,
   '仅创建平台兜底资金政策草稿并查询证据，不得复核',
   UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000,
   UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000),
  (0, 1, '平台兜底政策复核员', 'platform_backstop_policy_reviewer', 1,
   '仅独立复核平台兜底资金政策并查询证据，不得创建',
   UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000,
   UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000)
ON DUPLICATE KEY UPDATE
  name = VALUES(name),
  enabled = 1,
  remark = VALUES(remark),
  update_times = VALUES(update_times);

DELETE rm
FROM sys_role_menu rm
JOIN sys_role r ON r.id = rm.role_id
JOIN sys_menu m ON m.id = rm.menu_id
WHERE r.tenant_id = 0
  AND r.app_scope = 1
  AND r.code = 'platform_backstop_policy_creator'
  AND m.perms = 'asset:platform-backstop-policy:review';

DELETE rm
FROM sys_role_menu rm
JOIN sys_role r ON r.id = rm.role_id
JOIN sys_menu m ON m.id = rm.menu_id
WHERE r.tenant_id = 0
  AND r.app_scope = 1
  AND r.code = 'platform_backstop_policy_reviewer'
  AND m.perms = 'asset:platform-backstop-policy:create';

INSERT IGNORE INTO sys_role_menu (tenant_id, role_id, menu_id)
SELECT 0, r.id, m.id
FROM sys_role r
JOIN sys_menu m ON m.id IN (751, 752, 754)
WHERE r.tenant_id = 0
  AND r.app_scope = 1
  AND r.code = 'platform_backstop_policy_creator';

INSERT IGNORE INTO sys_role_menu (tenant_id, role_id, menu_id)
SELECT 0, r.id, m.id
FROM sys_role r
JOIN sys_menu m ON m.id IN (751, 753, 754)
WHERE r.tenant_id = 0
  AND r.app_scope = 1
  AND r.code = 'platform_backstop_policy_reviewer';
