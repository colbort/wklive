-- 行情快照列表已经提供当前行情详情和更新入口，不再单独展示“行情详情”页面。
-- 保留详情、更新接口权限，并作为按钮权限归属到行情快照列表。
UPDATE sys_menu
SET parent_id = 630,
    name = '获取行情详情',
    menu_type = 3,
    component = '',
    icon = '',
    sort = 631
WHERE id = 620;

UPDATE sys_menu
SET parent_id = 630,
    menu_type = 3,
    component = '',
    icon = '',
    sort = 632
WHERE id = 621;
