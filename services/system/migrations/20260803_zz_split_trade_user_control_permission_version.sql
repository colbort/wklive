-- 菜单结构变化后提升权限版本，使在线管理员重新加载菜单树。
UPDATE sys_user u
SET u.perms_ver = u.perms_ver + 1,
    u.update_times = CAST(UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000 AS UNSIGNED)
WHERE EXISTS (
  SELECT 1
  FROM sys_user_role ur
  JOIN sys_role_menu rm
    ON rm.tenant_id = ur.tenant_id AND rm.role_id = ur.role_id
  WHERE ur.user_id = u.id
    AND ur.tenant_id = u.tenant_id
    AND rm.menu_id = 1080
);
