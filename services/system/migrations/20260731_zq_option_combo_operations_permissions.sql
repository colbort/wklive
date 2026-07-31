-- 组合父单运营工作台权限；整组查询和整组强撤，不提供单腿操作权限。

INSERT INTO sys_menu
(id, parent_id, app_scope, name, menu_type, method, path, perms, component, icon, sort)
SELECT 748, 729, 1, '查询组合父单', 3, 'GET',
       '/option/combo-orders', 'option:operations:combo-list', '', '', 748
WHERE NOT EXISTS (
  SELECT 1 FROM sys_menu WHERE id = 748 OR perms = 'option:operations:combo-list'
);

INSERT INTO sys_menu
(id, parent_id, app_scope, name, menu_type, method, path, perms, component, icon, sort)
SELECT 749, 729, 1, '下钻组合父单', 3, 'GET',
       '/option/combo-orders/detail', 'option:operations:combo-view', '', '', 749
WHERE NOT EXISTS (
  SELECT 1 FROM sys_menu WHERE id = 749 OR perms = 'option:operations:combo-view'
);

INSERT INTO sys_menu
(id, parent_id, app_scope, name, menu_type, method, path, perms, component, icon, sort)
SELECT 750, 729, 1, '整组强撤组合父单', 3, 'POST',
       '/option/combo-orders/force-cancel', 'option:operations:combo-cancel', '', '', 750
WHERE NOT EXISTS (
  SELECT 1 FROM sys_menu WHERE id = 750 OR perms = 'option:operations:combo-cancel'
);
