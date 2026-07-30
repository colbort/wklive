-- dbinit:baseline-safe
-- 服务更名为 market 后，保留菜单 ID 与角色授权关系，只迁移展示名称、接口路径和权限标识。
UPDATE `sys_menu`
SET `name` = 'Market数据管理'
WHERE `id` = 400
  AND `name` = 'ITICK数据管理';

UPDATE `sys_menu`
SET `path` = REPLACE(`path`, '/itick/', '/market/'),
    `perms` = REPLACE(
        REPLACE(`perms`, 'itick:', 'market:'),
        'tenant-itick',
        'tenant-market'
    ),
    `component` = REPLACE(`component`, 'itick/', 'market/')
WHERE `app_scope` = 1
  AND (
      `path` LIKE '/itick/%'
      OR `perms` LIKE 'itick:%'
      OR `component` LIKE 'itick/%'
  );
