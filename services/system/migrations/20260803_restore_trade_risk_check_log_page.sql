-- 风控校验日志属于交易风控公共数据，不分别嵌入产品级和交易对级控制页面。
-- 将既有权限节点恢复为用户交易控制目录下的独立列表页面。

SET NAMES utf8mb4;

UPDATE sys_menu
SET parent_id = 1080,
    name = '风控校验日志',
    menu_type = 2,
    method = 'GET',
    path = '/trade/risk-order-check-logs',
    perms = 'trade:risk-order-check-log:list',
    component = 'trade/risk-order-check-logs',
    icon = 'DocumentChecked',
    sort = 30,
    visible = 1,
    enabled = 1
WHERE id = 1110;

-- 已有两个用户控制页面权限的角色，同时获得公共日志页面入口。
INSERT IGNORE INTO sys_role_menu (tenant_id, role_id, menu_id)
SELECT DISTINCT tenant_id, role_id, 1110
FROM sys_role_menu
WHERE menu_id IN (1085, 1086);
