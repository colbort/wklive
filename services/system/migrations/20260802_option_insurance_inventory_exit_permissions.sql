-- 保险接管库存受控主动退出：申请、四眼复核、执行与查询权限。

INSERT INTO sys_menu
(id, parent_id, app_scope, name, menu_type, method, path, perms, component, icon, sort)
SELECT 755, 710, 1, '创建保险库存退出申请', 3, 'POST',
       '/option/risk/insurance-inventory-exits', 'option:insurance-inventory-exit:create', '', '', 755
WHERE NOT EXISTS (
  SELECT 1 FROM sys_menu
  WHERE id = 755 OR perms = 'option:insurance-inventory-exit:create'
);

INSERT INTO sys_menu
(id, parent_id, app_scope, name, menu_type, method, path, perms, component, icon, sort)
SELECT 756, 710, 1, '复核保险库存退出申请', 3, 'POST',
       '/option/risk/insurance-inventory-exits/review', 'option:insurance-inventory-exit:review', '', '', 756
WHERE NOT EXISTS (
  SELECT 1 FROM sys_menu
  WHERE id = 756 OR perms = 'option:insurance-inventory-exit:review'
);

INSERT INTO sys_menu
(id, parent_id, app_scope, name, menu_type, method, path, perms, component, icon, sort)
SELECT 757, 710, 1, '执行保险库存退出订单', 3, 'POST',
       '/option/risk/insurance-inventory-exits/execute', 'option:insurance-inventory-exit:execute', '', '', 757
WHERE NOT EXISTS (
  SELECT 1 FROM sys_menu
  WHERE id = 757 OR perms = 'option:insurance-inventory-exit:execute'
);

INSERT INTO sys_menu
(id, parent_id, app_scope, name, menu_type, method, path, perms, component, icon, sort)
SELECT 758, 710, 1, '查询保险库存退出申请', 3, 'GET',
       '/option/risk/insurance-inventory-exits', 'option:insurance-inventory-exit:list', '', '', 758
WHERE NOT EXISTS (
  SELECT 1 FROM sys_menu
  WHERE id = 758 OR perms = 'option:insurance-inventory-exit:list'
);
