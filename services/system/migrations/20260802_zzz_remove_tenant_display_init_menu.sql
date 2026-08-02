-- “初始化租户展示配置”页面已废弃；先清理角色授权，再删除菜单/API权限资源。
DELETE FROM sys_role_menu
WHERE menu_id IN (
  SELECT id
  FROM sys_menu
  WHERE id = 470
     OR path = '/market/tenant-display/init'
     OR perms = 'market:tenant-display:init'
);

DELETE FROM sys_menu
WHERE id = 470
   OR path = '/market/tenant-display/init'
   OR perms = 'market:tenant-display:init';
