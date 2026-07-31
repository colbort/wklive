-- 期权风险工作台：交易控制、异常成交更正和 MMP 操作权限。

INSERT INTO sys_menu
(id, parent_id, app_scope, name, menu_type, method, path, perms, component, icon, sort)
SELECT 715, 710, 1, '解除期权 Kill Switch', 3, 'POST',
       '/option/trading-controls/release-kill-switch', 'option:trading-control:release', '', '', 715
WHERE NOT EXISTS (SELECT 1 FROM sys_menu WHERE id = 715 OR perms = 'option:trading-control:release');

INSERT INTO sys_menu
(id, parent_id, app_scope, name, menu_type, method, path, perms, component, icon, sort)
SELECT 716, 710, 1, '创建异常成交更正', 3, 'POST',
       '/option/trade-corrections', 'option:trade-correction:create', '', '', 716
WHERE NOT EXISTS (SELECT 1 FROM sys_menu WHERE id = 716 OR perms = 'option:trade-correction:create');

INSERT INTO sys_menu
(id, parent_id, app_scope, name, menu_type, method, path, perms, component, icon, sort)
SELECT 717, 710, 1, '复核异常成交更正', 3, 'POST',
       '/option/trade-corrections/review', 'option:trade-correction:review', '', '', 717
WHERE NOT EXISTS (SELECT 1 FROM sys_menu WHERE id = 717 OR perms = 'option:trade-correction:review');

INSERT INTO sys_menu
(id, parent_id, app_scope, name, menu_type, method, path, perms, component, icon, sort)
SELECT 718, 710, 1, '配置期权 MMP', 3, 'POST',
       '/option/mmp/config', 'option:mmp:config', '', '', 718
WHERE NOT EXISTS (SELECT 1 FROM sys_menu WHERE id = 718 OR perms = 'option:mmp:config');

INSERT INTO sys_menu
(id, parent_id, app_scope, name, menu_type, method, path, perms, component, icon, sort)
SELECT 719, 710, 1, '恢复期权 MMP', 3, 'POST',
       '/option/mmp/reset', 'option:mmp:reset', '', '', 719
WHERE NOT EXISTS (SELECT 1 FROM sys_menu WHERE id = 719 OR perms = 'option:mmp:reset');

INSERT INTO sys_menu
(id, parent_id, app_scope, name, menu_type, method, path, perms, component, icon, sort)
SELECT 720, 710, 1, '查询用户交易控制', 3, 'GET',
       '/option/trading-controls/detail', 'option:trading-control:detail', '', '', 720
WHERE NOT EXISTS (SELECT 1 FROM sys_menu WHERE id = 720 OR perms = 'option:trading-control:detail');

INSERT INTO sys_menu
(id, parent_id, app_scope, name, menu_type, method, path, perms, component, icon, sort)
SELECT 721, 710, 1, '查询交易控制审计', 3, 'GET',
       '/option/trading-controls/events', 'option:trading-control:event:list', '', '', 721
WHERE NOT EXISTS (SELECT 1 FROM sys_menu WHERE id = 721 OR perms = 'option:trading-control:event:list');

INSERT INTO sys_menu
(id, parent_id, app_scope, name, menu_type, method, path, perms, component, icon, sort)
SELECT 722, 710, 1, '查询异常成交更正', 3, 'GET',
       '/option/trade-corrections', 'option:trade-correction:list', '', '', 722
WHERE NOT EXISTS (SELECT 1 FROM sys_menu WHERE id = 722 OR perms = 'option:trade-correction:list');

INSERT INTO sys_menu
(id, parent_id, app_scope, name, menu_type, method, path, perms, component, icon, sort)
SELECT 723, 710, 1, '查询期权 MMP', 3, 'GET',
       '/option/mmp/configs', 'option:mmp:list', '', '', 723
WHERE NOT EXISTS (SELECT 1 FROM sys_menu WHERE id = 723 OR perms = 'option:mmp:list');

INSERT INTO sys_menu
(id, parent_id, app_scope, name, menu_type, method, path, perms, component, icon, sort)
SELECT 724, 710, 1, '创建组合保证金参数版本', 3, 'POST',
       '/option/risk/portfolio-configs', 'option:portfolio-risk:create', '', '', 724
WHERE NOT EXISTS (SELECT 1 FROM sys_menu WHERE id = 724 OR perms = 'option:portfolio-risk:create');

INSERT INTO sys_menu
(id, parent_id, app_scope, name, menu_type, method, path, perms, component, icon, sort)
SELECT 725, 710, 1, '复核组合保证金参数版本', 3, 'POST',
       '/option/risk/portfolio-configs/review', 'option:portfolio-risk:review', '', '', 725
WHERE NOT EXISTS (SELECT 1 FROM sys_menu WHERE id = 725 OR perms = 'option:portfolio-risk:review');

INSERT INTO sys_menu
(id, parent_id, app_scope, name, menu_type, method, path, perms, component, icon, sort)
SELECT 726, 710, 1, '查询组合保证金参数版本', 3, 'GET',
       '/option/risk/portfolio-configs', 'option:portfolio-risk:list', '', '', 726
WHERE NOT EXISTS (SELECT 1 FROM sys_menu WHERE id = 726 OR perms = 'option:portfolio-risk:list');
