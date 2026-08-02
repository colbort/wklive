INSERT INTO sys_menu
  (id, parent_id, app_scope, name, menu_type, method, path, perms, component, icon, sort)
SELECT 714, 680, 1, '重试结算资产指令', 3, 'POST',
       '/option/settlements/retry-instruction',
       'option:settlement-instruction:retry', '', '', 714
WHERE NOT EXISTS (
  SELECT 1 FROM sys_menu
  WHERE id = 714 OR perms = 'option:settlement-instruction:retry'
);
