INSERT INTO sys_menu
(id, parent_id, app_scope, name, menu_type, method, path, perms, component, icon, sort)
SELECT 713, 670, 1, '重试行权清算', 3, 'POST', '/option/exercises/retry',
       'option:exercise:retry', '', '', 713
WHERE NOT EXISTS (SELECT 1 FROM sys_menu WHERE id = 713 OR perms = 'option:exercise:retry');
