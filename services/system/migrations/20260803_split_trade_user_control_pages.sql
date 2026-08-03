-- 将产品级用户控制和交易对级覆盖控制拆分为两个独立页面。

SET NAMES utf8mb4;

UPDATE sys_menu
SET parent_id = 1000,
    name = '用户交易控制',
    menu_type = 1,
    method = '',
    path = '',
    perms = '',
    component = '',
    icon = 'Operation',
    sort = 1080,
    visible = 1,
    enabled = 1
WHERE id = 1080;

INSERT INTO sys_menu
(id, parent_id, app_scope, name, menu_type, method, path, perms, component, icon, sort)
VALUES
  (1085, 1080, 1, '用户产品级控制', 2, 'GET', '/trade/user-product-controls',
   'trade:user-trade-control:list', 'trade/user-product-controls', 'SetUp', 10),
  (1086, 1080, 1, '用户交易对级控制', 2, 'GET', '/trade/user-symbol-controls',
   'trade:user-trade-control:list', 'trade/user-symbol-controls', 'Switch', 20)
ON DUPLICATE KEY UPDATE
  parent_id=VALUES(parent_id), app_scope=VALUES(app_scope), name=VALUES(name),
  menu_type=VALUES(menu_type), method=VALUES(method), path=VALUES(path),
  perms=VALUES(perms), component=VALUES(component), icon=VALUES(icon), sort=VALUES(sort),
  visible=1, enabled=1;

UPDATE sys_menu SET parent_id = 1085 WHERE id = 1081;
UPDATE sys_menu SET parent_id = 1086 WHERE id IN (1090, 1091);

-- 公共列表、停用、审计和日志权限继续挂在父目录，供两个页面共同使用。
UPDATE sys_menu SET parent_id = 1080 WHERE id IN (1082, 1083, 1084, 1110);

INSERT IGNORE INTO sys_role_menu (tenant_id, role_id, menu_id)
SELECT tenant_id, role_id, 1085 FROM sys_role_menu WHERE menu_id = 1080;

INSERT IGNORE INTO sys_role_menu (tenant_id, role_id, menu_id)
SELECT tenant_id, role_id, 1086 FROM sys_role_menu WHERE menu_id = 1080;

INSERT IGNORE INTO sys_role_menu (tenant_id, role_id, menu_id)
SELECT DISTINCT tenant_id, role_id, 1080
FROM sys_role_menu
WHERE menu_id IN (1081, 1082, 1083, 1084, 1085, 1086, 1090, 1091, 1110);
