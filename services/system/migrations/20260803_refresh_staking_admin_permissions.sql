-- Role-menu inserts do not pass through the RPC grant logic, so cached login
-- permissions must move to a new versioned key exactly once.
INSERT IGNORE INTO sys_role_menu (tenant_id, role_id, menu_id)
SELECT r.tenant_id, r.id, m.id
FROM sys_role r
JOIN sys_menu m ON m.id IN (850, 851, 852, 860, 861)
WHERE r.app_scope = 1 AND r.code IN ('super_admin', 'tenant_super_admin');

UPDATE sys_user u
JOIN sys_user_role ur ON ur.user_id = u.id
JOIN sys_role r ON r.id = ur.role_id
SET u.perms_ver = GREATEST(u.perms_ver, 2026080301),
    u.update_times = UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000
WHERE r.app_scope = 1
  AND r.code IN ('super_admin', 'tenant_super_admin')
  AND u.perms_ver < 2026080301;
