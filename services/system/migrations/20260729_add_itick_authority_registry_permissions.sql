-- Authority Registry 查询和配置接口。
-- 挂在现有价格公式菜单下；INSERT IGNORE 支持基线初始化和升级重复执行。
INSERT IGNORE INTO `sys_menu`
(`id`, `parent_id`, `app_scope`, `name`, `menu_type`, `method`, `path`, `perms`, `component`, `icon`, `sort`)
VALUES
(483, 480, 1, '查询权威行情来源', 3, 'GET', '/market/authorities', 'market:authority:list', '', '', 483),
(484, 480, 1, '配置权威行情来源', 3, 'POST', '/market/authorities', 'market:authority:set', '', '', 484);
