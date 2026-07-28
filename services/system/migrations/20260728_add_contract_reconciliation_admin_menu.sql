-- 合约资产对账异常查询与人工忽略入口。
-- 菜单 ID 与 init.sql 保持一致；INSERT IGNORE 支持重复执行。
INSERT IGNORE INTO `sys_menu`
(`id`, `parent_id`, `app_scope`, `name`, `menu_type`, `method`, `path`, `perms`, `component`, `icon`, `sort`)
VALUES
(1193, 1000, 1, '合约资产对账异常', 2, 'GET', '/trade/operations/reconciliation-issues', 'trade:operation:reconciliation-issue:list', 'trade/reconciliation-issues', 'Warning', 1193),
(1194, 1193, 1, '忽略合约资产对账异常', 3, 'POST', '/trade/operations/reconciliation-issues/ignore', 'trade:operation:reconciliation-issue:ignore', '', '', 1194);
