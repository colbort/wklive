-- dbinit:baseline-safe
-- 服务更名为 market 后，保留菜单 ID 与角色授权关系，仅迁移接口路径和权限标识。
UPDATE `sys_menu`
SET `path` = '/market/authorities',
    `perms` = 'market:authority:list'
WHERE `id` = 483
  AND (`path` = '/itick/authorities' OR `perms` = 'itick:authority:list');

UPDATE `sys_menu`
SET `path` = '/market/authorities',
    `perms` = 'market:authority:set'
WHERE `id` = 484
  AND (`path` = '/itick/authorities' OR `perms` = 'itick:authority:set');
