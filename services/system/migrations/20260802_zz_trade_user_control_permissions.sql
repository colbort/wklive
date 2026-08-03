UPDATE sys_menu
SET name='用户交易控制', menu_type=2, method='GET', path='/trade/user-trade-controls',
    perms='trade:user-trade-control:list', component='trade/risk-controls', icon='Operation',
    parent_id=1000, sort=1080
WHERE id=1080;

INSERT INTO sys_menu
  (id, parent_id, app_scope, name, menu_type, method, path, perms, component, icon, sort)
VALUES
  (1082, 1080, 1, '查询统一用户交易控制', 3, 'GET', '/trade/user-trade-controls', 'trade:user-trade-control:list', '', '', 1082),
  (1083, 1080, 1, '停用用户交易控制', 3, 'POST', '/trade/user-trade-controls/disable', 'trade:user-trade-control:disable', '', '', 1083),
  (1084, 1080, 1, '查询用户交易控制审计', 3, 'GET', '/trade/user-trade-control-audits', 'trade:user-trade-control:audit', '', '', 1084)
ON DUPLICATE KEY UPDATE
  parent_id=VALUES(parent_id), app_scope=VALUES(app_scope), name=VALUES(name), menu_type=VALUES(menu_type),
  method=VALUES(method), path=VALUES(path), perms=VALUES(perms), component=VALUES(component), icon=VALUES(icon), sort=VALUES(sort);

INSERT IGNORE INTO sys_role_menu (tenant_id, role_id, menu_id)
SELECT tenant_id, role_id, 1082 FROM sys_role_menu WHERE menu_id=1080;

INSERT IGNORE INTO sys_role_menu (tenant_id, role_id, menu_id)
SELECT tenant_id, role_id, 1083 FROM sys_role_menu WHERE menu_id=1080;

INSERT IGNORE INTO sys_role_menu (tenant_id, role_id, menu_id)
SELECT tenant_id, role_id, 1084 FROM sys_role_menu WHERE menu_id=1080;
